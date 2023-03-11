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

func (v *jwtTokenVerifier) VerifyToken(c *gin.Context) (*jwt.Token, error) {
	bearerToken := c.Request.Header.Get("Authorization")
	tokenString := ""
	if len(strings.Split(bearerToken, " ")) == 2 {
		tokenString = strings.Split(bearerToken, " ")[1]
	}
	return v.parseToken(tokenString)
}

func (v *jwtTokenVerifier) GetUserId(token *jwt.Token) (uuid.UUID, error) {
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		uid, err := uuid.FromString(fmt.Sprintf("%v", claims["user_id"]))
		if err != nil {
			return uuid.Nil, err
		}
		return uid, nil
	}
	return uuid.Nil, fmt.Errorf("user id could not be extracted from token.")
}

func (v *jwtTokenVerifier) HasRole(token *jwt.Token, role string) (bool, error) {
	roles, err := v.GetRoles(token)
	if err != nil {
		return false, err
	}
	for _, r := range roles {
		if r == role {
			return true, nil
		}
	}
	return false, nil
}

func (v *jwtTokenVerifier) GetRoles(token *jwt.Token) ([]string, error) {
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		roleString := fmt.Sprintf("%v", claims["user_roles"])
		if roleString == "" {
			return nil, fmt.Errorf("could not extract user roles from token")
		}
		return strings.Split(roleString, ","), nil
	}
	return nil, fmt.Errorf("user roles could not be extracted from token.")
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
