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
