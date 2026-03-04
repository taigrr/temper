package temper

import (
	"testing"
)

func TestCToF(t *testing.T) {
	tests := []struct {
		name    string
		celsius float32
		wantF   float32
	}{
		{"freezing", 0, 32},
		{"boiling", 100, 212},
		{"body temp", 37, 98.6},
		{"negative", -40, -40},
		{"room temp", 22, 71.6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.celsius*9.0/5.0 + 32.0
			if diff := got - tt.wantF; diff > 0.01 || diff < -0.01 {
				t.Errorf("C to F conversion: got %f, want %f", got, tt.wantF)
			}
		})
	}
}

func TestNewInvalidPath(t *testing.T) {
	_, err := New("/dev/nonexistent-temper-device-test")
	if err == nil {
		t.Error("expected error for nonexistent device path, got nil")
	}
}

func TestTemperStringAndDescriptor(t *testing.T) {
	desc := "/dev/temper0"
	temp := Temper{descriptor: desc}

	if got := temp.Descriptor(); got != desc {
		t.Errorf("Descriptor() = %q, want %q", got, desc)
	}
	if got := temp.String(); got != desc {
		t.Errorf("String() = %q, want %q", got, desc)
	}
}

func TestIsInputDevice(t *testing.T) {
	if isInputDevice("temper999") {
		t.Error("expected false for non-existent hidraw device")
	}
}
