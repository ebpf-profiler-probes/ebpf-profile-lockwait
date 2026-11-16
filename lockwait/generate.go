// Copyright The ebpf-profile-lockwait Authors
// SPDX-License-Identifier: Apache-2.0

package lockwait // import "github.com/ebpf-profiler-probes/ebpf-profile-lockwait/lockwait"

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -target amd64,arm64 -no-strip lockwait ../bpf/lockwait.ebpf.c -- -I../bpf/headers
