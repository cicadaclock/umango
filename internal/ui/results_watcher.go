package ui

import (
	"log"
	"path/filepath"
	"time"

	"github.com/cicadaclock/umango/internal/data"
	"github.com/cicadaclock/umango/internal/races"
	"github.com/fsnotify/fsnotify"
)

// Quiet time after the last change to a saved race before the races load again, so a file still
// being written loads once complete
const resultsSettleDelay = time.Second

// watchTeamTrialResults loads the team trial races in dir again whenever a race file there
// changes, once dir has been quiet for resultsSettleDelay, and passes each load to deliver
//
// It runs on its own goroutine for the life of the app. Failures to start watching or to load
// are logged; a failed load leaves the page as it was until the next change.
func watchTeamTrialResults(dir string, dataStore *data.DataStore, deliver func(teamTrialsData)) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Printf("watch team trial results: %v", err)
		return
	}
	defer watcher.Close()
	if err := watcher.Add(dir); err != nil {
		log.Printf("watch %s: %v", dir, err)
		return
	}
	settle := time.NewTimer(resultsSettleDelay)
	settle.Stop()
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op != fsnotify.Chmod && filepath.Ext(event.Name) == ".json" {
				settle.Reset(resultsSettleDelay)
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Printf("watch %s: %v", dir, err)
		case <-settle.C:
			resultSet, err := races.LoadRacesFolder(dir)
			if err != nil {
				log.Printf("reload team trial results from %s: %v", dir, err)
				continue
			}
			deliver(newTeamTrialsData(dataStore, resultSet))
		}
	}
}
