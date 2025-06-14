package main

import (
	"testing"
)

func TestCleanInput(t *testing.T) {

	cases := []struct {
		input    string
		expected []string
	}{
		{
			input:    "	hello  world	",
			expected: []string{"hello", "world"},
		},
		{
			input:    "foo bar",
			expected: []string{"foo", "bar"},
		},
		{
			input:    "bang whizz",
			expected: []string{"bang", "whizz"},
		},
	}
}
