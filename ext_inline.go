package adf

import (
	"bytes"
	"io"

	"github.com/ajbeck/goldmark-adf/v2/astext"
	"github.com/yuin/goldmark/v2/ast"
	"github.com/yuin/goldmark/v2/parser"
	"github.com/yuin/goldmark/v2/text"
)

// delimiterDecoder consumes only the escapes defined by the round-trip
// interchange syntax. Other Markdown escapes remain literal payload bytes.
type delimiterDecoder struct{ delimiters string }

func (d delimiterDecoder) Decode(input []byte) []byte {
	i := bytes.IndexByte(input, '\\')
	if i < 0 {
		return input
	}
	out := make([]byte, 0, len(input))
	out = append(out, input[:i]...)
	for i < len(input) {
		if input[i] == '\\' && i+1 < len(input) && bytes.IndexByte([]byte(d.delimiters), input[i+1]) >= 0 {
			i++
		}
		out = append(out, input[i])
		i++
	}
	return out
}

func (d delimiterDecoder) DecodeTo(w io.Writer, input []byte) (int, error) {
	return w.Write(d.Decode(input))
}

func valueAt(seg text.Segment, start, stop int, decoder text.Decoder) text.SingleLineValue {
	return text.NewSingleLineValueFromIndex(text.NewIndex(seg.Start+start-seg.Padding, seg.Start+stop-seg.Padding), decoder)
}

func findClosingBracket(s string) int {
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' {
			i++
			continue
		}
		if s[i] == ']' {
			return i
		}
	}
	return -1
}

func findUnescaped(s string, ch byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' {
			i++
			continue
		}
		if s[i] == ch {
			return i
		}
	}
	return -1
}

func findLastUnescapedPipe(s string) int {
	last := -1
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' {
			i++
			continue
		}
		if s[i] == '|' {
			last = i
		}
	}
	return last
}

type statusParser struct{}

func (*statusParser) Trigger() []byte { return []byte{'['} }

func (*statusParser) Parse(_ ast.Node, block text.Reader, _ parser.Context) ast.Node {
	line, seg := block.PeekLine()
	if len(line) < 10 || string(line[:8]) != "[status:" {
		return nil
	}
	rest := string(line[8:])
	close := findClosingBracket(rest)
	if close < 0 {
		return nil
	}
	pipe := findLastUnescapedPipe(rest[:close])
	if pipe <= 0 {
		return nil
	}
	color := rest[pipe+1 : close]
	switch color {
	case "neutral", "purple", "blue", "red", "yellow", "green":
	default:
		return nil
	}
	block.Advance(8 + close + 1)
	n := astext.NewStatus(valueAt(seg, 8, 8+pipe, delimiterDecoder{"|]"}), color)
	n.SetPos(seg.Start - seg.Padding)
	return n
}

type mentionParser struct{}

func (*mentionParser) Trigger() []byte { return []byte{'@'} }

func (*mentionParser) Parse(_ ast.Node, block text.Reader, _ parser.Context) ast.Node {
	line, seg := block.PeekLine()
	if len(line) < 5 || !bytes.HasPrefix(line, []byte("@[")) {
		return nil
	}
	rest := string(line[2:])
	closeBracket := findClosingBracket(rest)
	if closeBracket < 0 || len(rest) <= closeBracket+2 || rest[closeBracket+1] != '(' {
		return nil
	}
	closeParen := findUnescaped(rest[closeBracket+2:], ')')
	if closeParen <= 0 {
		return nil
	}
	idStart := 2 + closeBracket + 2
	total := idStart + closeParen + 1
	block.Advance(total)
	n := astext.NewMention(
		valueAt(seg, 2, 2+closeBracket, delimiterDecoder{"]"}),
		valueAt(seg, idStart, idStart+closeParen, delimiterDecoder{")"}),
	)
	n.SetPos(seg.Start - seg.Padding)
	return n
}

type dateParser struct{}

func (*dateParser) Trigger() []byte { return []byte{'['} }

func (*dateParser) Parse(_ ast.Node, block text.Reader, _ parser.Context) ast.Node {
	line, seg := block.PeekLine()
	if len(line) < 8 || string(line[:6]) != "[date:" {
		return nil
	}
	close := 6
	for close < len(line) && line[close] != ']' {
		if line[close] < '0' || line[close] > '9' {
			return nil
		}
		close++
	}
	if close == 6 || close == len(line) {
		return nil
	}
	block.Advance(close + 1)
	n := astext.NewDate(valueAt(seg, 6, close, text.IdentityDecoder))
	n.SetPos(seg.Start - seg.Padding)
	return n
}

type placeholderParser struct{}

func (*placeholderParser) Trigger() []byte { return []byte{'{'} }

func (*placeholderParser) Parse(_ ast.Node, block text.Reader, _ parser.Context) ast.Node {
	line, seg := block.PeekLine()
	if len(line) < 5 || !bytes.HasPrefix(line, []byte("{{")) {
		return nil
	}
	for i := 2; i < len(line)-1; i++ {
		if line[i] == '\\' {
			i++
			continue
		}
		if line[i] == '}' && line[i+1] == '}' {
			block.Advance(i + 2)
			n := astext.NewPlaceholder(valueAt(seg, 2, i, delimiterDecoder{"}"}))
			n.SetPos(seg.Start - seg.Padding)
			return n
		}
	}
	return nil
}

type cardParser struct{}

func (*cardParser) Trigger() []byte { return []byte{'['} }

func (*cardParser) Parse(_ ast.Node, block text.Reader, _ parser.Context) ast.Node {
	line, seg := block.PeekLine()
	if len(line) < 8 || string(line[:6]) != "[card:" {
		return nil
	}
	close := findClosingBracket(string(line[6:]))
	if close < 1 {
		return nil
	}
	block.Advance(6 + close + 1)
	n := astext.NewCard(valueAt(seg, 6, 6+close, delimiterDecoder{"]"}))
	n.SetPos(seg.Start - seg.Padding)
	return n
}

type embedParser struct{}

func (*embedParser) Trigger() []byte { return []byte{'['} }

func (*embedParser) Parse(_ ast.Node, block text.Reader, _ parser.Context) ast.Node {
	line, seg := block.PeekLine()
	if len(line) < 9 || string(line[:7]) != "[embed:" {
		return nil
	}
	close := findClosingBracket(string(line[7:]))
	if close < 1 {
		return nil
	}
	block.Advance(7 + close + 1)
	n := astext.NewEmbed(valueAt(seg, 7, 7+close, delimiterDecoder{"]"}))
	n.SetPos(seg.Start - seg.Padding)
	return n
}

type emojiParser struct{}

func (*emojiParser) Trigger() []byte { return []byte{':'} }

func (*emojiParser) Parse(_ ast.Node, block text.Reader, _ parser.Context) ast.Node {
	line, seg := block.PeekLine()
	if len(line) < 3 || line[0] != ':' {
		return nil
	}
	for i := 1; i < len(line); i++ {
		if line[i] == ':' {
			if i == 1 {
				return nil
			}
			block.Advance(i + 1)
			n := astext.NewEmoji(valueAt(seg, 0, i+1, text.IdentityDecoder))
			n.SetPos(seg.Start - seg.Padding)
			return n
		}
		if !isShortcodeChar(line[i]) {
			return nil
		}
	}
	return nil
}

func isShortcodeChar(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '_' || c == '-' || c == '+'
}
