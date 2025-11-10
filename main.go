package main

import (
	"fmt"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Simulasi secret dari config
const SecretKey = "secret"

// Fungsi GenerateJWT (bisa pakai versi kamu di middleware juga)
func GenerateJWT(secretKey string, userID string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(), // kadaluarsa 24 jam
		"iat":     time.Now().Unix(),                     // waktu dibuat
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}

func main() {
	// Jalankan fungsi generate JWT
	token, err := GenerateJWT(SecretKey, "user-123")
	if err != nil {
		log.Fatalf("gagal membuat token: %v", err)
	}

	fmt.Println("Generated JWT:")
	fmt.Println(token)
}
