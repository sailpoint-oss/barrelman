package markdown

import (
	navigator "github.com/LukasParke/navigator"
)

// TranslatePosition converts a goldmark 1-based line number into an absolute
// range using the description's source location and geometry.
func TranslatePosition(desc navigator.DescriptionValue, goldmarkLine int) navigator.Range {
	absLine := desc.Loc.Range.Start.Line + uint32(desc.LineOffset) + uint32(goldmarkLine) - 1
	return navigator.Range{
		Start: navigator.Position{Line: absLine, Character: 0},
		End:   navigator.Position{Line: absLine, Character: 0},
	}
}

// TranslateRange converts a goldmark position with column and length into a
// precise absolute range. IndentCols is added to the column to account for
// block scalar indentation in the source file.
func TranslateRange(desc navigator.DescriptionValue, line, col, length int) navigator.Range {
	absLine := desc.Loc.Range.Start.Line + uint32(desc.LineOffset) + uint32(line) - 1
	absCol := uint32(col) + uint32(desc.IndentCols)
	return navigator.Range{
		Start: navigator.Position{Line: absLine, Character: absCol},
		End:   navigator.Position{Line: absLine, Character: absCol + uint32(length)},
	}
}
