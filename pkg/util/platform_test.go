package util

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestCiEnvironmentTrue(t *testing.T) {
	_ = os.Setenv("CI", "true")
	assert.Equal(t, true, IsCIEnvironment())
}

func TestCiEnvironmentFalse(t *testing.T) {
	_ = os.Unsetenv("CI")
	assert.Equal(t, false, IsCIEnvironment())
}
