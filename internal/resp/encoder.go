package resp

import (
	"fmt"
	"strconv"
	"strings"
)

func BulkString(raw string) string {
	// $<length>\r\n<data>\r\n
	return fmt.Sprintf("$%d\r\n%s\r\n", len(raw), raw)
}

func SimpleString(raw string) string {
	return fmt.Sprintf("+%s\r\n", raw)
}

func Integer(raw int) string {
	stringRaw := strconv.Itoa(raw)
	return fmt.Sprintf(":%s\r\n", stringRaw)
}

func Array(values []string) string {
	if len(values) == 0 {
		return "*0\r\n"
	}

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("*%d\r\n", len(values)))

	for _, value := range values {
		builder.WriteString(BulkString(value))
	}

	return builder.String()
}
