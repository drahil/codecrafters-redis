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
	line, _, err := r.readLineWithByteCount()
	return line, err
}

func (r *Reader) readLineWithByteCount() (string, int, error) {
	line, err := r.reader.ReadString('\n')
	if err != nil {
		return "", 0, err
	}

	if !strings.HasSuffix(line, "\r\n") {
		return "", 0, fmt.Errorf("invalid RESP line ending")
	}

	return strings.TrimSuffix(line, "\r\n"), len(line), nil
}

func (r *Reader) ReadBulkString() ([]byte, error) {
	value, _, err := r.readBulkStringWithByteCount()
	return value, err
}

func (r *Reader) readBulkStringWithByteCount() ([]byte, int, error) {
	length, byteCount, err := r.readBulkLengthWithByteCount()
	if err != nil {
		return nil, 0, err
	}

	if length < 0 {
		return nil, byteCount, nil
	}

	buf := make([]byte, length+2)
	if _, err := io.ReadFull(r.reader, buf); err != nil {
		return nil, 0, err
	}

	if string(buf[length:]) != "\r\n" {
		return nil, 0, fmt.Errorf("invalid bulk string ending")
	}

	return buf[:length], byteCount + len(buf), nil
}

func (r *Reader) ReadBulkPayload() ([]byte, error) {
	length, err := r.readBulkLength()
	if err != nil {
		return nil, err
	}

	if length < 0 {
		return nil, nil
	}

	buf := make([]byte, length)
	if _, err := io.ReadFull(r.reader, buf); err != nil {
		return nil, err
	}

	return buf, nil
}

func (r *Reader) readBulkLength() (int, error) {
	length, _, err := r.readBulkLengthWithByteCount()
	return length, err
}

func (r *Reader) readBulkLengthWithByteCount() (int, int, error) {
	line, byteCount, err := r.readLineWithByteCount()
	if err != nil {
		return 0, 0, err
	}

	if len(line) == 0 || line[0] != '$' {
		return 0, 0, fmt.Errorf("expected bulk string, got %q", line)
	}

	length, err := strconv.Atoi(line[1:])
	if err != nil {
		return 0, 0, err
	}

	return length, byteCount, nil
}

func (r *Reader) ReadArray() ([]string, error) {
	args, _, err := r.ReadArrayWithByteCount()
	return args, err
}

func (r *Reader) ReadArrayWithByteCount() ([]string, int, error) {
	line, byteCount, err := r.readLineWithByteCount()
	if err != nil {
		return nil, 0, err
	}

	if len(line) == 0 || line[0] != '*' {
		return nil, 0, fmt.Errorf("expected array, got %q", line)
	}

	numArgs, err := strconv.Atoi(line[1:])
	if err != nil {
		return nil, 0, err
	}

	if numArgs < 0 {
		return nil, byteCount, nil
	}

	args := make([]string, 0, numArgs)
	for i := 0; i < numArgs; i++ {
		value, argByteCount, err := r.readBulkStringWithByteCount()
		if err != nil {
			return nil, 0, err
		}

		args = append(args, string(value))
		byteCount += argByteCount
	}

	return args, byteCount, nil
}
