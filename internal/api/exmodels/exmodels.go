package exmodels

import (
	"context"
	"github.com/go-openapi/strfmt"
)

// IdentityProviderMap maps known IdPs to their enabled state. Need to create a dedicated type to address go-swagger
// limitations
type IdentityProviderMap map[string]bool

// Validate is required by go-openapi
func (m IdentityProviderMap) Validate(strfmt.Registry) error {
	return nil
}

// ContextValidate is required by go-openapi
func (m IdentityProviderMap) ContextValidate(context.Context, strfmt.Registry) error {
	return nil
}

// Clone returns a copy of this map
func (m IdentityProviderMap) Clone() IdentityProviderMap {
	r := make(IdentityProviderMap)
	for key, val := range m {
		r[key] = val
	}
	return r
}
