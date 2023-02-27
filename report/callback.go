package report

import (
	"errors"
	"net/http"
	"strings"
)

// SendCallback sends a callback to the client.
func (r *Report) SendCallback() error {
	if !IsEnabled {
		return nil
	}

	// Send JSON as POST request to callback URL
	dataDump := r.Dump()

	// create a reader from origin file
	reader := strings.NewReader(dataDump)

	if reader == nil {
		return errors.New("reader is nil")
	}

	// Send POST request to callback URL
	resp, err := http.Post(CallbackConf.URL, "application/json", reader)

	if err != nil {
		return err
	}

	// Check response status code
	if resp.StatusCode != http.StatusOK {
		return errors.New("response status code is not 200")
	}

	// Close response body
	err = resp.Body.Close()

	if err != nil {
		return err
	}

	return nil
}
