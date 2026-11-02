package racesim

import (
	"encoding/json"
	"fmt"
	"os"
)

// CourseSpan is a stretch of a course, in meters from the start
type CourseSpan struct {
	Start, End float64
}

// CourseSlope is a stretch of a course with a gradient. Positive gradients run uphill, and
// 10000 gradient units climb one meter in every meter of length.
type CourseSlope struct {
	CourseSpan
	Gradient float64
}

// CourseLayout holds the corners, straights and slopes of one course in file order
type CourseLayout struct {
	Corners   []CourseSpan
	Straights []CourseSpan
	Slopes    []CourseSlope
}

// courseLayoutFile holds the fields of the course layout file that this package reads
type courseLayoutFile struct {
	Courses []struct {
		CourseId int `json:"course_id"`
		Corners  []struct {
			Start  float64 `json:"start"`
			Length float64 `json:"length"`
		} `json:"corners"`
		Straights []struct {
			Start float64 `json:"start"`
			End   float64 `json:"end"`
		} `json:"straights"`
		Slopes []struct {
			Start    float64 `json:"start"`
			Length   float64 `json:"length"`
			Gradient float64 `json:"slope"`
		} `json:"slopes"`
	} `json:"courses"`
}

// LoadCourseLayouts reads the layout of every course from the course layout file. The error wraps
// fs.ErrNotExist when the file does not exist.
func LoadCourseLayouts(path string) (map[int]CourseLayout, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read course layouts: %w", err)
	}
	var file courseLayoutFile
	if err := json.Unmarshal(data, &file); err != nil {
		return nil, fmt.Errorf("decode course layouts: %w", err)
	}
	layouts := make(map[int]CourseLayout, len(file.Courses))
	for _, course := range file.Courses {
		layout := CourseLayout{
			Corners:   make([]CourseSpan, len(course.Corners)),
			Straights: make([]CourseSpan, len(course.Straights)),
			Slopes:    make([]CourseSlope, len(course.Slopes)),
		}
		for i, corner := range course.Corners {
			layout.Corners[i] = CourseSpan{corner.Start, corner.Start + corner.Length}
		}
		for i, straight := range course.Straights {
			layout.Straights[i] = CourseSpan{straight.Start, straight.End}
		}
		for i, slope := range course.Slopes {
			span := CourseSpan{slope.Start, slope.Start + slope.Length}
			layout.Slopes[i] = CourseSlope{span, slope.Gradient}
		}
		layouts[course.CourseId] = layout
	}
	return layouts, nil
}
