# rssprobe

Small runner to compare Go heap (`runtime.MemStats`) with process RSS while repeatedly calling `EncodeOneFrame`.

## Run

From repository root:

```bash
go run ./cmd/rssprobe -wav testdata/test.wav -duration 20s -report-every 1s -frame-bytes 4096
```

Example output fields:

- `goAlloc`: Go heap live bytes
- `goSys`: Go runtime reserved bytes
- `rss`: process RSS from `ps`

If `rss` keeps growing while `goAlloc` stays low/flat, the growth is likely outside Go heap (for example cgo/native allocator behavior).

## macOS cross-check

Start `rssprobe`, copy `pid=...`, then in another terminal:

```bash
vmmap -summary <PID>
```

Use this with the runner output to see whether RSS changes are mostly from native memory regions.

