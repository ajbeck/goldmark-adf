# Atlassian Image Handling Research

**Research Date:** January 29, 2026
**Purpose:** Understand how Atlassian's official TypeScript libraries convert Markdown images to ADF (Atlassian Document Format)

## Package Versions Analyzed

```json
{
  "@atlaskit/adf-schema": "^51.5.7",
  "@atlaskit/editor-json-transformer": "^8.31.4",
  "@atlaskit/editor-markdown-transformer": "^5.20.4"
}
```

## Key Source Files Examined

### 1. Media Plugin (`@atlaskit/editor-markdown-transformer/dist/esm/media.js`)

This file contains the core logic for converting Markdown images to ADF media nodes. Key points:

- Uses regex `/!\[([^\]]*)\]\(([^)]*?)\s*(?:"([^")]*)"\s*)?\)/g` to detect image syntax
- Transforms inline images into block-level `mediaSingle` + `media` tokens
- **Paragraph splitting**: When an image is found inside inline content, the plugin splits the content before and after the image into separate paragraphs

```javascript
// Creates media tokens with external type
var createMediaTokens = function createMediaTokens(url) {
  var alt = arguments.length > 1 && arguments[1] !== undefined ? arguments[1] : '';
  var mediaSingleOpen = new State.Token('media_single_open', '', 1);
  var media = new State.Token('media', '', 0);
  media.attrs = [['url', getUrl(url)], ['type', 'external'], ['alt', alt]];
  var mediaSingleClose = new State.Token('media_single_close', '', -1);
  return [mediaSingleOpen, media, mediaSingleClose];
};
```

### 2. Token Mapping (`@atlaskit/editor-markdown-transformer/dist/esm/index.js`)

Defines how markdown-it tokens map to ProseMirror/ADF nodes:

```javascript
media_single: {
  block: 'mediaSingle',
  attrs: function attrs() {
    return {};
  }
},
media: {
  node: 'media',
  attrs: function attrs(tok) {
    return {
      url: tok.attrGet('url'),
      alt: tok.attrGet('alt'),
      type: 'external'
    };
  }
}
```

## ADF Schema Definitions

### media_node

```json
{
  "type": "object",
  "properties": {
    "type": { "enum": ["media"] },
    "attrs": {
      "anyOf": [
        {
          "type": "object",
          "properties": {
            "type": { "enum": ["link", "file"] },
            "id": { "minLength": 1, "type": "string" },
            "collection": { "type": "string" },
            "alt": { "type": "string" },
            "width": { "type": "number" },
            "height": { "type": "number" }
          },
          "required": ["type", "id", "collection"]
        },
        {
          "type": "object",
          "properties": {
            "type": { "enum": ["external"] },
            "url": { "type": "string" },
            "alt": { "type": "string" },
            "width": { "type": "number" },
            "height": { "type": "number" }
          },
          "required": ["type", "url"]
        }
      ]
    }
  },
  "required": ["type", "attrs"]
}
```

### mediaSingle_node

```json
{
  "type": "object",
  "properties": {
    "type": { "enum": ["mediaSingle"] },
    "attrs": {
      "anyOf": [
        {
          "type": "object",
          "properties": {
            "layout": {
              "enum": ["wide", "full-width", "center", "wrap-right", "wrap-left", "align-end", "align-start"]
            },
            "width": { "type": "number", "minimum": 0, "maximum": 100 },
            "widthType": { "enum": ["percentage"] }
          },
          "required": ["layout"]
        },
        {
          "type": "object",
          "properties": {
            "layout": {
              "enum": ["wide", "full-width", "center", "wrap-right", "wrap-left", "align-end", "align-start"]
            },
            "width": { "type": "number", "minimum": 0 },
            "widthType": { "enum": ["pixel"] }
          },
          "required": ["width", "widthType", "layout"]
        }
      ]
    },
    "content": {
      "type": "array",
      "items": { "$ref": "#/definitions/media_node" },
      "minItems": 1,
      "maxItems": 1
    }
  },
  "required": ["type"]
}
```

### caption_node

The `mediaSingle_caption_node` variant allows a caption after the media:

```json
{
  "content": {
    "type": "array",
    "items": [
      { "$ref": "#/definitions/media_node" },
      { "$ref": "#/definitions/caption_node" }
    ],
    "minItems": 1,
    "maxItems": 2
  }
}
```

## Test Script

```typescript
/**
 * Test script to see how Atlassian converts images to ADF
 */

import { createSchema } from "@atlaskit/adf-schema";
import { JSONTransformer } from "@atlaskit/editor-json-transformer";
import { MarkdownTransformer } from "@atlaskit/editor-markdown-transformer";

// Create schema with media support
const schema = createSchema({
  nodes: [
    "doc",
    "paragraph",
    "text",
    "bulletList",
    "orderedList",
    "listItem",
    "heading",
    "blockquote",
    "codeBlock",
    "hardBreak",
    "rule",
    "table",
    "tableRow",
    "tableCell",
    "tableHeader",
    "panel",
    "media",
    "mediaSingle",
    "caption",
  ],
  marks: ["strong", "em", "code", "link", "strike"],
});

const jsonTransformer = new JSONTransformer();
const markdownTransformer = new MarkdownTransformer(schema);

function markdownToAdf(markdown: string) {
  const pmNode = markdownTransformer.parse(markdown);
  return jsonTransformer.encode(pmNode);
}

// Test cases
const testCases = [
  "![Alt text](https://example.com/image.png)",
  '![Alt text](https://example.com/image.png "Image Title")',
  "Here is some text ![image](https://example.com/image.png) and more text after",
  "![first](https://a.com/1.png)\n![second](https://b.com/2.png)",
  "- Item with ![image](https://example.com/image.png) inline",
  "> Quote with ![image](https://example.com/image.png)",
];

for (const md of testCases) {
  console.log("--- INPUT ---");
  console.log(md);
  console.log("\n--- OUTPUT ---");
  console.log(JSON.stringify(markdownToAdf(md), null, 2));
}
```

## Test Output

### Simple Image

**Input:**
```markdown
![Alt text](https://example.com/image.png)
```

**Output:**
```json
{
  "version": 1,
  "type": "doc",
  "content": [
    {
      "type": "mediaSingle",
      "attrs": { "layout": "center" },
      "content": [
        {
          "type": "media",
          "attrs": {
            "type": "external",
            "alt": "Alt text",
            "url": "https://example.com/image.png"
          }
        }
      ]
    }
  ]
}
```

### Inline Image in Paragraph (Paragraph Splitting)

**Input:**
```markdown
Here is some text ![image](https://example.com/image.png) and more text after
```

**Output:**
```json
{
  "version": 1,
  "type": "doc",
  "content": [
    {
      "type": "paragraph",
      "content": [{ "type": "text", "text": "Here is some text " }]
    },
    {
      "type": "mediaSingle",
      "attrs": { "layout": "center" },
      "content": [
        {
          "type": "media",
          "attrs": {
            "type": "external",
            "alt": "image",
            "url": "https://example.com/image.png"
          }
        }
      ]
    },
    {
      "type": "paragraph",
      "content": [{ "type": "text", "text": " and more text after" }]
    }
  ]
}
```

### Image in List Item

**Input:**
```markdown
- Item with ![image](https://example.com/image.png) inline
```

**Output:**
```json
{
  "version": 1,
  "type": "doc",
  "content": [
    {
      "type": "bulletList",
      "content": [
        {
          "type": "listItem",
          "content": [
            {
              "type": "paragraph",
              "content": [{ "type": "text", "text": "Item with " }]
            },
            {
              "type": "mediaSingle",
              "attrs": { "layout": "center" },
              "content": [
                {
                  "type": "media",
                  "attrs": {
                    "type": "external",
                    "alt": "image",
                    "url": "https://example.com/image.png"
                  }
                }
              ]
            },
            {
              "type": "paragraph",
              "content": [{ "type": "text", "text": " inline" }]
            }
          ]
        }
      ]
    }
  ]
}
```

### Image in Blockquote

**Input:**
```markdown
> Quote with ![image](https://example.com/image.png)
```

**Output:**
```json
{
  "version": 1,
  "type": "doc",
  "content": [
    {
      "type": "blockquote",
      "content": [
        {
          "type": "paragraph",
          "content": [{ "type": "text", "text": "Quote with " }]
        },
        {
          "type": "mediaSingle",
          "attrs": { "layout": "center" },
          "content": [
            {
              "type": "media",
              "attrs": {
                "type": "external",
                "alt": "image",
                "url": "https://example.com/image.png"
              }
            }
          ]
        }
      ]
    }
  ]
}
```

## Key Findings

1. **External Media Type**: Atlassian uses `type: "external"` for URL-based images (as opposed to `type: "file"` for Media Services-uploaded files)

2. **Paragraph Splitting**: When an image appears inline within a paragraph, Atlassian splits the paragraph into separate before/after paragraphs with the mediaSingle between them. This is because `mediaSingle` is a block-level element that cannot be nested inside a paragraph.

3. **Default Layout**: The default layout for mediaSingle is `"center"`

4. **Nested Contexts**: Images inside list items and blockquotes are handled the same way - paragraphs are split, and the mediaSingle becomes a sibling to the surrounding content within the container.

5. **Title/Caption**: The Atlassian transformer does NOT convert the image title (`"Image Title"`) to a caption - it's simply ignored. Caption support requires explicit ADF authoring.

6. **Alt Text**: The alt text is preserved in the media node's `alt` attribute.

## Re-running This Research

When Atlassian updates their packages, follow these steps to validate our implementation:

1. Create a new directory and initialize with bun:
   ```bash
   mkdir atlassian-test && cd atlassian-test
   bun init -y
   ```

2. Install the Atlassian packages:
   ```bash
   bun add @atlaskit/adf-schema @atlaskit/editor-json-transformer @atlaskit/editor-markdown-transformer
   ```

3. Copy the test script above and run it:
   ```bash
   bun run test-image.ts
   ```

4. Compare the output against our Go implementation:
   ```bash
   GOEXPERIMENT=jsonv2 go test ./... -run TestConvert_Image -v
   ```

5. Key things to check:
   - Media node structure (attrs: type, url, alt)
   - mediaSingle attrs (layout default)
   - Paragraph splitting behavior
   - Nested contexts (lists, blockquotes)

## Related Documentation

- [ADF Documentation](https://developer.atlassian.com/cloud/jira/platform/apis/document/structure/)
- [Atlaskit Packages](https://atlassian.design/components)
- [Go Implementation Notes](../future-improvements/image-handling.md)
