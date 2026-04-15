package server

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/guvense/lara/internal/model"
)

func TestFindMocksReturnsOnlyJSONFiles(t *testing.T) {
	tmp := t.TempDir()
	jsonFile := filepath.Join(tmp, "mocks.json")
	txtFile := filepath.Join(tmp, "notes.txt")

	if err := os.WriteFile(jsonFile, []byte("[]"), 0o600); err != nil {
		t.Fatalf("failed writing json file: %v", err)
	}
	if err := os.WriteFile(txtFile, []byte("ignored"), 0o600); err != nil {
		t.Fatalf("failed writing txt file: %v", err)
	}

	ch := make(chan string, 10)
	if err := FindMocks(tmp, ch); err != nil {
		t.Fatalf("FindMocks returned error: %v", err)
	}
	close(ch)

	var got []string
	for path := range ch {
		got = append(got, filepath.Base(path))
	}

	if len(got) != 1 || got[0] != "mocks.json" {
		t.Fatalf("expected only mocks.json, got %v", got)
	}
}

func TestUnmarshalMocs(t *testing.T) {
	tmp := t.TempDir()
	file := filepath.Join(tmp, "mocks.json")
	content := []byte(`[
		{"rest":{"request":{"endpoint":"/x","method":"GET"},"response":{"status":200,"body":{"ok":true}}}}
	]`)
	if err := os.WriteFile(file, content, 0o600); err != nil {
		t.Fatalf("failed writing mocks file: %v", err)
	}

	var mocs []model.Moc

	if err := UnmarshalMocs(file, &mocs); err != nil {
		t.Fatalf("UnmarshalMocs returned error: %v", err)
	}

	if len(mocs) != 1 {
		t.Fatalf("expected one mock, got %d", len(mocs))
	}
	if mocs[0].Rest.Request.Endpoint != "/x" || mocs[0].Rest.Request.Method != "GET" {
		t.Fatalf("unexpected request parsed: %+v", mocs[0].Rest.Request)
	}
	if mocs[0].Rest.Response.Status != 200 {
		t.Fatalf("unexpected response status: %d", mocs[0].Rest.Response.Status)
	}
}

func TestUnmarshalMocsInvalidJSON(t *testing.T) {
	tmp := t.TempDir()
	file := filepath.Join(tmp, "bad.json")
	if err := os.WriteFile(file, []byte("{not-json}"), 0o600); err != nil {
		t.Fatalf("failed writing invalid json: %v", err)
	}

	var mocs []interface{}
	if err := UnmarshalMocs(file, &mocs); err == nil {
		t.Fatalf("expected error for invalid json")
	}
}
