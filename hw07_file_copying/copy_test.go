package main

import (
	"crypto/rand"
	"encoding/hex"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

const TestDataDir = "testdata"

func TestCopy(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		offset   int64
		limit    int64
		expected string
	}{
		{
			"offset0_limit0",
			"input.txt",
			0,
			0,
			"out_offset0_limit0.txt",
		},
		{
			"offset0_limit10",
			"input.txt",
			0,
			10,
			"out_offset0_limit10.txt",
		},
		{
			"offset0_limit1000",
			"input.txt",
			0,
			1000,
			"out_offset0_limit1000.txt",
		},
		{
			"offset0_limit10000",
			"input.txt",
			0,
			10000,
			"out_offset0_limit10000.txt",
		},
		{
			"offset100_limit1000",
			"input.txt",
			100,
			1000,
			"out_offset100_limit1000.txt",
		},
		{
			"offset6000_limit1000",
			"input.txt",
			6000,
			1000,
			"out_offset6000_limit1000.txt",
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			fromPath := path.Join(TestDataDir, test.input)

			toPath := tempFileName("temp_output", ".tmp")

			require.NoError(t, Copy(fromPath, toPath, test.offset, test.limit))

			expectedFile := path.Join(TestDataDir, test.expected)

			expected, err := os.ReadFile(expectedFile)
			require.NoError(t, err)

			actual, err := os.ReadFile(toPath)
			require.NoError(t, err)

			require.True(t, reflect.DeepEqual(expected, actual))
			require.NoError(t, os.Remove(toPath))
		})
	}
}

func tempFileName(prefix, suffix string) string {
	randBytes := make([]byte, 16)
	if _, err := rand.Read(randBytes); err != nil {
		panic(err)
	}

	return filepath.Join(os.TempDir(), prefix+hex.EncodeToString(randBytes)+suffix)
}

func TestOffsetToLong(t *testing.T) {
	fromPath := path.Join(TestDataDir, "input.txt")

	fi, err := os.Stat(fromPath)
	require.NoError(t, err)

	tooLongOffset := fi.Size() + 100

	toPath := tempFileName("temp_output", ".tmp")

	err = Copy(fromPath, toPath, tooLongOffset, 0)

	// проверка что ошибка вообще есть;
	// возможно избыточно, но наткнулся на такую ошибку:
	// errors: Is(nil) behaves unexpectedly #40442
	// https://github.com/golang/go/issues/40442
	require.Error(t, err)

	require.ErrorIs(t, err, ErrOffsetExceedsFileSize)

	_, err = os.Stat(toPath)
	require.Error(t, err)
	require.ErrorIs(t, err, fs.ErrNotExist)
}

func TestUnknownFileSize(t *testing.T) {
	fromPath := "/dev/urandom"
	toPath := tempFileName("temp_output", ".tmp")

	err := Copy(fromPath, toPath, 0, 0)

	require.Error(t, err)

	require.ErrorIs(t, err, ErrUnsupportedFile)

	_, err = os.Stat(toPath)
	require.Error(t, err)
	require.ErrorIs(t, err, fs.ErrNotExist)
}
