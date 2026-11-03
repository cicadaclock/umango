package ui

import (
	"cmp"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"gioui.org/text"
	"gioui.org/unit"
	"github.com/cicadaclock/umango/internal/races"
	"github.com/cicadaclock/umango/internal/racesim"
)

const (
	placementRowHeight    = unit.Dp(34)
	placementHeaderHeight = unit.Dp(26)
	placementCellLines    = 2
	// Seconds below which a mode counts as not entered
	placementMinimumModeTime = 0.01
	placementEmptyCell       = "-"
)

// placementRow holds the values of one runner in the placement table
type placementRow struct {
	// Gate number, counted from 1
	gate        int
	skillPoints int
	runner      *racesim.Runner
	summary     *racesim.RunnerSummary
}

// placementColumns defines the placement table, whose columns follow the race analysis table of
// Hakuraku, with the win rate in place of its predicted win chance
var placementColumns = [...]struct {
	header string
	style  tableColumn
	cell   func(row *placementRow) string
}{
	{"Finish", tableColumn{alignment: text.End}, func(row *placementRow) string {
		return fmt.Sprintf("%.2f", row.summary.Finish)
	}},
	{"No.", tableColumn{alignment: text.End}, func(row *placementRow) string {
		return strconv.Itoa(row.gate)
	}},
	{"Uma", tableColumn{fill: true, fillFitsCells: true}, placementUmaCell},
	{"Time", tableColumn{alignment: text.End}, placementTimeCell},
	{"Win Rate", tableColumn{alignment: text.End}, func(row *placementRow) string {
		return fmt.Sprintf("%.1f%%", row.summary.WinRate*100)
	}},
	{"Style/Mood", tableColumn{}, placementStyleCell},
	{"Delay", tableColumn{alignment: text.End}, func(row *placementRow) string {
		return fmt.Sprintf("%.1fms\nLate %.0f%%", row.summary.StartDelay*1000, row.summary.LateStartRate*100)
	}},
	{"Last spurt", tableColumn{}, placementSpurtCell},
	{"HP Result", tableColumn{}, placementHPCell},
	{"Duel", tableColumn{alignment: text.End}, func(row *placementRow) string {
		if row.summary.DuelTime < placementMinimumModeTime {
			return placementEmptyCell
		}
		return fmt.Sprintf("%.1fs", row.summary.DuelTime)
	}},
	{"Downhill", tableColumn{alignment: text.End}, placementDownhillCell},
	{"Pace", tableColumn{alignment: text.End}, placementPaceCell},
	{"WT", tableColumn{alignment: text.End}, func(row *placementRow) string {
		return fmt.Sprintf("-%.2fm", max(row.summary.WorldTransformLoss, 0))
	}},
	{"Stats", tableColumn{}, func(row *placementRow) string {
		r := row.runner
		return fmt.Sprintf("%d/%d/%d\n%d/%d/%d SP", r.Speed, r.Stamina, r.Power, r.Guts, r.Wit, row.skillPoints)
	}},
}

// newPlacementTable returns the placement table without rows
func newPlacementTable() tableView {
	headers := make([]string, len(placementColumns))
	styles := make([]tableColumn, len(placementColumns))
	for col := range placementColumns {
		headers[col] = placementColumns[col].header
		styles[col] = placementColumns[col].style
	}
	table := newTableView(headers, styles, placementRowHeight, true)
	table.headerHeight = placementHeaderHeight
	table.cellLines = placementCellLines
	return table
}

// placementCells orders the rows by mean place and returns the cell text, one slice per column
func placementCells(rows []placementRow) [][]string {
	slices.SortStableFunc(rows, func(a, b placementRow) int {
		return cmp.Compare(a.summary.Finish, b.summary.Finish)
	})
	columns := make([][]string, len(placementColumns))
	for col := range placementColumns {
		cells := make([]string, len(rows))
		for i := range rows {
			cells[i] = placementColumns[col].cell(&rows[i])
		}
		columns[col] = cells
	}
	return columns
}

// placementUmaCell shows the chara above the trainer
func placementUmaCell(row *placementRow) string {
	if row.runner.Trainer == "" {
		return row.runner.Name
	}
	return row.runner.Name + "\n[" + row.runner.Trainer + "]"
}

// placementTimeCell shows the finish time in race simulation time above the distance to the
// previous finisher
func placementTimeCell(row *placementRow) string {
	seconds := row.summary.FinishTime
	minutes := int(seconds) / 60
	cell := fmt.Sprintf("%d:%07.4f", minutes, seconds-float64(minutes*60))
	if row.summary.FinishGap > 0 {
		cell += fmt.Sprintf("\n+%.1fm", row.summary.FinishGap)
	}
	return cell
}

// placementStyleCell shows the running style above the mood
func placementStyleCell(row *placementRow) string {
	style := races.RunStyle(row.runner.RunningStyle).String()
	if row.runner.Runaway {
		style = "Runaway"
	}
	return style + "\n" + moodName(row.runner.Mood)
}

// moodName returns the display name of a mood
func moodName(mood int) string {
	switch mood {
	case 1:
		return "Awful"
	case 2:
		return "Bad"
	case 3:
		return "Normal"
	case 4:
		return "Good"
	case 5:
		return "Great"
	}
	return ""
}

// placementSpurtCell shows the last spurt delay above its target speed, with the rate of races
// without a last spurt when there are any
func placementSpurtCell(row *placementRow) string {
	s := row.summary
	if s.NoSpurtRate == 1 {
		return "No spurt"
	}
	noSpurt := ""
	if s.NoSpurtRate > 0 {
		noSpurt = fmt.Sprintf(", No spurt %.0f%%", s.NoSpurtRate*100)
	}
	return fmt.Sprintf("Delay: %.1fm%s\nSpeed: %.2f", s.SpurtDelay, noSpurt, s.SpurtSpeed)
}

// placementHPCell shows whether HP lasted to the goal above the HP margin at the finish
func placementHPCell(row *placementRow) string {
	s := row.summary
	status := "Survived"
	if s.HPOutRate > 0 {
		status = fmt.Sprintf("Died %.0f%% (-%.0fm)", s.HPOutRate*100, s.HPOutDistance)
	}
	return fmt.Sprintf("%s\n%.0f HP (%.1f%%)", status, s.HPMargin, s.HPMargin/s.StartHP*100)
}

// placementDownhillCell shows the time in downhill mode above its split around the start of the
// late race
func placementDownhillCell(row *placementRow) string {
	s := row.summary
	total := s.DownhillTimeEarly + s.DownhillTimeLate
	if total < placementMinimumModeTime {
		return placementEmptyCell
	}
	if s.DownhillTimeLate < placementMinimumModeTime {
		return fmt.Sprintf("%.1fs", total)
	}
	return fmt.Sprintf("%.1fs\n%.1fs + %.1fs late", total, s.DownhillTimeEarly, s.DownhillTimeLate)
}

// placementPaceCell shows the time in the modes that speed up above the time in Pace Down mode
func placementPaceCell(row *placementRow) string {
	var lines []string
	if row.summary.PaceUpTime >= placementMinimumModeTime {
		lines = append(lines, fmt.Sprintf("↑%.1fs", row.summary.PaceUpTime))
	}
	if row.summary.PaceDownTime >= placementMinimumModeTime {
		lines = append(lines, fmt.Sprintf("↓%.1fs", row.summary.PaceDownTime))
	}
	if len(lines) == 0 {
		return placementEmptyCell
	}
	return strings.Join(lines, "\n")
}
