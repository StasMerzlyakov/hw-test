package main

import (
	"os"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunCmd(t *testing.T) {
	require.NoError(t, os.Setenv("TEST", "1"))
	defer func() {
		require.NoError(t, os.Unsetenv("TEST"))
	}()

	env := Environment{
		"TEST": EnvValue{
			Value:      "",
			NeedRemove: true,
		},

		"MY_ENV_VARS": EnvValue{
			Value: "1",
		},

		"MY_ANOTHER_ENV_VARS": EnvValue{
			Value: "test val",
		},
	}

	closer := envSetter(env)
	t.Cleanup(closer)
}

func TestMerge(t *testing.T) {
	osEnv := map[string]string{
		"TEST":                "123",
		"MY_ENV_VARS":         "abc",
		"MY_ANOTHER_ENV_VARS": "test",
		"A":                   "S",
	}

	env := Environment{
		"TEST": EnvValue{
			NeedRemove: true,
		},

		"MY_ENV_VARS": EnvValue{
			Value: "1",
		},

		"MY_ANOTHER_ENV_VARS": EnvValue{
			Value: "test val",
		},
	}

	result := merge(osEnv, env)

	require.Equal(t, 3, len(result))

	sort.Strings(result)

	assert.True(t, result[0] == "A=S")
	assert.True(t, result[1] == "MY_ANOTHER_ENV_VARS=test val")
	assert.True(t, result[2] == "MY_ENV_VARS=1")
}

func TestEnvToMap(t *testing.T) {
	input := []string{
		"GPG_AGENT_INFO=/run/user/1000/gnupg/S.gpg-agent:0:1",
		"NLSPATH=/opt/cprocsp/share/locale/%L/LC_MESSAGES/%N:/opt/cprocsp/share/locale/%L/LC_MESSAGES/%N",
	}

	result := envToMap(input)
	require.Equal(t, 2, len(result))

	require.Equal(t, "/run/user/1000/gnupg/S.gpg-agent:0:1", result["GPG_AGENT_INFO"])
	require.Equal(t, "/opt/cprocsp/share/locale/%L/LC_MESSAGES/%N:/opt/cprocsp/share/locale/%L/LC_MESSAGES/%N",
		result["NLSPATH"])
}

// The idea from https://dev.to/arxeiss/auto-reset-environment-variables-when-testing-in-go-5ec.
func envSetter(env Environment) (closer func()) {
	originalEnvs := envToMap(os.Environ())

	for name, envValue := range env {
		if originalValue, ok := os.LookupEnv(name); ok {
			originalEnvs[name] = originalValue
		}

		if envValue.NeedRemove {
			_ = os.Unsetenv(name)
		} else {
			_ = os.Setenv(name, envValue.Value)
		}
	}

	return func() {
		for name := range env {
			origValue, has := originalEnvs[name]
			if has {
				_ = os.Setenv(name, origValue)
			} else {
				_ = os.Unsetenv(name)
			}
		}
	}
}
