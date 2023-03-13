package data

import (
	"gitlab.com/comentario/comentario/internal/api/models"
	"testing"
)

func TestStringHexID(t *testing.T) {
	v := models.HexID("oneTwo")
	tests := []struct {
		name string
		v    *models.HexID
		want string
	}{
		{"nil  ", nil, ""},
		{"value", &v, "oneTwo"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StringHexID(tt.v); got != tt.want {
				t.Errorf("StringHexID() = %v, want %v", got, tt.want)
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
