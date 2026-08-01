package races

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"
	"math"
)

// Go implementation of github.com/ayaliz/hakuraku RaceDataParser

// RaceScenario represents the deserialized form of the RaceResult.RaceScenario blob
type RaceScenario struct {
	Version         int32
	DistanceDiffMax float32
	Frames          []RaceFrame
	// Per-horse outcomes, indexed by horse index
	// HorseResult indices correspond to positions in RaceStartParams.RaceHorseDataArray
	HorseResults []HorseResult
	// Point-in-time occurrences such as skill activations
	Events []RaceEvent
}

// RaceFrame represents a snapshot of every horse at any given frame
type RaceFrame struct {
	// Sim time
	Time   float32
	Horses []HorseFrame
}

// HorseFrame represents an uma's position data at any given frame
type HorseFrame struct {
	// Position from starting gate in meters
	Distance     float32
	LanePosition uint16
	Speed        float32 // In m/s
	HP           uint16
	// Non-zero while rushed
	TemptationMode int8
	// Horse index blocking this one in front, -1 if unblocked
	BlockFrontHorseIndex int8
}

// HorseResult represents a single uma's race results
type HorseResult struct {
	// 0-indexed placing
	FinishOrder int32
	// Time in seconds (sim time x ~1.2)
	FinishTimeRaw float32
	// Gap in seconds to the horse one place ahead, 0 for the winner
	FinishDiff float32
	// I lost CM16 to a double late start
	StartDelay float32
	GutsOrder  uint8
	WizOrder   uint8
	// Meters from the gate where the last spurt began
	LastSpurtStartDistance float32
	RunningStyle           RunStyle
	Defeat                 int32
	// Sim time
	FinishTimeSim float32
}

type RaceEventType int8

const (
	// Skill activation: Params[0] = horse index, Params[1] = skill ID
	EventSkill RaceEventType = 3
)

// RaceEvent represents any occurrence during the race
type RaceEvent struct {
	Time   float32 // Time in seconds
	Type   RaceEventType
	Params []int32
}

// A skill proc during the race
type SkillActivation struct {
	Time       float32
	HorseIndex int
	SkillId    int
}

// Decodes a raw race_scenario string
func DecodeRaceScenario(encoded string) (RaceScenario, error) {
	var scenario RaceScenario
	compressed, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return scenario, fmt.Errorf("base64 decode: %w", err)
	}
	gz, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		return scenario, fmt.Errorf("gzip: %w", err)
	}
	data, err := io.ReadAll(gz)
	if err != nil {
		return scenario, fmt.Errorf("read: %w", err)
	}
	return parseRaceScenario(data)
}

// Extracts skill activations from the event list
func (scenario RaceScenario) SkillActivations() []SkillActivation {
	activations := make([]SkillActivation, 0, len(scenario.Events))
	for _, event := range scenario.Events {
		if event.Type != EventSkill || len(event.Params) < 2 {
			continue
		}
		activations = append(activations, SkillActivation{
			Time:       event.Time,
			HorseIndex: int(event.Params[0]),
			SkillId:    int(event.Params[1]),
		})
	}
	return activations
}

// Minimum struct sizes used to validate data
const (
	minHorseFrameSize  = 12
	minHorseResultSize = 31
	minEventSize       = 6
)

type scenarioReader struct {
	data []byte
	off  int
	err  error
}

func parseRaceScenario(data []byte) (RaceScenario, error) {
	var scenario RaceScenario
	r := &scenarioReader{data: data}

	// Sizes of the various structs in the serialized data
	headerSize := int(r.i32())
	headerStart := r.off
	scenario.Version = r.i32()
	r.seek(headerStart + headerSize)
	scenario.DistanceDiffMax = r.f32()
	horseNum := int(r.i32())
	horseFrameSize := int(r.i32())
	horseResultSize := int(r.i32())

	// This is presumably a padding region that denotes the start of a major data
	// section, so that later data versions can append fields and still have older
	// parsers work
	r.skip(int(r.i32())) // Padding

	frameCount := int(r.i32())
	frameSize := int(r.i32())

	if r.err != nil {
		return scenario, r.err
	}
	if horseNum <= 0 || horseFrameSize < minHorseFrameSize || horseResultSize < minHorseResultSize {
		return scenario, fmt.Errorf("implausible scenario sizes: %d horses, %d byte frames, %d byte results",
			horseNum, horseFrameSize, horseResultSize)
	}
	if frameSize < 4+(horseNum*horseFrameSize) || frameCount < 0 || frameCount*frameSize > len(data)-r.off {
		return scenario, fmt.Errorf("implausible frame section: %d frames of %d bytes", frameCount, frameSize)
	}

	scenario.Frames = make([]RaceFrame, frameCount)
	for i := range scenario.Frames {
		frameStart := r.off
		frame := RaceFrame{
			Time:   r.f32(),
			Horses: make([]HorseFrame, horseNum),
		}
		for h := range frame.Horses {
			horseStart := r.off
			frame.Horses[h] = r.DeserializeHorseFrame()
			r.seek(horseStart + horseFrameSize)
		}
		scenario.Frames[i] = frame
		r.seek(frameStart + frameSize)
	}

	r.skip(int(r.i32())) // Padding

	scenario.HorseResults = make([]HorseResult, horseNum)
	for h := range scenario.HorseResults {
		resultStart := r.off
		scenario.HorseResults[h] = r.DeserializeHorseResult()
		r.seek(resultStart + horseResultSize)
	}

	r.skip(int(r.i32())) // Padding

	eventCount := int(r.i32())
	if eventCount < 0 || eventCount*minEventSize > len(data)-r.off {
		return scenario, fmt.Errorf("implausible event count %d", eventCount)
	}
	scenario.Events = make([]RaceEvent, eventCount)
	for i := range scenario.Events {
		// Each event record is prefixed with its body length
		bodySize := int(r.u16())
		next := r.off + bodySize
		event := r.DeserializeRaceEvent()
		event.Params = make([]int32, r.u8())
		for e := range event.Params {
			event.Params[e] = r.i32()
		}
		scenario.Events[i] = event
		r.seek(next)
	}

	if r.err != nil {
		return scenario, r.err
	}

	// A correct parse should land exactly on the end
	// Leftover bytes mean the parse may have failed
	if r.off != len(data) {
		return scenario, fmt.Errorf("scenario not fully parsed: stopped at %d of %d bytes", r.off, len(data))
	}

	return scenario, nil
}

// Struct layouts (all little-endian):
// header:      ii   = int32, int32                                         (8 bytes)
// raceStruct:  fiii = float32, int32, int32, int32                         (16 bytes)
// horseFrame:  fHHHbb = float32, uint16, uint16, uint16, int8, int8        (12 bytes)
// horseResult: ifffBBfBif = int32, f, f, f, uint8, uint8, f, uint8, i, f   (31 bytes)
// event:       fbb = float32, int8, int8                                   (6 bytes)
//
// These are the layouts as of data version 100000002.

func (r *scenarioReader) DeserializeHorseFrame() HorseFrame {
	result := HorseFrame{
		Distance:             r.f32(),
		LanePosition:         r.u16(),
		Speed:                float32(r.u16()) / 100,
		HP:                   r.u16(),
		TemptationMode:       r.i8(),
		BlockFrontHorseIndex: r.i8(),
	}
	return result
}

func (r *scenarioReader) DeserializeHorseResult() HorseResult {
	result := HorseResult{
		FinishOrder:            r.i32(),
		FinishTimeRaw:          r.f32(),
		FinishDiff:             r.f32(),
		StartDelay:             r.f32(),
		GutsOrder:              r.u8(),
		WizOrder:               r.u8(),
		LastSpurtStartDistance: r.f32(),
		RunningStyle:           RunStyle(r.u8()),
		Defeat:                 r.i32(),
		FinishTimeSim:          r.f32(),
	}
	return result
}

func (r *scenarioReader) DeserializeRaceEvent() RaceEvent {
	result := RaceEvent{
		Time: r.f32(),
		Type: RaceEventType(r.i8()),
	}
	return result
}

// take takes n bytes from the reader and advances the offset
func (r *scenarioReader) take(n int) []byte {
	if r.err != nil {
		return nil
	}
	if n < 0 || n > len(r.data)-r.off {
		r.err = fmt.Errorf("truncated scenario: need %d bytes at offset %d of %d", n, r.off, len(r.data))
		return nil
	}
	taken := r.data[r.off : r.off+n]
	r.off += n
	return taken
}

// seek sets the offset to n
func (r *scenarioReader) seek(n int) {
	if r.err != nil {
		return
	}
	if n < 0 || n > len(r.data) {
		r.err = fmt.Errorf("truncated scenario: seek to %d of %d", n, len(r.data))
		return
	}
	r.off = n
}

// skip advances the offset without returning anything
func (r *scenarioReader) skip(n int) {
	r.take(n)
}

func (r *scenarioReader) i32() int32 {
	b := r.take(4)
	if b == nil {
		return 0
	}
	return int32(binary.LittleEndian.Uint32(b))
}

func (r *scenarioReader) f32() float32 {
	b := r.take(4)
	if b == nil {
		return 0
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(b))
}

func (r *scenarioReader) u16() uint16 {
	b := r.take(2)
	if b == nil {
		return 0
	}
	return binary.LittleEndian.Uint16(b)
}

func (r *scenarioReader) u8() uint8 {
	b := r.take(1)
	if b == nil {
		return 0
	}
	return b[0]
}

func (r *scenarioReader) i8() int8 {
	return int8(r.u8())
}
