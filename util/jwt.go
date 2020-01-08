package util

import (
	"errors"
	jwtLib "github.com/dgrijalva/jwt-go"
)

type JWTBuilder struct {
	token       *jwtLib.Token
	algStrategy jwtLib.SigningMethod
	claims      jwtLib.Claims
	secretKey   string
}

func NewBuilder(signStrategy jwtLib.SigningMethod, secretKey string) *JWTBuilder {
	return &JWTBuilder{
		algStrategy: signStrategy,
		secretKey:   secretKey,
	}
}

func (b *JWTBuilder) SetClaims(claims jwtLib.Claims) *JWTBuilder {
	b.claims = claims
	return b
}

func (b *JWTBuilder) GenerateToken() (*string, error) {
	if b.claims == nil {
		return nil, errors.New("no claims provided")
	}
	token := jwtLib.NewWithClaims(
		b.algStrategy,
		b.claims,
	)
	secretKey := []byte(b.secretKey)
	signedToken, err := token.SignedString(secretKey)
	if err != nil {
		return nil, err
	}

	return &signedToken, nil
}

type JWTValidator struct {
	method    jwtLib.SigningMethod
	secretKey string
}

func NewValidator(method jwtLib.SigningMethod, secretKey string) *JWTValidator {
	return &JWTValidator{
		method:    method,
		secretKey: secretKey,
	}
}

func (v *JWTValidator) ValidateToken(token string) (jwtLib.MapClaims, error) {
	jwtToken, err := jwtLib.Parse(token, func(token *jwtLib.Token) (interface{}, error) {
		if token.Method != v.method {
			return nil, errors.New("signing method invalid")
		}
		return []byte(v.secretKey), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := jwtToken.Claims.(jwtLib.MapClaims)
	if !ok {
		return nil, errors.New("failed to retrieve claims data from token")
	}
	return claims, nil
}
