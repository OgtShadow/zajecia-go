package main

import (
	"fmt"
	"os"
	"path/filepath"

	"zajecia-go/JWT/internal/token_generator"
)

func main() {
	signedToken, err := tokengenerator.BuildSignedToken()
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to sign token:", err)
		os.Exit(1)
	}

	tokenPath := filepath.Join("token.txt")
	if err := tokengenerator.WriteTokenFile(tokenPath, signedToken); err != nil {
		fmt.Fprintln(os.Stderr, "failed to write token:", err)
		os.Exit(1)
	}

	fmt.Println("JWT token written to", tokenPath)
	fmt.Println(signedToken)
}
