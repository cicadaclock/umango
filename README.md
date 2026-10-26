# Umango

This is an UM:PD Team Trial analysis tool designed to read your team trials race data and present visualizations of aggregated data in order to help you decide your next uma upgrade. It uses your team trials folder, and your local UM:PD database to populate the data.

## Game icons

Character, stat, running style, mood and distance type icons are extracted from your local game install and embedded into the executable at build time; they are not part of the repository. Close the game, then run this once, and again after new characters are released:

```sh
uv run assets/extract_assets.py
```

[uv](https://docs.astral.sh/uv/) (`pip install uv`) fetches Python 3.14 and the script's dependencies on the first run. Builds without extracted icons still succeed.
