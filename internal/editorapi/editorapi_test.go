package editorapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gorilla/mux"
)

func TestNormalizeMockFile(t *testing.T) {
	base := t.TempDir()

	resolved, err := normalizeMockFile(base, "nested/mocks.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if filepath.Ext(resolved) != ".json" {
		t.Fatalf("expected json extension, got %s", resolved)
	}
}

func TestNormalizeMockFileRejectsOutsidePath(t *testing.T) {
	base := t.TempDir()
	_, err := normalizeMockFile(base, "../outside.json")
	if err == nil {
		t.Fatalf("expected error for outside path")
	}
}

func TestNormalizeMockFileRejectsNonJSON(t *testing.T) {
	base := t.TempDir()
	_, err := normalizeMockFile(base, "mocks.txt")
	if err == nil {
		t.Fatalf("expected error for non-json mock file")
	}
}

func TestAttachRoutesLoadAndSave(t *testing.T) {
	tmp := t.TempDir()
	configPath := filepath.Join(tmp, "config.yml")
	mocksDir := filepath.Join(tmp, "mocks")
	if err := os.MkdirAll(mocksDir, 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	mockFile := filepath.Join(mocksDir, "mocks.json")

	if err := os.WriteFile(configPath, []byte("watcher: true\n"), 0o600); err != nil {
		t.Fatalf("config write failed: %v", err)
	}
	if err := os.WriteFile(mockFile, []byte("[]"), 0o600); err != nil {
		t.Fatalf("mock write failed: %v", err)
	}

	r := mux.NewRouter()
	AttachRoutes(r, Options{ConfigPath: configPath, MocksPath: mocksDir})

	loadReq := httptest.NewRequest(http.MethodGet, "/api/editor", nil)
	loadRes := httptest.NewRecorder()
	r.ServeHTTP(loadRes, loadReq)

	if loadRes.Code != http.StatusOK {
		t.Fatalf("expected 200 from load, got %d", loadRes.Code)
	}

	body := map[string]string{
		"config":    "server:\n  port: 8899\n",
		"mocks":     "[{\"rest\":{}}]",
		"mocksFile": "mocks/second.json",
	}
	payload, _ := json.Marshal(body)
	saveReq := httptest.NewRequest(http.MethodPost, "/api/editor/save", bytes.NewReader(payload))
	saveRes := httptest.NewRecorder()
	r.ServeHTTP(saveRes, saveReq)

	if saveRes.Code != http.StatusOK {
		t.Fatalf("expected 200 from save, got %d body=%s", saveRes.Code, saveRes.Body.String())
	}

	updatedConfig, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("config read failed: %v", err)
	}
	if string(updatedConfig) != "server:\n  port: 8899\n" {
		t.Fatalf("unexpected config content: %q", string(updatedConfig))
	}

	secondFile := filepath.Join(mocksDir, "mocks", "second.json")
	updatedMocks, err := os.ReadFile(secondFile)
	if err != nil {
		t.Fatalf("updated mocks read failed: %v", err)
	}
	if string(updatedMocks) != "[{\"rest\":{}}]" {
		t.Fatalf("unexpected mocks content: %q", string(updatedMocks))
	}
}
