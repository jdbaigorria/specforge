# Requirements — add-kelvin

Add Kelvin conversions (c_to_k, k_to_c) to the existing tempconv library.

**Actors:** library caller

## R1 (event)

When c_to_k(celsius) is called, the system shall return the temperature in kelvin.

**Acceptance:**
- c_to_k(0) == 273.15

## R2 (event)

When k_to_c(kelvin) is called, the system shall return the temperature in celsius.

**Acceptance:**
- k_to_c(273.15) == 0

## R3 (error)

If a temperature below absolute zero (-273.15 C) is produced, then the system shall raise ValueError.

**Acceptance:**
- c_to_k(-300) raises ValueError

