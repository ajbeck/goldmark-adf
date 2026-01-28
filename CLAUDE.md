This is a Go module which is a goldmark (https://github.com/yuin/goldmark) renderer for rendering Markdown in Atlassian Document Format (ADF) https://developer.atlassian.com/cloud/jira/platform/apis/document/structure/

## Build Requirements

This package requires Go 1.25+ with the experimental json/v2 package enabled:

```bash
GOEXPERIMENT=jsonv2 go build ./...
GOEXPERIMENT=jsonv2 go test ./...
```
