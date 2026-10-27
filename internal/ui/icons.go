package ui

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/png"
	"io/fs"

	"gioui.org/op/paint"
	"github.com/cicadaclock/umango/internal/races"
)

// distanceTypeIconPaths locates each distance type's racing shoe icon in the embedded assets
var distanceTypeIconPaths = [races.Dirt + 1]string{
	races.Sprint: "assets/icons/distance_type/sprint.png",
	races.Mile:   "assets/icons/distance_type/mile.png",
	races.Medium: "assets/icons/distance_type/medium.png",
	races.Long:   "assets/icons/distance_type/long.png",
	races.Dirt:   "assets/icons/distance_type/dirt.png",
}

// distanceTypeIcons holds the racing shoe icon of each distance type, indexed by
// races.DistanceType; an icon whose file was not extracted stays the zero ImageOp
type distanceTypeIcons [races.Dirt + 1]paint.ImageOp

// loadDistanceTypeIcons decodes the racing shoe icons extracted by assets/extract_assets.py
//
// Inputs:
// - assets: File system holding the assets/icons tree
//
// Returns:
// - icons: Decoded icons, each zero when its file was not extracted
// - err: Failure reading or decoding an extracted icon
func loadDistanceTypeIcons(assets fs.FS) (icons distanceTypeIcons, err error) {
	for distance, path := range distanceTypeIconPaths {
		if path == "" {
			continue
		}
		if icons[distance], err = loadIcon(assets, path); err != nil {
			return distanceTypeIcons{}, err
		}
	}
	return icons, nil
}

// loadIcon decodes the PNG at path, or returns the zero ImageOp when the file does not exist
func loadIcon(assets fs.FS, path string) (paint.ImageOp, error) {
	data, err := fs.ReadFile(assets, path)
	if errors.Is(err, fs.ErrNotExist) {
		return paint.ImageOp{}, nil
	}
	if err != nil {
		return paint.ImageOp{}, fmt.Errorf("read icon %s: %w", path, err)
	}
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return paint.ImageOp{}, fmt.Errorf("decode icon %s: %w", path, err)
	}
	return paint.NewImageOp(img), nil
}

// iconLoaded reports whether icon holds an image
func iconLoaded(icon paint.ImageOp) bool {
	return icon.Size() != image.Point{}
}
