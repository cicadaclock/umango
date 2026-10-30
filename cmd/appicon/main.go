// Command appicon renders the Umango emblem at the Windows icon sizes and writes them as the
// application icon resource that go build links into the Windows executable.
//
// go generate runs it from the repository root; run it again whenever the emblem changes.
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

const resourcePath = "rsrc_windows_amd64.syso"

// Gio's Windows backend loads the window and taskbar icon from this resource ID
const iconResourceID = 1

// Sizes Windows picks from across display scales for the title bar, taskbar, Alt+Tab and
// Explorer, ascending to the 256px maximum of the icon format
var iconSizes = [...]int{16, 20, 24, 32, 40, 48, 64, 256}

func main() {
	if err := writeIconResource(); err != nil {
		log.Fatalf("write icon resource %s: %v", resourcePath, err)
	}
}

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

// renderIconImages draws the emblem at every icon size down one column of a single canvas and
// returns a view of each
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
