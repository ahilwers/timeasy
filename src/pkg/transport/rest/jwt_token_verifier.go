package rest

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"timeasy-server/pkg/domain/model"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
)

var secret string = "MyVerySecretString"
var token_hour_livespan string = "1"

type jwtTokenVerifier struct{}

func GenerateToken(userId uuid.UUID, roles model.RoleList) (string, error) {

	token_lifespan, err := strconv.Atoi(token_hour_livespan)

	if err != nil {
		return "", err
	}

	claims := jwt.MapClaims{}
	claims["authorized"] = true
	claims["user_id"] = userId
	claims["user_roles"] = strings.Join(roles, ",")
	claims["exp"] = time.Now().Add(time.Hour * time.Duration(token_lifespan)).Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(secret))
}

func NewJwtTokenVerifier() TokenVerifier {
	return &jwtTokenVerifier{}
}

func (v *jwtTokenVerifier) VerifyToken(c *gin.Context) (AuthToken, error) {
	bearerToken := c.Request.Header.Get("Authorization")
	tokenString := ""
	if len(strings.Split(bearerToken, " ")) == 2 {
		tokenString = strings.Split(bearerToken, " ")[1]
	}
	token, err := v.parseToken(tokenString)
	if err != nil {
		return nil, err
	}
	return NewJwtToken(*token), nil
}

func (v *jwtTokenVerifier) parseToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
	return token, err
}
