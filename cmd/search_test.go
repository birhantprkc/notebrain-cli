package cmd

import (
	"reflect"
	"testing"
)

func TestResolveQueries(t *testing.T) {
	tests := []struct {
		name    string
		queries []string
		want    []string
	}{
		{
			name:    "single query",
			queries: []string{"redis"},
			want:    []string{"redis"},
		},
		{
			name:    "multi positional queries",
			queries: []string{"redis", "message broker"},
			want:    []string{"redis", "message broker"},
		},
		{
			name:    "deduplication and whitespace trimming",
			queries: []string{"  redis ", "redis", "message broker"},
			want:    []string{"redis", "message broker"},
		},
		{
			name:    "empty strings ignored",
			queries: []string{"", "  "},
			want:    []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveQueries(tt.queries)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("resolveQueries() = %v, want %v", got, tt.want)
			}
		})
	}
}
