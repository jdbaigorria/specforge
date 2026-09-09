#!/usr/bin/env bash
# nuevo.sh — arma una corrida limpia.
#
# ─────────────────────────────────────────────────────────────────────────────
# ADENTRO DE LA CORRIDA VA SÓLO EL PROYECTO
# ─────────────────────────────────────────────────────────────────────────────
#
# Nada que hable del experimento: ni el guion, ni la idea con su explicación, ni
# corridas anteriores, ni el comparador. Todo eso vive acá, en el repo, fuera del
# árbol que el agente puede caminar.
#
# SE ROMPIÓ ASÍ EL 2026-09-08. El banco tenía las corridas viejas renombradas al
# lado, un IDEA.md que explicaba que la idea "seguro ya existe" —o sea, la
# respuesta— y un GUION.md que contaba qué se estaba probando y cómo se pasa. El
# agente las encontró y concluyó solo que ya era momento de implementar.
#
#   ./nuevo.sh A          arma corridas/A
#   ./nuevo.sh A --rehacer  la borra y la vuelve a armar

set -euo pipefail

# El directorio del repo se resuelve ANTES de moverse: más abajo se hace `cd` a
# la corrida y ahí `dirname $0` ya no vale nada.
AQUI="$(cd "$(dirname "$0")" && pwd)"

CORRIDAS="$HOME/projects/workspace/personal/sf-banco/corridas"
# ─────────────────────────────────────────────────────────────────────────────
# CADA CORRIDA VA EN UN DIRECTORIO NUEVO, Y NO ES MANÍA
# ─────────────────────────────────────────────────────────────────────────────
#
# MEDIDO EL 2026-09-08: ICM inventa un proyecto de memoria CON EL NOMBRE DEL
# DIRECTORIO. Correr en `corridas/A` creó el tópico `context-sf-banco-A`, y ahí
# quedó guardado el resultado de esa corrida — incluido un "la corrida A NO
# llegó al PRD, no queda nada".
#
# O sea: CADA CORRIDA EN corridas/A ENVENENA LA SIGUIENTE CORRIDA EN corridas/A.
# Para siempre, aunque cambies la semilla y aunque borres la carpeta: el veneno
# no está en el disco, está indexado bajo el nombre.
#
# Con un nombre nuevo cada vez, el proyecto de memoria arranca vacío. Es la
# única forma de que el arnés no se acuerde sin desinstalarle la memoria.
base="${1:?uso: ./nuevo.sh <a|b> [sufijo]}"
nombre="${base}-$(date +%Y%m%d-%H%M)${2:+-$2}"
d="$CORRIDAS/$nombre"

[ -e "$d" ] && { echo "ya existe $d"; exit 1; }

mkdir -p "$d"
cd "$d"
git init -q .
sf install --harness=claude-code,opencode >/dev/null
sf init >/dev/null
: > .specforge/registro.jsonl

# El diagnóstico de arranque se guarda ACÁ, en el repo — NO adentro de la corrida.
# Es material del experimento: nombra el banco y deja ver que esto es una prueba.
DIAG="$AQUI/diagnosticos"
mkdir -p "$DIAG"
sf doctor > "$DIAG/$nombre-$(date +%Y%m%d-%H%M).txt" 2>&1 || true

echo "listo: $d"
echo "     (nombre nuevo a propósito: la memoria del arnés se indexa por directorio)"
sed -n '/^buscar/,/^$/p' "$DIAG"/$nombre-*.txt | tail -5
