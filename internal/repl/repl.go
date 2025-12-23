package repl

import "strings"

// CleanInput splits the input string by spaces, trims whitespace characters,
// and converts all strings to lowercase.
func CleanInput(input string) []string {
	var out []string
	for w := range strings.FieldsSeq(input) {
		out = append(out, strings.ToLower(w))
	}
	return out
}
