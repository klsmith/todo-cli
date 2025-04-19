package errs

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestErrs_New(t *testing.T) {
	assert.Error(t, New("Message"), "Message\n")
}

func TestErrs_Wrap_nil(t *testing.T) {
	assert.Error(t, Wrap("Message", nil), "Message\n")
}

func TestErrs_Wrap_notNil(t *testing.T) {
	assert.Error(t, Wrap("Outer", New("Inner")), "Outer\n\t| Inner\n")
}

func TestErrs_MaybePanic_nil(t *testing.T) {
	assert.NotPanics(t, func() {
		MaybePanic("Message", nil)
	})
}

func TestErrs_MaybePanic_nonNil(t *testing.T) {
	assert.Panics(t, func() {
		MaybePanic("Outer", New("Inner"))
	}, "Outer\n\t| Inner\n")
}
