# Umango

This is an UM:PD Team Trial analysis tool designed to read your team trials race data and present visualizations of aggregated data in order to help you decide your next uma upgrade. It uses your team trials folder, and your local UM:PD database to populate the data.

## Race simulator

Race simulation runs on [our fork](https://github.com/cicadaclock/uma-race-simulator) of [uma-race-simulator](https://github.com/kachi-dev/uma-race-simulator) by Kachi (MIT). Build it with Rust and Git installed:

```sh
./build_engine.sh
```

This uses the fork in `../uma-race-simulator`, cloning it there if needed, then builds `racesim/racesim.dll` and installs its game data to `racesim/data`. The game data belongs to Cygames, Inc. and is not part of this repository.

`racesim/courses.json` holds the course layouts that the course view draws, which is not included with Umango. You will need to provide your own. Races simulate without it.

## Skill analysis

The Skill Analysis tab tests which skills suit one uma of a team on the team trial courses of a category. Pick a category and an uma from your latest saved team trial. The field of that round then runs on every course of the category's pool, with season, weather and ground drawn at the game's rates and the gates shuffled in each race.

Each skill that the uma could still learn and that can activate on some course of the pool first watches 100 races per course alongside all the others, without affecting them. Those races are the baseline and give every proc rate within seconds. A skill then races for real, one at a time, on only the races where it activated, because a race where it did not activate would not change. The table refreshes every 10 races per course, up to 100, and you can stop at any time. A full run takes a few minutes.

The table shows how often each skill activates, how much it changes the win rate, and how much team trial score it adds for each skill point. The score is what the uma earns in the race before the ace, opponent rating, streak and support bonuses. The skill point cost is before any hint discount, and an upgrade costs only what it adds beyond the tier that the uma owns.

master.mdb does not list the team trial pool. Umango derives it from race group 9 and the race instances that saved team trials use, so check it again after a game update adds courses.
