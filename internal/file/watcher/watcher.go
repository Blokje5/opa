package watcher

import (
	"context"

	"gopkg.in/fsnotify.v1"
)

// StartWatcher starts an filesystem watcher at the given paths.
// Whenever an File is created, written, removed or renamed the processWatcherUpdate func is triggered
// If the file is removed or renamed, the filename is passed to the processWatcherUpdate func.
// The close func is returned to close all watches and close the event channel
func StartWatcher(ctx context.Context, paths []string, processWatcherUpdate func(context.Context, []string, string)) (func() error, error) {
	watcher, err := getWatcher(paths)
	if err != nil {
		return nil, err
	}
	go readWatcher(ctx, watcher, paths, processWatcherUpdate)
	return watcher.Close, nil
}


func getWatcher(watchPaths []string) (*fsnotify.Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	for _, path := range watchPaths {
		if err := watcher.Add(path); err != nil {
			return nil, err
		}
	}

	return watcher, nil
}

func readWatcher(ctx context.Context, watcher *fsnotify.Watcher, paths []string, processWatcherUpdate func(context.Context, []string, string)) {
	for {
		select {
		case evt := <-watcher.Events:
			removalMask := (fsnotify.Remove | fsnotify.Rename)
			mask := (fsnotify.Create | fsnotify.Write | removalMask)
			if (evt.Op & mask) != 0 {
				removed := ""
				if (evt.Op & removalMask) != 0 {
					removed = evt.Name
				}
				processWatcherUpdate(ctx, paths, removed)
			}
		}
	}
}
