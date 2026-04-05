package temper

import (
	"context"
	"os"
	"sync"
)

// Temper represents a connection to a TEMPer USB thermometer device.
// It holds file descriptors for reading and writing to the HID device.
// Close must be called when the device is no longer needed to prevent
// file descriptor leaks.
type Temper struct {
	descriptor string
	reader     *os.File
	writer     *os.File
	lock       sync.Mutex
}

// New opens a new Temper device at the given descriptor path.
// It is the caller's responsibility to call Close()
// to prevent a file descriptor leak.
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
	t := Temper{reader: r, writer: w, descriptor: descriptor}
	return &t, nil
}

func (t *Temper) Descriptor() string {
	return t.descriptor
}

func (t *Temper) String() string {
	return t.Descriptor()
}

// Close releases the file descriptors for the Temper device.
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

// ReadC reads the internal sensor temperature in Celsius.
//
// This is a convenience wrapper around ReadCWithContext using a
// background context. For cancellation or timeout support, use
// ReadCWithContext directly.
func (t *Temper) ReadC() (float32, error) {
	return t.ReadCWithContext(context.Background())
}

// ReadF reads the internal sensor temperature in Fahrenheit.
//
// This is a convenience wrapper around ReadFWithContext using a
// background context. For cancellation or timeout support, use
// ReadFWithContext directly.
func (t *Temper) ReadF() (float32, error) {
	return t.ReadFWithContext(context.Background())
}
