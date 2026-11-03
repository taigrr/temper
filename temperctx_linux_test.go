package temper

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"testing"
	"time"
)

// Pipes exercise the pollable file I/O without requiring a USB thermometer.
func pipeTemper(t *testing.T) (*Temper, *os.File, *os.File) {
	t.Helper()
	responseReader, responseWriter, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { responseReader.Close(); responseWriter.Close() })
	requestReader, requestWriter, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { requestReader.Close(); requestWriter.Close() })
	device := &Temper{reader: responseReader, writer: requestWriter}
	t.Cleanup(func() { device.Close() })
	// Bound failed tests as well as successful transactions.
	if err := requestReader.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
		t.Fatal(err)
	}
	return device, requestReader, responseWriter
}

func readRequest(t *testing.T, reader *os.File) {
	t.Helper()
	request := make([]byte, 9)
	if _, err := io.ReadFull(reader, request); err != nil {
		t.Fatal(err)
	}
	if want := []byte{0, 1, 128, 51, 1, 0, 0, 0, 0}; !bytes.Equal(request, want) {
		t.Fatalf("request = %v, want %v", request, want)
	}
}

type temperatureResult struct {
	value float32
	err   error
}

func waitTemperature(t *testing.T, results <-chan temperatureResult) temperatureResult {
	t.Helper()
	select {
	case result := <-results:
		return result
	case <-time.After(5 * time.Second):
		t.Fatal("temperature read did not return")
		return temperatureResult{}
	}
}

func TestReadCWithContextFragmentedResponse(t *testing.T) {
	device, requests, responses := pipeTemper(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	results := make(chan temperatureResult, 1)
	go func() { value, err := device.ReadCWithContext(ctx); results <- temperatureResult{value, err} }()
	readRequest(t, requests)
	// A response may arrive in multiple reads; bytes 2-3 encode -0.20 C.
	for _, fragment := range [][]byte{{0, 0, 0xff}, {0xec, 0, 0, 0, 0}} {
		if _, err := responses.Write(fragment); err != nil {
			t.Fatal(err)
		}
	}
	result := waitTemperature(t, results)
	if result.err != nil || result.value != -0.20 {
		t.Fatalf("read = %v, %v; want -0.20, nil", result.value, result.err)
	}
}

func TestReadCWithContextCancellationAndRecovery(t *testing.T) {
	device, requests, responses := pipeTemper(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	results := make(chan temperatureResult, 1)
	go func() { value, err := device.ReadCWithContext(ctx); results <- temperatureResult{value, err} }()
	// Ensure the transaction has started before cancelling its blocked read.
	readRequest(t, requests)
	cancel()
	if result := waitTemperature(t, results); !errors.Is(result.err, context.Canceled) {
		t.Fatalf("cancelled read error = %v, want context.Canceled", result.err)
	}
	// The cancelled transaction must not leave a deadline that poisons reuse.
	recoveryCtx, recoveryCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer recoveryCancel()
	go func() { value, err := device.ReadFWithContext(recoveryCtx); results <- temperatureResult{value, err} }()
	readRequest(t, requests)
	if _, err := responses.Write([]byte{0, 0, 0x08, 0x98, 0, 0, 0, 0}); err != nil {
		t.Fatal(err)
	}
	result := waitTemperature(t, results)
	if result.err != nil || result.value != cToF(22) {
		t.Fatalf("recovery read = %v, %v; want %v, nil", result.value, result.err, cToF(22))
	}
}

func TestReadCWithContextTruncatedResponse(t *testing.T) {
	device, requests, responses := pipeTemper(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	results := make(chan temperatureResult, 1)
	go func() { value, err := device.ReadCWithContext(ctx); results <- temperatureResult{value, err} }()
	readRequest(t, requests)
	if _, err := responses.Write([]byte{0, 0, 0x08, 0x98}); err != nil {
		t.Fatal(err)
	}
	if err := responses.Close(); err != nil {
		t.Fatal(err)
	}
	if result := waitTemperature(t, results); !errors.Is(result.err, io.ErrUnexpectedEOF) {
		t.Fatalf("truncated response error = %v, want io.ErrUnexpectedEOF", result.err)
	}
}
