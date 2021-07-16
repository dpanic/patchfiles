package logger

import "bytes"

// minWidth will make string with minimal width
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
