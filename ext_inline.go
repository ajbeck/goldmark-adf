//go:build goexperiment.jsonv2

package adf

import (
	"strings"

	"github.com/ajbeck/goldmark-adf/astext"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

// unescapeDelimiters reverses backslash escaping applied by adf-to-markdown.
func unescapeDelimiters(s string, delims string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+1 < len(s) && strings.IndexByte(delims, s[i+1]) >= 0 {
			i++
		}
		b.WriteByte(s[i])
	}
	return b.String()
}

// findClosingBracket finds the index of the unescaped closing bracket ']' in s.
// Returns -1 if not found.
func findClosingBracket(s string) int {
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' {
			i++ // skip escaped char
			continue
		}
		if s[i] == ']' {
			return i
		}
	}
	return -1
}

// findUnescaped finds the index of the first unescaped occurrence of ch in s.
// Returns -1 if not found.
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

// --- Status parser: [status:text|color] ---

type statusParser struct{}

func (p *statusParser) Trigger() []byte { return []byte{'['} }

func (p *statusParser) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
	line, seg := block.PeekLine()
	if len(line) < 10 || string(line[:8]) != "[status:" {
		return nil
	}

	rest := string(line[8:])
	closeBracket := findClosingBracket(rest)
	if closeBracket < 0 {
		return nil
	}

	inner := rest[:closeBracket]
	pipeIdx := findLastUnescapedPipe(inner)
	if pipeIdx < 0 {
		return nil
	}

	statusText := unescapeDelimiters(inner[:pipeIdx], "|]")
	color := inner[pipeIdx+1:]

	// Validate color is a known ADF status color
	switch color {
	case "neutral", "purple", "blue", "red", "yellow", "green":
	default:
		return nil
	}

	block.Advance(8 + closeBracket + 1) // [status: + inner + ]
	_ = seg
	return astext.NewStatus(statusText, color)
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

// --- Mention parser: @[name](id) ---

type mentionParser struct{}

func (p *mentionParser) Trigger() []byte { return []byte{'@'} }

func (p *mentionParser) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
	line, _ := block.PeekLine()
	if len(line) < 5 || line[0] != '@' || line[1] != '[' {
		return nil
	}

	rest := string(line[2:])
	closeBracket := findClosingBracket(rest)
	if closeBracket < 0 {
		return nil
	}

	displayName := unescapeDelimiters(rest[:closeBracket], "]")

	after := rest[closeBracket+1:]
	if len(after) == 0 || after[0] != '(' {
		return nil
	}

	closeParen := findUnescaped(after[1:], ')')
	if closeParen < 0 {
		return nil
	}

	id := unescapeDelimiters(after[1:1+closeParen], ")")

	// Total consumed: @ + [ + name + ] + ( + id + )
	total := 2 + closeBracket + 1 + 1 + closeParen + 1
	block.Advance(total)
	return astext.NewMention(displayName, id)
}

// --- Date parser: [date:digits] ---

type dateParser struct{}

func (p *dateParser) Trigger() []byte { return []byte{'['} }

func (p *dateParser) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
	line, _ := block.PeekLine()
	if len(line) < 8 || string(line[:6]) != "[date:" {
		return nil
	}

	rest := line[6:]
	closeIdx := -1
	for i := 0; i < len(rest); i++ {
		if rest[i] == ']' {
			closeIdx = i
			break
		}
		if rest[i] < '0' || rest[i] > '9' {
			return nil
		}
	}
	if closeIdx < 1 {
		return nil
	}

	timestamp := string(rest[:closeIdx])
	block.Advance(6 + closeIdx + 1)
	return astext.NewDate(timestamp)
}

// --- Placeholder parser: {{text}} ---

type placeholderParser struct{}

func (p *placeholderParser) Trigger() []byte { return []byte{'{'} }

func (p *placeholderParser) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
	line, _ := block.PeekLine()
	if len(line) < 5 || line[0] != '{' || line[1] != '{' {
		return nil
	}

	rest := string(line[2:])
	// Find unescaped }}
	for i := 0; i < len(rest)-1; i++ {
		if rest[i] == '\\' {
			i++
			continue
		}
		if rest[i] == '}' && rest[i+1] == '}' {
			text := unescapeDelimiters(rest[:i], "}")
			block.Advance(2 + i + 2) // {{ + content + }}
			return astext.NewPlaceholder(text)
		}
	}
	return nil
}

// --- Card parser: [card:url] ---

type cardParser struct{}

func (p *cardParser) Trigger() []byte { return []byte{'['} }

func (p *cardParser) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
	line, _ := block.PeekLine()
	if len(line) < 8 || string(line[:6]) != "[card:" {
		return nil
	}

	rest := string(line[6:])
	closeBracket := findClosingBracket(rest)
	if closeBracket < 1 {
		return nil
	}

	url := unescapeDelimiters(rest[:closeBracket], "]")
	block.Advance(6 + closeBracket + 1)
	return astext.NewCard(url)
}

// --- Embed parser: [embed:url] ---

type embedParser struct{}

func (p *embedParser) Trigger() []byte { return []byte{'['} }

func (p *embedParser) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
	line, _ := block.PeekLine()
	if len(line) < 9 || string(line[:7]) != "[embed:" {
		return nil
	}

	rest := string(line[7:])
	closeBracket := findClosingBracket(rest)
	if closeBracket < 1 {
		return nil
	}

	url := unescapeDelimiters(rest[:closeBracket], "]")
	block.Advance(7 + closeBracket + 1)
	return astext.NewEmbed(url)
}

// --- Emoji parser: :shortcode: ---

type emojiParser struct{}

func (p *emojiParser) Trigger() []byte { return []byte{':'} }

func (p *emojiParser) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
	line, _ := block.PeekLine()
	if len(line) < 3 || line[0] != ':' {
		return nil
	}

	// Emoji shortcodes are alphanumeric with underscores and hyphens
	rest := line[1:]
	end := -1
	for i := 0; i < len(rest); i++ {
		c := rest[i]
		if c == ':' {
			end = i
			break
		}
		if !isShortcodeChar(c) {
			return nil
		}
	}
	if end < 1 {
		return nil
	}

	shortName := ":" + string(rest[:end]) + ":"
	block.Advance(end + 2) // : + content + :
	return astext.NewEmoji(shortName)
}

func isShortcodeChar(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' || c == '-' || c == '+'
}
