package adf

import (
	"bytes"
	"io"

	"github.com/yuin/goldmark/v2/parser"
	gmrenderer "github.com/yuin/goldmark/v2/renderer"
	"github.com/yuin/goldmark/v2/util"
)

// Markdown combines an ADF parser profile and renderer. It is safe for
// concurrent use provided any configured ImageHandler is safe for concurrent
// calls.
type Markdown struct {
	newParser func() parser.Parser
	renderer  gmrenderer.Renderer[io.Writer]
}

// New creates a reusable CommonMark-to-ADF converter.
func New(opts ...Option) *Markdown {
	return &Markdown{
		newParser: NewParser,
		renderer:  NewRenderer(opts...),
	}
}

// NewWithGFM creates a reusable GFM and ADF-round-trip-to-ADF converter.
func NewWithGFM(opts ...Option) *Markdown {
	return &Markdown{
		newParser: NewWithGFMParser,
		renderer:  NewRenderer(opts...),
	}
}

// Convert parses source and writes its ADF JSON representation to w.
func (m *Markdown) Convert(source []byte, w io.Writer) error {
	doc := m.newParser().Parse(source)
	return m.renderer.Render(w, source, doc)
}

// NewParser creates a CommonMark parser with no GFM or ADF-round-trip
// extensions.
func NewParser() parser.Parser {
	return parser.New()
}

// NewWithGFMParser creates a parser with GFM and ADF-round-trip extensions.
func NewWithGFMParser() parser.Parser {
	return parser.New(parser.WithExtensions(newSafeGFMParser(), NewRoundTripParser()))
}

// RoundTripParser is the parser extension for goldmark-adf's custom ADF
// Markdown interchange syntax.
var RoundTripParser parser.Extension = NewRoundTripParser()

// NewRoundTripParser returns a fresh parser extension for goldmark-adf's
// custom ADF Markdown interchange syntax.
func NewRoundTripParser() parser.Extension {
	return &roundTripParser{}
}

type roundTripParser struct{}

func (*roundTripParser) ParserOptions(*parser.Config) []parser.Option {
	return []parser.Option{
		parser.WithInlineParsers(
			util.Prioritized[parser.InlineParser](&statusParser{}, 90),
			util.Prioritized[parser.InlineParser](&mentionParser{}, 91),
			util.Prioritized[parser.InlineParser](&dateParser{}, 92),
			util.Prioritized[parser.InlineParser](&placeholderParser{}, 93),
			util.Prioritized[parser.InlineParser](&cardParser{}, 94),
			util.Prioritized[parser.InlineParser](&embedParser{}, 95),
			util.Prioritized[parser.InlineParser](&emojiParser{}, 96),
		),
		parser.WithASTTransformers(
			util.Prioritized[parser.ASTTransformer](&panelTransformer{}, 100),
			util.Prioritized[parser.ASTTransformer](&decisionTransformer{}, 101),
		),
	}
}

// Convert converts CommonMark source to ADF JSON.
func Convert(source []byte) ([]byte, error) {
	var output bytes.Buffer
	if err := New().Convert(source, &output); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}

// ConvertWithGFM converts GFM and ADF-round-trip source to ADF JSON.
func ConvertWithGFM(source []byte) ([]byte, error) {
	var output bytes.Buffer
	if err := NewWithGFM().Convert(source, &output); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}
