// Copyright The ebpf-profile-lockwait Authors
// SPDX-License-Identifier: Apache-2.0

package lockwait

import (
	"testing"
	"time"
)

func TestConfigValidate(t *testing.T) {
	t.Parallel()

	valid := &Config{SampleRate: 0.1, MinDuration: time.Millisecond}
	if err := valid.Validate(); err != nil {
		t.Fatalf("expected valid config, got %v", err)
	}

	cases := []struct {
		name string
		cfg  Config
	}{
		{"sample_rate zero", Config{SampleRate: 0}},
		{"sample_rate high", Config{SampleRate: 1.1}},
		{"negative min_duration", Config{SampleRate: 1, MinDuration: -time.Second}},
		{"tracked threads too large", Config{SampleRate: 1, MaxTrackedThreads: maxTrackedThreadsCap + 1}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if err := tc.cfg.Validate(); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}
