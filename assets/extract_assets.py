# /// script
# requires-python = ">=3.14,<3.15"
# dependencies = [
#     "apsw-sqlite3mc==3.54.0.0",
#     "Pillow==13.0.0",
#     "UnityPy==1.25.4",
# ]
# ///
"""Extract character and UI icons from the local Umamusume: Pretty Derby install into assets/.

Downloads UmaExporter at a pinned commit and uses its meta database reader and asset bundle
decryptor; UnityPy decodes the textures. Icons are written to assets/icons as
<kind>/<name>.png, downscaled to at most ICON_HEIGHT pixels tall, for main.go to embed:

- character/<chara_id>.png
- stat/{speed,stamina,power,guts,wit}.png
- motivation/{awful,bad,normal,good,great}.png
- distance_type/{sprint,mile,medium,long,dirt}.png

Close the game first, then run from any directory:

    uv run assets/extract_assets.py [--game-data DIR]
"""

import argparse
import hashlib
import io
import sys
import tempfile
import urllib.request
import zipfile
from pathlib import Path

import UnityPy
from PIL import Image

EXTRACTOR_COMMIT = "027ed178120b6a21f2bb43970b82e89d2f4ffc36"  # UmaExporter v0.5.0
EXTRACTOR_SHA256 = "8686e3c63e7a33c804b2b3d3814dc05fe8c426f8fc57be9cde814b188bc37c43"
EXTRACTOR_URL = f"https://codeload.github.com/mumu-lhl/UmaExporter/zip/{EXTRACTOR_COMMIT}"
REGION = "global"
ICON_HEIGHT = 64

ASSETS_DIR = Path(__file__).resolve().parent
ICON_DIR = ASSETS_DIR / "icons"
EXTRACTOR_DIR = ASSETS_DIR / ".extractor"

COMMON_ATLAS = "atlas/common/common_tex"
COMMON_ATLAS_SPRITES = {
    "utx_ico_charastatus_00": "stat/speed.png",
    "utx_ico_charastatus_01": "stat/stamina.png",
    "utx_ico_charastatus_02": "stat/power.png",
    "utx_ico_charastatus_03": "stat/guts.png",
    "utx_ico_charastatus_04": "stat/wit.png",
    "utx_ico_motivation_m_00": "motivation/awful.png",
    "utx_ico_motivation_m_01": "motivation/bad.png",
    "utx_ico_motivation_m_02": "motivation/normal.png",
    "utx_ico_motivation_m_03": "motivation/good.png",
    "utx_ico_motivation_m_04": "motivation/great.png",
}
# Racing shoe items by distance type; each item bundle holds one Texture2D of the same name
DISTANCE_TYPE_SHOE_ITEMS = {
    "sprint": "item_icon_00001",
    "mile": "item_icon_00004",
    "medium": "item_icon_00007",
    "long": "item_icon_00010",
    "dirt": "item_icon_00013",
}


def default_game_data() -> Path:
    """Return the Steam Global game data folder that umango reads master.mdb from.

    Returns:
    - path: Folder holding meta, master/master.mdb and dat/ on this platform
    """
    if sys.platform == "win32":
        return Path.home() / "AppData/LocalLow/Cygames/Umamusume"
    return Path.home() / (
        ".local/share/Steam/steamapps/compatdata/3224770/pfx/drive_c/users/steamuser"
        "/AppData/LocalLow/Cygames/Umamusume"
    )


def fetch_extractor() -> Path:
    """Download and unpack the pinned UmaExporter source unless EXTRACTOR_DIR already holds it.

    Returns:
    - root: Folder holding UmaExporter's src package

    Raises:
    - urllib.error.URLError: The download failed
    - ValueError: The archive does not match EXTRACTOR_SHA256
    """
    root = EXTRACTOR_DIR / f"UmaExporter-{EXTRACTOR_COMMIT}"
    if root.is_dir():
        return root
    with urllib.request.urlopen(EXTRACTOR_URL) as response:
        archive = response.read()
    digest = hashlib.sha256(archive).hexdigest()
    if digest != EXTRACTOR_SHA256:
        raise ValueError(f"UmaExporter archive SHA-256 is {digest}, expected {EXTRACTOR_SHA256}")
    EXTRACTOR_DIR.mkdir(exist_ok=True)
    # Unpack beside the target and rename, so an interrupted run never leaves a partial root
    with tempfile.TemporaryDirectory(dir=EXTRACTOR_DIR) as staging:
        with zipfile.ZipFile(io.BytesIO(archive)) as zip_file:
            zip_file.extractall(staging)
        (Path(staging) / root.name).rename(root)
    return root


def extract_icons(extractor_root: Path, game_data: Path) -> int:
    """Write the character icons, COMMON_ATLAS_SPRITES and the distance type shoes to ICON_DIR.

    Inputs:
    - extractor_root: Folder holding UmaExporter's src package, from fetch_extractor
    - game_data: Folder holding meta, master/master.mdb and dat/

    Returns:
    - failures: Number of asset bundles whose icons could not all be written; each is reported
      on stderr

    Raises:
    - LookupError: The meta database lists no COMMON_ATLAS or no racing shoe item
    - ValueError: UmaExporter could not decrypt the meta database
    """
    sys.path.insert(0, str(extractor_root))
    from src.core.config import Config
    from src.core.database import UmaDatabase
    from src.core.decryptor import DEFAULT_BASE_KEYS, DEFAULT_KEY, decrypt_bundle

    Config.REGION = REGION
    Config.DB_ENCRYPTED = Config.is_encrypted_region(REGION)
    Config.set_base_path(str(game_data))
    database = UmaDatabase()
    try:
        # Each job is a meta database row, its object type, and its object names to PNG paths
        jobs = [
            (entry, "Texture2D", {entry["texture_name"]: f"character/{entry['chara_id']}.png"})
            for entry in database.get_character_entries()
        ]
        named_jobs = [(COMMON_ATLAS, "Sprite", COMMON_ATLAS_SPRITES)] + [
            (f"item/{item}", "Texture2D", {item: f"distance_type/{name}.png"})
            for name, item in DISTANCE_TYPE_SHOE_ITEMS.items()
        ]
        for asset_path, object_type, outputs in named_jobs:
            asset = database.get_asset_by_path(asset_path)
            if asset is None:
                raise LookupError(f"meta database lists no {asset_path}")
            jobs.append((asset, object_type, outputs))
    finally:
        database.close()

    bundle_root = Path(Config.get_data_root())
    written = failures = 0
    for asset, object_type, outputs in jobs:
        bundle_hash = asset["hash"]
        try:
            key = DEFAULT_KEY if asset["key"] is None else int(asset["key"])
            bundle = bytearray((bundle_root / bundle_hash[:2] / bundle_hash).read_bytes())
            # Global encrypts meta with its own key but its bundles with the default base keys
            bundle = decrypt_bundle(bundle, REGION, key, DEFAULT_BASE_KEYS)
            written += write_icons(bytes(bundle), object_type, outputs)
        except (OSError, LookupError, ValueError) as err:
            failures += 1
            print(f"extracting {asset['full_path']} ({bundle_hash}): {err}", file=sys.stderr)
    print(f"Wrote {written} icons to {ICON_DIR}")
    return failures


def write_icons(bundle: bytes, object_type: str, outputs: dict[str, str]) -> int:
    """Save the named textures or sprites of a decrypted asset bundle as downscaled PNGs.

    Inputs:
    - bundle: Decrypted UnityFS asset bundle
    - object_type: Unity class of the objects to export, "Texture2D" or "Sprite"
    - outputs: m_Name of each object to export, mapped to its PNG path relative to ICON_DIR;
      existing files are overwritten

    Returns:
    - written: Number of PNGs written, len(outputs)

    Raises:
    - LookupError: The bundle lacks some of the named objects; the others are still written
    - OSError: A PNG could not be written
    """
    pending = dict(outputs)
    for obj in UnityPy.load(bundle).objects:
        if obj.type.name != object_type:
            continue
        path = pending.pop(obj.peek_name(), None)
        if path is None:
            continue
        image_object = obj.read()
        icon = image_object.image
        if object_type == "Sprite":
            # Tightly packed sprites decode trimmed; restore the full frame so icons align
            rect, offset = image_object.m_Rect, image_object.m_RD.textureRectOffset
            frame = Image.new("RGBA", (round(rect.width), round(rect.height)))
            frame.paste(icon, (round(offset.x), round(rect.height - offset.y) - icon.height))
            icon = frame
        icon.thumbnail((icon.width, ICON_HEIGHT), Image.Resampling.LANCZOS)
        (ICON_DIR / path).parent.mkdir(parents=True, exist_ok=True)
        icon.save(ICON_DIR / path, optimize=True)
    if pending:
        raise LookupError(f"no {object_type} named {', '.join(pending)}")
    return len(outputs)


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__.split("\n\n")[0])
    parser.add_argument(
        "--game-data",
        type=Path,
        default=default_game_data(),
        help="folder holding meta, master/master.mdb and dat/ (default: %(default)s)",
    )
    args = parser.parse_args()
    if not (args.game_data / "meta").is_file():
        print(f"no meta file in {args.game_data}; pass --game-data", file=sys.stderr)
        return 1
    return 1 if extract_icons(fetch_extractor(), args.game_data) else 0


if __name__ == "__main__":
    sys.exit(main())
