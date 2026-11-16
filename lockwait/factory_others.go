// Copyright The ebpf-profile-lockwait Authors
// SPDX-License-Identifier: Apache-2.0

//go:build !(linux && (amd64 || arm64))

package lockwait // import "github.com/ebpf-profiler-probes/ebpf-profile-lockwait/lockwait"

import (
	"context"
	"errors"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/extension"
)

// NewFactory returns an extension.Factory for the lockwait extension.
// The lockwait extension is only functional on Linux amd64/arm64.
func NewFactory() extension.Factory {
	return extension.NewFactory(
		Type,
		func() component.Config { return &Config{SampleRate: 1.0} },
		func(_ context.Context, _ extension.Settings, _ component.Config) (extension.Extension, error) {
			return nil, errors.New("lockwait extension is only supported on Linux amd64/arm64")
		},
		stability,
	)
}
