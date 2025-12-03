package logger

import "bytes"

// minWidth pads a string to a minimum width by appending the separator character.
// If the input string is shorter than the minimum width, it appends the separator
// character until the string reaches the desired minimum width.
func minWidth(in, separator string, min int) string {
	diff := min - len(in)

	if diff > 0 {
		var buffer bytes.Buffer
		buffer.WriteString(in)

		for i := 0; i < diff; i++ {
			buffer.WriteString(separator)
		}
		in = buffer.String()
	}

	return in
}
