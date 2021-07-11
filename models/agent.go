package models

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"
)

type Agent struct {
	gorm.Model
	ID     uuid.UUID `gorm:"primaryKey;type:char(36);"`
	Host   string    `gorm:"primaryKey;type:char(36);"`
	Port   uint
	Secret string `gorm:"primaryKey;type:char(36);"`
}

type AgentTag struct {
	gorm.Model
	ID      uuid.UUID `gorm:"primaryKey;type:char(36);"`
	AgentId uuid.UUID
	Agent   Agent `gorm:"foreignKey:AgentId" json:"-"`
	Key     string
	Value   string
}

func VerifyRegisterAgentToken(tokenString string) (jwt.Claims, error) {
	signingKey := []byte(os.Getenv("AGENT_JWT_SECRET_TOKEN"))

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return signingKey, nil
	})
	if err != nil {
		return nil, err
	}
	return token.Claims, err
}

func AuthRegisterAgentMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		tokenString := r.Header.Get("Authorization")

		if len(tokenString) == 0 {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Missing Authorization Header"))
			return
		}
		tokenString = strings.Replace(tokenString, "Bearer ", "", 1)
		claims, err := VerifyRegisterAgentToken(tokenString)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Error verifying JWT token: " + err.Error()))
			return
		}

		exp := claims.(jwt.MapClaims)["exp"].(float64)

		if exp < float64(time.Now().Unix()) {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Error verifying JWT token: " + err.Error()))
			return
		}

		next.ServeHTTP(w, r)
	})
}

func CreateRegisterAgentToken() (string, error) {
	var err error

	atClaims := jwt.MapClaims{}
	atClaims["authorized"] = true
	atClaims["exp"] = time.Now().Add(time.Minute * 10).Unix()
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	token, err := at.SignedString([]byte(os.Getenv("AGENT_JWT_SECRET_TOKEN")))
	if err != nil {
		return "", err
	}

	return token, nil
}
