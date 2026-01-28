---
name: go-doc
description: Research Go package and symbol documentation. Use when exploring Go codebases, understanding APIs, or looking up how functions/types/methods work.
---

# Go Documentation Research

Use this skill when you need to understand Go packages, types, functions, or methods. Combine local `go doc` lookups with pkg.go.dev and web searches for comprehensive understanding.

## When to Use

- Exploring unfamiliar Go packages or standard library
- Looking up function/type/method signatures and documentation
- Understanding what symbols a package exports
- Checking unexported internals (use `-u` flag)
- Viewing implementation source code (use `-src` flag)
- Discovering methods available on a type

## Command Syntax

```bash
go doc                          # docs for current package
go doc <pkg>                    # package docs (e.g., encoding/json)
go doc <sym>                    # symbol in current package (e.g., Handler)
go doc <pkg>.<sym>              # symbol in package (e.g., json.Decoder)
go doc <pkg>.<type>.<method>    # method docs (e.g., json.Decoder.Decode)
```

### Package Path Shortcuts

Go doc accepts partial package paths:

```bash
go doc json                     # shorthand for encoding/json
go doc template.new             # html/template.New (lexically first match)
go doc text/template.new        # explicit: text/template.New
```

### Case Sensitivity

- Lowercase arguments match either case
- Uppercase arguments match exactly
- Use `-c` flag to force case-sensitive matching

## Key Flags

| Flag | Purpose |
|------|---------|
| `-all` | Show all documentation for the package (verbose) |
| `-short` | One-line representation for each symbol (overview) |
| `-src` | Show source code for the symbol |
| `-u` | Include unexported symbols and methods |
| `-c` | Case-sensitive symbol matching |
| `-cmd` | Show symbols even for command packages (package main) |

## Common Research Patterns

### 1. Package Overview

Get a quick overview of what a package provides:

```bash
go doc -short encoding/json     # one-line summary of each symbol
go doc -all encoding/json       # complete documentation
```

### 2. Type and Methods

Understand a type and what you can do with it:

```bash
go doc json.Decoder             # type docs + method summary
go doc json.Decoder.Decode      # specific method documentation
go doc -all json.Decoder        # type with all method docs
```

### 3. Interface Discovery

See what methods an interface requires:

```bash
go doc io.Reader                # interface definition
go doc io.ReadWriter            # composed interfaces
```

### 4. Unexported Internals

When you need to understand internal implementation:

```bash
go doc -u -src <pkg>.<symbol>   # unexported symbols with source
```

### 5. Current Package Exploration

When working in a Go project:

```bash
go doc                          # current package summary
go doc -all                     # full current package docs
go doc MyType                   # local type (starts with capital)
```

## Complementary Research

### pkg.go.dev

For deeper understanding, check https://pkg.go.dev/<import-path>:

- **Examples** - Runnable code examples not shown by `go doc`
- **Source browsing** - Navigate implementation with cross-references
- **Import graph** - See dependencies and dependents
- **Version history** - Track API changes across versions
- **License info** - Check licensing for third-party packages

Example: For `encoding/json`, visit https://pkg.go.dev/encoding/json

### Web Searches

Complement local docs with searches for:

- Tutorials and usage patterns: `"golang json.Decoder" tutorial`
- Common issues: `"golang json.Decoder" gotchas`
- Best practices: `"golang encoding/json" best practices`
- Related packages: `"golang json" alternatives`

## Examples

### Standard Library Research

```bash
# HTTP server setup
go doc net/http
go doc http.Server
go doc http.Handler
go doc http.HandlerFunc

# Context usage
go doc context
go doc context.WithCancel
go doc context.Context

# Error handling
go doc errors
go doc errors.Is
go doc errors.As
```

### Understanding Interfaces

```bash
# What does io.Reader require?
go doc io.Reader

# What about io.ReadCloser?
go doc io.ReadCloser

# Find types that commonly implement Reader
# (use web search: "golang io.Reader implementations")
```

### Third-Party Packages

```bash
# If installed in your module
go doc github.com/some/package
go doc github.com/some/package.Type

# For uninstalled packages, use pkg.go.dev directly
```

## Tips

1. **Start broad, then narrow**: Use `-short` for overview, then drill into specific symbols
2. **Check the source**: Use `-src` when docs are sparse or you need implementation details
3. **Don't forget unexported**: Use `-u` to understand internal helpers
4. **Combine with grep**: Pipe to grep for large packages: `go doc -all fmt | grep -i print`
5. **Use in any directory**: Specify full import paths to query any package from anywhere
