package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	validTokenPath := filepath.Join("JWT", "token.txt")
	invalidTokenPath := filepath.Join("JWT", "invalid-token.txt")

	if err := os.WriteFile(invalidTokenPath, []byte("not-a-valid-jwt-token\n"), 0o600); err != nil {
		fmt.Fprintln(os.Stderr, "failed to create invalid token file:", err)
		os.Exit(1)
	}

	if err := checkVerifier(validTokenPath, true); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err := checkVerifier(invalidTokenPath, false); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println("tester: all checks passed")
}

func checkVerifier(tokenPath string, shouldPass bool) error {
	cmd := exec.Command("go", "run", "./JWT/verifier", tokenPath)
	output, err := cmd.CombinedOutput()
	outputText := string(output)

	if shouldPass {
		if err != nil {
			return fmt.Errorf("expected verifier to accept %s, got error: %v\noutput:\n%s", tokenPath, err, outputText)
		}
		if !strings.Contains(outputText, "token accepted") {
			return fmt.Errorf("expected acceptance output for %s, got:\n%s", tokenPath, outputText)
		}
		return nil
	}

	if err == nil {
		return fmt.Errorf("expected verifier to reject %s, but it succeeded\noutput:\n%s", tokenPath, outputText)
	}
	if !strings.Contains(outputText, "token rejected") {
		return fmt.Errorf("expected rejection output for %s, got:\n%s", tokenPath, outputText)
	}
	return nil
}
