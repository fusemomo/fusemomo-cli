package output

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/fusemomo/fusemomo-cli/internal/api"
)

// Table renders a human-readable table to stdout using lipgloss.
// Falls back to JSON automatically when stdout is not a TTY.
func Table(format string, v interface{}) error {
	if !IsTTY(os.Stdout) {
		return JSON(v)
	}
	switch format {
	case "entity":
		return renderEntityTable(v)
	case "entity-list":
		return renderEntityListTable(v)
	case "interaction":
		return renderInteractionTable(v)
	case "interaction-batch":
		return renderBatchTable(v)
	case "recommend":
		return renderRecommendTable(v)
	case "recommend-outcome":
		return renderOutcomeTable(v)
	default:
		return JSON(v)
	}
}

//  Styles

var (
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39")).
			Padding(0, 1)

	cellStyle = lipgloss.NewStyle().
			Padding(0, 1)
)

func renderRow(cells ...string) string {
	parts := make([]string, len(cells))
	for i, c := range cells {
		parts[i] = cellStyle.Render(c)
	}
	return strings.Join(parts, " │ ")
}

func renderHeader(cols ...string) string {
	parts := make([]string, len(cols))
	for i, c := range cols {
		parts[i] = headerStyle.Render(c)
	}
	return strings.Join(parts, " │ ")
}

func ptrStr(s *string) string {
	if s == nil {
		return "—"
	}
	return *s
}

func ptrFloat(f *float64) string {
	if f == nil {
		return "—"
	}
	return fmt.Sprintf("%.2f", *f)
}

func shortID(id string) string {
	if len(id) > 8 {
		return id[:8] + "..."
	}
	return id
}

//  Per-command table renderers

func renderEntityTable(v interface{}) error {
	header := renderHeader("ENTITY ID", "DISPLAY NAME", "TYPE", "INTERACTIONS", "SUCCESS RATE", "BEHAVIORAL SCORE")
	fmt.Println(header)
	fmt.Println(strings.Repeat("", 90))

	switch e := v.(type) {
	case *api.ResolveEntityResponse:
		rate := "—"
		if e.TotalInteractions > 0 {
			rate = fmt.Sprintf("%.1f%%", float64(e.SuccessfulInteractions)/float64(e.TotalInteractions)*100)
		}
		fmt.Println(renderRow(
			shortID(e.EntityID),
			ptrStr(e.DisplayName),
			ptrStr(e.EntityType),
			fmt.Sprintf("%d", e.TotalInteractions),
			rate,
			ptrFloat(e.BehavioralScore),
		))
	case *api.EntityDetailResponse:
		rate := "—"
		if e.TotalInteractions > 0 {
			rate = fmt.Sprintf("%.1f%%", float64(e.SuccessfulInteractions)/float64(e.TotalInteractions)*100)
		}
		fmt.Println(renderRow(
			shortID(e.ID),
			e.DisplayName,
			e.EntityType,
			fmt.Sprintf("%d", e.TotalInteractions),
			rate,
			ptrFloat(e.BehavioralScore),
		))
	}
	return nil
}

func renderEntityListTable(v interface{}) error {
	header := renderHeader("ENTITY ID", "DISPLAY NAME", "TYPE", "TOTAL", "LAST INTERACTION")
	fmt.Println(header)
	fmt.Println(strings.Repeat("", 90))

	if list, ok := v.(*api.EntitiesListResponse); ok {
		for _, e := range list.Entities {
			last := "—"
			if e.LastInteractionAt != nil {
				last = e.LastInteractionAt.Format("2006-01-02 15:04")
			}
			fmt.Println(renderRow(
				shortID(e.ID),
				e.DisplayName,
				e.EntityType,
				fmt.Sprintf("%d", e.TotalInteractions),
				last,
			))
		}
		fmt.Printf("\n  Total: %d | Showing %d–%d\n", list.Total, list.Offset+1, list.Offset+len(list.Entities))
	}
	return nil
}

func renderInteractionTable(v interface{}) error {
	header := renderHeader("INTERACTION ID", "ENTITY ID", "LOGGED AT")
	fmt.Println(header)
	fmt.Println(strings.Repeat("", 70))
	if r, ok := v.(*api.InteractionLogResponse); ok {
		fmt.Println(renderRow(shortID(r.InteractionID), shortID(r.EntityID), r.LoggedAt.Format(time.RFC3339)))
	}
	return nil
}

func renderBatchTable(v interface{}) error {
	header := renderHeader("LOGGED", "FIRST ID", "LAST ID", "LOGGED AT")
	fmt.Println(header)
	fmt.Println(strings.Repeat("", 70))
	if r, ok := v.(*api.BatchInteractionLogResponse); ok {
		fmt.Println(renderRow(
			fmt.Sprintf("%d", r.LoggedCount),
			shortID(r.FirstID),
			shortID(r.LastID),
			r.LoggedAt.Format(time.RFC3339),
		))
	}
	return nil
}

func renderRecommendTable(v interface{}) error {
	header := renderHeader("REC ID", "ACTION TYPE", "CONFIDENCE", "REASON", "SAMPLE SIZE")
	fmt.Println(header)
	fmt.Println(strings.Repeat("", 90))
	if r, ok := v.(*api.RecommendResponse); ok {
		fmt.Println(renderRow(
			ptrStr(r.RecommendationID),
			ptrStr(r.RecommendedActionType),
			ptrFloat(r.Confidence),
			r.Reason,
			fmt.Sprintf("%d", r.SampleSize),
		))
	}
	return nil
}

func renderOutcomeTable(v interface{}) error {
	header := renderHeader("REC ID", "FOLLOWED", "OUTCOME", "UPDATED AT")
	fmt.Println(header)
	fmt.Println(strings.Repeat("", 70))
	if r, ok := v.(*api.RecommendOutcomeResponse); ok {
		fmt.Println(renderRow(
			shortID(r.RecommendationID),
			fmt.Sprintf("%v", r.WasFollowed),
			ptrStr(r.Outcome),
			r.UpdatedAt.Format(time.RFC3339),
		))
	}
	return nil
}

// Print outputs to stdout using the specified format ("json" or "table").
// Falls back to JSON when not TTY or format is not "table".
func Print(format string, tableKey string, v interface{}, noColor bool) error {
	if noColor {
		lipgloss.SetHasDarkBackground(false)
	}
	if format == "table" {
		return Table(tableKey, v)
	}
	return JSON(v)
}
