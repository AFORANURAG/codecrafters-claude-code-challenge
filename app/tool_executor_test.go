package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExecuteRead(t *testing.T) {
	path := filepath.Join(t.TempDir(), "notes.txt")
	if err := os.WriteFile(path, []byte("hello from file"), 0644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	got := executeRead(`{"file_path":"` + path + `"}`)
	if got != "hello from file" {
		t.Fatalf("executeRead = %q, want file contents", got)
	}
}

func TestExecuteReadReportsInvalidArguments(t *testing.T) {
	got := executeRead(`{`)
	if !strings.Contains(got, "Invalid Read arguments") {
		t.Fatalf("executeRead = %q, want invalid arguments error", got)
	}
}

func TestExecuteWrite(t *testing.T) {
	path := filepath.Join(t.TempDir(), "notes.txt")

	got := executeWrite(`{"file_path":"` + path + `","content":"saved"}`)
	if got != "Successfully wrote to file" {
		t.Fatalf("executeWrite = %q, want success message", got)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}
	if string(content) != "saved" {
		t.Fatalf("file content = %q, want %q", string(content), "saved")
	}
}

func TestExecuteBash(t *testing.T) {
	got := executeBash(context.Background(), `{"command":"printf hello"}`)
	if got != "hello" {
		t.Fatalf("executeBash = %q, want command output", got)
	}
}

func TestExecuteBashReportsFailures(t *testing.T) {
	got := executeBash(context.Background(), `{"command":"printf nope && exit 7"}`)
	if !strings.Contains(got, "error while executing the bash command") {
		t.Fatalf("executeBash = %q, want execution error", got)
	}
	if !strings.Contains(got, "nope") {
		t.Fatalf("executeBash = %q, want combined command output", got)
	}
}
