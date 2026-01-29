//go:build goexperiment.jsonv2

package adf

import (
	"github.com/yuin/goldmark/renderer"
)

// Config holds configuration options for the ADF renderer.
type Config struct {
	// ImageHandler specifies how to handle image nodes.
	// By default, images are converted to links.
	ImageHandler ImageHandler

	// TableLayout specifies the default table layout.
	// Valid values: "default", "center", "wide", "full-width"
	TableLayout string

	// ExternalMedia enables external media image handling.
	// When true, images are rendered as mediaSingle nodes with external media.
	// When false (default), images are converted to text with link marks.
	ExternalMedia bool

	// ImageLayout specifies the default layout for mediaSingle nodes.
	// Valid values: "center", "wide", "full-width", "wrap-left", "wrap-right", "align-start", "align-end"
	// Defaults to "center" if not specified.
	ImageLayout string
}

// ImageHandler is a function that handles image rendering.
type ImageHandler func(dest, alt, title string) *Node

// NewConfig creates a new Config with default values.
func NewConfig() Config {
	return Config{
		TableLayout: "default",
		ImageLayout: "center",
	}
}

// Option is a functional option for configuring the ADF renderer.
type Option interface {
	SetADFOption(*Config)
}

// withImageHandler implements Option.
type withImageHandler struct {
	handler ImageHandler
}

func (o *withImageHandler) SetADFOption(c *Config) {
	c.ImageHandler = o.handler
}

func (o *withImageHandler) SetConfig(c *renderer.Config) {
	// No-op for renderer.Config
}

// WithImageHandler sets a custom image handler.
func WithImageHandler(handler ImageHandler) Option {
	return &withImageHandler{handler: handler}
}

// withTableLayout implements Option.
type withTableLayout struct {
	layout string
}

func (o *withTableLayout) SetADFOption(c *Config) {
	c.TableLayout = o.layout
}

func (o *withTableLayout) SetConfig(c *renderer.Config) {
	// No-op for renderer.Config
}

// WithTableLayout sets the default table layout.
// Valid values: "default", "center", "wide", "full-width"
func WithTableLayout(layout string) Option {
	return &withTableLayout{layout: layout}
}

// withExternalMedia implements Option.
type withExternalMedia struct {
	enabled bool
}

func (o *withExternalMedia) SetADFOption(c *Config) {
	c.ExternalMedia = o.enabled
}

func (o *withExternalMedia) SetConfig(c *renderer.Config) {
	// No-op for renderer.Config
}

// WithExternalMedia enables or disables external media image handling.
// When enabled, images are rendered as mediaSingle nodes containing external media.
// When disabled (default), images are converted to text with link marks.
func WithExternalMedia(enabled bool) Option {
	return &withExternalMedia{enabled: enabled}
}

// withImageLayout implements Option.
type withImageLayout struct {
	layout string
}

func (o *withImageLayout) SetADFOption(c *Config) {
	c.ImageLayout = o.layout
}

func (o *withImageLayout) SetConfig(c *renderer.Config) {
	// No-op for renderer.Config
}

// WithImageLayout sets the default layout for mediaSingle nodes.
// Valid values: "center", "wide", "full-width", "wrap-left", "wrap-right", "align-start", "align-end"
// Defaults to "center" if not specified.
func WithImageLayout(layout string) Option {
	return &withImageLayout{layout: layout}
}
