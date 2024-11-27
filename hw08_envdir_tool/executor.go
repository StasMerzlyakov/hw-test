package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// RunCmd runs a command + arguments (cmd) with environment variables from env.
func RunCmd(cmd []string, env Environment) (returnCode int) {
	sysEnv := envToMap(os.Environ())

	command := cmd[0]
	args := cmd[1:]
	process := exec.Command(command, args...)
	process.Stdin = os.Stdin
	process.Stdout = os.Stdout
	process.Stderr = os.Stderr

	process.Env = merge(sysEnv, env)

	if err := process.Run(); err != nil {
		var e *exec.ExitError

		if errors.As(err, &e) {
			return e.ExitCode()
		}
	}

	return 0
}

func envToMap(env []string) map[string]string {
	resultMap := make(map[string]string, len(env))
	for _, keyAndVal := range env {
		keyAndVal = strings.TrimRight(keyAndVal, "=")
		eqIndx := strings.Index(keyAndVal, "=")
		if (eqIndx) > 0 {
			resultMap[keyAndVal[:eqIndx]] = keyAndVal[eqIndx+1:]
		}
	}
	return resultMap
}

func merge(envKeyValues map[string]string, env Environment) []string {
	for envKey, envValue := range env {
		if envValue.NeedRemove {
			delete(envKeyValues, envKey)
		} else {
			envKeyValues[envKey] = envValue.Value
		}
	}

	result := make([]string, 0, len(envKeyValues))
	for k, v := range envKeyValues {
		result = append(result, fmt.Sprintf("%s=%s", k, v))
	}
	return result
}
