import pytest

from tempconv.convert import c_to_f, c_to_k, k_to_c


def test_c_to_f():            # pre-existing
    assert c_to_f(100) == 212


def test_c_to_k():
    assert c_to_k(0) == pytest.approx(273.15)


def test_k_to_c():
    assert k_to_c(273.15) == pytest.approx(0)


def test_abs_zero():
    with pytest.raises(ValueError):
        c_to_k(-300)
