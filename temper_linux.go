package temper

import (
	"encoding/hex"
	"os"
	"strconv"
	"sync"
)

type Temper struct {
	reader *os.File
	writer *os.File
	lock   sync.Mutex
}

type reading struct {
	value float32
	error error
}

func New(descriptor string) (*Temper, error) {
	if _, statErr := os.Stat(descriptor); statErr != nil {
		return &Temper{}, statErr
	}
	r, readErr := os.Open(descriptor)
	if readErr != nil {
		return &Temper{}, readErr
	}
	w, writeErr := os.OpenFile(descriptor,
		os.O_APPEND|os.O_WRONLY, os.ModeDevice)
	if writeErr != nil {
		r.Close()
		return &Temper{}, writeErr
	}
	t := Temper{reader: r, writer: w}
	return &t, nil
}

func (t *Temper) Close() error {
	t.lock.Lock()
	defer t.lock.Unlock()
	rErr := t.reader.Close()
	wErr := t.writer.Close()
	if rErr != nil {
		return rErr
	}
	return wErr
}

func (t *Temper) ReadC() (float32, error) {
	t.lock.Lock()
	defer t.lock.Unlock()
	tempChan := make(chan reading)
	go func() {
		// prepare a buffer and get ready to read
		// from the temper hid device
		response := make([]byte, 8)
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
	_, err := t.writer.Write([]byte{0, 1, 128, 51, 1, 0, 0, 0, 0})
	if err != nil {
		return 0, err
	}
	read := <-tempChan
	return read.value, read.error
}

func (t *Temper) ReadF() (float32, error) {
	c, err := t.ReadC()
	if err != nil {
		return 0, err
	}
	f := c*9.0/5.0 + 32.0
	return f, err
}
