package temper

import (
	"context"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"
)

// ReadCWithContext reads the internal sensor temperature in Celsius,
// respecting the provided context's deadline for cancellation/timeout.
func (t *Temper) ReadCWithContext(ctx context.Context) (float32, error) {
	t.lock.Lock()
	defer t.lock.Unlock()

	deadline, ok := ctx.Deadline()
	if !ok {
		deadline = time.Time{}
	}

	if err := t.writer.SetWriteDeadline(deadline); err != nil {
		return 0, fmt.Errorf("setting write deadline: %w", err)
	}
	defer t.writer.SetDeadline(time.Time{})

	if err := t.reader.SetReadDeadline(deadline); err != nil {
		return 0, fmt.Errorf("setting read deadline: %w", err)
	}
	defer t.reader.SetDeadline(time.Time{})

	tempChan := make(chan reading)
	go func() {
		// prepare a buffer and get ready to read
		// from the temper hid device
		response := make([]byte, 8)

		if err := t.reader.SetDeadline(deadline); err != nil {
			tempChan <- reading{0, fmt.Errorf("setting reader deadline: %w", err)}
			return
		}

		_, err := t.reader.Read(response)
		if err != nil {
			tempChan <- reading{0, err}
			return
		}

		// interpret the bytes as hex
		hexStr := hex.EncodeToString(response)
		// extract the temperature fields from the string
		temp := hexStr[4:8]
		// convert the hex ints to an integer
		tempInt, err := strconv.ParseInt(temp, 16, 64)
		if err != nil {
			tempChan <- reading{0, err}
			return
		}

		// divide the result by 100 and send to chan
		float := float32(tempInt) / 100
		tempChan <- reading{error: nil, value: float}
	}()

	// send magic byte sequence to request a temperature reading
	_, wErr := t.writer.Write([]byte{0, 1, 128, 51, 1, 0, 0, 0, 0})
	if wErr != nil {
		return 0, fmt.Errorf("writing temperature request: %w", wErr)
	}

	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	case read := <-tempChan:
		return read.value, read.error
	}
}

// ReadFWithContext reads the internal sensor temperature in Fahrenheit,
// respecting the provided context's deadline for cancellation/timeout.
func (t *Temper) ReadFWithContext(ctx context.Context) (float32, error) {
	c, err := t.ReadCWithContext(ctx)
	if err != nil {
		return 0, err
	}

	f := c*9.0/5.0 + 32.0
	return f, nil
}
