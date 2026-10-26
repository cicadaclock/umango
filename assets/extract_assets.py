# /// script
# requires-python = ">=3.14,<3.15"
# dependencies = [
#     "apsw-sqlite3mc==3.54.0.0",
#     "Pillow==13.0.0",
#     "UnityPy==1.25.4",
# ]
# ///
"""Extract the character icons from the local Umamusume: Pretty Derby install into assets/.

Downloads UmaExporter at a pinned commit and uses its meta database reader and asset bundle
decryptor; UnityPy decodes the textures. Each icon is written to assets/character_icons as
<chara_id>.png, downscaled to fit ICON_SIZE pixels, for main.go to embed into the executable.

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
ICON_SIZE = 64

ASSETS_DIR = Path(__file__).resolve().parent
ICON_DIR = ASSETS_DIR / "character_icons"
EXTRACTOR_DIR = ASSETS_DIR / ".extractor"


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
    """Write every character icon listed in the game's meta database to ICON_DIR.

    Inputs:
    - extractor_root: Folder holding UmaExporter's src package, from fetch_extractor
    - game_data: Folder holding meta, master/master.mdb and dat/

    Returns:
    - failures: Number of icons that could not be written; each is reported on stderr

    Raises:
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
        entries = database.get_character_entries()
    finally:
        database.close()

    ICON_DIR.mkdir(exist_ok=True)
    bundle_root = Path(Config.get_data_root())
    failures = 0
    for entry in entries:
        bundle_hash = entry["hash"]
        try:
            key = DEFAULT_KEY if entry["key"] is None else int(entry["key"])
            bundle = bytearray((bundle_root / bundle_hash[:2] / bundle_hash).read_bytes())
            # Global encrypts meta with its own key but its bundles with the default base keys
            bundle = decrypt_bundle(bundle, REGION, key, DEFAULT_BASE_KEYS)
            write_icon(bytes(bundle), entry["texture_name"], ICON_DIR / f"{entry['chara_id']}.png")
        except (OSError, LookupError, ValueError) as err:
            failures += 1
            print(f"extracting {entry['full_path']} ({bundle_hash}): {err}", file=sys.stderr)
    print(f"Wrote {len(entries) - failures} of {len(entries)} icons to {ICON_DIR}")
    return failures


def write_icon(bundle: bytes, texture_name: str, path: Path) -> None:
    """Decode the named texture from a decrypted asset bundle and save it as a downscaled PNG.

    Inputs:
    - bundle: Decrypted UnityFS asset bundle
    - texture_name: m_Name of the Texture2D to export
    - path: Destination PNG file, overwritten if present

    Raises:
    - LookupError: The bundle holds no Texture2D named texture_name
    - OSError: The PNG could not be written
    """
    for obj in UnityPy.load(bundle).objects:
        if obj.type.name != "Texture2D":
            continue
        texture = obj.read()
        if texture.m_Name != texture_name:
            continue
        icon = texture.image
        icon.thumbnail((ICON_SIZE, ICON_SIZE), Image.Resampling.LANCZOS)
        icon.save(path, optimize=True)
        return
    raise LookupError(f"no Texture2D named {texture_name}")


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
