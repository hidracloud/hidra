package ftp

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/hidracloud/hidra/src/models"
	"github.com/hidracloud/hidra/src/scenarios"
	ftpclient "github.com/jlaffaye/ftp"
)

// Scenario Represent an ssl scenario
type Scenario struct {
	models.Scenario
	ftpConn *ftpclient.ServerConn
}

// RCA generate RCAs for scenario
func (h *Scenario) RCA(result *models.ScenarioResult) error {
	log.Println("ftp RCA")
	return nil
}

func (h *Scenario) connectTo(c map[string]string) ([]models.Metric, error) {
	ftpConn, err := ftpclient.Dial(c["to"], ftpclient.DialWithTimeout(5*time.Second))

	if err != nil {
		return nil, err
	}

	h.ftpConn = ftpConn

	return nil, nil
}

func (h *Scenario) login(c map[string]string) ([]models.Metric, error) {
	err := h.ftpConn.Login(c["user"], c["password"])
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (h *Scenario) write(c map[string]string) ([]models.Metric, error) {
	data := bytes.NewBufferString(c["data"])
	err := h.ftpConn.Stor(c["test-file"], data)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (h *Scenario) read(c map[string]string) ([]models.Metric, error) {
	r, err := h.ftpConn.Retr(c["test-file"])
	if err != nil {
		return nil, err
	}
	defer r.Close()

	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	if string(data) != c["data"] {
		return nil, fmt.Errorf("data is not %s is %s", c["data"], string(data))
	}

	return nil, nil
}

func (h *Scenario) delete(c map[string]string) ([]models.Metric, error) {
	err := h.ftpConn.Delete(c["test-file"])
	if err != nil {
		return nil, err
	}
	return nil, nil
}

// Description return the description of the scenario
func (s *Scenario) Description() string {
	return "Run a ftp scenario"
}

// Close closes the scenario
func (s *Scenario) Close() {
	if s.ftpConn != nil {
		s.ftpConn.Quit()
	}
}

// Init initialize the scenario
func (s *Scenario) Init() {
	s.StartPrimitives()

	s.RegisterStep("connectTo", models.StepDefinition{
		Description: "Connect to a host",
		Params: []models.StepParam{
			{
				Name:        "to",
				Description: "Host to connect to",
				Optional:    false,
			},
		},
		Fn: s.connectTo,
	})

	s.RegisterStep("login", models.StepDefinition{
		Description: "Login to a host",
		Params: []models.StepParam{
			{
				Name:        "user",
				Description: "User to login",
				Optional:    false,
			},
			{
				Name:        "password",
				Description: "Password to login",
				Optional:    false,
			},
		},
		Fn: s.login,
	})

	s.RegisterStep("write", models.StepDefinition{
		Description: "Write a file",
		Params: []models.StepParam{
			{
				Name:        "test-file",
				Description: "File to write",
				Optional:    false,
			},
			{
				Name:        "data",
				Description: "Data to write",
				Optional:    false,
			},
		},
		Fn: s.write,
	})

	s.RegisterStep("read", models.StepDefinition{
		Description: "Read a file",
		Params: []models.StepParam{
			{
				Name:        "test-file",
				Description: "File to read",
				Optional:    false,
			},
			{
				Name:        "data",
				Description: "Data to read",
				Optional:    false,
			},
		},
		Fn: s.read,
	})

	s.RegisterStep("delete", models.StepDefinition{
		Description: "Delete a file",
		Params: []models.StepParam{
			{
				Name:        "test-file",
				Description: "File to delete",
				Optional:    false,
			},
		},
		Fn: s.delete,
	})

}

func init() {
	scenarios.Add("ftp", func() models.IScenario {
		return &Scenario{}
	})
}
