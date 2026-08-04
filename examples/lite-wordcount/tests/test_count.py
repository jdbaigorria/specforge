from textstats.count import word_count


def test_basic():
    assert word_count("hello world") == 2


def test_empty():            # the bug this lite change fixes
    assert word_count("") == 0


def test_whitespace_only():
    assert word_count("   ") == 0
