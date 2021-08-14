package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/hidracloud/hidra/models"
)

// List sample response struct
type ListSampleResponse struct {
	Id          string
	Name        string
	Description string
	UpdatedAt   time.Time
	Kind        string
}

// Get a list of samples by id and checksum
func (a *API) ListSamples(w http.ResponseWriter, r *http.Request) {
	var err error
	var samples []models.Sample

	search := r.URL.Query().Get("search")

	if search == "" {
		samples, err = models.GetSamples()
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	} else {
		samples, err = models.SearchSamples(search)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	sampleResponse := make([]ListSampleResponse, len(samples))

	for sample := range samples {
		sampleResponse[sample] = ListSampleResponse{
			Id:          samples[sample].ID.String(),
			Name:        samples[sample].Name,
			UpdatedAt:   samples[sample].UpdatedAt,
			Description: samples[sample].Description,
			Kind:        samples[sample].Kind,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sampleResponse)
}
