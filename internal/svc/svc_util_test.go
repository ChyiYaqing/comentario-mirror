package svc

import (
	"database/sql"
	"errors"
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
			if err := translateErrors(tt.errs...); err != tt.wantErr {
				t.Errorf("translateErrors() error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}
