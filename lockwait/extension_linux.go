// Copyright The ebpf-profile-lockwait Authors
// SPDX-License-Identifier: Apache-2.0

//go:build linux && (amd64 || arm64)

package lockwait // import "github.com/ebpf-profiler-probes/ebpf-profile-lockwait/lockwait"

import (
	"context"

	"go.opentelemetry.io/collector/component"

	"go.opentelemetry.io/ebpf-profiler/tracer"
)

// lockwaitExtension is an OTel Collector extension that exposes a tracer.Probe
// for the ebpf-profiler receiver (Probes API).
type lockwaitExtension struct {
	p *probe
}

func (e *lockwaitExtension) Start(_ context.Context, _ component.Host) error { return nil }

// Shutdown is a no-op: the profiling receiver calls Probe.Unload on teardown.
func (e *lockwaitExtension) Shutdown(_ context.Context) error { return nil }

// Probe returns the tracer.Probe that this extension provides.
func (e *lockwaitExtension) Probe() tracer.Probe { return e.p }
