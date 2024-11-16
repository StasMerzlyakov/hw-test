package main

import (
	"crypto/rand"
	"encoding/hex"
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

			output := tempFileName("temp_output", ".tmp")

			require.NoError(t, Copy(fromPath, output, test.offset, test.limit))

			expectedFile := path.Join(TestDataDir, test.expected)

			expected, err := os.ReadFile(expectedFile)
			require.NoError(t, err)

			actual, err := os.ReadFile(output)
			require.NoError(t, err)

			require.True(t, reflect.DeepEqual(expected, actual))
			require.NoError(t, os.Remove(output))
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
