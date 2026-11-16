// Copyright The ebpf-profile-lockwait Authors
// SPDX-License-Identifier: Apache-2.0

//go:build linux && (amd64 || arm64)

package lockwait // import "github.com/ebpf-profiler-probes/ebpf-profile-lockwait/lockwait"

import (
	"context"
	"errors"
	"fmt"
	"time"

	cebpf "github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"

	"go.opentelemetry.io/ebpf-profiler/reporter/samples"
	"go.opentelemetry.io/ebpf-profiler/tracer"
)

const futexWaitSymbol = "futex_wait"

// probe implements tracer.Probe for futex wait enter/leave.
type probe struct {
	sampleThreshold   uint32
	minDuration       time.Duration
	maxTrackedThreads uint32

	ref        *tracer.CollectTrampolineRef
	waitStart  *cebpf.Map
	trampoline *cebpf.Map
	links      []link.Link
}

func (p *probe) Load(_ context.Context, _ tracer.ProbeRegistrar, probeCtx *tracer.ProbeContext) error {
	ref, err := probeCtx.RegisterCollectTrampoline(&samples.TypeMetadata{
		SampleTypes:  []samples.SampleType{{Type: "lock_wait", Unit: "nanoseconds"}},
		ReportValues: true,
		AppendValues: samples.AppendFirstContextValue,
	}, tracer.ProbeKProbe)
	if err != nil {
		return fmt.Errorf("registering collect trampoline: %w", err)
	}

	spec, err := loadLockwait()
	if err != nil {
		_ = ref.Close()
		return fmt.Errorf("loading lockwait BPF spec: %w", err)
	}

	if v, ok := spec.Variables["sample_threshold"]; ok {
		if err := v.Set(p.sampleThreshold); err != nil {
			_ = ref.Close()
			return fmt.Errorf("setting sample_threshold: %w", err)
		}
	}
	if v, ok := spec.Variables["min_duration_ns"]; ok {
		if err := v.Set(uint64(p.minDuration.Nanoseconds())); err != nil {
			_ = ref.Close()
			return fmt.Errorf("setting min_duration_ns: %w", err)
		}
	}

	waitStartSpec, ok := spec.Maps["wait_start"]
	if !ok {
		_ = ref.Close()
		return fmt.Errorf("map wait_start missing from BPF spec")
	}
	waitStartSpec.MaxEntries = p.maxTrackedThreads
	waitStart, err := cebpf.NewMap(waitStartSpec)
	if err != nil {
		_ = ref.Close()
		return fmt.Errorf("creating wait_start: %w", err)
	}

	trampSpec, ok := spec.Maps["lockwait_trampoline"]
	if !ok {
		waitStart.Close()
		_ = ref.Close()
		return fmt.Errorf("map lockwait_trampoline missing from BPF spec")
	}
	trampMap, err := cebpf.NewMap(trampSpec)
	if err != nil {
		waitStart.Close()
		_ = ref.Close()
		return fmt.Errorf("creating lockwait_trampoline: %w", err)
	}
	if err := ref.InsertInto(trampMap, 0); err != nil {
		trampMap.Close()
		waitStart.Close()
		_ = ref.Close()
		return fmt.Errorf("inserting collect trampoline: %w", err)
	}

	coll, err := cebpf.NewCollectionWithOptions(spec, cebpf.CollectionOptions{
		MapReplacements: map[string]*cebpf.Map{
			"wait_start":           waitStart,
			"lockwait_probe_value": ref.CtxMap,
			"lockwait_trampoline":  trampMap,
		},
	})
	if err != nil {
		trampMap.Close()
		waitStart.Close()
		_ = ref.Close()
		return fmt.Errorf("loading lockwait collection: %w", err)
	}
	defer coll.Close()

	enterProg, ok := coll.Programs["lockwait_enter"]
	if !ok {
		trampMap.Close()
		waitStart.Close()
		_ = ref.Close()
		return fmt.Errorf("program lockwait_enter not found")
	}
	leaveProg, ok := coll.Programs["lockwait_leave"]
	if !ok {
		trampMap.Close()
		waitStart.Close()
		_ = ref.Close()
		return fmt.Errorf("program lockwait_leave not found")
	}

	enterLink, err := link.Kprobe(futexWaitSymbol, enterProg, nil)
	if err != nil {
		trampMap.Close()
		waitStart.Close()
		_ = ref.Close()
		return fmt.Errorf("attaching kprobe %s: %w", futexWaitSymbol, err)
	}
	leaveLink, err := link.Kretprobe(futexWaitSymbol, leaveProg, nil)
	if err != nil {
		_ = enterLink.Close()
		trampMap.Close()
		waitStart.Close()
		_ = ref.Close()
		return fmt.Errorf("attaching kretprobe %s: %w", futexWaitSymbol, err)
	}

	p.ref = ref
	p.waitStart = waitStart
	p.trampoline = trampMap
	p.links = []link.Link{enterLink, leaveLink}
	return nil
}

func (p *probe) Unload() error {
	var errs error
	for i := range p.links {
		errs = errors.Join(errs, p.links[i].Close())
	}
	p.links = nil
	if p.trampoline != nil {
		errs = errors.Join(errs, p.trampoline.Close())
		p.trampoline = nil
	}
	if p.waitStart != nil {
		errs = errors.Join(errs, p.waitStart.Close())
		p.waitStart = nil
	}
	if p.ref != nil {
		errs = errors.Join(errs, p.ref.Close())
		p.ref = nil
	}
	return errs
}
