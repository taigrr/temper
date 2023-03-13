package temper

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Function which scans /dev for temperXX devices (configured by udev)
//
// A timeout of 250ms is recommended but as YMMV, this function allows
// for an arbitrary timeout.
func FindTempersWithTimeout(timeout time.Duration) ([]*Temper, error) {
	// list over dev folder for temperXX devices
	dirEnts, err := os.ReadDir("/dev")
	if err != nil {
		return []*Temper{}, err
	}

	tempers := []*Temper{}
	for _, d := range dirEnts {
		if name := d.Name(); strings.HasPrefix(name, "temper") {
			if isInputDevice(name) {
				continue
			}
			temper, err := New(filepath.Join("/dev", name))
			if err != nil {
				continue
			}
			// attempt to take a reading from the temper
			// if the reading times out, assume it's a false positive
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			_, err = temper.ReadCWithContext(ctx)
			if err == nil {
				tempers = append(tempers, temper)
			} else {
				// prevent file descriptor leaks
				temper.Close()
			}
			cancel()
		}
	}
	return tempers, nil
}

// Helper function to return list of temper devices available in /dev
//
// Uses the recommended default timeout of 250ms. See
// FindTempersWithTimeout for more details
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
	hidrawDesc := strings.ReplaceAll(temperDescriptor, "temper", "hidraw")
	inputPath := filepath.Join("/sys/class/hidraw", hidrawDesc, "device/input")
	if _, statErr := os.Stat(inputPath); statErr == nil {
		return true
	}
	return false
}
