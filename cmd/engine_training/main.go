// Command engine_training runs a sample of skill surveys through the race engine. The engine build
// runs it on a library that records which paths it takes, then optimizes the final library for
// those paths.
package main

import (
	"log"
	"os"
	"sync/atomic"

	"github.com/cicadaclock/umango/internal/data"
	"github.com/cicadaclock/umango/internal/db"
	"github.com/cicadaclock/umango/internal/races"
	"github.com/cicadaclock/umango/internal/racesim"
)

// Courses of each category's pool that a sample survey runs. A recording library runs slowly, and
// more courses change little about which paths a race takes.
const coursesPerCategory = 2

func main() {
	if len(os.Args) != 2 {
		log.Fatal("usage: engine_training <folder of the engine files>")
	}
	if err := run(racesim.Files{Directory: os.Args[1]}); err != nil {
		log.Fatal(err)
	}
}

// run surveys the first runner of each round of the latest team trial result
func run(files racesim.Files) error {
	engine, err := racesim.Open(files)
	if err != nil {
		return err
	}
	dataStore, err := data.Init()
	if err != nil {
		return err
	}
	directory, err := races.TeamTrialResultsDirectory()
	if err != nil {
		return err
	}
	rounds, err := racesim.ReadLatestTeamRounds(directory)
	if err != nil {
		return err
	}
	for i := range rounds {
		if err := surveyRound(engine, dataStore, &rounds[i]); err != nil {
			return err
		}
	}
	return nil
}

// surveyRound runs the first batch of a survey on the first courses of the round's category
func surveyRound(engine *racesim.Engine, dataStore *data.DataStore, round *racesim.TeamRound) error {
	config := racesim.SkillSurveyConfig{
		Engine:      engine,
		Round:       round,
		Runner:      &round.Team[0],
		Courses:     categoryCourses(dataStore.TeamTrialPool(), races.DistanceType(round.Category)),
		Conditions:  dataStore.RaceConditions(),
		Skills:      dataStore.GeneralSkills(),
		ScoreRules:  dataStore.ScoreRules(),
		SkillPoints: dataStore.TotalSkillPoints,
	}
	var stop atomic.Bool
	return racesim.RunSkillSurvey(&config, &stop, func(report racesim.SkillSurveyReport) {
		stop.Store(report.RacesPerCourse > 0)
	})
}

// categoryCourses returns the first courses of a category
func categoryCourses(pool []db.TeamTrialRace, category races.DistanceType) []db.TeamTrialRace {
	courses := make([]db.TeamTrialRace, 0, coursesPerCategory)
	for i := range pool {
		if len(courses) < coursesPerCategory && races.TeamTrialCategory(pool[i].Distance, pool[i].Ground) == category {
			courses = append(courses, pool[i])
		}
	}
	return courses
}
