// Package agent provide support for hidra-agents
package agent

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/hidracloud/hidra/models"
	"github.com/hidracloud/hidra/scenarios"
)

// Agent Represent one agent configuration
type Agent struct {
	APIURL  string
	Secret  string
	DataDir string
}

var sampleScrapeInterval map[string]time.Time

// DoAPICall Make a request to hidra API
func (a *Agent) DoAPICall(endpoint, method string, body io.Reader) (*http.Response, error) {
	client := &http.Client{}
	req, err := http.NewRequest(method, a.APIURL+endpoint, body)

	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+a.Secret)

	resp, err := client.Do(req)

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("error code not 200, %d", resp.StatusCode)
	}

	return resp, err
}

// ListSamples List all samples related to current agent
func (a *Agent) ListSamples() []models.Sample {
	samples := make([]models.Sample, 0)

	res, err := a.DoAPICall("/agent_list_samples", "GET", nil)
	if err != nil {
		log.Fatal(err)
	}

	defer res.Body.Close()

	json.NewDecoder(res.Body).Decode(&samples)

	return samples
}

// GetSample Get one sample
func (a *Agent) GetSample(id string) []byte {
	res, err := a.DoAPICall("/agent_get_sample/"+id, "GET", nil)
	if err != nil {
		log.Fatal(err)
	}

	defer res.Body.Close()

	b, err := ioutil.ReadAll(res.Body)

	if err != nil {
		log.Fatal(err)
	}

	return b
}

// PushMetrics Push metrics to API
func (a *Agent) PushMetrics(sampleID string, metrics *models.ScenarioResult) error {
	if metrics.Error != nil {
		metrics.ErrorString = metrics.Error.Error()
	}

	payloadBuf := new(bytes.Buffer)
	fmt.Println(metrics)
	err := json.NewEncoder(payloadBuf).Encode(metrics)

	if err != nil {
		return err
	}

	_, err = a.DoAPICall("/agent_push_metrics/"+sampleID, "POST", payloadBuf)
	return err
}

// RemoveDeprecatedSamples Clean up old samples
func (a *Agent) RemoveDeprecatedSamples(samples []models.Sample, files []fs.FileInfo) {
	for _, file := range files {
		found := false
		for _, sample := range samples {
			if file.Name() == sample.ID.String() {
				found = true
			}
		}

		if !found {
			log.Println("Agent has found a old sample, deleting...")
			os.Remove(a.DataDir + "/" + file.Name())
		}
	}
}

// Calculate checksum from a local file
func calculateLocalChecksum(filePath string) (string, error) {
	var returnMD5String string
	file, err := os.Open(filePath)
	if err != nil {
		return returnMD5String, err
	}
	defer file.Close()
	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return returnMD5String, err
	}
	hashInBytes := hash.Sum(nil)[:16]
	returnMD5String = hex.EncodeToString(hashInBytes)
	return returnMD5String, nil

}

// UpdateSamplesIfNeeded Check if current sample should be updated
func (a *Agent) UpdateSamplesIfNeeded(samples []models.Sample, files []fs.FileInfo) {
	for _, sample := range samples {
		needupdate := true

		sampleLocalFile := a.DataDir + "/" + sample.ID.String()
		if _, err := os.Stat(sampleLocalFile); err == nil {
			localChecksum, err := calculateLocalChecksum(sampleLocalFile)
			if err != nil {
				log.Fatal(err)
			}
			needupdate = localChecksum != sample.Checksum
		}

		if needupdate {
			log.Println("Updating " + sample.ID.String())
			sampleData := a.GetSample(sample.ID.String())
			os.Remove(sampleLocalFile)
			os.WriteFile(sampleLocalFile, sampleData, 0660)
		}
	}
}

// UpdateLocalResources Try to update local resources
func (a *Agent) UpdateLocalResources() {
	if time.Since(sampleScrapeInterval["foobar"]) < time.Minute*5 {
		return
	}

	samples := a.ListSamples()

	files, err := ioutil.ReadDir(a.DataDir)

	if err != nil {
		log.Fatal(err)
	}

	// Remove deprecated samples
	a.RemoveDeprecatedSamples(samples, files)

	// Update samples that need it
	a.UpdateSamplesIfNeeded(samples, files)

	// Update local resources timer
	sampleScrapeInterval["foobar"] = time.Now()
}

// RunAllSamples Run all samples in current agent
func (a *Agent) RunAllSamples() {
	files, err := ioutil.ReadDir(a.DataDir)

	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		data, err := ioutil.ReadFile(a.DataDir + "/" + file.Name())

		if err != nil {
			log.Fatal(err)
		}

		slist, err := models.ReadScenariosYAML(data)

		if err != nil {
			log.Fatal(err)
		}

		if _, ok := sampleScrapeInterval[file.Name()]; !ok {
			sampleScrapeInterval[file.Name()] = time.Unix(0, 0)
		}

		if time.Since(sampleScrapeInterval[file.Name()]) < slist.ScrapeInterval {
			continue
		}

		m := scenarios.RunScenario(slist.Scenario, slist.Name, slist.Description)
		err = a.PushMetrics(file.Name(), m)

		if err != nil {
			log.Fatal(err)
		}

		sampleScrapeInterval[file.Name()] = time.Now()
	}
}

// StartAgent Initialize an agent
func StartAgent(apiURL, secretToken, datadir string) {
	agent := Agent{APIURL: apiURL, Secret: secretToken, DataDir: datadir}

	sampleScrapeInterval = make(map[string]time.Time)
	sampleScrapeInterval["foobar"] = time.Unix(0, 0)

	for {
		agent.UpdateLocalResources()
		agent.RunAllSamples()
		time.Sleep(time.Second * 5)
	}
}
