package test

import (
	"os"
	"path/filepath"
	"testing"
)

func TempCopyDir(t *testing.T, source string) string {
	t.Helper()

	source, err := filepath.Abs(source)
	if err != nil {
		t.Fatal(err)
	}

	files, err := os.ReadDir(source)
	if err != nil {
		t.Fatal(err)
	}

	destination := t.TempDir()

	for _, file := range files {
		data, err := os.ReadFile(filepath.Join(source, file.Name()))
		if err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile(filepath.Join(destination, file.Name()), data, 0o644); err != nil {
			t.Fatal(err)
		}
	}

	return destination
}

func TempReadOnlyCopyDir(t *testing.T, source string) (string, func()) {
	t.Helper()

	destination := TempCopyDir(t, source)
	if err := os.Chmod(destination, 0o555); err != nil {
		t.Fatal(err)
	}

	files, err := os.ReadDir(destination)
	if err != nil {
		t.Fatal(err)
	}

	for _, file := range files {
		if err = os.Chmod(filepath.Join(destination, file.Name()), 0o444); err != nil {
			t.Fatal(err)
		}
	}

	return destination, func() {
		if err := os.Chmod(destination, 0o755); err != nil {
			t.Fatal(err)
		}
	}
}
