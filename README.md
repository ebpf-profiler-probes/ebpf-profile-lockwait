# ebpf-profile-lockwait

Lock-wait profiling probe for OpenTelemetry eBPF Profiler. Determine **which call sites waited on locks, and for how long**.

## Why

On- and off-CPU profiles don’t pinpoint lock-wait locations or duration.

| | offcpu | **lock-wait** |
|---|---|---|
| Question | How long off CPU? | How long waiting on a lock? |
| Hook | `sched_switch` | futex wait enter/leave |

## Approach

We own **when/what**; otel-ebpf-profiler handles unwinding and OTLP export. \
Cost: sample on enter; drop short waits on leave; `max_tracked_threads` caps concurrent tracked waiters.

```
enter wait → sample_rate → store start (no stack)
leave wait → duration → drop if < min_duration
unwind once → export lock_wait / nanoseconds
```

## Config

```yaml
extensions:
  lockwait:
    min_duration: 1ms
    sample_rate: 0.1
    max_tracked_threads: 4096

receivers:
  profiling:
    probes: [lockwait]

service:
  extensions: [lockwait]
  pipelines:
    profiles:
      receivers: [profiling]
      exporters: [otlp]
```

## Build

Requires Go 1.26+, clang/LLVM with BPF target, and OCB:

```sh
go install go.opentelemetry.io/collector/cmd/builder@v0.165.0   # binary name: builder
make generate   
make build     
```

Run on Linux (eBPF):

```sh
sudo ./dist/otelcol-lockwait \
  --config=config/collector.yaml \
  --feature-gates=+service.profilesSupport
```

## Links

- [OpenTelemetry eBPF Profiler](https://github.com/open-telemetry/opentelemetry-ebpf-profiler)
- Probes API [#1658](https://github.com/open-telemetry/opentelemetry-ebpf-profiler/issues/1658) / [#1739](https://github.com/open-telemetry/opentelemetry-ebpf-profiler/pull/1739)
- Proposal [#2114](https://github.com/open-telemetry/opentelemetry-ebpf-profiler/issues/2114)

## License

- Go / docs: Apache-2.0
- eBPF C: GPL-2.0
