// Copyright The ebpf-profile-lockwait Authors
// SPDX-License-Identifier: Apache-2.0

//go:build linux && (amd64 || arm64)

package lockwait // import "github.com/ebpf-profiler-probes/ebpf-profile-lockwait/lockwait"

import (
	"context"
	"math"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/extension"
)

// NewFactory returns an extension.Factory for the lockwait extension.
func NewFactory() extension.Factory {
	return extension.NewFactory(
		Type,
		func() component.Config { return &Config{SampleRate: 1.0} },
		createExtension,
		stability,
	)
}

func createExtension(_ context.Context, _ extension.Settings, cfg component.Config) (extension.Extension, error) {
	c := cfg.(*Config)
	return &lockwaitExtension{p: &probe{
		sampleThreshold:   uint32(c.SampleRate * math.MaxUint32),
		minDuration:       c.MinDuration,
		maxTrackedThreads: c.trackedThreadsOrDefault(),
	}}, nil
}
