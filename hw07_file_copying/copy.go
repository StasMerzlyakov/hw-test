package main

import (
	"errors"
	"io"
	"os"
)

var (
	ErrUnsupportedFile       = errors.New("unsupported file")
	ErrOffsetExceedsFileSize = errors.New("offset exceeds file size")
)

func Copy(fromPath, toPath string, offset, limit int64) error {
	input, err := os.Open(fromPath)
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := input.Close(); err != nil {
			panic(err)
		}
	}()

	if _, err := input.Seek(offset, io.SeekStart); err != nil {
		panic(err) // offset больше, чем размер файла - невалидная ситуация
	}

	output, err := os.OpenFile(toPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o666)
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := output.Close(); err != nil {
			panic(err)
		}
	}()

	var inReader io.Reader
	if limit > 0 {
		inReader = io.LimitReader(input, limit)
	} else {
		inReader = input
	}

	_, err = io.Copy(output, inReader)

	return err
}
