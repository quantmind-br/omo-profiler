package diff

import (
	"strings"

	dmp "github.com/sergi/go-diff/diffmatchpatch"
)

// DiffType represents the type of difference for a line
type DiffType int

const (
	DiffEqual DiffType = iota
	DiffAdded
	DiffRemoved
)

// DiffLine represents a single line in the diff result
type DiffLine struct {
	Text    string
	Type    DiffType
	LineNum int
}

// DiffResult contains the side-by-side diff representation
type DiffResult struct {
	Left  []DiffLine
	Right []DiffLine
}

// ComputeDiff computes a line-based diff between two JSON byte slices
func ComputeDiff(json1, json2 []byte) (*DiffResult, error) {
	differ := dmp.New()

	text1 := string(json1)
	text2 := string(json2)

	chars1, chars2, lineArray := differ.DiffLinesToChars(text1, text2)
	diffs := differ.DiffMain(chars1, chars2, false)
	diffs = differ.DiffCharsToLines(diffs, lineArray)
	diffs = differ.DiffCleanupSemantic(diffs)

	return buildDiffResult(diffs), nil
}

// buildDiffResult converts dmp.Diff slices to DiffResult with side-by-side representation
func buildDiffResult(diffs []dmp.Diff) *DiffResult {
	result := &DiffResult{
		Left:  make([]DiffLine, 0),
		Right: make([]DiffLine, 0),
	}

	leftLineNum := 1
	rightLineNum := 1

	for _, d := range diffs {
		lines := splitLines(d.Text)

		switch d.Type {
		case dmp.DiffEqual:
			for _, line := range lines {
				result.Left = append(result.Left, DiffLine{
					Text:    line,
					Type:    DiffEqual,
					LineNum: leftLineNum,
				})
				result.Right = append(result.Right, DiffLine{
					Text:    line,
					Type:    DiffEqual,
					LineNum: rightLineNum,
				})
				leftLineNum++
				rightLineNum++
			}
		case dmp.DiffDelete:
			for _, line := range lines {
				result.Left = append(result.Left, DiffLine{
					Text:    line,
					Type:    DiffRemoved,
					LineNum: leftLineNum,
				})
				result.Right = append(result.Right, DiffLine{
					Text:    "",
					Type:    DiffRemoved,
					LineNum: 0,
				})
				leftLineNum++
			}
		case dmp.DiffInsert:
			for _, line := range lines {
				result.Left = append(result.Left, DiffLine{
					Text:    "",
					Type:    DiffAdded,
					LineNum: 0,
				})
				result.Right = append(result.Right, DiffLine{
					Text:    line,
					Type:    DiffAdded,
					LineNum: rightLineNum,
				})
				rightLineNum++
			}
		}
	}

	return result
}

// splitLines splits text into lines, handling trailing newlines properly
func splitLines(text string) []string {
	if text == "" {
		return nil
	}

	lines := strings.Split(text, "\n")

	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	return lines
}
