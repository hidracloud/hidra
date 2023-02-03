package report

import (
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

// SaveFile saves the report to a file.
func (r *Report) SaveFile() error {
	if r == nil {
		return nil
	}

	rDump := r.Dump()

	if err := os.MkdirAll(BasePath, 0755); err != nil {
		return err
	}

	filePath := filepath.Join(BasePath, r.Name+".json")

	log.Debugf("Saving report to file %s", filePath)

	for dest, content := range r.Attachments {
		attachmentPath := filepath.Join(BasePath, r.Name+".more", dest)

		if err := os.MkdirAll(filepath.Dir(attachmentPath), 0755); err != nil {
			log.Errorf("Error creating attachment directory %s: %s", filepath.Dir(attachmentPath), err)
			continue
		}

		if err := os.WriteFile(attachmentPath, content, 0644); err != nil {
			log.Errorf("Error writing attachment %s: %s", attachmentPath, err)
			continue
		}
	}

	return os.WriteFile(filePath, []byte(rDump), 0644)
}
