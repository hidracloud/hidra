package models

import (
	"net/http"
	"strings"

	"github.com/hidracloud/hidra/database"
	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"
)

// Agent model
type Agent struct {
	gorm.Model
	Name        string
	Description string
	ID          uuid.UUID `gorm:"primaryKey;type:char(36);"`
	Secret      string    `json:"-" gorm:"primaryKey;type:char(36);"`
}

// AgentTag model
type AgentTag struct {
	gorm.Model
	ID      uuid.UUID `gorm:"primaryKey;type:char(36);"`
	AgentID uuid.UUID
	Agent   Agent `gorm:"foreignKey:AgentID" json:"-"`
	Key     string
	Value   string
}

// CreateAgent create a new agent
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

// CreateAgentTagByAgentID create a new agent tag
func CreateAgentTagByAgentID(agentID uuid.UUID, key string, value string) error {
	newAgentTag := AgentTag{ID: uuid.NewV4(), Key: key, Value: value, AgentID: agentID}
	if result := database.ORM.Create(&newAgentTag); result.Error != nil {
		return result.Error
	}
	return nil
}

// GetAgent get agent by agent id
func GetAgent(agentID uuid.UUID) (Agent, error) {
	var agent Agent
	if result := database.ORM.First(&agent, "id = ?", agentID); result.Error != nil {
		return agent, result.Error
	}
	return agent, nil
}

// SearchAgentByName search agent by name
func SearchAgentByName(name string) ([]Agent, error) {
	var agents []Agent
	if result := GetAgentQuery().Where("name LIKE ?", "%"+name+"%").Find(&agents); result.Error != nil {
		return agents, result.Error
	}
	return agents, nil
}

// GetAgents get all agents
func GetAgents() ([]Agent, error) {
	var agents []Agent
	if result := GetAgentQuery().Find(&agents); result.Error != nil {
		return nil, result.Error
	}
	return agents, nil
}

// GetAgentQuery get agent query
func GetAgentQuery() *gorm.DB {
	return database.ORM.Order("updated_at desc")
}

// UpdateAgent update agent
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

// DeleteAgentTags delete agent tags
func DeleteAgentTags(agentID uuid.UUID) error {
	if result := database.ORM.Where("agent_id = ?", agentID).Delete(&AgentTag{}); result.Error != nil {
		return result.Error
	}
	return nil
}

// GetAgentTags get agent tags
func GetAgentTags(agentID uuid.UUID) ([]AgentTag, error) {
	var agentTags []AgentTag
	if result := database.ORM.Where("agent_id = ?", agentID).Find(&agentTags); result.Error != nil {
		return nil, result.Error
	}
	return agentTags, nil
}

// AuthSecretAgentMiddleware auth secret agent middleware
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
