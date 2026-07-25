package config

import (
	"bufio"
	"os"
	"strings"
)

func LoadEnv(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		key, value, ok := strings.Cut(scanner.Text(), "=")
		if ok {
			if err := os.Setenv(strings.TrimSpace(key), strings.Trim(strings.TrimSpace(value), `"'`)); err != nil {
				return err
			}
		}
	}
	return scanner.Err()
}
