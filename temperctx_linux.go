package temper

import (
	"context"
	"encoding/hex"
	"fmt"
	"strconv"
)

// ReadCWithContext reads the internal sensor temperature in Celsius,
// respecting the provided context for cancellation/timeout.
func (t *Temper) ReadCWithContext(ctx context.Context) (float32, error) {
	t.lock.Lock()
	defer t.lock.Unlock()

	type result struct {
		value float32
		err   error
	}

	ch := make(chan result, 1)
	go func() {
		// send magic byte sequence to request a temperature reading
		_, wErr := t.writer.Write([]byte{0, 1, 128, 51, 1, 0, 0, 0, 0})
		if wErr != nil {
			ch <- result{0, fmt.Errorf("writing temperature request: %w", wErr)}
			return
		}

		// read response from the temper HID device
		response := make([]byte, 8)
		_, rErr := t.reader.Read(response)
		if rErr != nil {
			ch <- result{0, fmt.Errorf("reading temperature response: %w", rErr)}
			return
		}

		// interpret the bytes as hex and extract temperature
		hexStr := hex.EncodeToString(response)
		temp := hexStr[4:8]

		tempInt, err := strconv.ParseInt(temp, 16, 64)
		if err != nil {
			ch <- result{0, fmt.Errorf("parsing temperature value %q: %w", temp, err)}
			return
		}

		ch <- result{float32(tempInt) / 100, nil}
	}()

	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	case r := <-ch:
		return r.value, r.err
	}
}

// ReadFWithContext reads the internal sensor temperature in Fahrenheit,
// respecting the provided context for cancellation/timeout.
func (t *Temper) ReadFWithContext(ctx context.Context) (float32, error) {
	c, err := t.ReadCWithContext(ctx)
	if err != nil {
		return 0, err
	}

	return c*9.0/5.0 + 32.0, nil
}
