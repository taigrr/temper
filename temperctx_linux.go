package temper

import (
	"context"
	"encoding/hex"
	"strconv"
	"time"
)

func (t *Temper) ReadCWithContext(ctx context.Context) (float32, error) {
	t.lock.Lock()
	defer t.lock.Unlock()

	deadline, ok := ctx.Deadline()
	if !ok {
		deadline = time.Time{}
	}
	err := t.writer.SetWriteDeadline(deadline)
	if err != nil {
		panic(err)
	}
	defer t.writer.SetDeadline(time.Time{})
	err = t.reader.SetReadDeadline(deadline)
	if err != nil {
		panic(err)
	}
	defer t.reader.SetDeadline(time.Time{})
	tempChan := make(chan reading)
	go func() {
		// prepare a buffer and get ready to read
		// from the temper hid device
		response := make([]byte, 8)
		err := t.reader.SetDeadline(deadline)
		if err != nil {
			panic(err)
		}
		_, err = t.reader.Read(response)
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
		return 0, err
	}
	read := <-tempChan
	return read.value, read.error
}

// Read the internal sensor temperature in Fahrenheit
func (t *Temper) ReadFWithContext(ctx context.Context) (float32, error) {
	c, err := t.ReadCWithContext(ctx)
	if err != nil {
		return 0, err
	}
	f := c*9.0/5.0 + 32.0
	return f, err
}
