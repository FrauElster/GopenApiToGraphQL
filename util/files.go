package util

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
)

func ToAbsolutePath(rawPath string) (string, error) {
	if strings.HasPrefix(rawPath, "/") {
		return rawPath, nil
	}

	expath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("could not determine executable path: %w", err)
	}
	return path.Join(path.Dir(expath), rawPath), nil
}

func FileExists(path string) (bool, error) {
	if _, err := os.Stat(path); err == nil {
		return true, nil
	} else if errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else {
		return false, fmt.Errorf("could not check %s: %w", path, err)
	}
}
