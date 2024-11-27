package main

import (
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const TestData = "./testdata"

func TestReadDirEmpty(t *testing.T) {
	envData := path.Join(TestData, "env")
	env, err := ReadDir(envData)
	require.NoError(t, err)

	assert.Equal(t, EnvValue{Value: "\"hello\"", NeedRemove: false}, env["HELLO"])
	assert.Equal(t, EnvValue{Value: "bar", NeedRemove: false}, env["BAR"])
	assert.Equal(t, EnvValue{Value: "   foo\nwith new line", NeedRemove: false}, env["FOO"])
	assert.Equal(t, EnvValue{NeedRemove: true}, env["EMPTY"])
	assert.Equal(t, EnvValue{NeedRemove: true}, env["UNSET"])
}

func TestProcessBytes(t *testing.T) {
	testCases := []struct {
		Name     string
		Input    []byte
		Expected string
	}{
		{
			"empty",
			[]byte{},
			"",
		},
		{
			"simple",
			[]byte("test value"),
			"test value",
		},
		{
			"tabs",
			[]byte("test value\t\t    \t123\t\t  \t "),
			"test value\t\t    \t123",
		},
		{
			"space",
			[]byte{' '},
			"",
		},
	}

	for _, test := range testCases {
		t.Run(test.Name, func(t *testing.T) {
			result := ProcessBytes(test.Input)
			assert.Equal(t, test.Expected, result)
		})
	}
}
