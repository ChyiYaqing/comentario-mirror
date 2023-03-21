package svc

import (
	"database/sql"
	"errors"
	"gitlab.com/comentario/comentario/internal/api/models"
	"testing"
)

func Test_checkErrors(t *testing.T) {
	tests := []struct {
		name    string
		errs    []error
		wantErr error
	}{
		{"No error       ", nil, nil},
		{"Multiple nils  ", []error{nil, nil, nil, nil}, nil},
		{"Single error   ", []error{sql.ErrNoRows}, sql.ErrNoRows},
		{"Mix nils/errors", []error{nil, nil, nil, nil, sql.ErrNoRows, nil, sql.ErrConnDone}, sql.ErrNoRows},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := checkErrors(tt.errs...); err != tt.wantErr {
				t.Errorf("checkErrors() error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func Test_fixCommenterHex(t *testing.T) {
	tests := []struct {
		name string
		id   models.HexID
		want string
	}{
		{"empty", "", ""},
		{"anonymous", "0000000000000000000000000000000000000000000000000000000000000000", "anonymous"},
		{"non-anonymous", "0000000000000000000000000000000000000000000000000000000000000001", "0000000000000000000000000000000000000000000000000000000000000001"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fixCommenterHex(tt.id); got != tt.want {
				t.Errorf("fixCommenterHex() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_fixIdP(t *testing.T) {
	tests := []struct {
		name string
		idp  string
		want string
	}{
		{"empty", "", "commento"},
		{"non-empty", "google", "google"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fixIdP(tt.idp); got != tt.want {
				t.Errorf("fixIdP() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_fixNone(t *testing.T) {
	tests := []struct {
		name string
		id   models.HexID
		want string
	}{
		{"empty", "", "none"},
		{"non-empty", "foo", "foo"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fixNone(tt.id); got != tt.want {
				t.Errorf("fixNone() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_fixUndefined(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want string
	}{
		{"empty", "", "undefined"},
		{"non-empty", "foo", "foo"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fixUndefined(tt.s); got != tt.want {
				t.Errorf("fixUndefined() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_translateError(t *testing.T) {
	tests := []struct {
		name    string
		errs    []error
		wantErr error
	}{
		{"No error", nil, nil},
		{"Empty errors", []error{}, nil},
		{"Multiple nils", []error{nil, nil, nil, nil, nil, nil, nil}, nil},
		{"NotFound error", []error{sql.ErrNoRows}, ErrNotFound},
		{"Other Mongo error", []error{sql.ErrConnDone}, ErrDB},
		{"Custom error", []error{errors.New("ouch")}, ErrDB},
		{"Multiple errors", []error{errors.New("ouch"), sql.ErrNoRows, sql.ErrConnDone}, ErrDB},
		{"Mix of nils and errors", []error{nil, nil, nil, nil, nil, nil, nil, sql.ErrNoRows, nil}, ErrNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := translateDBErrors(tt.errs...); err != tt.wantErr {
				t.Errorf("translateDBErrors() error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func Test_unfixCommenterHex(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want models.HexID
	}{
		{"empty", "", ""},
		{"anonymous", "anonymous", "0000000000000000000000000000000000000000000000000000000000000000"},
		{"non-anonymous", "0000000000000000000000000000000000000000000000000000000000000001", "0000000000000000000000000000000000000000000000000000000000000001"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := unfixCommenterHex(tt.s); got != tt.want {
				t.Errorf("unfixCommenterHex() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_unfixIdP(t *testing.T) {
	tests := []struct {
		name string
		idp  string
		want string
	}{
		{"commento", "commento", ""},
		{"non-empty", "google", "google"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := unfixIdP(tt.idp); got != tt.want {
				t.Errorf("unfixIdP() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_unfixNone(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want models.HexID
	}{
		{"undefined", "none", ""},
		{"non-empty", "foo", "foo"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := unfixNone(tt.s); got != tt.want {
				t.Errorf("unfixNone() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_unfixUndefined(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want string
	}{
		{"undefined", "undefined", ""},
		{"non-empty", "foo", "foo"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := unfixUndefined(tt.s); got != tt.want {
				t.Errorf("unfixUndefined() = %v, want %v", got, tt.want)
			}
		})
	}
}
