# Requirements — slugify

Convert arbitrary text into a URL-safe slug: lowercase, hyphen-joined, accents folded to ASCII, punctuation removed.

**Actors:** library caller

## R1 (event)

When slugify is called with a string of words separated by spaces, the system shall return the words lowercased and joined by single hyphens.

**Acceptance:**
- **R1.1** — slugify("Hello World") == "hello-world"

## R2 (event)

When the input contains accented or non-ASCII Latin characters, the system shall fold them to their closest ASCII equivalent.

**Acceptance:**
- **R2.1** — slugify("Café Olé") == "cafe-ole"

## R3 (event)

When the input contains punctuation or repeated separators, the system shall drop the punctuation and collapse runs of separators into a single hyphen, with no leading or trailing hyphen.

**Acceptance:**
- **R3.1** — slugify("  --A, B & C!! ") == "a-b-c"

## R4 (error)

If the input contains no slug-able characters, then the system shall return an empty string (never raise).

**Acceptance:**
- **R4.1** — slugify("!!!") == ""
- **R4.2** — slugify("") == ""

