"""slugify — turn arbitrary text into a URL-safe slug."""

from __future__ import annotations

import re
import unicodedata

_WORD = re.compile(r"[a-z0-9]+")


def slugify(text: str, sep: str = "-") -> str:
    """Convert text into a URL-safe slug.

    Lowercase, words joined by ``sep``, accents folded to ASCII, punctuation
    dropped. Returns "" when there is nothing slug-able.

    >>> slugify("Hello World")
    'hello-world'
    >>> slugify("Café Olé")
    'cafe-ole'
    >>> slugify("  --A, B & C!! ")
    'a-b-c'
    >>> slugify("!!!")
    ''
    """
    # NFKD decomposes accented chars; drop the combining marks (R2).
    folded = unicodedata.normalize("NFKD", text)
    ascii_text = "".join(c for c in folded if not unicodedata.combining(c))
    words = _WORD.findall(ascii_text.lower())  # drops punctuation (R3)
    return sep.join(words)  # single separators, no leading/trailing (R1, R3, R4)
