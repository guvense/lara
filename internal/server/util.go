package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/guvense/lara/internal/model"
)

func FindMocks(mocsDirectory string, filePathsCh chan string) error {

	err := filepath.Walk(mocsDirectory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("%w: Error occurred while finding mocs", err)
		}
		if !info.IsDir() {
			if filepath.Ext(path) == ".json" {
				filePathsCh <- path
			}
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("%w: Error occurred while finding mocs", err)
	}
	return nil
}

func UnmarshalMocs(filePath string, mocs *[]model.Moc) error {
	mockDefinitionFile, _ := os.Open(filePath)

	defer mockDefinitionFile.Close()

	bytes, err := ioutil.ReadAll(mockDefinitionFile)

	if err != nil {
		return fmt.Errorf("%w: error while unmarshalling mocs %s", err, filePath)

	}

	parseError := json.Unmarshal(bytes, mocs)

	if parseError != nil {
		return fmt.Errorf("%w: error while unmarshalling mocs %s", parseError, filePath)
	}

	return nil
}
