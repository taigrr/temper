package temper

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// FindTempersWithContext scans /dev for temperXX devices (configured by udev),
// respecting the provided context for cancellation. Each candidate device is
// probed with a temperature read; if the read fails or the context is cancelled,
// the device is skipped and its file descriptors are closed.
func FindTempersWithContext(ctx context.Context) ([]*Temper, error) {
	dirEnts, err := os.ReadDir("/dev")
	if err != nil {
		return []*Temper{}, err
	}

	tempers := []*Temper{}
	for _, d := range dirEnts {
		// bail out early if the caller cancelled
		if ctx.Err() != nil {
			break
		}

		name := d.Name()
		if !strings.HasPrefix(name, "temper") {
			continue
		}
		if isInputDevice(name) {
			continue
		}

		temperDev, err := New(filepath.Join("/dev", name))
		if err != nil {
			continue
		}

		// attempt to take a reading; if it fails, skip the device
		_, err = temperDev.ReadCWithContext(ctx)
		if err == nil {
			tempers = append(tempers, temperDev)
		} else {
			// prevent file descriptor leaks
			temperDev.Close()
		}
	}

	return tempers, nil
}

// FindTempersWithTimeout scans /dev for temperXX devices (configured by udev).
//
// A timeout of 250ms is recommended but as YMMV, this function allows
// for an arbitrary timeout. Each candidate device is probed with a
// temperature read; devices that fail or time out are skipped.
func FindTempersWithTimeout(timeout time.Duration) ([]*Temper, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return FindTempersWithContext(ctx)
}

// FindTempers returns a list of temper devices available in /dev.
//
// Uses the recommended default timeout of 250ms. See
// FindTempersWithTimeout and FindTempersWithContext for more control.
func FindTempers() ([]*Temper, error) {
	return FindTempersWithTimeout(time.Millisecond * 250)
}

// Determines if the current hidraw device also doubles as a virtual keyboard
//
// some temper devices also have a keyboard emulation mode.
// The regular discovery function can trigger data entry mode, and cause
// annoying and distracting typing to happen, so this function allows us to
// skip the check on devices we know aren't temper sensors
func isInputDevice(temperDescriptor string) bool {
	hidrawDesc := strings.ReplaceAll(filepath.Base(temperDescriptor), "temper", "hidraw")
	inputPath := filepath.Join("/sys/class/hidraw", hidrawDesc, "device/input")
	if _, statErr := os.Stat(inputPath); statErr == nil {
		return true
	}
	return false
}
