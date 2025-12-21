package main

import "testing"

func TestCleanInput(t *testing.T) {
	cases := []struct {
		input    string
		expected []string
	}{
		{
			input:    "  hello world  ",
			expected: []string{"hello", "world"},
		},
		{
			input:    "Charmander Bulbasaur PIKACHU",
			expected: []string{"charmander", "bulbasaur", "pikachu"},
		},
		{
			input:    "   Multiple   Spaces   Between   Words   ",
			expected: []string{"multiple", "spaces", "between", "words"},
		},
		{
			input:    "MiXeD CaSe WoRdS",
			expected: []string{"mixed", "case", "words"},
		},
	}

	for _, c := range cases {
		actual := cleanInput(c.input)
		if len(actual) != len(c.expected) {
			t.Errorf("lengths of actual and expected do not match: %v != %v", len(actual), len(c.expected))
		}
		for i := range actual {
			word := actual[i]
			expectedWord := c.expected[i]
			if word != expectedWord {
				t.Errorf("words at index %v do not match: %v != %v", i, word, expectedWord)
			}
		}
	}
}
