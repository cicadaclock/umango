// Command appicon writes the Windows application icon resource of the Umango executable.
// go generate runs it from the repository root.
package main

import (
	"fmt"
	"image"
	"log"
	"os"

	"gioui.org/gpu/headless"
	"gioui.org/op"
	"github.com/cicadaclock/umango/internal/logo"
	"github.com/tc-hib/winres"
)

// resourcePath is the object file that go build links into the Windows executable
const resourcePath = "rsrc_windows_amd64.syso"

// iconResourceID is the resource ID from which the Gio Windows backend loads the window icon
const iconResourceID = 1

// iconSizes lists the icon sizes in pixels from which Windows selects at each display scale
var iconSizes = [...]int{16, 20, 24, 32, 40, 48, 64, 256}

func main() {
	if err := writeIconResource(); err != nil {
		log.Fatalf("write icon resource %s: %v", resourcePath, err)
	}
}

// writeIconResource renders the icon images and writes them to resourcePath
func writeIconResource() error {
	images, err := renderIconImages()
	if err != nil {
		return err
	}
	icon, err := winres.NewIconFromImages(images)
	if err != nil {
		return fmt.Errorf("encode icon: %w", err)
	}
	var resources winres.ResourceSet
	if err := resources.SetIcon(winres.ID(iconResourceID), icon); err != nil {
		return fmt.Errorf("add icon: %w", err)
	}
	file, err := os.Create(resourcePath)
	if err != nil {
		return err
	}
	defer file.Close()
	if err := resources.WriteObject(file, winres.ArchAMD64); err != nil {
		return err
	}
	return file.Close()
}

// renderIconImages renders the emblem at each icon size into one column of a single canvas
// and returns one subimage per size
func renderIconImages() ([]image.Image, error) {
	width, height := iconSizes[len(iconSizes)-1], 0
	for _, size := range iconSizes {
		height += size
	}
	window, err := headless.NewWindow(width, height)
	if err != nil {
		return nil, fmt.Errorf("create canvas: %w", err)
	}
	defer window.Release()

	var ops op.Ops
	y := 0
	for _, size := range iconSizes {
		offset := op.Offset(image.Pt(0, y)).Push(&ops)
		logo.Draw(&ops, float32(size))
		offset.Pop()
		y += size
	}
	if err := window.Frame(&ops); err != nil {
		return nil, fmt.Errorf("render emblem: %w", err)
	}
	canvas := image.NewRGBA(image.Rect(0, 0, width, height))
	if err := window.Screenshot(canvas); err != nil {
		return nil, fmt.Errorf("read canvas: %w", err)
	}

	images := make([]image.Image, 0, len(iconSizes))
	y = 0
	for _, size := range iconSizes {
		images = append(images, canvas.SubImage(image.Rect(0, y, size, y+size)))
		y += size
	}
	return images, nil
}
