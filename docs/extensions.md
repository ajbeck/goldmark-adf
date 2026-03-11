# Markdown Extensions

goldmark-adf includes parser extensions that recognize custom markdown syntax for ADF-specific nodes. These extensions parse the output of [adf-to-markdown](https://github.com/ajbeck/adf-to-markdown) back into valid ADF JSON, enabling lossless round-tripping.

For the full round-trip specification shared with adf-to-markdown, see the [roundtrip-extensions.md](https://github.com/ajbeck/adf-to-markdown/blob/main/docs/roundtrip-extensions.md) document.

## Enabling Extensions

All custom extension parsers are registered automatically when using `NewWithGFM()`:

```go
md := adf.NewWithGFM()

var buf bytes.Buffer
if err := md.Convert(markdown, &buf); err != nil {
    log.Fatal(err)
}
```

`NewWithGFM()` enables:
- GFM parsing (tables, strikethrough, autolinks, task lists)
- Custom inline parsers (status, mention, date, placeholder, card, embed, emoji)
- Custom block transformers (panels from GitHub alerts, decision lists)

`New()` provides only standard CommonMark parsing with no extensions.

## Inline Extensions

### Status

Parses `[status:TEXT|COLOR]` into an ADF `status` node.

```markdown
[status:In Progress|yellow]
```

```json
{"type": "status", "attrs": {"text": "In Progress", "color": "yellow"}}
```

Backslash-escaped `\|` and `\]` in the text field are unescaped during parsing.

### Mentions

Parses `@[NAME](ID)` into an ADF `mention` node.

```markdown
@[Jane Smith](abc-123)
```

```json
{"type": "mention", "attrs": {"id": "abc-123", "text": "Jane Smith"}}
```

Backslash-escaped `\]` in the name and `\)` in the ID are unescaped during parsing.

### Dates

Parses `[date:DIGITS]` into an ADF `date` node.

```markdown
[date:1582152559]
```

```json
{"type": "date", "attrs": {"timestamp": "1582152559"}}
```

### Placeholders

Parses `{{TEXT}}` into an ADF `placeholder` node.

```markdown
{{Enter your name}}
```

```json
{"type": "placeholder", "attrs": {"text": "Enter your name"}}
```

Backslash-escaped `\}` in the text is unescaped during parsing.

### Cards

Parses `[card:URL]` into an ADF `inlineCard` node.

```markdown
[card:https://atlassian.com/project]
```

```json
{"type": "inlineCard", "attrs": {"url": "https://atlassian.com/project"}}
```

### Embed Cards

Parses `[embed:URL]` into an ADF `embedCard` node.

```markdown
[embed:https://youtube.com/watch?v=abc]
```

```json
{"type": "embedCard", "attrs": {"url": "https://youtube.com/watch?v=abc"}}
```

### Emoji

Parses `:shortcode:` into an ADF `emoji` node.

```markdown
:smile:
```

```json
{"type": "emoji", "attrs": {"shortName": ":smile:"}}
```

The shortcode must contain only alphanumeric characters, underscores, hyphens, and plus signs.

## Block Extensions

### Panels (GitHub Alerts)

Blockquotes starting with `[!KEYWORD]` are converted to ADF `panel` nodes.

```markdown
> [!WARNING]
> Be careful with this operation.
```

```json
{
  "type": "panel",
  "attrs": {"panelType": "warning"},
  "content": [
    {
      "type": "paragraph",
      "content": [{"type": "text", "text": "Be careful with this operation."}]
    }
  ]
}
```

Alert keyword mapping:

| Alert Keyword | ADF `panelType` |
|---|---|
| `NOTE` | `info` |
| `TIP` | `success` |
| `IMPORTANT` | `info` |
| `WARNING` | `warning` |
| `CAUTION` | `error` |

Regular blockquotes (without an alert keyword) are unaffected.

### Decision Lists

Unordered lists where **all** items start with `[!]` or `[?]` are converted to ADF `decisionList` / `decisionItem` nodes.

```markdown
- [!] Use json/v2 for performance
- [?] Pending design review
```

```json
{
  "type": "decisionList",
  "attrs": {"localId": ""},
  "content": [
    {
      "type": "decisionItem",
      "attrs": {"localId": "", "state": "DECIDED"},
      "content": [{"type": "paragraph", "content": [{"type": "text", "text": "Use json/v2 for performance"}]}]
    },
    {
      "type": "decisionItem",
      "attrs": {"localId": "", "state": "UNDECIDED"},
      "content": [{"type": "paragraph", "content": [{"type": "text", "text": "Pending design review"}]}]
    }
  ]
}
```

| Marker | ADF `state` |
|---|---|
| `[!]` | `DECIDED` |
| `[?]` | `UNDECIDED` |

If any item in a list does not start with `[!]` or `[?]`, the entire list is treated as a regular bullet list.

### Task Lists

GFM task list checkboxes are converted to ADF `taskList` / `taskItem` nodes.

```markdown
- [x] Completed task
- [ ] Incomplete task
```

```json
{
  "type": "taskList",
  "attrs": {"localId": ""},
  "content": [
    {
      "type": "taskItem",
      "attrs": {"localId": "", "state": "DONE"},
      "content": [{"type": "text", "text": "Completed task"}]
    },
    {
      "type": "taskItem",
      "attrs": {"localId": "", "state": "TODO"},
      "content": [{"type": "text", "text": "Incomplete task"}]
    }
  ]
}
```

A list is treated as a task list only if **all** items contain a checkbox.

## Not Yet Implemented

The following extensions are planned but not yet available in goldmark-adf:

- `<details><summary>` parsing for ADF `expand` nodes
- `[layout-section]` / `[layout-column N]` markers for layout nodes
- `[extension:TYPE:KEY]` markers for extension nodes
- Non-external media URL schemes (`atlassian-media://`)
- Inline HTML mark parsing (`<u>`, `<sub>`, `<sup>`)
