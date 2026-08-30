package adf

import (
	"fmt"
	"io"

	"github.com/yuin/goldmark/v2/renderer"
)

// TableLayout controls the layout of generated ADF tables.
type TableLayout string

const (
	// TableLayoutDefault uses the standard table width.
	TableLayoutDefault TableLayout = "default"
	// TableLayoutCenter centers the table.
	TableLayoutCenter TableLayout = "center"
	// TableLayoutWide extends the table into the page margins.
	TableLayoutWide TableLayout = "wide"
	// TableLayoutFullWidth makes the table full width.
	TableLayoutFullWidth TableLayout = "full-width"
	// TableLayoutAlignStart aligns the table to the logical start edge.
	TableLayoutAlignStart TableLayout = "align-start"
	// TableLayoutAlignEnd aligns the table to the logical end edge.
	TableLayoutAlignEnd TableLayout = "align-end"
)

// ImageLayout controls the layout of block-level image results.
type ImageLayout string

const (
	// ImageLayoutCenter centers the image.
	ImageLayoutCenter ImageLayout = "center"
	// ImageLayoutWide extends the image into the page margins.
	ImageLayoutWide ImageLayout = "wide"
	// ImageLayoutFullWidth makes the image full width.
	ImageLayoutFullWidth ImageLayout = "full-width"
	// ImageLayoutWrapLeft floats the image to the logical left.
	ImageLayoutWrapLeft ImageLayout = "wrap-left"
	// ImageLayoutWrapRight floats the image to the logical right.
	ImageLayoutWrapRight ImageLayout = "wrap-right"
	// ImageLayoutAlignStart aligns the image to the logical start edge.
	ImageLayoutAlignStart ImageLayout = "align-start"
	// ImageLayoutAlignEnd aligns the image to the logical end edge.
	ImageLayoutAlignEnd ImageLayout = "align-end"
)

// Image is the source information provided to an ImageHandler.
type Image struct {
	Destination string
	Alt         string
	Title       string
}

// ImagePlacement controls where an ImageHandler result is inserted.
type ImagePlacement uint8

const (
	// ImageInline inserts the node into the current inline container.
	ImageInline ImagePlacement = iota
	// ImageBlock inserts the node as a block, splitting a surrounding paragraph.
	ImageBlock
)

// ImageResult is the ADF node produced by an ImageHandler.
type ImageResult struct {
	Node      Node
	Placement ImagePlacement
}

// ImageHandler overrides the built-in image conversion behavior. It must be
// safe for concurrent calls when used with a concurrently reused Markdown or
// Renderer instance. The returned node becomes renderer-owned after return.
type ImageHandler func(Image) (ImageResult, error)

// Config holds configuration options for the ADF renderer.
type Config struct {
	renderer.Config[io.Writer, Config]

	ImageHandler  ImageHandler
	TableLayout   TableLayout
	ExternalMedia bool
	ImageLayout   ImageLayout
}

// Default returns a Config with default values.
func (Config) Default() Config {
	return Config{
		TableLayout: TableLayoutDefault,
		ImageLayout: ImageLayoutCenter,
	}
}

// NewConfig creates a new Config with default values.
func NewConfig() Config {
	return Config{}.Default()
}

func (c Config) validate() error {
	if !validTableLayout(c.TableLayout) {
		return fmt.Errorf("adf: invalid table layout %q", c.TableLayout)
	}
	if !validImageLayout(c.ImageLayout) {
		return fmt.Errorf("adf: invalid image layout %q", c.ImageLayout)
	}
	return nil
}

func validTableLayout(layout TableLayout) bool {
	switch layout {
	case TableLayoutDefault, TableLayoutCenter, TableLayoutWide, TableLayoutFullWidth, TableLayoutAlignStart, TableLayoutAlignEnd:
		return true
	default:
		return false
	}
}

func validImageLayout(layout ImageLayout) bool {
	switch layout {
	case ImageLayoutCenter, ImageLayoutWide, ImageLayoutFullWidth, ImageLayoutWrapLeft, ImageLayoutWrapRight, ImageLayoutAlignStart, ImageLayoutAlignEnd:
		return true
	default:
		return false
	}
}

// Option configures an ADF Renderer.
type Option interface {
	renderer.Option[Config]
}

type optionFunc func(*Config)

func (f optionFunc) SetFormatOption(c *Config) { f(c) }

// WithImageHandler sets a custom image handler. It takes precedence over
// WithExternalMedia.
func WithImageHandler(handler ImageHandler) Option {
	return optionFunc(func(c *Config) { c.ImageHandler = handler })
}

// WithTableLayout sets the default layout for generated tables.
func WithTableLayout(layout TableLayout) Option {
	return optionFunc(func(c *Config) { c.TableLayout = layout })
}

// WithExternalMedia enables or disables external-media image handling. When
// disabled, images are rendered as linked text unless an ImageHandler is set.
func WithExternalMedia(enabled bool) Option {
	return optionFunc(func(c *Config) { c.ExternalMedia = enabled })
}

// WithImageLayout sets the default layout for generated mediaSingle nodes.
func WithImageLayout(layout ImageLayout) Option {
	return optionFunc(func(c *Config) { c.ImageLayout = layout })
}
