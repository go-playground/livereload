package livereload

import (
	"os"
	"path/filepath"
	"strconv"

	"github.com/go-playground/log"
	"github.com/go-playground/log/handlers/console"
	"github.com/jaschaephraim/lrserver"
	"gopkg.in/fsnotify.v1"
)

// emoji's used in logg messages for at a glance changes.
const (
	errorEmoji        = "‼️ "
	unknownErrorEmoji = "⁉️ "
	thumbsUpEmoji     = "👍 "
	DefaultPort       = lrserver.DefaultPort
)

func init() {

	if !log.HasHandlers() {
		log.RegisterHandler(console.New(), log.AllLevels...)
	}
}

// PreReloadFunc is the reload function, mapped to a given extension
// just in case you wish to do something before notifying the browser/listener
// of the change. reload indicates if the reload notification should proceed
// eg. PreReloadFunc may just compile .sass to .css so no need to notify
// for the .sass conversion because the .css change will be caught and
// will trigger the notify.
type PreReloadFunc func(name string) (reload bool, err error)

// ReloadMapping is a map of file extensions
// to their PreReloadFunc's. PreReloadFunc mapping can be
// nil if nor function needs to be run.
type ReloadMapping map[string]PreReloadFunc

// ListenAndServe sets up an asset livereload monitor and notification instance.
// default livereload port is 35729.
//
// if you wish to stop the listener just close the returned 'done' channel
func ListenAndServe(livereloadPort uint16, paths []string, mappings ReloadMapping) (done chan struct{}, err error) {

	done = make(chan struct{})

	go func() {
		var watcher *fsnotify.Watcher

		watcher, err = fsnotify.NewWatcher()
		if err != nil {
			log.WithFields(log.F("error", err)).Fatal("failed to initialize fsnotify watcher")
		}

		defer watcher.Close()

		walker := func(path string, info os.FileInfo, err error) error {

			if info.IsDir() {
				err = watcher.Add(path)
				if err != nil {
					return err
				}
			}

			return nil
		}

		for _, path := range paths {

			err = filepath.Walk(path, walker)
			if err != nil {
				log.WithFields(log.F("error", err)).Fatal("failure walking path(s) to monitor")
			}
		}

		lr := lrserver.New(lrserver.DefaultName+":"+strconv.Itoa(int(livereloadPort)), livereloadPort)

		go func() {
			err := lr.ListenAndServe()
			if err != nil {
				log.WithFields(log.F("error", err)).Error("error with livereload")
			}
		}()

		go func() {
			for {
				select {
				case event := <-watcher.Events:

					if event.Op&fsnotify.Remove == fsnotify.Remove || event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {

						ext := filepath.Ext(event.Name)

						fn, ok := mappings[ext]
						if !ok {
							continue
						}

						if fn != nil {

							reload, err := fn(event.Name)
							if err != nil {
								log.WithFields(log.F("error", err)).Errorf("%s preload function error", errorEmoji)
								continue
							}

							if !reload {
								continue
							}
						}

						log.WithFields(log.F("file", event.Name)).Infof("%s %s file updated", thumbsUpEmoji, ext)

						lr.Reload(event.Name)
					}

				case err := <-watcher.Errors:
					log.WithFields(log.F("error", err)).Errorf("%s unknown event watch error", unknownErrorEmoji)
				}
			}
		}()

		<-done
	}()

	return
}
