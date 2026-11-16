// Copyright The ebpf-profile-lockwait Authors
// SPDX-License-Identifier: Apache-2.0

package lockwait // import "github.com/ebpf-profiler-probes/ebpf-profile-lockwait/lockwait"

import "go.opentelemetry.io/collector/component"

var (
	Type      = component.MustNewType("lockwait")
	stability = component.StabilityLevelDevelopment
)
