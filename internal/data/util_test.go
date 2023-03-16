package data

import (
	"github.com/go-openapi/strfmt"
	"testing"
)

func TestEmailToString(t *testing.T) {
	v1 := strfmt.Email("whatever@foo.bar")
	v2 := strfmt.Email("  spaces@foo.bar\n ")
	tests := []struct {
		name string
		v    *strfmt.Email
		want string
	}{
		{"nil       ", nil, ""},
		{"value     ", &v1, "whatever@foo.bar"},
		{"whitespace", &v2, "spaces@foo.bar"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EmailToString(tt.v); got != tt.want {
				t.Errorf("EmailToString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRandomHexID(t *testing.T) {
	t.Run("randomness test", func(t *testing.T) {
		// Generate first ID
		h1, err1 := RandomHexID()
		if err1 != nil {
			t.Errorf("RandomHexID() invocation 1 errored with %v", err1)
		}

		// Generate second ID
		h2, err2 := RandomHexID()
		if err2 != nil {
			t.Errorf("RandomHexID() invocation 2 errored with %v", err2)
		}

		// The IDs must differ
		if h1 == h2 {
			t.Errorf("RandomHexID() generated 2 duplicate IDs = %x", h1)
		}
	})
}

func TestTrimmedString(t *testing.T) {
	v1 := "You see, it's complicated"
	v2 := "  \nBut not as complicated\t"
	tests := []struct {
		name string
		v    *string
		want string
	}{
		{"nil            ", nil, ""},
		{"regular value  ", &v1, "You see, it's complicated"},
		{"with whitespace", &v2, "But not as complicated"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TrimmedString(tt.v); got != tt.want {
				t.Errorf("TrimmedString() = %v, want %v", got, tt.want)
			}
		})
	}
}
