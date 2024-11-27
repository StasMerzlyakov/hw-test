package main

import (
	"fmt"
	"io"
	"os"
	"path"
	"strings"
)

type Environment map[string]EnvValue

// EnvValue helps to distinguish between empty files and files with the first empty line.
type EnvValue struct {
	Value      string
	NeedRemove bool
}

// ReadDir reads a specified directory and returns map of env variables.
// Variables represented as files where filename is name of variable, file first line is a value.
func ReadDir(dir string) (Environment, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read dir err %w", err)
	}

	env := make(Environment)

	for _, entry := range entries {
		name := entry.Name()
		if strings.Contains(name, "=") || entry.IsDir() {
			continue
		}

		envVar, err := ProcessFile(path.Join(dir, name))
		if err != nil {
			return nil, err
		}

		env[name] = envVar
	}

	return env, nil
}

func ProcessFile(name string) (EnvValue, error) {
	info, err := os.Stat(name)
	if err != nil {
		return EnvValue{}, fmt.Errorf("can't get stat for file %w", err)
	}

	if info.Size() == 0 {
		return EnvValue{NeedRemove: true}, nil
	}

	f, err := os.Open(name)
	if err != nil {
		return EnvValue{}, fmt.Errorf("can't open file %w", err)
	}

	defer func() {
		if err := f.Close(); err != nil {
			panic(err)
		}
	}()

	fileBytes, err := io.ReadAll(f)
	if err != nil {
		return EnvValue{}, fmt.Errorf("can't read file bytes %w", err)
	}

	processedText := ProcessBytes(fileBytes)
	if len(processedText) > 0 {
		return EnvValue{Value: processedText}, nil
	}
	return EnvValue{NeedRemove: true}, nil
}

func ProcessBytes(bytesScanned []byte) string {
	// Решил что возможны случаю когда на конце встретится последовательность вида "\t \t \t ".
	// В этом случае придется применять в цикле strings.TrimRight(.., "\t") и strings.TrimRight(.., " ")
	// пока будут измененния.
	// Быстрее будет один раз пройтись по массиву

	rightPos := -1

	for pos := range bytesScanned {
		if bytesScanned[pos] == '\n' {
			if rightPos == -1 {
				rightPos = pos
			}
			break
		}

		// replace 0x00 -> '\n'
		if bytesScanned[pos] == 0x00 {
			bytesScanned[pos] = '\n'
		}

		if bytesScanned[pos] == '\t' || bytesScanned[pos] == ' ' {
			if rightPos == -1 {
				rightPos = pos
			}
		} else {
			if rightPos >= 0 {
				rightPos = -1
			}
		}
	}

	if rightPos >= 0 {
		return string(bytesScanned[:rightPos])
	}
	return string(bytesScanned)
}
