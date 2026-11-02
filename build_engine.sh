#!/usr/bin/env bash
# Builds the race engine library from the simulator source beside this repository and installs it
# with its game data.
set -euo pipefail
cd "$(dirname "$0")"

simulator=../uma-race-simulator
if [ ! -d "$simulator" ]; then
	git clone https://github.com/cicadaclock/uma-race-simulator.git "$simulator"
fi

cargo build --release --manifest-path engine/Cargo.toml

mkdir -p racesim/data
cp engine/target/release/racesim.dll racesim/racesim.dll
for item in master_sim.db race_paramdefine.json courseeventparam coursepath; do
	rm -rf "racesim/data/$item"
	cp -r "$simulator/data/$item" racesim/data/
done
