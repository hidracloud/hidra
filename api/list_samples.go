package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/hidracloud/hidra/models"
)

// List sample response struct
type listSampleResponse struct {
	ID          string
	Name        string
	Description string
	UpdatedAt   time.Time
	Kind        string
	Error       string
}

// List sample response with pages
type listSampleResponseWithPages struct {
	Total    int64
	Page     int
	PageSize int
	Items    []listSampleResponse
}

// ListSamples is a method to list samples
func (a *API) ListSamples(w http.ResponseWriter, r *http.Request) {
	var err error
	var samples []models.Sample

	search := r.URL.Query().Get("search")

	page := 0
	pageSize := 20

	if r.URL.Query().Get("page") != "" {
		page, err = strconv.Atoi(r.URL.Query().Get("page"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	if search == "" {
		samples, err = models.GetSamplesWithPagination(page, pageSize)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	} else {
		samples, err = models.SearchSamplesWithPagination(search, page, pageSize)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	sampleResponse := make([]listSampleResponse, len(samples))

	for sample := range samples {
		sampleResult, err := models.GetLastSampleResultBySampleID(samples[sample].ID.String())

		sampleResponse[sample] = listSampleResponse{
			ID:          samples[sample].ID.String(),
			Name:        samples[sample].Name,
			UpdatedAt:   samples[sample].UpdatedAt,
			Description: samples[sample].Description,
			Kind:        samples[sample].Kind,
		}

		if err == nil {
			sampleResponse[sample].Error = sampleResult.Error
		}

	}

	response := listSampleResponseWithPages{
		Total:    models.GetTotalSamples(),
		Page:     page,
		PageSize: pageSize,
		Items:    sampleResponse,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
