package resp

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type Reader struct {
	reader *bufio.Reader
}

func NewReader(reader io.Reader) *Reader {
	return &Reader{
		reader: bufio.NewReader(reader),
	}
}

func (r *Reader) ReadLine() (string, error) {
	line, err := r.reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	if !strings.HasSuffix(line, "\r\n") {
		return "", fmt.Errorf("invalid RESP line ending")
	}

	return strings.TrimSuffix(line, "\r\n"), nil
}

func (r *Reader) ReadBulkString() ([]byte, error) {
	line, err := r.ReadLine()
	if err != nil {
		return nil, err
	}

	if len(line) == 0 || line[0] != '$' {
		return nil, fmt.Errorf("expected bulk string, got %q", line)
	}

	length, err := strconv.Atoi(line[1:])
	if err != nil {
		return nil, err
	}

	if length < 0 {
		return nil, nil
	}

	buf := make([]byte, length+2)
	if _, err := io.ReadFull(r.reader, buf); err != nil {
		return nil, err
	}

	if string(buf[length:]) != "\r\n" {
		return nil, fmt.Errorf("invalid bulk string ending")
	}

	return buf[:length], nil
}

func (r *Reader) ReadArray() ([]string, error) {
	line, err := r.ReadLine()
	if err != nil {
		return nil, err
	}

	if len(line) == 0 || line[0] != '*' {
		return nil, fmt.Errorf("expected array, got %q", line)
	}

	numArgs, err := strconv.Atoi(line[1:])
	if err != nil {
		return nil, err
	}

	if numArgs < 0 {
		return nil, nil
	}

	args := make([]string, 0, numArgs)
	for i := 0; i < numArgs; i++ {
		value, err := r.ReadBulkString()
		if err != nil {
			return nil, err
		}

		args = append(args, string(value))
	}

	return args, nil
}
