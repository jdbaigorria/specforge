#!/bin/sh
#
# Instala el binario `sf`.
#
#   curl -fsSL https://raw.githubusercontent.com/jdbaigorria/specforge/main/install.sh | sh
#
# ─────────────────────────────────────────────────────────────────────────────
# POR QUÉ ESTO ES SHELL Y NO UN COMANDO DE sf
# ─────────────────────────────────────────────────────────────────────────────
#
# Porque `sf` no puede instalar a `sf`. Todo lo que ponga el binario tiene que
# correr ANTES de que el binario exista, y lo único que se puede dar por
# supuesto en una máquina ajena es un shell.
#
# Por eso mismo esto instala UNA sola cosa. Los 22 skills son Markdown y los
# trae el plugin del harness, que ya sabe versionarlos y actualizarlos:
#
#   /plugin marketplace add jdbaigorria/specforge
#   /plugin install specforge
#
# Las dos mitades pueden quedar desalineadas, y de eso se ocupa `sf doctor`.
#
# ─────────────────────────────────────────────────────────────────────────────
# VARIABLES
# ─────────────────────────────────────────────────────────────────────────────
#
#   SF_VERSION    la versión a instalar. Por defecto, la última publicada.
#   SF_BIN_DIR    dónde ponerlo. Por defecto ~/.local/bin.
#
set -eu

REPO="jdbaigorria/specforge"
BIN_DIR="${SF_BIN_DIR:-$HOME/.local/bin}"

# ─────────────────────────────────────────────────────────────────────────────
# Salida
# ─────────────────────────────────────────────────────────────────────────────

di()    { printf '%s\n' "$*"; }
paso()  { printf '\033[2m·\033[0m %s\n' "$*"; }
ok()    { printf '\033[32m✓\033[0m %s\n' "$*"; }

# muero imprime en stderr y corta. El mensaje dice qué pasó, no "error".
muero() {
  printf '\033[31m✗\033[0m %s\n' "$*" >&2
  exit 1
}

hay() { command -v "$1" >/dev/null 2>&1; }

# ─────────────────────────────────────────────────────────────────────────────
# El temporal, uno solo
# ─────────────────────────────────────────────────────────────────────────────
#
# Antes cada funcion armaba el suyo con su propio `trap ... EXIT`. Un shell
# tiene UN trap de EXIT: el segundo pisa al primero, asi que cuando la descarga
# fallaba y se caia a compilar, el temporal de la descarga quedaba en /tmp para
# siempre. Y ese es justamente el camino normal mientras no haya releases.
TMP=""
limpiar() { [ -n "$TMP" ] && rm -rf "$TMP"; }

# ─────────────────────────────────────────────────────────────────────────────
# Qué máquina es ésta
# ─────────────────────────────────────────────────────────────────────────────

detectar_plataforma() {
  os=$(uname -s | tr '[:upper:]' '[:lower:]')
  arch=$(uname -m)

  case "$os" in
    linux|darwin) ;;
    # Git Bash y MSYS reportan cosas distintas y las dos son Windows.
    mingw*|msys*|cygwin*) os=windows ;;
    *) muero "no tengo binario para $os. Compilá desde el repo: https://github.com/$REPO" ;;
  esac

  case "$arch" in
    x86_64|amd64)  arch=amd64 ;;
    arm64|aarch64) arch=arm64 ;;
    *) muero "no tengo binario para $arch. Compilá desde el repo: https://github.com/$REPO" ;;
  esac

  PLATAFORMA="${os}_${arch}"
  EXT=tar.gz
  SUFIJO=""
  if [ "$os" = windows ]; then
    EXT=zip
    SUFIJO=".exe"
  fi
}

# ─────────────────────────────────────────────────────────────────────────────
# Qué versión
# ─────────────────────────────────────────────────────────────────────────────

# ultima_version pregunta a la API de GitHub cuál es el último release.
#
# Sin `jq`: se resuelve con sed porque pedirle a alguien que instale jq para
# poder instalar sf es exactamente el tipo de fricción que este script existe
# para sacar.
ultima_version() {
  bajar_a_stdout "https://api.github.com/repos/$REPO/releases/latest" |
    sed -n 's/.*"tag_name" *: *"\([^"]*\)".*/\1/p' |
    head -1
}

bajar_a_stdout() {
  if hay curl; then
    curl -fsSL "$1"
  elif hay wget; then
    wget -qO- "$1"
  else
    muero "necesito curl o wget"
  fi
}

bajar_a_archivo() {
  if hay curl; then
    curl -fsSL -o "$2" "$1"
  elif hay wget; then
    wget -qO "$2" "$1"
  else
    muero "necesito curl o wget"
  fi
}

# ─────────────────────────────────────────────────────────────────────────────
# El plan B: compilar
# ─────────────────────────────────────────────────────────────────────────────
#
# Si no hay release para esta plataforma pero sí hay Go, se compila. No es un
# caso raro: mientras el proyecto no publique releases, ES el único camino.
compilar() {
  hay go || muero "no encontré un release para $PLATAFORMA y tampoco Go para compilar.
  Instalá Go 1.22+ (https://go.dev/dl/) o compilá en otra máquina:
      GOOS=${PLATAFORMA%_*} GOARCH=${PLATAFORMA#*_} go build -o sf ./sf/cmd/sf"

  paso "compilando desde el código (no hay release para $PLATAFORMA)"
  mkdir -p "$BIN_DIR"

  # Primero por el proxy de módulos, que es lo más liviano.
  if GOBIN="$BIN_DIR" go install "github.com/$REPO/sf/cmd/sf@${SF_VERSION:-latest}" 2>/dev/null; then
    return 0
  fi

  # Y si el proxy todavía no lo tiene —el caso normal antes del primer
  # release—, clonando. Es más lento y funciona siempre, que es lo que hace
  # falta de un plan B.
  paso "el proxy de módulos no lo tiene todavía; clonando"
  hay git || muero "necesito git para clonar. O instalá Go y esperá al primer release."

  tmp="$TMP"

  rama="${SF_VERSION:-}"
  if [ -n "$rama" ]; then
    git clone --depth 1 --branch "$rama" "https://github.com/$REPO" "$tmp/repo" >/dev/null 2>&1 ||
      muero "no pude clonar $REPO en la versión $rama"
  else
    git clone --depth 1 "https://github.com/$REPO" "$tmp/repo" >/dev/null 2>&1 ||
      muero "no pude clonar https://github.com/$REPO"
  fi

  (cd "$tmp/repo/sf" && go build -o "$BIN_DIR/sf$SUFIJO" ./cmd/sf) ||
    muero "cloné el repo y no compiló. Es un bug: contalo en https://github.com/$REPO/issues"
}

# ─────────────────────────────────────────────────────────────────────────────
# Instalar
# ─────────────────────────────────────────────────────────────────────────────

verificar_suma() {
  # tarball · sumas · nombre dentro del archivo de sumas
  esperada=$(grep " $3\$" "$2" 2>/dev/null | cut -d' ' -f1) || true
  [ -n "${esperada:-}" ] || { paso "el release no trae checksums; sigo sin verificar"; return 0; }

  if hay sha256sum;   then real=$(sha256sum "$1" | cut -d' ' -f1)
  elif hay shasum;    then real=$(shasum -a 256 "$1" | cut -d' ' -f1)
  else paso "no tengo sha256sum ni shasum; sigo sin verificar"; return 0
  fi

  [ "$real" = "$esperada" ] ||
    muero "el checksum no da: bajé algo que no es lo que el release publica.
  esperado  $esperada
  bajado    $real"
  paso "checksum ok"
}

desde_release() {
  tag="$1"
  nombre="sf_${tag}_${PLATAFORMA}.${EXT}"
  base="https://github.com/$REPO/releases/download/$tag"

  tmp="$TMP"

  # El 2>/dev/null va ACA y no envolviendo a la funcion entera.
  #
  # `curl -fsSL` lleva -S —show errors—, asi que grita cuando el release no
  # existe para esta plataforma. Ese grito hay que taparlo: no encontrar el
  # release es normal y el que sigue es el plan B.
  #
  # Pero cuando el silencio envolvia a toda la funcion se llevaba puesto TODO
  # lo que `muero` tiene para decir aca adentro —el checksum que no da, el
  # release sin ejecutable, el unzip que falta—, y como `muero` ademas hace
  # exit, el que instalaba veia "bajando ..." y despues nada. Codigo 1, cero
  # explicacion, sobre el unico chequeo de seguridad del script.
  paso "bajando $nombre"
  bajar_a_archivo "$base/$nombre" "$tmp/$nombre" 2>/dev/null || return 1

  bajar_a_archivo "$base/SHA256SUMS" "$tmp/SHA256SUMS" 2>/dev/null || true
  verificar_suma "$tmp/$nombre" "$tmp/SHA256SUMS" "$nombre"

  if [ "$EXT" = zip ]; then
    hay unzip || muero "necesito unzip para abrir el release de Windows"
    unzip -qo "$tmp/$nombre" -d "$tmp"
  else
    tar -xzf "$tmp/$nombre" -C "$tmp"
  fi

  [ -f "$tmp/sf$SUFIJO" ] || muero "el release no trae un ejecutable \`sf$SUFIJO\` adentro"

  mkdir -p "$BIN_DIR"
  # Se instala con `mv` sobre el destino final y no editando el archivo en su
  # lugar: si alguien está corriendo el sf viejo justo ahora, reemplazar el
  # inodo no le rompe la corrida.
  chmod +x "$tmp/sf$SUFIJO"
  mv -f "$tmp/sf$SUFIJO" "$BIN_DIR/sf$SUFIJO"
}

# ─────────────────────────────────────────────────────────────────────────────
# Después
# ─────────────────────────────────────────────────────────────────────────────

# avisar_del_path es la mitad que se olvida y es la que más molesta.
#
# Instalar el binario donde el shell no lo busca produce exactamente el síntoma
# más confuso posible: `sf: command not found` justo después de un instalador
# que dijo ✓.
avisar_del_path() {
  case ":$PATH:" in
    *":$BIN_DIR:"*) return 0 ;;
  esac

  di ""
  di "⚠ $BIN_DIR no está en tu PATH. Agregá esta línea a tu ~/.zshrc o ~/.bashrc:"
  di ""
  di "    export PATH=\"$BIN_DIR:\$PATH\""
  di ""
  di "  y abrí una terminal nueva."
}

# avisar_del_viejo cubre el caso que motivó `sf doctor`: hay OTRO sf adelante en
# el PATH, así que el que va a correr el agente no es el que se acaba de
# instalar. Es silencioso y confunde durante horas.
avisar_del_viejo() {
  otro=$(command -v sf 2>/dev/null) || return 0
  [ -n "$otro" ] || return 0
  [ "$otro" != "$BIN_DIR/sf$SUFIJO" ] || return 0

  di ""
  di "⚠ ya había otro \`sf\` adelante en el PATH, y es el que se va a usar:"
  di ""
  di "    $otro     ← éste gana"
  di "    $BIN_DIR/sf$SUFIJO     ← el que acabo de instalar"
  di ""
  di "  Sacá el viejo, o poné $BIN_DIR primero en el PATH."
}

main() {
  detectar_plataforma

  TMP=$(mktemp -d)
  trap limpiar EXIT INT TERM

  tag="${SF_VERSION:-}"
  if [ -z "$tag" ]; then
    paso "buscando la última versión"
    tag=$(ultima_version 2>/dev/null) || true
  fi

  if [ -n "${tag:-}" ] && desde_release "$tag"; then
    ok "sf $tag → $BIN_DIR/sf$SUFIJO"
  else
    compilar
    ok "sf → $BIN_DIR/sf$SUFIJO"
  fi

  avisar_del_path
  avisar_del_viejo

  di ""
  di "Falta la otra mitad — los 22 skills. En Claude Code:"
  di ""
  di "    /plugin marketplace add $REPO"
  di "    /plugin install specforge"
  di ""
  di "Y después, para comprobar que las dos mitades se ven:"
  di ""
  di "    sf doctor"
}

main "$@"
