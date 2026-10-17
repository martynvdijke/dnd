# wiki-content-safety Specification

## Purpose
Defines the trust boundary for rendering user-authored content: markdown is parsed and sanitized before it reaches the DOM, and the shared toast and modal helpers default to escaping.

## Requirements

### Requirement: Safe markdown rendering for user content

The system SHALL render user-authored markdown (wiki pages, notes, and any other `marked.parse` consumer) through a shared `renderMarkdown()` helper that parses with `marked` and sanitizes the resulting HTML with DOMPurify before insertion into the DOM. Raw HTML injection via markdown input SHALL NOT execute.

#### Scenario: Script tag stripped

- **WHEN** a wiki page containing `<script>alert(1)</script>` is rendered via `renderMarkdown()`
- **THEN** the resulting HTML contains no `<script>` element and no script execution occurs

#### Scenario: Image onerror handler stripped

- **WHEN** markdown containing `<img src=x onerror=alert(1)>` (or markdown image with crafted attributes) is rendered
- **THEN** the `onerror` attribute is removed from the output and no handler executes

#### Scenario: javascript: link sanitized

- **WHEN** markdown containing `[click](javascript:alert(1))` is rendered
- **THEN** the `javascript:` href is stripped or neutralized and no navigation to a javascript URL is possible

#### Scenario: Legitimate markdown preserved

- **WHEN** markdown containing headings, bold, links (`https://`), lists, code blocks, tables, and images with `https://` or relative `/uploads/` sources is rendered
- **THEN** the corresponding HTML elements are present and styled correctly

#### Scenario: All marked.parse sites use the helper

- **WHEN** the codebase is searched for `marked.parse`
- **THEN** no call site outside `ts/lib/markdown.ts` invokes `marked.parse` directly; all go through `renderMarkdown()`

### Requirement: Safe toast helper defaults to escaping

The `toast()` helper in `ts/lib/dom.ts` SHALL escape its message content by default so that HTML in the message is rendered as text, not interpreted. Callers that intentionally need HTML SHALL opt in explicitly via `{ html: true }`.

#### Scenario: Toast escapes HTML by default

- **WHEN** `toast('<b>hi</b><img src=x onerror=alert(1)>')` is called without an HTML opt-in
- **THEN** the toast displays the literal text including angle brackets and no HTML is interpreted or executed

#### Scenario: Toast renders HTML only with explicit opt-in

- **WHEN** `toast('<b>hi</b>', { html: true })` is called
- **THEN** the toast renders bold markup as HTML

#### Scenario: Existing HTML toast callers migrated

- **WHEN** the codebase is audited for `toast(` calls that pass intentional HTML
- **THEN** each such call uses the `{ html: true }` opt-in and no caller relies on implicit HTML interpretation

### Requirement: Modal helper trust boundary documented and enforced

Callers of `showModal(bodyHtml)` SHALL ensure any user-authored content interpolated into `bodyHtml` is escaped via `esc()` (or via `renderMarkdown()` for markdown) before calling `showModal`. The helper itself SHALL NOT double-escape its input.

#### Scenario: User content in modal is escaped by caller

- **WHEN** a modal body includes a wiki page title or other user-supplied string
- **THEN** that string is escaped before interpolation so that `<script>` in the title does not execute when the modal opens

### Requirement: Sanitizer helper has unit test coverage

The `renderMarkdown()` helper SHALL have vitest unit tests that verify dangerous payloads are stripped and legitimate markdown is preserved.

#### Scenario: Sanitizer unit tests pass

- **WHEN** `npm run test:unit` is executed
- **THEN** tests for `renderMarkdown` covering `<script>` injection, `onerror` attribute stripping, `javascript:` href neutralization, and preservation of allowed markdown elements all pass
