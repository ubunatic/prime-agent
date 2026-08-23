package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLIVersionAndHelp(t *testing.T) {
	// Build the fork-sync binary in a temporary directory
	tmpDir := t.TempDir()
	binPath := filepath.Join(tmpDir, "fork-sync")

	cmd := exec.Command("go", "build", "-o", binPath, ".")
	cmd.Dir = "."
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go build failed: %v: %s", err, string(out))
	}

	// 1. Test --help
	cmdHelp := exec.Command(binPath, "--help")
	var stdout bytes.Buffer
	cmdHelp.Stdout = &stdout
	if err := cmdHelp.Run(); err != nil {
		t.Errorf("fork-sync --help failed: %v", err)
	}
	if !strings.Contains(stdout.String(), "fork-sync") || !strings.Contains(stdout.String(), "Usage:") {
		t.Errorf("Unexpected help output: %s", stdout.String())
	}

	// 2. Test --version
	cmdVer := exec.Command(binPath, "--version")
	stdout.Reset()
	cmdVer.Stdout = &stdout
	if err := cmdVer.Run(); err != nil {
		t.Errorf("fork-sync --version failed: %v", err)
	}
	if !strings.Contains(stdout.String(), version) {
		t.Errorf("Unexpected version output: %s", stdout.String())
	}

	// 3. Test verify subcommand
	repoRoot, err := FindRepoRoot(".")
	if err != nil {
		t.Fatalf("FindRepoRoot failed: %v", err)
	}
	cmdVerify := exec.Command(binPath, "verify", "--repo-root", repoRoot)
	stdout.Reset()
	cmdVerify.Stdout = &stdout
	if err := cmdVerify.Run(); err != nil {
		t.Errorf("fork-sync verify failed: %v", err)
	}
	if !strings.Contains(stdout.String(), "[PASS]") {
		t.Errorf("Expected [PASS] in verify output: %s", stdout.String())
	}
}

func TestPrintUsage(t *testing.T) {
	// Simple test to ensure printUsage runs without panics
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	printUsage()

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	if !strings.Contains(buf.String(), "fork-sync") {
		t.Errorf("printUsage didn't output expected text")
	}
}
