package temper

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// ReadCWithContext reads the internal sensor temperature in Celsius,
// respecting the provided context for cancellation/timeout.
//
// The device I/O is genuine blocking I/O on a hidraw node; cancellation is
// implemented by driving the underlying file's read/write deadlines from the
// context, so an in-flight, unresponsive read is actually unblocked rather
// than abandoned. The device lock is held for the whole transaction, so reads
// are serialized and the descriptors are never touched concurrently.
func (t *Temper) ReadCWithContext(ctx context.Context) (float32, error) {
	t.lock.Lock()
	defer t.lock.Unlock()

	// Honor an already-cancelled context before touching the device.
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	// Guard against use after Close (which nils the descriptors under the
	// lock), so callers get a clear error instead of a misleading one.
	if t.reader == nil || t.writer == nil {
		return 0, os.ErrClosed
	}

	// Clear any deadline left behind by a previous cancelled call so this
	// read starts with a clean slate.
	t.reader.SetReadDeadline(time.Time{})
	t.writer.SetWriteDeadline(time.Time{})

	// Watch the context and, on cancel/timeout, push an immediate deadline so
	// any blocking Write/Read returns promptly. The watcher is joined (wg.Wait)
	// before we reset the deadlines and release the lock, so it can never touch
	// the descriptors after this function returns — in particular it can never
	// race Close(), which nils them under the same lock.
	stop := make(chan struct{})
	var wg sync.WaitGroup
	wg.Go(func() {
		select {
		case <-ctx.Done():
			t.reader.SetReadDeadline(time.Now())
			t.writer.SetWriteDeadline(time.Now())
		case <-stop:
		}
	})
	defer func() {
		close(stop)
		wg.Wait()
		// Reset only after the watcher has exited, so it cannot re-arm a stale
		// past deadline that would spuriously fail the next read.
		t.reader.SetReadDeadline(time.Time{})
		t.writer.SetWriteDeadline(time.Time{})
	}()

	// send magic byte sequence to request a temperature reading
	if _, wErr := t.writer.Write([]byte{0, 1, 128, 51, 1, 0, 0, 0, 0}); wErr != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return 0, ctxErr
		}
		return 0, fmt.Errorf("writing temperature request: %w", wErr)
	}

	// read response from the temper HID device
	response := make([]byte, 8)
	if _, rErr := io.ReadFull(t.reader, response); rErr != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return 0, ctxErr
		}
		return 0, fmt.Errorf("reading temperature response: %w", rErr)
	}

	return decodeCelsius(response)
}

// decodeCelsius extracts the Celsius temperature from a raw 8-byte TEMPer
// response. The reading is a signed, big-endian 16-bit value in bytes 2-3
// expressed in hundredths of a degree, so sub-zero temperatures decode
// correctly instead of wrapping to large positives.
func decodeCelsius(response []byte) (float32, error) {
	if len(response) < 4 {
		return 0, fmt.Errorf("temperature response too short: %d bytes", len(response))
	}
	raw := int16(uint16(response[2])<<8 | uint16(response[3]))
	return float32(raw) / 100, nil
}

// ReadFWithContext reads the internal sensor temperature in Fahrenheit,
// respecting the provided context for cancellation/timeout.
func (t *Temper) ReadFWithContext(ctx context.Context) (float32, error) {
	c, err := t.ReadCWithContext(ctx)
	if err != nil {
		return 0, err
	}

	return cToF(c), nil
}

// cToF converts a Celsius temperature to Fahrenheit.
func cToF(c float32) float32 {
	return c*9.0/5.0 + 32.0
}
