"""Tiny text statistics."""


def word_count(text: str) -> int:
    """Count whitespace-separated words.

    Empty or whitespace-only input returns 0. (Lite fix `fix-empty-wordcount`:
    the previous `text.split(" ")` counted "" as one word — now `text.split()`
    collapses runs of whitespace and yields 0 on empty input.)
    """
    return len(text.split())
