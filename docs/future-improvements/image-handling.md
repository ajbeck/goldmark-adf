# Future Improvements: Image Handling

## Current Implementation

Currently, Markdown images are converted to links:

```markdown
![Alt text](https://example.com/image.png "Title")
```

Becomes:

```json
{
  "type": "text",
  "text": "Alt text",
  "marks": [
    {
      "type": "link",
      "attrs": {
        "href": "https://example.com/image.png",
        "title": "Title"
      }
    }
  ]
}
```

This approach was chosen for maximum compatibility since ADF media nodes require integration with Atlassian Media Services.

## Alternative Approaches

### Option 1: External Media Nodes

ADF supports external media via `mediaSingle` with `media` nodes using `type: "external"`:

```json
{
  "type": "mediaSingle",
  "attrs": {
    "layout": "center"
  },
  "content": [
    {
      "type": "media",
      "attrs": {
        "type": "external",
        "url": "https://example.com/image.png",
        "alt": "Alt text",
        "width": 800,
        "height": 600
      }
    }
  ]
}
```

**Pros:**
- Preserves image rendering intent
- Supports layout options (center, wide, full-width, wrap-left, wrap-right)
- Can include width/height hints

**Cons:**
- External URLs may be blocked by CSP in some Atlassian products
- Not all Atlassian contexts fully support external media
- May require additional permissions or trust settings
- Width/height not available from Markdown syntax

**Layout Options:**
- `center`: Block-aligned, centered on page
- `wide`: Centered, extending into margins
- `full-width`: Edge-to-edge stretch
- `wrap-left`: Floated left with text wrap
- `wrap-right`: Floated right with text wrap
- `align-start`: Aligned to start (left in LTR)
- `align-end`: Aligned to end (right in LTR)

### Option 2: Inline Card Nodes

For URLs that point to known services, use `inlineCard`:

```json
{
  "type": "inlineCard",
  "attrs": {
    "url": "https://example.com/image.png"
  }
}
```

**Pros:**
- Atlassian's Smart Links may render image previews
- Works well for links to Confluence pages, Jira issues, etc.

**Cons:**
- Preview depends on Atlassian resolving the URL
- May just show as a link card rather than an image
- Not suitable for arbitrary image URLs

### Option 3: Media Services Integration

For full media support, integrate with Atlassian Media Services:

1. Upload image to Media Services API
2. Receive media ID and collection
3. Use proper `media` node structure:

```json
{
  "type": "mediaSingle",
  "attrs": { "layout": "center" },
  "content": [
    {
      "type": "media",
      "attrs": {
        "id": "4478e39c-cf9b-41d1-ba92-68589487cd75",
        "type": "file",
        "collection": "MediaServicesSample",
        "width": 800,
        "height": 600,
        "alt": "Alt text"
      }
    }
  ]
}
```

**Pros:**
- Full native support in all Atlassian products
- Thumbnail generation, preview support
- Proper access control

**Cons:**
- Requires API access and authentication
- Significant implementation complexity
- May require separate upload step before rendering
- Not suitable for offline/static conversion

### Option 4: Configurable Handler

Provide a callback mechanism for custom image handling:

```go
type ImageHandler func(destination, alt, title string) (Node, error)

type Config struct {
    ImageHandler ImageHandler
}

// Default handler
func DefaultImageHandler(dest, alt, title string) (Node, error) {
    return &TextNode{
        Text: alt,
        Marks: []Mark{{Type: "link", Attrs: map[string]any{"href": dest, "title": title}}},
    }, nil
}

// External media handler
func ExternalMediaHandler(dest, alt, title string) (Node, error) {
    return &MediaSingleNode{
        Attrs: MediaSingleAttrs{Layout: "center"},
        Content: []Node{
            &MediaNode{
                Attrs: MediaAttrs{
                    Type: "external",
                    URL:  dest,
                    Alt:  alt,
                },
            },
        },
    }, nil
}
```

**Pros:**
- Maximum flexibility for users
- Can integrate with any backend
- Allows runtime decision based on URL patterns

**Cons:**
- More API complexity
- Users need to understand ADF structure

## Recommendations

### Short Term
Keep the current link-based approach as the default. It's the safest and most compatible option.

### Medium Term
Add a configuration option for external media nodes:

```go
adf.NewRenderer(
    adf.WithExternalMediaImages(true),
    adf.WithDefaultImageLayout("center"),
)
```

### Long Term
Implement the configurable handler pattern to allow:
- Custom Media Services integration
- URL pattern matching (e.g., Confluence/Jira links → inlineCard)
- Image dimension fetching
- Base64 data URL handling

## Related ADF Nodes

### mediaSingle
Container for single media items. Layout options control presentation.

### mediaGroup
Container for multiple media items shown as an attachment list. Not suitable for inline images.

### mediaInline
Inline media within text. Limited support and use cases.

### media Node Attributes

For `type: "file"`:
- `id`: Media Services identifier (required)
- `type`: "file" (required)
- `collection`: Media Services collection name (required)
- `width`: Display width in pixels
- `height`: Display height in pixels
- `alt`: Alternative text
- `occurrenceKey`: For file management

For `type: "external"`:
- `type`: "external" (required)
- `url`: Image URL (required)
- `width`: Display width in pixels
- `height`: Display height in pixels
- `alt`: Alternative text

For `type: "link"`:
- `id`: Media Services identifier (required)
- `type`: "link" (required)
- `collection`: Media Services collection name (required)

## Caption Support

`mediaSingle` can include a caption:

```json
{
  "type": "mediaSingle",
  "attrs": { "layout": "center" },
  "content": [
    {
      "type": "media",
      "attrs": { "type": "external", "url": "..." }
    },
    {
      "type": "caption",
      "content": [
        { "type": "text", "text": "Figure 1: Description" }
      ]
    }
  ]
}
```

Markdown title could be converted to caption in future.
