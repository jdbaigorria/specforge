"""Temperature conversions.

`c_to_f` pre-dates SpecForge (this is a brownfield project). The `add-kelvin`
feature added `c_to_k`, `k_to_c`, and the absolute-zero guard.
"""

ABSOLUTE_ZERO_C = -273.15


def c_to_f(celsius: float) -> float:
    """Convert Celsius to Fahrenheit. (Pre-existing.)"""
    return celsius * 9 / 5 + 32


def c_to_k(celsius: float) -> float:
    """Convert Celsius to Kelvin."""
    _guard_abs_zero(celsius)
    return celsius - ABSOLUTE_ZERO_C


def k_to_c(kelvin: float) -> float:
    """Convert Kelvin to Celsius."""
    celsius = kelvin + ABSOLUTE_ZERO_C
    _guard_abs_zero(celsius)
    return celsius


def _guard_abs_zero(celsius: float) -> None:
    if celsius < ABSOLUTE_ZERO_C:
        raise ValueError("temperature below absolute zero")
