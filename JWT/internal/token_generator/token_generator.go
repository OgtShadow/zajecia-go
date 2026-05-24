package tokengenerator

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const SecretKey = "demo-secret-key"

type PlantClaims struct {
	PlantName         string `json:"plant_name"`
	Species           string `json:"species"`
	WateringFrequency string `json:"watering_frequency"`
	Temperature       string `json:"temperature"`
	jwt.RegisteredClaims
}

func DemoClaims() PlantClaims {
	return PlantClaims{
		PlantName:         "Monstera",
		Species:           "Monstera deliciosa",
		WateringFrequency: "Once a week",
		Temperature:       "18-27 C",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "JWTManager",
			Subject:   "plant-demo",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}
}

func BuildSignedToken() (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, DemoClaims())
	return token.SignedString([]byte(SecretKey))
}

func WriteTokenFile(path string, token string) error {
	return os.WriteFile(path, []byte(token+"\n"), 0o600)
}