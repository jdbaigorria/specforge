#!/usr/bin/env bash
# comparar.sh — el veredicto de un tramo, en una corrida.
#
#   ./comparar.sh              compara A y B  (el default)
#   ./comparar.sh A B          lo mismo, explícito
#   ./comparar.sh T2-A T2-B    compara otro par
#
# ─────────────────────────────────────────────────────────────────────────────
# LA VARA DE T1 ES DE EVIDENCIA, NO DE VEREDICTO  (por-tramos.md §4, la excepción)
# ─────────────────────────────────────────────────────────────────────────────
#
# Hasta el 2026-09-07 este script decía que un tramo PASA si el `estado.json`
# quedó igual con los dos modelos. Ese criterio sirve de T2 en adelante, donde
# los campos que se mueven son mecánicos.
#
# En T1 NO. El único campo que se mueve es `brief_sellado`, y ése ES el
# veredicto: juicio puro. Dos modelos honestos pueden leer la misma evidencia y
# decidir distinto sin que nada esté roto — con la vara vieja T1 fallaba SIEMPRE
# y la corrida se declaraba fallida por la razón equivocada.

set -uo pipefail
cd "$(dirname "$0")"

a="corridas/${1:-A}"
b="corridas/${2:-B}"

for d in "$a" "$b"; do
  [ -d "$d" ] || { echo "no existe $d"; exit 1; }
done

# campo saca un número del frontmatter de un artefacto. Vacío si no está —
# y "no está" NO es cero: son cosas distintas y el informe las distingue.
campo() { # campo <archivo> <clave>
  [ -f "$1" ] || return 1
  sed -n "s/^$2: *//p" "$1" | head -1
}

echo "════════════════════════════════════════════════════════"
echo " ① LA VARA DE T1 — evidencia, no veredicto"
echo "════════════════════════════════════════════════════════"
pasa=1
for d in "$a" "$b"; do
  n=$(basename "$d")
  ent="$d/.docs/entrevista.md"
  evi="$d/.docs/evidencia.md"

  abiertas=$(campo "$ent" abiertas)
  links=$(grep -o 'https\?://' "$evi" 2>/dev/null | wc -l | tr -d ' ')
  baja=$(campo "$evi" evidencia)

  # ① la entrevista cerró
  if [ ! -f "$ent" ]; then
    echo "  $n  ✗ no hay entrevista.md"; pasa=0
  elif [ -z "$abiertas" ]; then
    echo "  $n  ✗ entrevista.md no declara \`abiertas\` — no se sabe si terminó"; pasa=0
  elif [ "$abiertas" != "0" ]; then
    echo "  $n  ✗ la entrevista cerró con $abiertas rama(s) abierta(s)"; pasa=0
  else
    echo "  $n  ✓ la entrevista cerró con la frontera vacía"
  fi

  # ② la evidencia cita
  if [ ! -f "$evi" ]; then
    echo "      ✗ no hay evidencia.md"; pasa=0
  elif [ "$baja" = "baja" ]; then
    echo "      ⚠ evidencia declarada BAJA — pasada degradada, no cuenta como investigar"
  elif [ "$links" -eq 0 ]; then
    echo "      ✗ evidencia.md sin un solo link"; pasa=0
  else
    echo "      ✓ evidencia.md cita $links link(s)"
  fi
done
echo
if [ "$pasa" -eq 1 ]; then
  echo "  ✓ T1 PASA — los dos investigaron y los dos cerraron la entrevista."
  echo "    Era la herramienta, no el modelo. Causa raíz del 05-09 cerrada."
else
  echo "  ✗ T1 NO PASA. Y la pregunta es siempre la misma:"
  echo "    ¿el skill pidió mal, o la compuerta exige de más?"
  echo "    Lo que NO se hace es aflojar la compuerta para que pase."
fi

echo
echo "════════════════════════════════════════════════════════"
echo " ② EL VEREDICTO — se mira, y NO es la vara"
echo "════════════════════════════════════════════════════════"
echo "  Que uno diga hacelo y el otro no-lo-hagas sobre la misma evidencia"
echo "  no es un fallo: es el ⑥ haciendo su trabajo."
for d in "$a" "$b"; do
  v=$(campo "$d/.docs/brief.md" veredicto)
  echo "  $(basename "$d"): ${v:-(no hay brief.md)}"
done

echo
echo "════════════════════════════════════════════════════════"
echo " ③ LA EVIDENCIA — ¿investigó, o inventó?"
echo "════════════════════════════════════════════════════════"
echo "  El frontmatter DECLARA; el cuerpo es el HECHO. Si no cierran, el que"
echo "  vale es el cuerpo — y la diferencia es en sí misma un hallazgo."
for d in "$a" "$b"; do
  n=$(basename "$d")
  evi="$d/.docs/evidencia.md"
  if [ ! -f "$evi" ]; then echo "  $n: (no hay evidencia.md)"; continue; fi

  r=$(campo "$evi" retrieved);   r=${r:-—}
  m=$(campo "$evi" model_prior); m=${m:-—}
  p=$(campo "$evi" probado);     p=${p:-—}
  ld=$(campo "$evi" links);      ld=${ld:-—}
  lr=$(grep -o 'https\?://' "$evi" | wc -l | tr -d ' ')

  aviso=""
  [ "$ld" != "—" ] && [ "$ld" != "$lr" ] && aviso="   ← ⚠ declara $ld y hay $lr"
  echo "  $n: retrieved=$r  model-prior=$m  probado=$p  links(reales)=$lr$aviso"
done

echo
echo "════════════════════════════════════════════════════════"
echo " ④ LOS ARTEFACTOS — qué dejó cada uno"
echo "════════════════════════════════════════════════════════"
echo "  vocabulario.md es PEREZOSO: que no esté NO es una falta."
for d in "$a" "$b"; do
  echo -n "  $(basename "$d"): "
  for f in brief.md entrevista.md evidencia.md vocabulario.md; do
    [ -f "$d/.docs/$f" ] && echo -n "✓ $f  " || echo -n "·  $f  "
  done
  n=$(ls -d "$d/.docs/prototipos"/*/ 2>/dev/null | wc -l | tr -d ' ')
  echo "· prototipos: $n"
done

echo
echo "════════════════════════════════════════════════════════"
echo " ⑤ LA PELÍCULA — cuántas vueltas dio cada uno"
echo "════════════════════════════════════════════════════════"
for d in "$a" "$b"; do
  n=$(wc -l < "$d/.specforge/registro.jsonl" 2>/dev/null || echo 0)
  echo "  --- $(basename "$d") — $n líneas de registro ---"
  (cd "$d" && sf log 2>/dev/null | tail -15 | sed 's/^/    /')
done

echo
echo "════════════════════════════════════════════════════════"
echo " ⑥ EL ESTADO — informativo en T1, la vara de T2 en adelante"
echo "════════════════════════════════════════════════════════"
if diff -u "$a/.docs/estado.json" "$b/.docs/estado.json"; then
  echo "  ✓ idénticos"
else
  echo
  echo "  En T1 esto es NORMAL: el único campo que se mueve es el veredicto."
  echo "  De T2 en adelante un diff acá SÍ es el hallazgo."
fi

echo
echo "════════════════════════════════════════════════════════"
echo " ⑦ PARA LEER CON LOS OJOS"
echo "════════════════════════════════════════════════════════"
echo "  diff -y --width=200 $a/.docs/brief.md      $b/.docs/brief.md      | less"
echo "  diff -y --width=200 $a/.docs/evidencia.md  $b/.docs/evidencia.md  | less"
echo "  diff -y --width=200 $a/.docs/entrevista.md $b/.docs/entrevista.md | less"
