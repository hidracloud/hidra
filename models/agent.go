package models

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/JoseCarlosGarcia95/hidra/database"
	"github.com/golang-jwt/jwt"
	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"
)

// Represent one agent in db.
type Agent struct {
	gorm.Model
	ID     uuid.UUID `gorm:"primaryKey;type:char(36);"`
	Secret string    `json:"-" gorm:"primaryKey;type:char(36);"`
}

// Represent agent tags
type AgentTag struct {
	gorm.Model
	ID      uuid.UUID `gorm:"primaryKey;type:char(36);"`
	AgentId uuid.UUID
	Agent   Agent `gorm:"foreignKey:AgentId" json:"-"`
	Key     string
	Value   string
}

// Verify that agent token is correct
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

// Create a new agent.
func CreateAgent(secret string, tags map[string]string) error {
	newAgent := Agent{ID: uuid.NewV4(), Secret: secret}

	if result := database.ORM.Create(&newAgent); result.Error != nil {
		return result.Error
	}

	for k, v := range tags {
		newAgentTag := AgentTag{ID: uuid.NewV4(), Key: k, Value: v, Agent: newAgent}
		if result := database.ORM.Create(&newAgentTag); result.Error != nil {
			return result.Error
		}
	}

	return nil
}

// Check if agent secret is correct
func AuthSecretAgentMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		tokenString := r.Header.Get("Authorization")

		if len(tokenString) == 0 {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Missing Authorization Header"))
			return
		}
		secret := strings.Replace(tokenString, "Bearer ", "", 1)

		newAgent := Agent{}

		database.ORM.First(&newAgent, "secret = ?", secret)

		if newAgent.ID == uuid.Nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Generate a new temporal token
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
