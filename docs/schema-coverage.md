# ADF Schema Coverage

This document tracks which ADF nodes goldmark-adf can produce from markdown input. For extension usage details, see [extensions.md](extensions.md). For ADF node reference (schema-only nodes without official docs), see the [adf-nodes.md](https://github.com/ajbeck/adf-to-markdown/blob/main/docs/adf-nodes.md) document in adf-to-markdown.

Status labels:

- `implemented`: parser extension and renderer produce correct ADF output
- `partial`: produced via fallback behavior, not full ADF semantics
- `native`: handled by standard CommonMark or GFM parsing, no custom extension needed
- `planned`: not yet implemented, will be added in a future release
- `n/a`: not applicable for markdown-to-ADF conversion

## Nodes

| ADF Node | Status | Markdown Source | Notes |
|---|---|---|---|
| `blockCard` | implemented | `[card:url]` (sole paragraph content) | Custom extension |
| `blockquote` | native | `> text` | Standard CommonMark |
| `blockTaskItem` | n/a | - | Jira-specific variant; `taskItem` is used instead |
| `bodiedExtension` | planned | `[extension:bodiedExtension:key]` | Phase 4 |
| `bodiedSyncBlock` | planned | `[extension:bodiedSyncBlock:key]` | Phase 4 |
| `bulletList` | native | `- item` | Standard CommonMark |
| `caption` | native | `![alt](url "caption")` | Image title becomes caption |
| `codeBlock` | native | `` ```language `` | Standard CommonMark |
| `date` | implemented | `[date:timestamp]` | Custom inline extension |
| `decisionItem` | implemented | `- [!] text` / `- [?] text` | Custom block extension |
| `decisionList` | implemented | (list of decision items) | Custom block extension |
| `doc` | native | (document root) | Always produced |
| `embedCard` | implemented | `[embed:url]` | Custom inline extension |
| `emoji` | implemented | `:shortcode:` | Inline extension |
| `expand` | planned | `<details><summary>` | Phase 4 |
| `extension` | planned | `[extension:extension:key]` | Phase 4 |
| `hardBreak` | native | two spaces + newline | Standard CommonMark |
| `heading` | native | `# heading` | Standard CommonMark |
| `inlineCard` | implemented | `[card:url]` (within inline content) | Custom inline extension |
| `inlineExtension` | planned | `[extension:inlineExtension:key]` | Phase 4 |
| `layoutColumn` | planned | `[layout-column N]` | Phase 4 |
| `layoutSection` | planned | `[layout-section]` | Phase 4 |
| `listItem` | native | `- item` or `1. item` | Standard CommonMark |
| `media` | native | `![alt](url)` | Requires `WithExternalMedia(true)` |
| `mediaGroup` | planned | consecutive `![alt](url)` | Phase 4 |
| `mediaInline` | planned | `![alt](url)` inline | Phase 4 |
| `mediaSingle` | native | `![alt](url)` | Requires `WithExternalMedia(true)` |
| `mention` | implemented | `@[name](id)` | Custom inline extension |
| `nestedExpand` | planned | `<details><summary>` nested | Phase 4 |
| `orderedList` | native | `1. item` | Standard CommonMark |
| `panel` | implemented | `> [!TYPE]` | GitHub alert syntax |
| `paragraph` | native | (paragraph text) | Standard CommonMark |
| `placeholder` | implemented | `{{text}}` | Custom inline extension |
| `rule` | native | `---` | Standard CommonMark |
| `status` | implemented | `[status:text\|color]` | Custom inline extension |
| `syncBlock` | planned | `[extension:syncBlock:key]` | Phase 4 |
| `table` | native | GFM table syntax | Requires `NewWithGFM()` |
| `tableCell` | native | `\| cell \|` | Requires `NewWithGFM()` |
| `tableHeader` | native | `\| header \|` | Requires `NewWithGFM()` |
| `tableRow` | native | `\| row \|` | Requires `NewWithGFM()` |
| `taskItem` | implemented | `- [x]` / `- [ ]` | GFM task list detection |
| `taskList` | implemented | (list of task items) | GFM task list detection |
| `text` | native | (inline text) | Standard CommonMark |

## Marks

| ADF Mark | Status | Markdown Source | Notes |
|---|---|---|---|
| `strong` | native | `**text**` | Standard CommonMark |
| `em` | native | `*text*` | Standard CommonMark |
| `code` | native | `` `text` `` | Standard CommonMark |
| `link` | native | `[text](url)` | Standard CommonMark |
| `strike` | native | `~~text~~` | Requires `NewWithGFM()` |
| `underline` | planned | `<u>text</u>` | Phase 4 |
| `subsup` | planned | `<sub>text</sub>` / `<sup>text</sup>` | Phase 4 |
| `textColor` | n/a | - | Loss accepted; no markdown equivalent |
| `backgroundColor` | n/a | - | Loss accepted; no markdown equivalent |
| `annotation` | n/a | - | Confluence-specific |

## Summary

| Category | Count |
|---|---|
| Native (CommonMark/GFM) | 18 nodes, 5 marks |
| Implemented extensions | 10 nodes |
| Planned (Phase 4) | 11 nodes, 2 marks |
| Not applicable | 2 nodes, 3 marks |
