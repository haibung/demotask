package jwt

import (
	"fmt"

	"github.com/golang-jwt/jwt"
)

type ClaimData struct {
	UserID int    `json:"user_id,omitempty"`
	UUID   string `json:"uuid,omitempty"`
}

type InternalClaimData struct {
	UserID int
}

// Claim struct
type Claim struct {
	Data ClaimData `json:"data"`
	jwt.StandardClaims
}

// GenerateToken :
func GenerateToken(c Claim, secretString string) (*string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	tokenString, err := token.SignedString([]byte(secretString))
	if err != nil {
		return nil, err
	}

	return &tokenString, nil
}

// ParseToken :
func ParseToken(tokenString string, secretString string) (*jwt.Token, error) {
	t, err := jwt.Parse(tokenString, func(jt *jwt.Token) (interface{}, error) {
		if _, ok := jt.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", jt.Header["alg"])
		}
		return []byte(secretString), nil
	})
	return t, err
}

// IsValidToken : validate JWT Token
func IsValidToken(tokenString string, secretString string) (bool, error) {
	token, err := ParseToken(tokenString, secretString)
	if err != nil {
		return false, err
	}
	return token.Valid, nil
}

// ParseClaim :
func ParseClaim(tokenString string, secretString string) (*Claim, error) {
	claims := Claim{}
	_, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretString), nil
	})
	if err != nil {
		return nil, err
	}
	return &claims, nil
}
