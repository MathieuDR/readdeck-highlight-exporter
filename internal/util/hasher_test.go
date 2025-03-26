package util_test

import "fmt"

import (
	"reflect"
	"testing"

	"github.com/mathieudr/readdeck-highlight-exporter/internal/util"
)

func TestGobHasher_Encode(t *testing.T) {
	tests := []struct {
		name    string
		input   []string
		want    string
		wantErr bool
	}{
		{
			name:    "empty slice",
			input:   []string{},
			wantErr: false,
		},
		{
			name:    "single element",
			input:   []string{"hello"},
			wantErr: false,
		},
		{
			name:    "multiple elements",
			input:   []string{"hello", "world", "test"},
			wantErr: false,
		},
		{
			name:    "with empty string",
			input:   []string{"", "test", ""},
			wantErr: false,
		},
		{
			name:    "with special characters",
			input:   []string{"hello!", "@world", "#test$"},
			wantErr: false,
		},
		{
			name:    "with unicode characters",
			input:   []string{"こんにちは", "世界", "测试"},
			wantErr: false,
		},
		{
			name:    "with very long string",
			input:   []string{"a" + string(make([]byte, 1000))},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := util.NewGobHasher()
			got, gotErr := h.Encode(tt.input)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Encode() failed: %v", gotErr)
				}
				return
			}

			if tt.wantErr {
				t.Fatal("Encode() succeeded unexpectedly")
			}

			// We don't check specific encoded values because the gob encoding
			// might change between Go versions. Instead, we'll verify the
			// round-trip functionality in an additional test.

			if got == "" {
				t.Errorf("Encode() returned empty string")
			}
		})
	}
}

func TestGobHasher_Decode(t *testing.T) {
	tests := []struct {
		name    string
		encoded string
		want    []string
		wantErr bool
	}{
		{
			name:    "invalid base64",
			encoded: "this is not base64!",
			wantErr: true,
		},
		{
			name:    "valid base64 but invalid gob",
			encoded: "SGVsbG8gV29ybGQh", // "Hello World!" in base64
			wantErr: true,
		},
	}

	// We'll add more test cases using actual encoded values
	h := util.NewGobHasher()

	testInputs := [][]string{
		{},
		{"hello"},
		{"hello", "world", "test"},
		{"", "test", ""},
		{"こんにちは", "世界", "测试"},
	}

	for i, input := range testInputs {
		encoded, err := h.Encode(input)
		if err != nil {
			t.Fatalf("Failed to encode test input #%d: %v", i, err)
		}

		tests = append(tests, struct {
			name    string
			encoded string
			want    []string
			wantErr bool
		}{
			name:    "validly encoded #" + fmt.Sprint(1+i),
			encoded: encoded,
			want:    input,
			wantErr: false,
		})
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := util.NewGobHasher()
			got, gotErr := h.Decode(tt.encoded)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Decode() failed: %v", gotErr)
				}
				return
			}

			if tt.wantErr {
				t.Fatal("Decode() succeeded unexpectedly")
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Decode() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestRoundTrip verifies that encoding and then decoding preserves the original data
func TestRoundTrip(t *testing.T) {
	testCases := []struct {
		name  string
		input []string
	}{
		{"empty", []string{}},
		{"single", []string{"hello"}},
		{"multiple", []string{"hello", "world", "test"}},
		{"with empty strings", []string{"", "middle", ""}},
		{"unicode", []string{"こんにちは", "世界", "测试"}},
		{"special chars", []string{"!@#$%^&*()", "\\\"'\n\t"}},
		{"long string", []string{"a" + string(make([]byte, 5000))}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h := util.NewGobHasher()

			encoded, err := h.Encode(tc.input)
			if err != nil {
				t.Fatalf("Encode failed: %v", err)
			}

			decoded, err := h.Decode(encoded)
			if err != nil {
				t.Fatalf("Decode failed: %v", err)
			}

			if !reflect.DeepEqual(decoded, tc.input) {
				t.Errorf("Round trip failed: got %v (%T, len: %d, nil: %t), want %v (%T, len: %d, nil: %t)",
					decoded, decoded, len(decoded), decoded == nil,
					tc.input, tc.input, len(tc.input), tc.input == nil)
			}
		})
	}
}

// TestReuse verifies that the GobHasher can be reused for multiple operations
func TestReuse(t *testing.T) {
	h := util.NewGobHasher()

	inputs := [][]string{
		{"first", "encode"},
		{"second", "encode", "operation"},
		{"third", "encode", "with", "more", "strings"},
	}

	// Test multiple encodes with the same hasher
	var encodedResults []string
	for _, input := range inputs {
		encoded, err := h.Encode(input)
		if err != nil {
			t.Fatalf("Encode failed: %v", err)
		}
		encodedResults = append(encodedResults, encoded)
	}

	// Test multiple decodes with the same hasher
	for i, encoded := range encodedResults {
		decoded, err := h.Decode(encoded)
		if err != nil {
			t.Fatalf("Decode failed: %v", err)
		}

		if !reflect.DeepEqual(decoded, inputs[i]) {
			t.Errorf("Decode after reuse failed: got %v, want %v", decoded, inputs[i])
		}
	}
}

