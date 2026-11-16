// Copyright The ebpf-profile-lockwait Authors
// SPDX-License-Identifier: Apache-2.0

package lockwait // import "github.com/ebpf-profiler-probes/ebpf-profile-lockwait/lockwait"

import (
	"fmt"
	"time"
)

const (
	maxTrackedThreadsCap  = 1 << 20
	defaultTrackedThreads = 4096
)

// Config holds the YAML configuration for the lockwait extension.
//
//	extensions:
//	  lockwait:
//	    min_duration: 1ms
//	    sample_rate: 0.1
//	    max_tracked_threads: 4096
type Config struct {
	// MinDuration drops shorter waits before stack unwinding. 0 keeps all sampled waits.
	MinDuration time.Duration `mapstructure:"min_duration"`

	// SampleRate is the capture probability in ]0.0, 1.0], applied on wait enter.
	SampleRate float64 `mapstructure:"sample_rate"`

	// MaxTrackedThreads bounds concurrent in-flight waits. 0 uses 4096.
	MaxTrackedThreads uint32 `mapstructure:"max_tracked_threads"`
}

// Validate implements confmap.Validator.
func (cfg *Config) Validate() error {
	if cfg.SampleRate <= 0.0 || cfg.SampleRate > 1.0 {
		return fmt.Errorf("lockwait: sample_rate %f is out of range ]0.0, 1.0]", cfg.SampleRate)
	}
	if cfg.MinDuration < 0 {
		return fmt.Errorf("lockwait: min_duration %s must be >= 0", cfg.MinDuration)
	}
	if cfg.MaxTrackedThreads > maxTrackedThreadsCap {
		return fmt.Errorf("lockwait: max_tracked_threads %d exceeds maximum %d",
			cfg.MaxTrackedThreads, maxTrackedThreadsCap)
	}
	return nil
}

func (cfg *Config) trackedThreadsOrDefault() uint32 {
	if cfg.MaxTrackedThreads == 0 {
		return defaultTrackedThreads
	}
	return cfg.MaxTrackedThreads
}
