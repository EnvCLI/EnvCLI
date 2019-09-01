package common

import (
	"testing"
)

func TestParseAndEscapeArgs(t *testing.T) {
	var testArgs []string
	testArgs = append(testArgs, "go")
	testArgs = append(testArgs, "build")
	testArgs = append(testArgs, "-ldflags=-w -X main.Example=common")

	AssertStringEquals(t, ParseAndEscapeArgs(testArgs), "\"go\" \"build\" \"-ldflags=-w -X main.Example=common\"")

}

func AssertStringEquals(t *testing.T, value string, expected string) {
	if value != expected {
		t.Errorf("Failed to correctly parse the provided arguments! Expected: " + expected + ", got " + value)
	}
}
