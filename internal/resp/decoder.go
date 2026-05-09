package resp

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

func parseMessage(message string) []string {
	lines := strings.Split(message, "\r\n")
	var args []string

	i := 0
	if len(lines) == 0 {
		return args
	}

	// First line should be *N (array with N elements)
	if len(lines[0]) == 0 || lines[0][0] != '*' {
		return args
	}
	numArgs, err := strconv.Atoi(lines[0][1:])
	if err != nil {
		return args
	}
	i++

	for j := 0; j < numArgs; j++ {
		if i >= len(lines) {
			break
		}
		// Skip the $N bulk string length prefix
		if len(lines[i]) > 0 && lines[i][0] == '$' {
			i++
		}
		if i >= len(lines) {
			break
		}
		args = append(args, strings.ToLower(lines[i]))
		i++
	}

	return args
}

func GetArgs(conn net.Conn) ([]string, error) {
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)

	raw := string(buf[:n])
	args := parseMessage(raw)
	fmt.Printf("%#v\n", args)

	return args, err
}
