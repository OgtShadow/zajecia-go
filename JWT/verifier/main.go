package main

import (
	"fmt"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

const secretKey = "demo-secret-key"

type PlantClaims struct {
	PlantName         string `json:"plant_name"`
	Species           string `json:"species"`
	WateringFrequency string `json:"watering_frequency"`
	Temperature       string `json:"temperature"`
	jwt.RegisteredClaims
}

func main() {
	tokenPath := "token.txt"
	if len(os.Args) > 1 {
		tokenPath = os.Args[1]
	}

	tokenBytes, err := os.ReadFile(tokenPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "token rejected:", err)
		os.Exit(1)
	}

	claims := &PlantClaims{}
	token, err := jwt.ParseWithClaims(string(tokenBytes), claims, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(secretKey), nil
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "token rejected:", err)
		os.Exit(1)
	}
	if !token.Valid {
		fmt.Fprintln(os.Stderr, "token rejected: token is not valid")
		os.Exit(1)
	}

	fmt.Println("token accepted")
	fmt.Printf("plant=%s species=%s watering=%s temperature=%s\n", claims.PlantName, claims.Species, claims.WateringFrequency, claims.Temperature)
}
