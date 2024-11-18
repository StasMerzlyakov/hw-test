package main

import (
	"errors"
	"fmt"
	"io"
	"os"
)

var (
	ErrUnsupportedFile       = errors.New("unsupported file")
	ErrOffsetExceedsFileSize = errors.New("offset exceeds file size")
)

func Copy(fromPath, toPath string, offset, limit int64) error {
	var fi os.FileInfo
	var err error
	if fi, err = os.Stat(fromPath); err != nil {
		return fmt.Errorf("%w - %v", ErrUnsupportedFile, err.Error())
	}

	if fi.Size() == 0 {
		return fmt.Errorf("%w - unknown input file size", ErrUnsupportedFile)
	}

	if fi.Size() < offset {
		return fmt.Errorf("%w - size: %d, offset: %d", ErrOffsetExceedsFileSize, fi.Size(), offset)
	}

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
		panic(err)
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
