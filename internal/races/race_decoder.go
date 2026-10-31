package races

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"math"

	"github.com/klauspost/compress/gzip"
)

// The race scenario format follows the RaceDataParser of github.com/ayaliz/hakuraku

// RaceScenario is the decoded race scenario of one round
type RaceScenario struct {
	Version         int32
	DistanceDiffMax float32
	Frames          []RaceFrame
	// Result of each uma, indexed by RaceHorseData.FrameOrder - 1
	HorseResults []HorseResult
	// Race events in time order
	Events []RaceEvent
}

// RaceFrame is the state of every uma at one simulation time
type RaceFrame struct {
	Time   float32
	Horses []HorseFrame
}

// HorseFrame is the state of one uma in a RaceFrame
type HorseFrame struct {
	// Meters from the start
	Distance     float32
	LanePosition uint16
	// Meters per second
	Speed float32
	HP    uint16
	// Nonzero while rushed
	TemptationMode int8
	// Horse index of the uma blocking in front, -1 when unblocked
	BlockFrontHorseIndex int8
}

// HorseResult is the race result of one uma
type HorseResult struct {
	// Placing, counted from 0
	FinishOrder int32
	// Displayed finish time in seconds, about 1.2 times the simulation time
	FinishTimeRaw float32
	// Seconds behind the uma one place ahead, 0 for the winner
	FinishDiff float32
	StartDelay float32
	GutsOrder  uint8
	WizOrder   uint8
	// Meters from the start at which the last spurt began
	LastSpurtStartDistance float32
	RunningStyle           RunStyle
	Defeat                 int32
	// Finish time in simulation seconds
	FinishTimeSim float32
}

// spurtStartTolerance is the distance in meters into the final leg within which a last spurt
// counts as a full spurt. Full spurts start within a few meters, late spurts 30 m or more.
const spurtStartTolerance = 10

// fullSpurt reports whether the uma started its last spurt on entering the final leg of a
// course courseDistance meters long
func (r HorseResult) fullSpurt(courseDistance int) bool {
	pastFinalLeg := r.LastSpurtStartDistance - float32(courseDistance)*2/3
	return pastFinalLeg >= 0 && pastFinalLeg < spurtStartTolerance
}

// RaceEventType is the kind of a race event
type RaceEventType int8

const (
	// Skill activation; Params[0] is the horse index and Params[1] is the skill ID
	EventSkill RaceEventType = 3
)

// RaceEvent is one occurrence at a simulation time, such as a skill activation
type RaceEvent struct {
	Time   float32
	Type   RaceEventType
	Params []int32
}

// SkillActivation is one skill activation by one uma
type SkillActivation struct {
	Time       float32
	HorseIndex int
	SkillId    int
}

// DecodeRaceScenario decodes one race scenario string; a ScenarioDecoder decodes many
func DecodeRaceScenario(encoded string) (RaceScenario, error) {
	var decoder ScenarioDecoder
	return decoder.Decode(encoded)
}

// ScenarioDecoder decodes race scenario strings and reuses its buffers between calls. It is not
// safe for concurrent use, so each goroutine needs its own.
type ScenarioDecoder struct {
	// When set, Decode omits RaceScenario.Frames, which the summary does not read
	SkipFrames bool
	compressed []byte
	source     bytes.Reader
	inflater   gzip.Reader
	inflated   bytes.Buffer
}

// Decode decodes one race scenario string: base64, then gzip, then the binary scenario format
func (d *ScenarioDecoder) Decode(encoded string) (RaceScenario, error) {
	var err error
	d.compressed, err = base64.StdEncoding.AppendDecode(d.compressed[:0], []byte(encoded))
	if err != nil {
		return RaceScenario{}, fmt.Errorf("base64 decode: %w", err)
	}
	d.source.Reset(d.compressed)
	if err := d.inflater.Reset(&d.source); err != nil {
		return RaceScenario{}, fmt.Errorf("gzip: %w", err)
	}
	d.inflated.Reset()
	if _, err := d.inflated.ReadFrom(&d.inflater); err != nil {
		return RaceScenario{}, fmt.Errorf("read: %w", err)
	}
	return parseRaceScenario(d.inflated.Bytes(), d.SkipFrames)
}

// SkillActivations returns the skill activations among the race events
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

// Minimum record sizes in bytes, as of data version 100000002
const (
	minHorseFrameSize  = 12
	minHorseResultSize = 31
	minEventSize       = 6
)

// scenarioReader reads little-endian values from a scenario. After the first error, every
// read returns zero and err holds that error.
type scenarioReader struct {
	data []byte
	off  int
	err  error
}

// parseRaceScenario parses the binary scenario format. Record sizes are read from the data,
// so newer data versions with longer records still parse.
func parseRaceScenario(data []byte, skipFrames bool) (RaceScenario, error) {
	var scenario RaceScenario
	r := &scenarioReader{data: data}

	headerSize := int(r.i32())
	headerStart := r.off
	scenario.Version = r.i32()
	r.seek(headerStart + headerSize)
	scenario.DistanceDiffMax = r.f32()
	horseNum := int(r.i32())
	horseFrameSize := int(r.i32())
	horseResultSize := int(r.i32())

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

	if skipFrames {
		r.skip(frameCount * frameSize)
		frameCount = 0
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

	// A correct parse ends exactly at the end of the data
	if r.off != len(data) {
		return scenario, fmt.Errorf("scenario not fully parsed: stopped at %d of %d bytes", r.off, len(data))
	}

	return scenario, nil
}

// DeserializeHorseFrame reads one horse frame record
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

// DeserializeHorseResult reads one horse result record
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

// DeserializeRaceEvent reads the fixed part of one race event record
func (r *scenarioReader) DeserializeRaceEvent() RaceEvent {
	result := RaceEvent{
		Time: r.f32(),
		Type: RaceEventType(r.i8()),
	}
	return result
}

// take returns the next n bytes and advances past them, or records an error when fewer remain
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

// seek moves to offset n, or records an error when n is outside the data
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

// skip advances past the next n bytes
func (r *scenarioReader) skip(n int) {
	r.take(n)
}

// i32 reads a little-endian int32
func (r *scenarioReader) i32() int32 {
	b := r.take(4)
	if b == nil {
		return 0
	}
	return int32(binary.LittleEndian.Uint32(b))
}

// f32 reads a little-endian float32
func (r *scenarioReader) f32() float32 {
	b := r.take(4)
	if b == nil {
		return 0
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(b))
}

// u16 reads a little-endian uint16
func (r *scenarioReader) u16() uint16 {
	b := r.take(2)
	if b == nil {
		return 0
	}
	return binary.LittleEndian.Uint16(b)
}

// u8 reads a uint8
func (r *scenarioReader) u8() uint8 {
	b := r.take(1)
	if b == nil {
		return 0
	}
	return b[0]
}

// i8 reads an int8
func (r *scenarioReader) i8() int8 {
	return int8(r.u8())
}
