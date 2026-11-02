# Umango

This is an UM:PD Team Trial analysis tool designed to read your team trials race data and present visualizations of aggregated data in order to help you decide your next uma upgrade. It uses your team trials folder, and your local UM:PD database to populate the data.

## Race simulator

Race simulation runs on [our fork](https://github.com/cicadaclock/uma-race-simulator) of [uma-race-simulator](https://github.com/kachi-dev/uma-race-simulator) by Kachi (MIT). Build it with Rust and Git installed:

```sh
./build_engine.sh
```

This uses the fork in `../uma-race-simulator`, cloning it there if needed, then builds `racesim/racesim.dll` and installs its game data to `racesim/data`. The game data belongs to Cygames, Inc. and is not part of this repository.

`racesim/courses.json` holds the course layouts that the course view draws, which is not included with Umango. You will need to provide your own. Races simulate without it.
