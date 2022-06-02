package mocwatcher

import (
	"fmt"
	"log"
	"time"

	"github.com/radovskyb/watcher"
)

func InitializeWatcher(pathToWatch string) (*watcher.Watcher, error) {
	w := watcher.New()
	w.SetMaxEvents(1)
	w.FilterOps(watcher.Rename, watcher.Move, watcher.Create, watcher.Write)

	if err := w.AddRecursive(pathToWatch); err != nil {
		return nil, fmt.Errorf("%w: error trying to watch change on %s directory", err, pathToWatch)
	}

	return w, nil
}

func AttachWatcher(w *watcher.Watcher, fn func()) {
	go func() {
		if err := w.Start(time.Millisecond * 100); err != nil {
			log.Fatalln(err)
		}
	}()

	readEventsFromWatcher(w, fn)
}

func readEventsFromWatcher(w *watcher.Watcher, fn func()) {
	go func() {
		for {
			select {
			case evt := <-w.Event:
				log.Println("File modified:", evt.Name())
				fn()
			case err := <-w.Error:
				log.Printf("Error occurred while listening read event: %+v", err)
			case <-w.Closed:
				return
			}
		}
	}()
}