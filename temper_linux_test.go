package temper

import (
	"context"
	"os"
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
			got := cToF(tt.celsius)
			if diff := got - tt.wantF; diff > 0.01 || diff < -0.01 {
				t.Errorf("C to F conversion: got %f, want %f", got, tt.wantF)
			}
		})
	}
}

func TestNewInvalidPath(t *testing.T) {
	got, err := New("/dev/nonexistent-temper-device-test")
	if err == nil {
		t.Error("expected error for nonexistent device path, got nil")
	}
	if got == nil {
		t.Fatal("expected non-nil Temper on error for safe cleanup")
	}
	if closeErr := got.Close(); closeErr != nil {
		t.Fatalf("Close() on failed New result returned error: %v", closeErr)
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

func TestCloseSafeOnNilAndZeroValue(t *testing.T) {
	var nilTemper *Temper
	if err := nilTemper.Close(); err != nil {
		t.Fatalf("Close() on nil Temper returned error: %v", err)
	}

	var zero Temper
	if err := zero.Close(); err != nil {
		t.Fatalf("Close() on zero-value Temper returned error: %v", err)
	}

	// double close must remain a no-op
	if err := zero.Close(); err != nil {
		t.Fatalf("second Close() on zero-value Temper returned error: %v", err)
	}
}

func TestDecodeCelsius(t *testing.T) {
	tests := []struct {
		name    string
		b2, b3  byte
		want    float32
		wantErr bool
	}{
		{name: "room temp 22.00", b2: 0x08, b3: 0x98, want: 22.00},
		{name: "zero", b2: 0x00, b3: 0x00, want: 0},
		{name: "negative -0.20", b2: 0xFF, b3: 0xEC, want: -0.20},
		{name: "negative -40.00", b2: 0xF0, b3: 0x60, want: -40.00},
		{name: "too short", b2: 0x00, b3: 0x00, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resp []byte
			if tt.wantErr {
				resp = []byte{0, 1}
			} else {
				resp = []byte{0, 0, tt.b2, tt.b3, 0, 0, 0, 0}
			}
			got, err := decodeCelsius(resp)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil (value %f)", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if diff := got - tt.want; diff > 0.01 || diff < -0.01 {
				t.Errorf("decodeCelsius = %f, want %f", got, tt.want)
			}
		})
	}
}

func TestReadCWithContextSafeOnZeroValue(t *testing.T) {
	var zero Temper
	if _, err := zero.ReadCWithContext(context.Background()); err != errUninitializedTemper {
		t.Errorf("ReadCWithContext on closed device = %v, want %v", err, errUninitializedTemper)
	}
}

func TestReadCSafeOnNilAndZeroValue(t *testing.T) {
	var nilTemper *Temper
	if _, err := nilTemper.ReadC(); err != errUninitializedTemper {
		t.Fatalf("ReadC() on nil Temper error = %v, want %v", err, errUninitializedTemper)
	}

	var zero Temper
	if _, err := zero.ReadC(); err != errUninitializedTemper {
		t.Fatalf("ReadC() on zero-value Temper error = %v, want %v", err, errUninitializedTemper)
	}
}

func TestReadCSafeOnPartiallyInitializedTemper(t *testing.T) {
	tests := []struct {
		name string
		dev  *Temper
	}{
		{"missing reader", &Temper{writer: os.Stdout}},
		{"missing writer", &Temper{reader: os.Stdin}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := tt.dev.ReadC(); err != errUninitializedTemper {
				t.Fatalf("ReadC() error = %v, want %v", err, errUninitializedTemper)
			}
		})
	}
}

func TestReadFWithContextSafeOnUninitializedTemper(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name string
		dev  *Temper
	}{
		{"nil", nil},
		{"zero value", &Temper{}},
		{"missing reader", &Temper{writer: os.Stdout}},
		{"missing writer", &Temper{reader: os.Stdin}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := tt.dev.ReadFWithContext(ctx); err != errUninitializedTemper {
				t.Fatalf("ReadFWithContext() error = %v, want %v", err, errUninitializedTemper)
			}
		})
	}
}

func TestIsInputDevice(t *testing.T) {
	tests := []string{
		"temper999",
		"/dev/temper999",
	}

	for _, descriptor := range tests {
		t.Run(descriptor, func(t *testing.T) {
			if isInputDevice(descriptor) {
				t.Error("expected false for non-existent hidraw device")
			}
		})
	}
}

func TestFindTempersWithContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	tempers, err := FindTempersWithContext(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// with a cancelled context, no devices should be probed
	if len(tempers) != 0 {
		for _, dev := range tempers {
			dev.Close()
		}
		t.Errorf("expected 0 tempers with cancelled context, got %d", len(tempers))
	}
}

func TestFindTempersDoesNotPanic(t *testing.T) {
	// FindTempers should not panic even without devices present
	tempers, err := FindTempers()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, dev := range tempers {
		dev.Close()
	}
}
