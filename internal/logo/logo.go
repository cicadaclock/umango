// Package logo draws the Umango emblem, a glowing enchanted mango, from vector shapes so that it
// stays sharp at every size it is shown at.
package logo

import (
	"image/color"
	"math"

	"gioui.org/f32"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
)

// The halo around the fruit, leaf and stem is glowLayers translucent fills, each grown glowStep
// design units past the next, under a rim grown rimWidth past the outline. Filled outlines stand
// in for strokes, whose coverage Gio's renderer leaks in streaks below them.
const (
	glowLayers = 6
	glowStep   = 0.45
	rimWidth   = 0.8
)

// curve is a closed contour of cubic Bézier segments from start, in design units with y pointing
// down; each segment lists its two control points and its end point
type curve struct {
	start    f32.Point
	segments [][3]f32.Point
}

// shape is a smooth closed outline with the grown outlines of its halo layers, outermost first,
// and of its rim
type shape struct {
	outline curve
	halo    [glowLayers]curve
	rim     curve
}

// shading is a linear gradient between two points in design units
type shading struct {
	from, to           f32.Point
	fromColor, toColor color.NRGBA
}

var (
	// Anchor points traced around each part, which the outlines pass through
	body = newShape(
		f32.Pt(5.3, 72.8), f32.Pt(6.5, 77.8), f32.Pt(12.3, 84.4), f32.Pt(23.9, 89.4),
		f32.Pt(38.8, 89.9), f32.Pt(52.1, 85.3), f32.Pt(62.0, 75.3), f32.Pt(67.8, 63.7),
		f32.Pt(70.0, 50.5), f32.Pt(67.0, 38.8), f32.Pt(60.4, 31.4), f32.Pt(50.4, 29.7),
		f32.Pt(38.8, 32.2), f32.Pt(27.6, 37.6), f32.Pt(20.2, 47.4), f32.Pt(17.0, 59.8),
		f32.Pt(12.6, 66.8),
	)
	leaf = newShape(
		f32.Pt(8.5, 31.2), f32.Pt(14.8, 27.0), f32.Pt(22.2, 23.6), f32.Pt(32.2, 20.9),
		f32.Pt(42.1, 19.2), f32.Pt(50.4, 18.9), f32.Pt(57.1, 20.3), f32.Pt(58.4, 23.1),
		f32.Pt(55.4, 25.9), f32.Pt(45.5, 28.3), f32.Pt(33.8, 31.0), f32.Pt(23.9, 33.4),
		f32.Pt(15.6, 34.0), f32.Pt(11.2, 32.9),
	)
	stem = newShape(
		f32.Pt(56.2, 26.8), f32.Pt(58.7, 22.8), f32.Pt(61.7, 19.5), f32.Pt(65.0, 17.7),
		f32.Pt(66.6, 19.2), f32.Pt(63.3, 21.0), f32.Pt(60.3, 24.3), f32.Pt(57.8, 28.3),
	)
	// Glossy reflection on the fruit's upper right shoulder, its halo softening its edge
	highlight = newShape(
		f32.Pt(51.0, 37.4), f32.Pt(53.8, 36.2), f32.Pt(56.2, 38.4), f32.Pt(56.8, 42.0),
		f32.Pt(55.6, 45.6), f32.Pt(53.4, 46.2), f32.Pt(51.8, 43.6), f32.Pt(50.8, 40.4),
	)
	// Bounds of the outermost halo, which Draw centers and fits to its square
	extentMin, extentMax = haloExtent(&body, &leaf, &stem)

	// The fruit ripens from maroon at the stem through orange to olive at the lower lobe, with a
	// mauve shadow along its underside, painted as one gradient per pair of stops
	bodyShadings = [...]shading{
		{f32.Pt(66, 32), f32.Pt(40, 60), rgba(0x801c0eff), rgba(0xda6612ff)},
		{f32.Pt(38, 60), f32.Pt(12, 84), rgba(0xda661200), rgba(0xb6b320ff)},
		{f32.Pt(48, 66), f32.Pt(56, 88), rgba(0x93314e00), rgba(0x93314ec0)},
	}
	leafShading        = shading{f32.Pt(30, 18), f32.Pt(30, 34), rgba(0x9cfabdff), rgba(0x12a66aff)}
	colorGlow          = rgba(0x30c0ff26)
	colorRim           = rgba(0x7fe6ffff)
	colorStem          = rgba(0x0a4e78ff)
	colorHighlight     = rgba(0xfdcdb960)
	colorHighlightHalo = rgba(0xfdcdb914)
)

// Draw records the emblem into ops, centered in and fitted to the square from the origin to
// (side, side)
//
// Inputs:
// - ops: Operation list to record into, under the caller's offset and clip
// - side: Side of the square in pixels
func Draw(ops *op.Ops, side float32) {
	// Points are fitted on the CPU, since Gio misrenders gradients under an enlarging op.Affine
	size := extentMax.Sub(extentMin)
	scale := side / max(size.X, size.Y)
	fit := f32.AffineId().
		Offset(extentMin.Add(extentMax).Mul(-0.5)).
		Scale(f32.Point{}, f32.Pt(scale, scale)).
		Offset(f32.Pt(side/2, side/2))

	for layer := range glowLayers {
		for _, part := range [...]*shape{&body, &leaf, &stem} {
			fillCurve(ops, part.halo[layer], fit, colorGlow)
		}
	}
	fillCurve(ops, body.rim, fit, colorRim)
	bodyPath := body.outline.path(ops, fit)
	for _, s := range bodyShadings {
		fillShading(ops, bodyPath, fit, s)
	}
	fruit := clip.Outline{Path: bodyPath}.Op().Push(ops)
	for _, layer := range highlight.halo {
		fillCurve(ops, layer, fit, colorHighlightHalo)
	}
	fillCurve(ops, highlight.outline, fit, colorHighlight)
	fruit.Pop()
	fillCurve(ops, leaf.rim, fit, colorRim)
	fillShading(ops, leaf.outline.path(ops, fit), fit, leafShading)
	fillCurve(ops, stem.rim, fit, colorRim)
	fillCurve(ops, stem.outline, fit, colorStem)
}

// newShape smooths anchors into a closed outline and grows it along its normals into the
// outlines of its halo layers and rim
func newShape(anchors ...f32.Point) shape {
	normals := outwardNormals(anchors)
	grown := func(distance float32) curve {
		points := make([]f32.Point, len(anchors))
		for i, p := range anchors {
			points[i] = p.Add(normals[i].Mul(distance))
		}
		return smoothLoop(points)
	}
	s := shape{outline: smoothLoop(anchors), rim: grown(rimWidth)}
	for layer := range s.halo {
		s.halo[layer] = grown(rimWidth + glowStep*float32(glowLayers-layer))
	}
	return s
}

// outwardNormals returns the unit normal at each anchor of the loop through anchors, taken across
// the chord between its neighbors and pointing away from the enclosed area
func outwardNormals(anchors []f32.Point) []f32.Point {
	n := len(anchors)
	// The shoelace sum is positive when the normal right of the direction of travel points out
	var area float32
	for i, p := range anchors {
		q := anchors[(i+1)%n]
		area += p.X*q.Y - q.X*p.Y
	}
	side := float32(1)
	if area < 0 {
		side = -1
	}
	normals := make([]f32.Point, n)
	for i := range anchors {
		chord := anchors[(i+1)%n].Sub(anchors[(i+n-1)%n])
		length := float32(math.Hypot(float64(chord.X), float64(chord.Y)))
		normals[i] = f32.Pt(chord.Y, -chord.X).Mul(side / length)
	}
	return normals
}

// smoothLoop returns the closed curve through points whose segments take their tangents from the
// neighboring points, as a uniform Catmull-Rom spline does
func smoothLoop(points []f32.Point) curve {
	n := len(points)
	at := func(i int) f32.Point { return points[(i+n)%n] }
	c := curve{start: points[0], segments: make([][3]f32.Point, n)}
	for i := range n {
		p0, p1, p2, p3 := at(i-1), at(i), at(i+1), at(i+2)
		c.segments[i] = [3]f32.Point{p1.Add(p2.Sub(p0).Div(6)), p2.Sub(p3.Sub(p1).Div(6)), p2}
	}
	return c
}

// haloExtent returns the minimum and maximum corners of the control points of the parts'
// outermost halo layers, whose bounds contain the curves
func haloExtent(parts ...*shape) (lo, hi f32.Point) {
	lo, hi = parts[0].halo[0].start, parts[0].halo[0].start
	for _, part := range parts {
		for _, segment := range part.halo[0].segments {
			for _, p := range segment {
				lo = f32.Pt(min(lo.X, p.X), min(lo.Y, p.Y))
				hi = f32.Pt(max(hi.X, p.X), max(hi.Y, p.Y))
			}
		}
	}
	return lo, hi
}

// path records the curve mapped from design units to pixels by fit
func (c curve) path(ops *op.Ops, fit f32.Affine2D) clip.PathSpec {
	var p clip.Path
	p.Begin(ops)
	p.MoveTo(fit.Transform(c.start))
	for _, s := range c.segments {
		p.CubeTo(fit.Transform(s[0]), fit.Transform(s[1]), fit.Transform(s[2]))
	}
	p.Close()
	return p.End()
}

func fillCurve(ops *op.Ops, c curve, fit f32.Affine2D, fill color.NRGBA) {
	paint.FillShape(ops, fill, clip.Outline{Path: c.path(ops, fit)}.Op())
}

// fillShading paints the area inside path with s, its stops mapped from design units to pixels
// by fit
func fillShading(ops *op.Ops, path clip.PathSpec, fit f32.Affine2D, s shading) {
	defer clip.Outline{Path: path}.Op().Push(ops).Pop()
	paint.LinearGradientOp{
		Stop1:  fit.Transform(s.from),
		Color1: s.fromColor,
		Stop2:  fit.Transform(s.to),
		Color2: s.toColor,
	}.Add(ops)
	paint.PaintOp{}.Add(ops)
}

// rgba converts a 0xRRGGBBAA literal to a color
func rgba(hex uint32) color.NRGBA {
	return color.NRGBA{R: uint8(hex >> 24), G: uint8(hex >> 16), B: uint8(hex >> 8), A: uint8(hex)}
}
