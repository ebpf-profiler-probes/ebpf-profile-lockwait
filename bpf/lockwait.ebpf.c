// SPDX-License-Identifier: GPL-2.0
//
// Futex wait enter/leave probe for lock-wait profiling.
// Enter: sample and store start time (no stack).
// Leave: compute duration, filter, tail-call collect trampoline.

#include "common.h"

// Start times for in-flight futex waits, keyed by pid_tgid.
// max_entries is resized at load time from max_tracked_threads.
struct {
	__uint(type, BPF_MAP_TYPE_LRU_HASH);
	__uint(max_entries, 4096);
	__type(key, __u64);   // pid_tgid
	__type(value, __u64); // CLOCK_MONOTONIC start ns
} wait_start SEC(".maps");

// Replaced at load time with CollectTrampolineRef.CtxMap (ext_probe_value).
struct {
	__uint(type, BPF_MAP_TYPE_PERCPU_ARRAY);
	__uint(max_entries, 1);
	__type(key, __u32);
	__type(value, __u64);
} lockwait_probe_value SEC(".maps");

// Populated at load time with CollectTrampolineRef via InsertInto(..., 0).
struct {
	__uint(type, BPF_MAP_TYPE_PROG_ARRAY);
	__uint(max_entries, 1);
	__type(key, __u32);
	__type(value, __u32);
} lockwait_trampoline SEC(".maps");

// sample_threshold: keep event when bpf_get_prandom_u32() <= threshold.
// Set from sample_rate in ]0.0, 1.0] at load time.
volatile const __u32 sample_threshold = 0;

// Drop waits shorter than this (nanoseconds). 0 keeps all sampled waits.
volatile const __u64 min_duration_ns = 0;

SEC("kprobe/lockwait_enter")
int lockwait_enter(struct pt_regs *ctx)
{
	(void)ctx;

	if (bpf_get_prandom_u32() > sample_threshold)
		return 0;

	__u64 pid_tgid = bpf_get_current_pid_tgid();
	__u32 pid = pid_tgid >> 32;
	__u32 tid = (__u32)pid_tgid;
	if (pid == 0 || tid == 0)
		return 0;

	__u64 ts = bpf_ktime_get_ns();
	bpf_map_update_elem(&wait_start, &pid_tgid, &ts, BPF_ANY);
	return 0;
}

SEC("kretprobe/lockwait_leave")
int lockwait_leave(struct pt_regs *ctx)
{
	__u64 pid_tgid = bpf_get_current_pid_tgid();
	__u64 *start_ts = bpf_map_lookup_elem(&wait_start, &pid_tgid);
	if (!start_ts || *start_ts == 0)
		return 0;

	__u64 now = bpf_ktime_get_ns();
	__u64 diff = now - *start_ts;
	bpf_map_delete_elem(&wait_start, &pid_tgid);

	if (diff < min_duration_ns)
		return 0;

	__u32 key = 0;
	bpf_map_update_elem(&lockwait_probe_value, &key, &diff, BPF_ANY);
	bpf_tail_call(ctx, &lockwait_trampoline, 0);
	return 0;
}

char LICENSE[] SEC("license") = "GPL";
