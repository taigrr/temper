# AGENTS.md

## Overview

`github.com/taigrr/temper` is a zero-dependency Go library for reading USB
TEMPer thermometers. It talks to the devices' `hidraw` interface directly
through `/dev` device files — there is no cgo, no HID library, just `os.File`
reads/writes of a magic byte sequence.

## Critical: Linux-only build constraint

Every source file is suffixed `_linux.go`, so the package **only compiles on
`GOOS=linux`**. On macOS/Windows the build constraint excludes all files and
you will see gopls/`go build` errors like "build constraints exclude all Go
files" or "no required module provides package github.com/taigrr/temper".
This is expected, not a bug.

To build, vet, or test off-Linux, cross-compile with `GOOS=linux`:

```sh
GOOS=linux go build ./...
GOOS=linux go vet ./...
GOOS=linux go test -c        # compile the test binary (can't run it off Linux)
```

On an actual Linux host the normal commands work:

```sh
go build ./...
go test ./...
go vet ./...
```

Note: `go test` here is safe to run without hardware — `FindTempers` just
scans `/dev` and finds nothing, and the reading logic is only exercised
against real devices. There are no build tags beyond the implicit `_linux`
filename constraint, and the module targets `go 1.26`.

## Layout

- `temper_linux.go` — `Temper` struct, `New`, `Close`, `Descriptor`/`String`,
  and the context-free `ReadC`/`ReadF` convenience wrappers.
- `temperctx_linux.go` — the real read logic: `ReadCWithContext` (writes the
  magic request bytes, reads 8 bytes back, parses hex) and `ReadFWithContext`.
- `discovery_linux.go` — `FindTempers`, `FindTempersWithTimeout`,
  `FindTempersWithContext`, plus `isInputDevice`.
- `temper_linux_test.go` — unit tests.
- `examples/main.go` — runnable usage example (`package main`).

## How it works (non-obvious details)

- **Device discovery** scans `/dev` for entries prefixed `temper` — these
  symlinks are created by the udev rules documented in `README.md`, not by the
  kernel. Without those udev rules there is nothing to find.
- **Keyboard-emulation guard**: some TEMPer devices double as virtual
  keyboards. `isInputDevice` maps `/dev/temperN` back to
  `/sys/class/hidraw/hidrawN/device/input` and skips any device that has an
  `input` node. Probing such a device can trigger phantom keystrokes, so do
  **not** remove this check.
- **Reading protocol**: `ReadCWithContext` writes the 9-byte sequence
  `{0,1,128,51,1,0,0,0,0}` to request a reading, reads exactly 8 bytes
  (`io.ReadFull`), and decodes bytes 2-3 as a **signed** big-endian 16-bit
  value in hundredths of a degree (`decodeCelsius`), so sub-zero readings
  decode correctly. `ReadF` derives from `ReadC`.
- **Context handling**: the read holds the device lock for the whole
  transaction and drives the underlying file's read/write deadlines from the
  context (a watcher goroutine sets an immediate deadline on `ctx.Done()`).
  This actually unblocks an outstanding read on an unresponsive device instead
  of abandoning it, so a stuck device can't deadlock `Close` or leak
  goroutines. hidraw nodes are pollable, which is what makes deadlines work —
  if a device's fd were *not* pollable, `SetReadDeadline` would be a no-op and
  a truly wedged read could hang; hidraw implements `.poll` in-kernel, so this
  is the expected, working case. The watcher goroutine is joined (`wg.Wait`)
  before the deadlines are reset and the lock released, so it can never touch
  the descriptors after the call returns or race `Close`. `ReadCWithContext`
  also guards against uninitialized/zero-value devices and use-after-`Close`
  (nil descriptors → `errUninitializedTemper`), checking the descriptors under
  the lock so the guard can't race `Close`.
- **Concurrency**: each `Temper` has a `sync.Mutex` serializing reads; a single
  device must not be read concurrently.
- **Discovery probes every candidate** by doing a real read and discards
  (and `Close()`s) any device whose read fails, to avoid FD leaks. Default
  discovery timeout is 250ms.

## Conventions

- `New` returns a non-nil `*Temper` even on error so callers can always safely
  call `Close()`; `Close` is also nil-, zero-value-, and double-close-safe (it
  nils the descriptors under the lock). Preserve this.
- Callers own the returned `*Temper` slice/values and must `Close()` them (see
  `examples/main.go`).
- Errors are wrapped with `fmt.Errorf("...: %w", err)` context strings.
- Every new device-facing file should keep the `_linux.go` suffix.
