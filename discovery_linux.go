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
