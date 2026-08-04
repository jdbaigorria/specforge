"""Tests for slugify — one acceptance criterion per test (maps to trace.json)."""

from texttools import slugify


def test_basic():  # R1
    assert slugify("Hello World") == "hello-world"


def test_accents():  # R2
    assert slugify("Café Olé") == "cafe-ole"


def test_punctuation_collapse():  # R3
    assert slugify("  --A, B & C!! ") == "a-b-c"


def test_empty():  # R4
    assert slugify("!!!") == ""
    assert slugify("") == ""
