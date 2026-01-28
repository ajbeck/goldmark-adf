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
}

// ImageHandler is a function that handles image rendering.
type ImageHandler func(dest, alt, title string) *Node

// NewConfig creates a new Config with default values.
func NewConfig() Config {
	return Config{
		TableLayout: "default",
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
