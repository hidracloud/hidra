package models

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/hidracloud/hidra/database"
	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"
)

// Represent one agent in db.
type Agent struct {
	gorm.Model
	Name        string
	Description string
	ID          uuid.UUID `gorm:"primaryKey;type:char(36);"`
	Secret      string    `json:"-" gorm:"primaryKey;type:char(36);"`
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
func CreateAgent(secret, name, description string, tags map[string]string) error {
	newAgent := Agent{ID: uuid.NewV4(), Secret: secret, Name: name, Description: description}

	if result := database.ORM.Create(&newAgent); result.Error != nil {
		return result.Error
	}

	for k, v := range tags {
		err := CreateAgentTagByAgentID(newAgent.ID, k, v)
		if err != nil {
			return err
		}
	}

	return nil
}

func CreateAgentTagByAgentID(agentID uuid.UUID, key string, value string) error {
	newAgentTag := AgentTag{ID: uuid.NewV4(), Key: key, Value: value, AgentId: agentID}
	if result := database.ORM.Create(&newAgentTag); result.Error != nil {
		return result.Error
	}
	return nil
}

// Get agent by id
func GetAgent(agentID uuid.UUID) (Agent, error) {
	var agent Agent
	if result := database.ORM.First(&agent, "id = ?", agentID); result.Error != nil {
		return agent, result.Error
	}
	return agent, nil
}

// Search agent by name
func SearchAgentByName(name string) ([]Agent, error) {
	var agents []Agent
	if result := GetAgentQuery().Where("name LIKE ?", "%"+name+"%").Find(&agents); result.Error != nil {
		return agents, result.Error
	}
	return agents, nil
}

// Get all agents
func GetAgents() ([]Agent, error) {
	var agents []Agent
	if result := GetAgentQuery().Find(&agents); result.Error != nil {
		return nil, result.Error
	}
	return agents, nil
}

// Get common get agent query
func GetAgentQuery() *gorm.DB {
	return database.ORM.Order("updated_at desc")
}

// Update agent by agent id
func UpdateAgent(agentID uuid.UUID, name, description string) error {
	var agent Agent
	if result := database.ORM.First(&agent, "id = ?", agentID); result.Error != nil {
		return result.Error
	}
	agent.Name = name
	agent.Description = description
	if result := database.ORM.Save(&agent); result.Error != nil {
		return result.Error
	}
	return nil
}

// Delete agent tags by agent id
func DeleteAgentTags(agentID uuid.UUID) error {
	if result := database.ORM.Where("agent_id = ?", agentID).Delete(&AgentTag{}); result.Error != nil {
		return result.Error
	}
	return nil
}

// Get agent tags
func GetAgentTags(agentID uuid.UUID) ([]AgentTag, error) {
	var agentTags []AgentTag
	if result := database.ORM.Where("agent_id = ?", agentID).Find(&agentTags); result.Error != nil {
		return nil, result.Error
	}
	return agentTags, nil
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

		r.Header.Set("agent_id", newAgent.ID.String())

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
