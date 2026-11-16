# Example

Mutex contention workload for verifying lockwait on Linux.

```sh
go run ./example
go run ./example -workers=16 -hold=2ms -duration=30s
```

| Flag | Default | Meaning |
|---|---|---|
| `-workers` | 8 | goroutines sharing one mutex (≥ 2) |
| `-hold` | 1ms | how long the holder keeps the lock |
| `-duration` | 0 | run time; 0 means until Ctrl-C |

Prefer `-hold` ≥ 1ms so waits are not dropped by collector `min_duration: 1ms`.
