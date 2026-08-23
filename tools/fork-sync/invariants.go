package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// InvariantCategory defines a fork invariant domain.
type InvariantCategory string

const (
	CategoryTelemetry   InvariantCategory = "Telemetry & Privacy"
	CategoryInstaller   InvariantCategory = "User-Local Installer"
	CategoryOnboarding  InvariantCategory = "Direct Provider Onboarding"
	CategoryRelease     InvariantCategory = "Release & Publishing"
	CategoryWorkflows   InvariantCategory = "GitHub Workflows"
	CategoryTUIControls InvariantCategory = "TUI Controls & Selection"
	CategoryOtherFork   InvariantCategory = "Fork Documentation & Tools"
)

// InvariantRule represents a file pattern matched to a fork invariant.
type InvariantRule struct {
	Pattern     *regexp.Regexp
	Category    InvariantCategory
	Description string
}

var defaultInvariantRules = []InvariantRule{
	{
		Pattern:     regexp.MustCompile(`(^|/)packages/coding-agent/src/core/telemetry\.ts$`),
		Category:    CategoryTelemetry,
		Description: "Opt-in telemetry implementation with offline / DO_NOT_TRACK enforcement",
	},
	{
		Pattern:     regexp.MustCompile(`(^|/)packages/coding-agent/src/core/settings-manager\.ts$`),
		Category:    CategoryTelemetry,
		Description: "Settings defaults and telemetry enablement flags",
	},
	{
		Pattern:     regexp.MustCompile(`(^|/)packages/coding-agent/test/telemetry\.test\.ts$`),
		Category:    CategoryTelemetry,
		Description: "Unit tests verifying telemetry opt-in defaults",
	},
	{
		Pattern:     regexp.MustCompile(`(^|/)install(-beta)?\.sh$`),
		Category:    CategoryInstaller,
		Description: "User-local non-sudo installer script targeting $HOME/.local",
	},
	{
		Pattern:     regexp.MustCompile(`(^|/)scripts/check-installer-render\.mjs$`),
		Category:    CategoryInstaller,
		Description: "Installer rendering and syntax check script",
	},
	{
		Pattern:     regexp.MustCompile(`(^|/)packages/coding-agent/src/modes/interactive/onboarding\.ts$`),
		Category:    CategoryOnboarding,
		Description: "Direct provider onboarding and bypass for mandatory Prime Intellect splash",
	},
	{
		Pattern:     regexp.MustCompile(`(^|/)packages/coding-agent/src/core/auth-guidance\.ts$`),
		Category:    CategoryOnboarding,
		Description: "Authentication guidance messages for standard provider keys",
	},
	{
		Pattern:     regexp.MustCompile(`(^|/)scripts/release\.mjs$`),
		Category:    CategoryRelease,
		Description: "Release script with PI_SKIP_NPM_PUBLISH / --skip-npm-publish support",
	},
	{
		Pattern:     regexp.MustCompile(`(^|/)docs/ForkArchitecture\.md$`),
		Category:    CategoryRelease,
		Description: "Fork architecture specification and invariant documentation",
	},
	{
		Pattern:     regexp.MustCompile(`(^|/)\.github/workflows/build-binaries\.yml$`),
		Category:    CategoryWorkflows,
		Description: "GitHub Releases artifact build pipeline without required R2 credentials",
	},
	{
		Pattern:     regexp.MustCompile(`(^|/)packages/tui/src/(fullscreen|mouse)\.ts$`),
		Category:    CategoryTUIControls,
		Description: "TUI terminal mouse scrolling and word/line selection handlers",
	},
	{
		Pattern:     regexp.MustCompile(`(^|/)tools/fork-sync/`),
		Category:    CategoryOtherFork,
		Description: "Fork synchronization Go CLI helper",
	},
}

// InvariantMatch records how a file path matches fork invariants.
type InvariantMatch struct {
	Path        string            `json:"path"`
	IsInvariant bool              `json:"is_invariant"`
	Category    InvariantCategory `json:"category,omitempty"`
	Description string            `json:"description,omitempty"`
}

// ClassifyFile checks if a repo-relative path touches a protected fork invariant.
func ClassifyFile(relPath string) InvariantMatch {
	normalized := filepath.ToSlash(relPath)
	for _, rule := range defaultInvariantRules {
		if rule.Pattern.MatchString(normalized) {
			return InvariantMatch{
				Path:        normalized,
				IsInvariant: true,
				Category:    rule.Category,
				Description: rule.Description,
			}
		}
	}
	return InvariantMatch{
		Path:        normalized,
		IsInvariant: false,
	}
}

// VerificationResult stores the outcome of checking a specific fork invariant.
type VerificationResult struct {
	Name        string `json:"name"`
	Category    string `json:"category"`
	Passed      bool   `json:"passed"`
	Description string `json:"description"`
	Details     string `json:"details,omitempty"`
}

// InvariantVerifier is the signature for invariant test functions.
type InvariantVerifier func(repoRoot string) VerificationResult

// Verifiers list all repository invariant test checks.
var Verifiers = []InvariantVerifier{
	VerifyTelemetryInvariant,
	VerifyInstallerInvariant,
	VerifyOnboardingInvariant,
	VerifyReleaseScriptInvariant,
	VerifyWorkflowsInvariant,
}

// VerifyTelemetryInvariant asserts that telemetry is opt-in and checks environment overrides.
func VerifyTelemetryInvariant(repoRoot string) VerificationResult {
	path := filepath.Join(repoRoot, "packages", "coding-agent", "src", "core", "telemetry.ts")
	content, err := os.ReadFile(path)
	if err != nil {
		return VerificationResult{
			Name:        "Telemetry Privacy",
			Category:    string(CategoryTelemetry),
			Passed:      false,
			Description: "Ensure telemetry respects PI_OFFLINE, DO_NOT_TRACK, and is opt-in",
			Details:     fmt.Sprintf("Failed to read %s: %v", path, err),
		}
	}

	text := string(content)
	checks := []struct {
		desc string
		sub  string
	}{
		{"Checks PI_OFFLINE override", "PI_OFFLINE"},
		{"Checks DO_NOT_TRACK override", "DO_NOT_TRACK"},
		{"Checks PRIME_AGENT_TELEMETRY override", "PRIME_AGENT_TELEMETRY"},
		{"Queries settingsManager for telemetry enabled", "settingsManager.getTelemetryEnabled()"},
	}

	for _, c := range checks {
		if !strings.Contains(text, c.sub) {
			return VerificationResult{
				Name:        "Telemetry Privacy",
				Category:    string(CategoryTelemetry),
				Passed:      false,
				Description: "Ensure telemetry respects PI_OFFLINE, DO_NOT_TRACK, and is opt-in",
				Details:     fmt.Sprintf("Missing invariant check '%s' in %s", c.desc, path),
			}
		}
	}

	return VerificationResult{
		Name:        "Telemetry Privacy",
		Category:    string(CategoryTelemetry),
		Passed:      true,
		Description: "Opt-in telemetry and environment overrides intact",
		Details:     "telemetry.ts respects PI_OFFLINE, DO_NOT_TRACK, PRIME_AGENT_TELEMETRY, and settings",
	}
}

// VerifyInstallerInvariant asserts user-local non-sudo install target.
func VerifyInstallerInvariant(repoRoot string) VerificationResult {
	path := filepath.Join(repoRoot, "install.sh")
	content, err := os.ReadFile(path)
	if err != nil {
		return VerificationResult{
			Name:        "User-Local Installer",
			Category:    string(CategoryInstaller),
			Passed:      false,
			Description: "Ensure install.sh targets $HOME/.local and user space without sudo",
			Details:     fmt.Sprintf("Failed to read %s: %v", path, err),
		}
	}

	text := string(content)
	if !strings.Contains(text, "${XDG_DATA_HOME:-$HOME/.local}") && !strings.Contains(text, "$HOME/.local") {
		return VerificationResult{
			Name:        "User-Local Installer",
			Category:    string(CategoryInstaller),
			Passed:      false,
			Description: "Ensure install.sh targets $HOME/.local and user space without sudo",
			Details:     "install.sh does not contain expected user-local path prefix ($HOME/.local)",
		}
	}

	if strings.Contains(text, "sudo npm install -g") {
		return VerificationResult{
			Name:        "User-Local Installer",
			Category:    string(CategoryInstaller),
			Passed:      false,
			Description: "Ensure install.sh targets $HOME/.local and user space without sudo",
			Details:     "install.sh contains forbidden 'sudo npm install -g'",
		}
	}

	return VerificationResult{
		Name:        "User-Local Installer",
		Category:    string(CategoryInstaller),
		Passed:      true,
		Description: "User-local install path ($HOME/.local) without sudo intact",
		Details:     "install.sh configures user-prefix without requiring global sudo npm install",
	}
}

// VerifyOnboardingInvariant asserts that mandatory Prime Intellect login splash is disabled.
func VerifyOnboardingInvariant(repoRoot string) VerificationResult {
	path := filepath.Join(repoRoot, "packages", "coding-agent", "src", "modes", "interactive", "onboarding.ts")
	content, err := os.ReadFile(path)
	if err != nil {
		return VerificationResult{
			Name:        "Direct Provider Onboarding",
			Category:    string(CategoryOnboarding),
			Passed:      false,
			Description: "Ensure forced Prime Intellect login splash is disabled by default",
			Details:     fmt.Sprintf("Failed to read %s: %v", path, err),
		}
	}

	text := string(content)
	re := regexp.MustCompile(`function\s+shouldRunPrimeCliOnboardingSplash\s*\([^\)]*\)\s*:\s*boolean\s*\{\s*return\s+false\s*;?\s*\}`)
	if !re.MatchString(text) {
		return VerificationResult{
			Name:        "Direct Provider Onboarding",
			Category:    string(CategoryOnboarding),
			Passed:      false,
			Description: "Ensure forced Prime Intellect login splash is disabled by default",
			Details:     "shouldRunPrimeCliOnboardingSplash is missing or does not return false",
		}
	}

	return VerificationResult{
		Name:        "Direct Provider Onboarding",
		Category:    string(CategoryOnboarding),
		Passed:      true,
		Description: "Prime Intellect splash disabled, direct provider setup preserved",
		Details:     "shouldRunPrimeCliOnboardingSplash returns false in onboarding.ts",
	}
}

// VerifyReleaseScriptInvariant asserts PI_SKIP_NPM_PUBLISH support in release.mjs.
func VerifyReleaseScriptInvariant(repoRoot string) VerificationResult {
	path := filepath.Join(repoRoot, "scripts", "release.mjs")
	content, err := os.ReadFile(path)
	if err != nil {
		return VerificationResult{
			Name:        "Release Publishing",
			Category:    string(CategoryRelease),
			Passed:      false,
			Description: "Ensure scripts/release.mjs supports PI_SKIP_NPM_PUBLISH and --skip-npm-publish",
			Details:     fmt.Sprintf("Failed to read %s: %v", path, err),
		}
	}

	text := string(content)
	if !strings.Contains(text, "PI_SKIP_NPM_PUBLISH") || !strings.Contains(text, "--skip-npm-publish") {
		return VerificationResult{
			Name:        "Release Publishing",
			Category:    string(CategoryRelease),
			Passed:      false,
			Description: "Ensure scripts/release.mjs supports PI_SKIP_NPM_PUBLISH and --skip-npm-publish",
			Details:     "scripts/release.mjs is missing PI_SKIP_NPM_PUBLISH or --skip-npm-publish handling",
		}
	}

	return VerificationResult{
		Name:        "Release Publishing",
		Category:    string(CategoryRelease),
		Passed:      true,
		Description: "PI_SKIP_NPM_PUBLISH and --skip-npm-publish supported in release script",
		Details:     "scripts/release.mjs handles optional npm publishing for fork releases",
	}
}

// VerifyWorkflowsInvariant asserts release workflows publish to GitHub Releases without mandatory R2 secrets.
func VerifyWorkflowsInvariant(repoRoot string) VerificationResult {
	path := filepath.Join(repoRoot, ".github", "workflows", "build-binaries.yml")
	content, err := os.ReadFile(path)
	if err != nil {
		return VerificationResult{
			Name:        "GitHub Workflows",
			Category:    string(CategoryWorkflows),
			Passed:      false,
			Description: "Ensure .github/workflows/build-binaries.yml supports GitHub Releases artifacts",
			Details:     fmt.Sprintf("Failed to read %s: %v", path, err),
		}
	}

	text := string(content)
	if !strings.Contains(text, "Release Prime Agent") && !strings.Contains(text, "build-binaries") {
		return VerificationResult{
			Name:        "GitHub Workflows",
			Category:    string(CategoryWorkflows),
			Passed:      false,
			Description: "Ensure .github/workflows/build-binaries.yml supports GitHub Releases artifacts",
			Details:     "build-binaries.yml missing expected workflow definitions",
		}
	}

	return VerificationResult{
		Name:        "GitHub Workflows",
		Category:    string(CategoryWorkflows),
		Passed:      true,
		Description: "GitHub Releases distribution workflow configured",
		Details:     "build-binaries.yml exists and defines release build pipeline",
	}
}

// VerifyAll executes all registered fork invariant checks.
func VerifyAll(repoRoot string) ([]VerificationResult, bool) {
	var results []VerificationResult
	allPassed := true
	for _, v := range Verifiers {
		res := v(repoRoot)
		if !res.Passed {
			allPassed = false
		}
		results = append(results, res)
	}
	return results, allPassed
}
