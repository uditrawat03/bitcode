package editor

import (
	"regexp"

	"github.com/gdamore/tcell/v2"
)

var (
	keywordRegex = regexp.MustCompile(`\b(func|var|if|else|for|return|package|import|echo)\b`)
	stringRegex  = regexp.MustCompile(`".*?"`)
	commentRegex = regexp.MustCompile(`//.*`)
	numberRegex  = regexp.MustCompile(`\b\d+(\.\d+)?\b`)
)

func highlightLine(line string) []struct {
	ch    rune
	style tcell.Style
} {
	result := make([]struct {
		ch    rune
		style tcell.Style
	}, len([]rune(line)))

	// Default style
	for i, r := range line {
		result[i].ch = r
		result[i].style = tcell.StyleDefault.Foreground(tcell.ColorWhite)
	}

	applyRegex := func(re *regexp.Regexp, style tcell.Style) {
		matches := re.FindAllStringIndex(line, -1)
		for _, m := range matches {
			for i := m[0]; i < m[1]; i++ {
				result[i].style = style
			}
		}
	}

	applyRegex(keywordRegex, getTokenStyle("keyword"))
	applyRegex(stringRegex, getTokenStyle("string"))
	applyRegex(commentRegex, getTokenStyle("comment"))
	applyRegex(numberRegex, getTokenStyle("number"))

	return result
}

func GetHighlightLine(line string) []struct {
	Ch    rune
	Style tcell.Style
} {
	result := make([]struct {
		Ch    rune
		Style tcell.Style
	}, len([]rune(line)))

	// Default style
	for i, r := range line {
		result[i].Ch = r
		result[i].Style = tcell.StyleDefault.Foreground(tcell.ColorWhite)
	}

	applyRegex := func(re *regexp.Regexp, style tcell.Style) {
		matches := re.FindAllStringIndex(line, -1)
		for _, m := range matches {
			for i := m[0]; i < m[1]; i++ {
				result[i].Style = style
			}
		}
	}

	applyRegex(keywordRegex, getTokenStyle("keyword"))
	applyRegex(stringRegex, getTokenStyle("string"))
	applyRegex(commentRegex, getTokenStyle("comment"))
	applyRegex(numberRegex, getTokenStyle("number"))

	return result
}

func getTokenStyle(tokenType string) tcell.Style {
	switch tokenType {
	case "keyword":
		return tcell.StyleDefault.Foreground(tcell.ColorOrange)
	case "string":
		return tcell.StyleDefault.Foreground(tcell.ColorGreen)
	case "comment":
		return tcell.StyleDefault.Foreground(tcell.ColorGray)
	case "number":
		return tcell.StyleDefault.Foreground(tcell.ColorPurple)
	case "function":
		return tcell.StyleDefault.Foreground(tcell.ColorBlue)
	default:
		return tcell.StyleDefault.Foreground(tcell.ColorWhite)
	}
}
