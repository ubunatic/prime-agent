package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// CommitInfo holds basic commit metadata.
type CommitInfo struct {
	Hash    string `json:"hash"`
	Author  string `json:"author"`
	Date    string `json:"date"`
	Subject string `json:"subject"`
}

// MergeResult holds the result of a git merge operation.
type MergeResult struct {
	Clean           bool     `json:"clean"`
	Output          string   `json:"output"`
	ConflictedFiles []string `json:"conflicted_files,omitempty"`
}

// GitRunner abstracts git command execution for testing and runtime.
type GitRunner interface {
	Run(dir string, args ...string) (string, error)
}

// ExecGitRunner executes git commands via os/exec.
type ExecGitRunner struct{}

func (e ExecGitRunner) Run(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		combined := strings.TrimSpace(stderr.String())
		if combined == "" {
			combined = strings.TrimSpace(stdout.String())
		}
		return stdout.String(), fmt.Errorf("git %s failed: %w: %s", strings.Join(args, " "), err, combined)
	}
	return stdout.String(), nil
}

var defaultGitRunner GitRunner = ExecGitRunner{}

// FindRepoRoot returns the top-level repository root directory.
func FindRepoRoot(startDir string) (string, error) {
	out, err := defaultGitRunner.Run(startDir, "rev-parse", "--show-toplevel")
	if err != nil {
		return "", err
	}
	root := strings.TrimSpace(out)
	return filepath.Clean(root), nil
}

// GetRemotes returns a map of remote name to URL.
func GetRemotes(repoRoot string) (map[string]string, error) {
	out, err := defaultGitRunner.Run(repoRoot, "remote", "-v")
	if err != nil {
		return nil, err
	}
	remotes := make(map[string]string)
	lines := strings.Split(out, "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			name := fields[0]
			url := fields[1]
			remotes[name] = url
		}
	}
	return remotes, nil
}

// EnsureUpstreamRemote checks if the upstream remote is configured, and adds it if missing.
func EnsureUpstreamRemote(repoRoot, remoteName, defaultURL string) (string, bool, error) {
	remotes, err := GetRemotes(repoRoot)
	if err != nil {
		return "", false, err
	}

	if url, exists := remotes[remoteName]; exists {
		return url, false, nil
	}

	_, err = defaultGitRunner.Run(repoRoot, "remote", "add", remoteName, defaultURL)
	if err != nil {
		return "", false, fmt.Errorf("failed to add remote %s (%s): %w", remoteName, defaultURL, err)
	}
	return defaultURL, true, nil
}

// FetchRemote fetches references from the specified remote.
func FetchRemote(repoRoot, remoteName string) error {
	_, err := defaultGitRunner.Run(repoRoot, "fetch", remoteName)
	if err != nil {
		return fmt.Errorf("failed to fetch %s: %w", remoteName, err)
	}
	return nil
}

// GetCurrentBranch returns the name of the currently checked out branch.
func GetCurrentBranch(repoRoot string) (string, error) {
	out, err := defaultGitRunner.Run(repoRoot, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// RefExists checks if a git reference or commit exists in the repository.
func RefExists(repoRoot, ref string) bool {
	_, err := defaultGitRunner.Run(repoRoot, "rev-parse", "--verify", "--quiet", ref)
	return err == nil
}

// IsWorkingTreeClean returns true if there are no staged or unstaged changes.
func IsWorkingTreeClean(repoRoot string) (bool, string, error) {
	out, err := defaultGitRunner.Run(repoRoot, "status", "--porcelain")
	if err != nil {
		return false, "", err
	}
	status := strings.TrimSpace(out)
	return status == "", status, nil
}

// GetIncomingCommits returns the commits between local baseRef and upstream targetRef.
func GetIncomingCommits(repoRoot, baseRef, targetRef string) ([]CommitInfo, error) {
	rangeSpec := fmt.Sprintf("%s..%s", baseRef, targetRef)
	out, err := defaultGitRunner.Run(repoRoot, "log", rangeSpec, "--pretty=format:%h%x09%an%x09%ad%x09%s", "--date=short")
	if err != nil {
		return nil, err
	}

	out = strings.TrimSpace(out)
	if out == "" {
		return nil, nil
	}

	var commits []CommitInfo
	lines := strings.Split(out, "\n")
	for _, line := range lines {
		parts := strings.Split(line, "\t")
		if len(parts) >= 4 {
			commits = append(commits, CommitInfo{
				Hash:    parts[0],
				Author:  parts[1],
				Date:    parts[2],
				Subject: parts[3],
			})
		}
	}
	return commits, nil
}

// GetDiffFiles returns the list of files modified between baseRef and targetRef.
func GetDiffFiles(repoRoot, baseRef, targetRef string) ([]string, error) {
	rangeSpec := fmt.Sprintf("%s...%s", baseRef, targetRef)
	out, err := defaultGitRunner.Run(repoRoot, "diff", "--name-only", rangeSpec)
	if err != nil {
		// Fallback to two-dot diff if three-dot fails
		out, err = defaultGitRunner.Run(repoRoot, "diff", "--name-only", fmt.Sprintf("%s..%s", baseRef, targetRef))
		if err != nil {
			return nil, err
		}
	}

	out = strings.TrimSpace(out)
	if out == "" {
		return nil, nil
	}

	var files []string
	for _, line := range strings.Split(out, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			files = append(files, trimmed)
		}
	}
	return files, nil
}

// CreateAndCheckoutBranch creates and checks out a new branch.
func CreateAndCheckoutBranch(repoRoot, branchName, startPoint string) error {
	args := []string{"checkout", "-b", branchName}
	if startPoint != "" {
		args = append(args, startPoint)
	}
	_, err := defaultGitRunner.Run(repoRoot, args...)
	return err
}

// GetConflictedFiles returns a list of files with unresolved merge conflicts.
func GetConflictedFiles(repoRoot string) ([]string, error) {
	out, err := defaultGitRunner.Run(repoRoot, "diff", "--name-only", "--diff-filter=U")
	if err != nil {
		return nil, err
	}
	out = strings.TrimSpace(out)
	if out == "" {
		return nil, nil
	}
	var files []string
	for _, line := range strings.Split(out, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			files = append(files, trimmed)
		}
	}
	return files, nil
}

// MergeBranch attempts to merge targetRef into current branch.
func MergeBranch(repoRoot, targetRef string) (MergeResult, error) {
	out, err := defaultGitRunner.Run(repoRoot, "merge", "--no-ff", targetRef)
	if err != nil {
		// Check if there are conflicted files
		conflicts, conflictErr := GetConflictedFiles(repoRoot)
		if conflictErr == nil && len(conflicts) > 0 {
			return MergeResult{
				Clean:           false,
				Output:          out,
				ConflictedFiles: conflicts,
			}, nil
		}
		return MergeResult{Clean: false, Output: out}, err
	}
	return MergeResult{
		Clean:  true,
		Output: out,
	}, nil
}
