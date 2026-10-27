package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplitCSV(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{"uma origem", "http://localhost:3000", []string{"http://localhost:3000"}},
		{"duas com espaço", "http://a.com, http://b.com", []string{"http://a.com", "http://b.com"}},
		{"vírgula sobrando no fim", "http://a.com,", []string{"http://a.com"}},
		{"vazio", "", []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, splitCSV(tt.input))
		})
	}
}
