---
id: m-04
nacio_en: 1                  # la vuelta que lo escribió. No cambia nunca.
objetivo: internal/suite/suite.go:71
intencion: "no cierra el archivo — el defer se pierde"
historial:                   # una línea por vuelta, append. Nunca se reescribe.
  - vuelta: 1
    resultado: murio         # murio | sobrevivio | no-aplica
  - vuelta: 2
    resultado: sobrevivio    # ← esto es una REGRESIÓN, y va como hallazgo
---

## El parche

```diff
-	defer f.Close()
+
```

## Qué test tendría que agarrarlo

`internal/suite/suite_test.go::TestCierraElArchivo` — y en la vuelta 1 lo agarraba.
