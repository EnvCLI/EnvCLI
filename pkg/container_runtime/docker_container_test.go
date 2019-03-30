package container_runtime

import (
	"testing"
)

func TestSetName(t *testing.T) {
	container := Container{}
	container.SetName("testCase")

	if container.GetName() != "testCase" {
		t.Errorf("Name doesn't equal to is set")
	}
}
