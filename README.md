# Umango

This is an UM:PD Team Trial analysis tool designed to read your team trials race data and present visualizations of aggregated data in order to help you decide your next uma upgrade. It uses your team trials folder, and your local UM:PD database to populate the data.

## Race simulator

Race simulation runs on [our fork](https://github.com/cicadaclock/uma-race-simulator) of [uma-race-simulator](https://github.com/kachi-dev/uma-race-simulator) by Kachi (MIT). Build it with Rust and Git installed:

```sh
./build_engine.sh
```

This uses the fork in `../uma-race-simulator`, cloning it there if needed, then builds `racesim/racesim.dll` and installs its game data to `racesim/data`. With `rustup component add llvm-tools-preview` and a saved team trial result, the build is profile guided: it runs sample skill surveys for a few minutes, and the engine runs about 10% faster. `./build_engine.sh --quick` skips that. The game data belongs to Cygames, Inc. and is not part of this repository.

`racesim/courses.json` holds the course layouts that the course view draws, which is not included with Umango. You will need to provide your own. Races simulate without it.

## Skill analysis

The Skill Analysis tab tests every possible skill on an existing TT uma. Pick a TT category and an uma from your latest saved team trial race. The field of that round runs on every racecourse of the category's pool, with randomized season, weather, and ground condition, and gates shuffled per race.

Skills are filtered out based on whether they can activate on the course and strategy of your uma. Remaining skills are simulated 100 times to get the average score gain per skill. The table runs 10 races per course, up to 100, and can be stopped at any time. A full run takes a few minutes.

Score is what the uma earns in the race minus the ace, opponent rating, streak and support bonus scores. This means if your uma runs out of stamina consistently, then stamina skills may add more score than what the skill is worth as your uma may now be able to finish 1/2/3. The skill point cost is before any hint discount, and an upgrade costs only what it adds beyond the previous skill tier.
