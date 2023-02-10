package api

import (
	"testing"
)

func TestCommenterTokenNewBasics(t *testing.T) {
	FailTestOnError(t, SetupTestEnv())

	if _, err := commenterTokenNew(); err != nil {
		t.Errorf("unexpected error creating new commenterToken: %v", err)
		return
	}
}
