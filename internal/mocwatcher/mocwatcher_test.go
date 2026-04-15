package mocwatcher

import (
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

func TestInitializeWatcher(t *testing.T) {
	tmp := t.TempDir()
	mocsDir := filepath.Join(tmp, "mocks")
	if err := os.MkdirAll(mocsDir, 0o755); err != nil {
		t.Fatalf("failed creating mocks dir: %v", err)
	}
	configPath := filepath.Join(tmp, "config.yml")
	if err := os.WriteFile(configPath, []byte("watcher: true\n"), 0o600); err != nil {
		t.Fatalf("failed writing config file: %v", err)
	}

	w, err := InitializeWatcher(mocsDir, configPath)
	if err != nil {
		t.Fatalf("expected watcher initialization to succeed: %v", err)
	}
	defer w.Close()
}

func TestInitializeWatcherInvalidMocksPath(t *testing.T) {
	tmp := t.TempDir()
	configPath := filepath.Join(tmp, "config.yml")
	if err := os.WriteFile(configPath, []byte("watcher: true\n"), 0o600); err != nil {
		t.Fatalf("failed writing config file: %v", err)
	}

	_, err := InitializeWatcher(filepath.Join(tmp, "does-not-exist"), configPath)
	if err == nil {
		t.Fatalf("expected error for invalid mocks path")
	}
}

func TestAttachWatcherTriggersCallbackOnWrite(t *testing.T) {
	tmp := t.TempDir()
	mocsDir := filepath.Join(tmp, "mocks")
	if err := os.MkdirAll(mocsDir, 0o755); err != nil {
		t.Fatalf("failed creating mocks dir: %v", err)
	}
	configPath := filepath.Join(tmp, "config.yml")
	if err := os.WriteFile(configPath, []byte("watcher: true\n"), 0o600); err != nil {
		t.Fatalf("failed writing config file: %v", err)
	}

	w, err := InitializeWatcher(mocsDir, configPath)
	if err != nil {
		t.Fatalf("watcher init failed: %v", err)
	}
	defer w.Close()

	var called int32
	AttachWatcher(w, func() {
		atomic.StoreInt32(&called, 1)
	})

	mockFile := filepath.Join(mocsDir, "mocks.json")
	if err := os.WriteFile(mockFile, []byte("[]"), 0o600); err != nil {
		t.Fatalf("failed writing mock file: %v", err)
	}

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if atomic.LoadInt32(&called) == 1 {
			return
		}
		time.Sleep(25 * time.Millisecond)
	}

	t.Fatalf("expected callback to be triggered by watcher event")
}
