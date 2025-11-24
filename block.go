package mathjax

import (
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

type mathJaxBlockParser struct {
}

var defaultMathJaxBlockParser = &mathJaxBlockParser{}

type mathBlockData struct {
	indent int
}

var mathBlockInfoKey = parser.NewContextKey()

func NewMathJaxBlockParser() parser.BlockParser {
	return defaultMathJaxBlockParser
}

func (b *mathJaxBlockParser) Open(parent ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	line, segment := reader.PeekLine()
	pos := pc.BlockOffset()
	if pos == -1 {
		return nil, parser.NoChildren
	}
	if pos >= len(line) || line[pos] != '$' {
		return nil, parser.NoChildren
	}

	// Count opening $$
	i := pos
	for ; i < len(line) && line[i] == '$'; i++ {
	}
	if i-pos < 2 {
		return nil, parser.NoChildren
	}

	remainingLine := line[i:]

	// Check if closing $$ exists on the same line
	// Look for at least 2 consecutive $ followed by blank/newline
	closingPos := -1
	for j := 0; j < len(remainingLine)-1; j++ {
		if remainingLine[j] == '$' {
			k := j
			for k < len(remainingLine) && remainingLine[k] == '$' {
				k++
			}
			closingLen := k - j
			if closingLen >= 2 && util.IsBlank(remainingLine[k:]) {
				// Found valid closing delimiter
				closingPos = j
				break
			}
			j = k - 1 // Skip the $ sequence we just checked
		}
	}

	if closingPos > 0 {
		// Same-line format: $$content$$
		node := NewMathBlock()
		content := remainingLine[:closingPos]
		if len(content) > 0 {
			// Add content to node (excluding opening and closing $$)
			contentSegment := text.NewSegment(segment.Start+i, segment.Start+i+closingPos)
			node.Lines().Append(contentSegment)
		}
		// Don't advance reader - goldmark will do it automatically
		// Return Close to indicate this block is complete
		return node, parser.Close
	}

	// Multi-line format: opening $$ on its own line or with content on first line
	pc.Set(mathBlockInfoKey, &mathBlockData{indent: pos})
	node := NewMathBlock()

	// If there's content after opening $$, save it as the first line
	if len(remainingLine) > 0 && !util.IsBlank(remainingLine) {
		contentSegment := text.NewSegment(segment.Start+i, segment.Stop)
		node.Lines().Append(contentSegment)
	}

	return node, parser.NoChildren
}

func (b *mathJaxBlockParser) Continue(node ast.Node, reader text.Reader, pc parser.Context) parser.State {
	line, segment := reader.PeekLine()

	// Safe get: check for nil before type assertion
	dataInterface := pc.Get(mathBlockInfoKey)
	if dataInterface == nil {
		// Context has been cleared, this block should already be closed
		return parser.Close
	}
	data := dataInterface.(*mathBlockData)

	// Check for closing $$ at the beginning of the line
	w, pos := util.IndentWidth(line, 0)
	if w < 4 {
		i := pos
		for ; i < len(line) && line[i] == '$'; i++ {
		}
		length := i - pos
		if length >= 2 && util.IsBlank(line[i:]) {
			reader.Advance(segment.Stop - segment.Start - segment.Padding)
			return parser.Close
		}
	}

	// Check for closing $$ anywhere in the line (for same-line ending format)
	// Search for $$ followed by blank/newline
	closingPos := -1
	for j := 0; j < len(line)-1; j++ {
		if line[j] == '$' {
			k := j
			for k < len(line) && line[k] == '$' {
				k++
			}
			closingLen := k - j
			if closingLen >= 2 && util.IsBlank(line[k:]) {
				// Found valid closing delimiter
				closingPos = j
				break
			}
			j = k - 1 // Skip the $ sequence we just checked
		}
	}

	if closingPos >= 0 {
		// Found closing $$ on this line - add content before $$ and close
		pos, padding := util.DedentPosition(line, 0, data.indent)
		if closingPos > pos {
			// Add content before the closing $$
			contentEnd := segment.Start + closingPos
			seg := text.NewSegmentPadding(segment.Start+pos, contentEnd, padding)
			node.Lines().Append(seg)
		}
		reader.Advance(segment.Stop - segment.Start - segment.Padding)
		return parser.Close
	}

	// No closing delimiter found - continue adding this line to the block
	pos, padding := util.DedentPosition(line, 0, data.indent)
	seg := text.NewSegmentPadding(segment.Start+pos, segment.Stop, padding)
	node.Lines().Append(seg)
	reader.AdvanceAndSetPadding(segment.Stop-segment.Start-pos-1, padding)
	return parser.Continue | parser.NoChildren
}

func (b *mathJaxBlockParser) Close(node ast.Node, reader text.Reader, pc parser.Context) {
	pc.Set(mathBlockInfoKey, nil)
}

func (b *mathJaxBlockParser) CanInterruptParagraph() bool {
	return true
}

func (b *mathJaxBlockParser) CanAcceptIndentedLine() bool {
	return false
}

func (b *mathJaxBlockParser) Trigger() []byte {
	return nil
}
