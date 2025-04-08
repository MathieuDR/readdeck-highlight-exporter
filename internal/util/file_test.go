package util_test

import (
	"testing"
	"time"

	"github.com/mathieudr/readdeck-highlight-exporter/internal/util"
	"github.com/stretchr/testify/assert"
)

func TestGenerateId(t *testing.T) {
	tests := []struct {
		// LEARNING
		name string // description of this test case
		// Named input parameters for target function.
		title     string
		timestamp string
		want      string
	}{
		{
			name:      "empty string",
			title:     "",
			timestamp: "2025-03-29T00:30:00Z",
			want:      "1743208200",
		},
		{
			name:      "normal title",
			title:     "my cool project",
			timestamp: "2023-05-25T19:37:59Z",
			want:      "1685043479-my-cool-project",
		},
		{
			name:      "removes multiple dashes",
			title:     "my - cool-project",
			timestamp: "2025-03-29T00:30:00Z",
			want:      "1743208200-my-cool-project",
		},
		{
			name:      "removes bad characters",
			title:     "my%co0ül project",
			timestamp: "2025-03-29T00:30:00Z",
			want:      "1743208200-my-co0-l-project",
		},
	}
	for _, tt := range tests {
		parsedTime, err := time.Parse(time.RFC3339, tt.timestamp)
		assert.NoError(t, err, "Can not parse time for test")

		t.Run(tt.name, func(t *testing.T) {
			got := util.GenerateId(tt.title, parsedTime)
			assert.Equal(t, tt.want, got)
		})
	}
}
