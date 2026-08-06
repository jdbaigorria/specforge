"""Tests for slugify — one acceptance criterion per test (maps to trace.json)."""

from texttools import slugify


def test_basic():  # R1.1
    assert slugify("Hello World") == "hello-world"


def test_accents():  # R2.1
    assert slugify("Café Olé") == "cafe-ole"


def test_punctuation_collapse():  # R3.1
    assert slugify("  --A, B & C!! ") == "a-b-c"


def test_punctuation_only_is_empty():  # R4.1
    assert slugify("!!!") == ""


def test_empty_input_is_empty():  # R4.2
    assert slugify("") == ""
