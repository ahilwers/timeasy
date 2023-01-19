package http

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

func ExtractToken(c *gin.Context) string {
	token := c.Query("token")
	if token != "" {
		return token
	}
	bearerToken := c.Request.Header.Get("Authorization")
	if len(strings.Split(bearerToken, " ")) == 2 {
		return strings.Split(bearerToken, " ")[1]
	}
	return ""
}

func TokenValid(tokenString string) error {
	_, err := parseToken(tokenString)
	if err != nil {
		return err
	}
	return nil
}

func ExtractTokenUserId(tokenString string) (uuid.UUID, error) {
	token, err := parseToken(tokenString)
	if err != nil {
		return uuid.Nil, err
	}
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

func TokenHasRole(tokenString string, role string) (bool, error) {
	roles, err := ExtractTokenRoles(tokenString)
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

func ExtractTokenRoles(tokenString string) ([]string, error) {
	token, err := parseToken(tokenString)
	if err != nil {
		return nil, err
	}
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

func parseToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
	return token, err
}
