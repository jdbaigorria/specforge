# Requirements — slugify

Status: approved

## Overview

Provide `slugify(text)` — convert arbitrary text into a URL-safe slug: lowercase,
words joined by single hyphens, accents folded to ASCII, punctuation removed.

## Requirements (EARS notation)

**R1 — Basic slug**
WHEN `slugify` is called with a string of words separated by spaces,
THE SYSTEM SHALL return the words lowercased and joined by single hyphens.
*Acceptance:* `slugify("Hello World")` → `"hello-world"`.

**R2 — Accent folding**
WHEN the input contains accented or non-ASCII Latin characters,
THE SYSTEM SHALL fold them to their closest ASCII equivalent.
*Acceptance:* `slugify("Café Olé")` → `"cafe-ole"`.

**R3 — Punctuation and collapsing**
WHEN the input contains punctuation or repeated separators,
THE SYSTEM SHALL drop the punctuation and collapse runs of separators into a
single hyphen, with no leading or trailing hyphen.
*Acceptance:* `slugify("  --A, B & C!! ")` → `"a-b-c"`.

**R4 — Empty result**
IF the input contains no slug-able characters,
THEN THE SYSTEM SHALL return an empty string (never raise).
*Acceptance:* `slugify("!!!")` → `""`; `slugify("")` → `""`.
