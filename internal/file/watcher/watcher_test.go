package watcher

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/open-policy-agent/opa/util/test"
)

func TestStartWatcher(t *testing.T) {
	tests := []struct {
		note            string
		fs              map[string]string
		updateFunc      func(string)
		expectedRemoval bool
	}{
		{
			note: "Add file",
			fs:   map[string]string{},
			updateFunc: func(rootDir string) {
				os.Create(filepath.Join(rootDir, "test.txt"))
			},
			expectedRemoval: false,
		},
		{
			note: "Remove file",
			fs: map[string]string{
				"test.txt": ``,
			},
			updateFunc: func(rootDir string) {
				os.Remove(filepath.Join(rootDir, "test.txt"))
			},
			expectedRemoval: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.note, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			test.WithTempFS(tc.fs, func(rootDir string) {
				triggered := make(chan interface{})
				processWatcherUpdate := func(ctx context.Context, paths []string, removed string) {
					defer close(triggered)
					fileRemoved := (removed != "")
					if fileRemoved != tc.expectedRemoval {
						t.Fatalf("Expected file removal to be %t, but instead it was %t", tc.expectedRemoval, fileRemoved)
					}
				}

				paths := []string{rootDir}
				close, err := StartWatcher(ctx, paths, processWatcherUpdate)
				if err != nil {
					t.Fatalf("Unexpected error starting watcher %v", err)
				}
				defer close()

				tc.updateFunc(rootDir)

				select {
				case <-triggered:
					t.Logf("Triggered test succesfully")
				case <-time.After(1 * time.Second):
					t.Fatalf("Failed to trigger the watcher update within the timeout.")
				}
			})
		})
	}
}
