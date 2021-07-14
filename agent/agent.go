package agent

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"io"
	"io/fs"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/JoseCarlosGarcia95/hidra/models"
	"github.com/JoseCarlosGarcia95/hidra/scenarios"
)

type Agent struct {
	ApiURL  string
	Secret  string
	DataDir string
}

var sampleScrapeInterval map[string]time.Time

func (a *Agent) DoApiCall(endpoint, method string, body io.Reader) (*http.Response, error) {
	client := &http.Client{}
	req, err := http.NewRequest(method, a.ApiURL+endpoint, body)

	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+a.Secret)

	return client.Do(req)
}

func (a *Agent) ListSamples() []models.Sample {
	samples := make([]models.Sample, 0)

	res, err := a.DoApiCall("/agent_list_samples", "GET", nil)
	if err != nil {
		log.Fatal(err)
	}

	defer res.Body.Close()

	json.NewDecoder(res.Body).Decode(&samples)

	return samples
}

func (a *Agent) GetSample(id string) []byte {
	res, err := a.DoApiCall("/agent_get_sample/"+id, "GET", nil)
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

func (a *Agent) PushMetrics(sampleId string, metrics *models.ScenarioMetric) error {
	if metrics.Error != nil {
		metrics.ErrorString = metrics.Error.Error()
	}

	payloadBuf := new(bytes.Buffer)
	err := json.NewEncoder(payloadBuf).Encode(metrics)

	if err != nil {
		return err
	}

	_, err = a.DoApiCall("/agent_push_metrics/"+sampleId, "POST", payloadBuf)
	return err
}

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

		for _, s := range slist.Scenarios {
			m := scenarios.RunScenario(s)

			err := a.PushMetrics(file.Name(), m)
			log.Println(err)
		}

		sampleScrapeInterval[file.Name()] = time.Now()
	}
}

func StartAgent(apiUrl, secretToken, datadir string) {
	agent := Agent{ApiURL: apiUrl, Secret: secretToken, DataDir: datadir}

	sampleScrapeInterval = make(map[string]time.Time)
	sampleScrapeInterval["foobar"] = time.Unix(0, 0)

	for {
		agent.UpdateLocalResources()
		agent.RunAllSamples()
		time.Sleep(time.Second * 5)
	}
}
