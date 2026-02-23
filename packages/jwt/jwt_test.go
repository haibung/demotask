package jwt_test

import (
	"testing"
	"time"

	"github.com/demotask/backend/packages/jwt"
	golangjwt "github.com/golang-jwt/jwt"
)

func BenchmarkGenerateToken(b *testing.B) {
	secret := "super-secret"
	claim := jwt.Claim{
		Data: jwt.ClaimData{
			UserID: 1234,
			UUID:   "test-uuid",
		},
		StandardClaims: golangjwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour).Unix(),
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := jwt.GenerateToken(claim, secret)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkValidateToken(b *testing.B) {
	secret := "super-secret"
	claim := jwt.Claim{
		Data: jwt.ClaimData{UserID: 1234, UUID: "test-uuid"},
		StandardClaims: golangjwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour).Unix(),
		},
	}
	token, _ := jwt.GenerateToken(claim, secret)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		valid, err := jwt.IsValidToken(*token, secret)
		if err != nil || !valid {
			b.Fatal("invalid token")
		}
	}
}
