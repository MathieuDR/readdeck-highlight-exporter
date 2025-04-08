package model

import (
	"fmt"
	"strings"
	"time"
)

type SimpleTime struct {
	time.Time
}

// The expected time format for our application
const ExpectedTimeFormat = "2006-01-02 15:04"

// UnmarshalYAML implements the yaml.Unmarshaler interface for SimpleTime
func (st *SimpleTime) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}

	s = strings.TrimSpace(s)
	if s == "" {
		// Return a zero time for empty strings
		st.Time = time.Time{}
		return nil
	}

	// Try to parse with our expected format
	parsedTime, err := time.Parse(ExpectedTimeFormat, s)
	if err != nil {
		return fmt.Errorf("time must be in format %s, got: %s", ExpectedTimeFormat, s)
	}

	st.Time = parsedTime
	return nil
}

func (st SimpleTime) MarshalYAML() (interface{}, error) {
	// If it's a zero time, return empty string
	if st.Time.IsZero() {
		return "", nil
	}
	return st.Time.Format(ExpectedTimeFormat), nil
}
