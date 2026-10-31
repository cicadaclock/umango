package ui

import (
	"log"
	"path/filepath"
	"time"

	"github.com/cicadaclock/umango/internal/data"
	"github.com/cicadaclock/umango/internal/races"
	"github.com/fsnotify/fsnotify"
)

// resultsSettleDelay is the quiet time after the last file change before the results reload,
// so a file still being written reloads once it is complete
const resultsSettleDelay = time.Second

// watchTeamTrialResults reloads the team trial results in dir after each change to a result
// file and passes each reload to deliver. A failed reload is logged and the page keeps its data.
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
