// internal/model/time_test.go
package model

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

type TestTimeStruct struct {
	Time SimpleTime `yaml:"time"`
}

func TestSimpleTimeUnmarshalYAML(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      string // Expected time in the format "2006-01-02 15:04"
		wantErr   bool
		errString string
	}{
		{
			name:    "valid time format",
			input:   "time: 2023-10-27 10:48",
			want:    "2023-10-27 10:48",
			wantErr: false,
		},
		{
			name:    "valid time with whitespace",
			input:   "time: '  2023-10-27 10:48  '",
			want:    "2023-10-27 10:48",
			wantErr: false,
		},
		{
			name:    "empty time",
			input:   "time: ''",
			want:    "",
			wantErr: false,
		},
		{
			name:      "invalid format with seconds",
			input:     "time: 2023-10-27 10:48:30",
			wantErr:   true,
			errString: "time must be in format 2006-01-02 15:04",
		},
		{
			name:      "invalid format with T separator",
			input:     "time: 2023-10-27T10:48",
			wantErr:   true,
			errString: "time must be in format 2006-01-02 15:04",
		},
		{
			name:      "invalid format RFC3339",
			input:     "time: 2023-10-27T10:48:30Z",
			wantErr:   true,
			errString: "time must be in format 2006-01-02 15:04",
		},
		{
			name:      "invalid month",
			input:     "time: 2023-13-27 10:48",
			wantErr:   true,
			errString: "time must be in format 2006-01-02 15:04",
		},
		{
			name:      "invalid day",
			input:     "time: 2023-10-32 10:48",
			wantErr:   true,
			errString: "time must be in format 2006-01-02 15:04",
		},
		{
			name:      "invalid hour",
			input:     "time: 2023-10-27 24:48",
			wantErr:   true,
			errString: "time must be in format 2006-01-02 15:04",
		},
		{
			name:      "just date",
			input:     "time: 2023-10-27",
			wantErr:   true,
			errString: "time must be in format 2006-01-02 15:04",
		},
		{
			name:      "non-time string",
			input:     "time: not-a-time",
			wantErr:   true,
			errString: "time must be in format 2006-01-02 15:04",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ts TestTimeStruct
			err := yaml.Unmarshal([]byte(tt.input), &ts)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errString != "" {
					assert.Contains(t, err.Error(), tt.errString)
				}
				return
			}

			assert.NoError(t, err)

			// For empty time, expect zero time
			if tt.want == "" {
				assert.True(t, ts.Time.IsZero())
				return
			}

			gotStr := ts.Time.Format(ExpectedTimeFormat)
			assert.Equal(t, tt.want, gotStr)
		})
	}
}

func TestSimpleTimeMarshalYAML(t *testing.T) {
	tests := []struct {
		name string
		time time.Time
		want string // Expected YAML output
	}{
		{
			name: "valid time",
			time: parseTime(t, "2023-10-27 10:48"),
			want: "time: 2023-10-27 10:48\n",
		},
		{
			name: "zero time",
			time: time.Time{},
			want: "time: \"\"\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := SimpleTime{Time: tt.time}
			ts := TestTimeStruct{Time: st}

			out, err := yaml.Marshal(ts)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, string(out))
		})
	}
}

func TestRoundTripMarshalUnmarshal(t *testing.T) {
	// Test that marshaling and then unmarshaling preserves the time value
	originalTime := parseTime(t, "2023-10-27 10:48")
	original := TestTimeStruct{
		Time: SimpleTime{Time: originalTime},
	}

	// Marshal to YAML
	yamlData, err := yaml.Marshal(original)
	assert.NoError(t, err)

	// Unmarshal back
	var restored TestTimeStruct
	err = yaml.Unmarshal(yamlData, &restored)
	assert.NoError(t, err)

	// Compare
	assert.Equal(t, original.Time.Format(ExpectedTimeFormat),
		restored.Time.Format(ExpectedTimeFormat))
}

func parseTime(t *testing.T, timeStr string) time.Time {
	parsed, err := time.Parse(ExpectedTimeFormat, timeStr)
	if err != nil {
		t.Fatalf("Failed to parse test time %s: %v", timeStr, err)
	}
	return parsed
}

type ComplexTimeStruct struct {
	Title     string     `yaml:"title"`
	Created   SimpleTime `yaml:"created"`
	Modified  SimpleTime `yaml:"modified"`
	Published SimpleTime `yaml:"published,omitempty"`
	Tags      []string   `yaml:"tags,omitempty"`
	EmptyTime SimpleTime `yaml:"empty_time"`
}

func TestComplexTimeStruct(t *testing.T) {
	input := `
title: Test Document
created: 2023-01-15 09:30
modified: 2023-02-20 14:45
published: 2023-03-10 12:00
tags:
  - test
  - time
empty_time: ""
`

	var complex ComplexTimeStruct
	err := yaml.Unmarshal([]byte(input), &complex)
	assert.NoError(t, err)

	assert.Equal(t, "Test Document", complex.Title)
	assert.Equal(t, "2023-01-15 09:30", complex.Created.Format(ExpectedTimeFormat))
	assert.Equal(t, "2023-02-20 14:45", complex.Modified.Format(ExpectedTimeFormat))
	assert.Equal(t, "2023-03-10 12:00", complex.Published.Format(ExpectedTimeFormat))
	assert.Equal(t, []string{"test", "time"}, complex.Tags)
	assert.True(t, complex.EmptyTime.IsZero())

	yamlData, err := yaml.Marshal(complex)
	assert.NoError(t, err)

	// Unmarshal again to verify round-trip
	var restored ComplexTimeStruct
	err = yaml.Unmarshal(yamlData, &restored)
	assert.NoError(t, err)

	// Verify fields match
	assert.Equal(t, complex.Title, restored.Title)
	assert.Equal(t, complex.Created.Format(ExpectedTimeFormat),
		restored.Created.Format(ExpectedTimeFormat))
	assert.Equal(t, complex.Modified.Format(ExpectedTimeFormat),
		restored.Modified.Format(ExpectedTimeFormat))
	assert.Equal(t, complex.Published.Format(ExpectedTimeFormat),
		restored.Published.Format(ExpectedTimeFormat))
	assert.Equal(t, complex.Tags, restored.Tags)
	assert.True(t, restored.EmptyTime.IsZero())
}

// Test that invalid time fields cause parsing failure
func TestInvalidTimeInComplexStruct(t *testing.T) {
	input := `
title: Test Document
created: 2023-01-15 09:30
modified: invalid-time-format
tags:
  - test
`

	var complex ComplexTimeStruct
	err := yaml.Unmarshal([]byte(input), &complex)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "time must be in format 2006-01-02 15:04")
}
