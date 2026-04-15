package editorapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gorilla/mux"
)

type Options struct {
	ConfigPath string
	MocksPath  string
}

type editorResponse struct {
	Config    string `json:"config"`
	Mocks     string `json:"mocks"`
	MocksFile string `json:"mocksFile"`
}

type saveRequest struct {
	Config    string `json:"config"`
	Mocks     string `json:"mocks"`
	MocksFile string `json:"mocksFile"`
}

func AttachRoutes(router *mux.Router, options Options) {
	routes := []string{"/api/editor", "/_lara/api/editor"}
	for _, route := range routes {
		router.HandleFunc(route, loadHandler(options)).Methods(http.MethodGet)
	}

	saveRoutes := []string{"/api/editor/save", "/_lara/api/editor/save"}
	for _, route := range saveRoutes {
		router.HandleFunc(route, saveHandler(options)).Methods(http.MethodPost)
	}
}

func loadHandler(options Options) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		config, _ := readFileIfExists(options.ConfigPath)
		mockFile := resolveMockFile(options.MocksPath)
		mocks, _ := readFileIfExists(mockFile)

		writeJSON(w, http.StatusOK, editorResponse{
			Config:    config,
			Mocks:     mocks,
			MocksFile: mockFile,
		})
	}
}

func saveHandler(options Options) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req saveRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON payload"})
			return
		}

		mockFile := req.MocksFile
		if strings.TrimSpace(mockFile) == "" {
			mockFile = resolveMockFile(options.MocksPath)
		}

		safeMockFile, err := normalizeMockFile(options.MocksPath, mockFile)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}

		if err := ensureParent(options.ConfigPath); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		if err := ensureParent(safeMockFile); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}

		if err := os.WriteFile(options.ConfigPath, []byte(req.Config), 0o600); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("config write failed: %v", err)})
			return
		}

		if err := os.WriteFile(safeMockFile, []byte(req.Mocks), 0o600); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("mocks write failed: %v", err)})
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{"status": "saved", "mocksFile": safeMockFile})
	}
}

func resolveMockFile(mocksPath string) string {
	entries, err := os.ReadDir(mocksPath)
	if err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			if strings.HasSuffix(strings.ToLower(entry.Name()), ".json") {
				return filepath.Join(mocksPath, entry.Name())
			}
		}
	}
	return filepath.Join(mocksPath, "mocks.json")
}

func normalizeMockFile(mocksPath, mockFile string) (string, error) {
	absBase, err := filepath.Abs(mocksPath)
	if err != nil {
		return "", err
	}

	resolved := mockFile
	if !filepath.IsAbs(resolved) {
		resolved = filepath.Join(absBase, resolved)
	}

	absResolved, err := filepath.Abs(resolved)
	if err != nil {
		return "", err
	}

	baseWithSep := absBase + string(os.PathSeparator)
	if absResolved != absBase && !strings.HasPrefix(absResolved, baseWithSep) {
		return "", errors.New("mocksFile must stay under mocks directory")
	}

	if filepath.Ext(absResolved) != ".json" {
		return "", errors.New("mocksFile must be a .json file")
	}

	return absResolved, nil
}

func readFileIfExists(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return string(content), nil
}

func ensureParent(path string) error {
	return os.MkdirAll(filepath.Dir(path), 0o755)
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
