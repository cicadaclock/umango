#!/usr/bin/env bash
# Builds the race engine library from the simulator source beside this repository and installs it
# with its game data. The Rust LLVM tools and a saved team trial result make the build profile
# guided: a recording library runs sample skill surveys, and the final library is optimized for the
# paths that they took. --quick builds without that.
set -euo pipefail
cd "$(dirname "$0")"

simulator=../uma-race-simulator
if [ ! -d "$simulator" ]; then
	git clone https://github.com/cicadaclock/uma-race-simulator.git "$simulator"
fi

mkdir -p racesim/data
for item in master_sim.db race_paramdefine.json courseeventparam coursepath; do
	rm -rf "racesim/data/$item"
	cp -r "$simulator/data/$item" racesim/data/
done

library=racesim/racesim.dll
profile_directory="$(pwd -W)/engine/target/profile"
merge_tool="$(rustc --print sysroot)/lib/rustlib/$(rustc -vV | sed -n 's/^host: //p')/bin/llvm-profdata.exe"
recording_directory=engine/target/recording
final_flags=""

if [ "${1:-}" != "--quick" ] && [ -x "$merge_tool" ]; then
	rm -rf "$profile_directory"
	RUSTFLAGS="-Cprofile-generate=$profile_directory" CARGO_TARGET_DIR="$recording_directory" \
		cargo build --release --manifest-path engine/Cargo.toml
	cp "$recording_directory/release/racesim.dll" "$library"
	if go run ./cmd/engine_training racesim; then
		"$merge_tool" merge -o "$profile_directory/merged.profdata" "$profile_directory"
		final_flags="-Cprofile-use=$profile_directory/merged.profdata"
	else
		echo "The sample surveys did not run, so the engine is built without a profile." >&2
	fi
elif [ "${1:-}" != "--quick" ]; then
	echo "rustup component add llvm-tools-preview makes the engine faster through a profile guided build." >&2
fi

RUSTFLAGS="$final_flags" cargo build --release --manifest-path engine/Cargo.toml
cp engine/target/release/racesim.dll "$library"
