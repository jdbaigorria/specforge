// Package arranque es `sf init`: lo único que hay que correr antes que nada.
//
// ────────────────────────────────────────────────────────────────────────────
// SON DOS COSAS, Y LAS DOS SON MECÁNICAS
// ────────────────────────────────────────────────────────────────────────────
//
//	el scaffold      2 directorios y un estado.json vacío
//	detectStack()    mirar qué manifiesto hay y llenar lenguaje · manifiesto ·
//	                 test_cmd
//
// Las dos entran por R1 —"sf hace todo lo que tiene UNA SOLA respuesta
// correcta"—: qué directorios crea el flujo no es opinable, y "hay un go.mod,
// entonces el comando de tests es `go test ./...`" tampoco.
//
// ────────────────────────────────────────────────────────────────────────────
// POR QUÉ ESCRIBE LA CONSTITUCIÓN A MEDIAS, Y NO ES UN ATAJO
// ────────────────────────────────────────────────────────────────────────────
//
// `sf init` deja `.docs/constitucion.md` con la CABECERA LLENA y el cuerpo en
// blanco. El cuerpo lo escribe el ⑧ (el skill sfp-constitucion), conversando.
//
// El reparto sale de la regla, no del gusto:
//
//	"hay un go.mod"      →  una sola respuesta correcta   →  sf     (R1)
//	"¿qué arquitectura?" →  varias respuestas             →  el LLM
//
// Y hay una consecuencia práctica: sin este archivo, el sobre del ⑧ no tendría
// qué servir. El sobre dice "la cabecera técnica ya está llena" — este paquete
// es lo que hace que eso sea verdad (H18).
//
// ────────────────────────────────────────────────────────────────────────────
// DOS O TRES DIRECTORIOS, NO SIETE
// ────────────────────────────────────────────────────────────────────────────
//
// El CLI viejo creaba siete. Acá se crean DOS —.docs/ y .docs/backlog/— y los
// otros aparecen cuando hacen falta:
//
//	.docs/features/    lo crea el ⑫, cuando hay una feature que planificar
//	.docs/archivado/   lo crea `sf approve`, cuando hay algo que archivar
//
// Un directorio vacío que existe desde el día uno es una promesa que el
// proyecto todavía no puede cumplir: alguien lo abre, lo ve vacío y no sabe si
// está roto o si no llegó.
package arranque

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jdbaigorria/specforge/sf/internal/docs"
	"github.com/jdbaigorria/specforge/sf/internal/estado"
)

// Resultado es qué hizo el arranque, para poder contárselo a quien lo corrió.
type Resultado struct {
	// Creados son las rutas que este arranque escribió, en orden.
	Creados []string

	// Stack es lo que se detectó, o el cero si no se reconoció el proyecto.
	Stack Stack
}

// Stack es lo que detectStack() puede contestar mirando el disco.
//
// Los tres campos son exactamente los tres del frontmatter que sf necesita para
// no quedar ciego. No hay un cuarto: `mutacion` y `git` son criterio, y el
// criterio es del ⑧.
type Stack struct {
	Lenguaje   string
	Manifiesto string
	TestCmd    string
}

// Reconocido dice si el proyecto se pudo identificar.
func (s Stack) Reconocido() bool { return s.Lenguaje != "" }

// ErrYaIniciado es que este proyecto ya tiene estado.
//
// Se distingue con errors.Is en vez de devolver un error de texto porque quien
// llama tiene algo útil que decir en ese caso —"ya está, seguí con sf next"— y
// eso no es lo mismo que "no pude escribir el archivo".
var ErrYaIniciado = errors.New("este proyecto ya está iniciado")

// Iniciar arma el andamio del proyecto en raiz.
//
// Es idempotente en lo que importa: si ya hay un estado.json, NO TOCA NADA y
// devuelve ErrYaIniciado. Pisar el estado sería borrar todo lo aprobado —los
// sellos del ⑥ y del ⑧, los lotes cerrados— y ninguno de esos datos se puede
// reconstruir mirando los archivos: por eso viven ahí (maquina-estados.md §9).
func Iniciar(raiz string) (*Resultado, error) {
	var r Resultado

	// El chequeo va PRIMERO, antes de crear ningún directorio. Si abortamos a
	// mitad de camino dejaríamos el proyecto con carpetas nuevas y sin estado,
	// que es un estado intermedio que después hay que limpiar a mano.
	if _, err := os.Stat(filepath.Join(raiz, estado.Archivo)); err == nil {
		return nil, ErrYaIniciado
	}

	// ① los dos directorios
	for _, d := range []string{docs.Base, docs.Backlog} {
		if err := os.MkdirAll(filepath.Join(raiz, d), 0o755); err != nil {
			return nil, fmt.Errorf("creando %s: %w", d, err)
		}
		r.Creados = append(r.Creados, d+"/")
	}

	// ② la constitución con la cabecera llena
	//
	// Se saltea si ya existe, y el caso es real: un proyecto que ya tenía
	// constitución y perdió el estado.json (lo borró, cambió de máquina). Pisar
	// el archivo ahí sería tirar el trabajo del ⑧.
	r.Stack = Detectar(raiz)
	if _, err := os.Stat(filepath.Join(raiz, docs.Constitucion)); os.IsNotExist(err) {
		if err := os.WriteFile(
			filepath.Join(raiz, docs.Constitucion),
			[]byte(esqueletoConstitucion(r.Stack)),
			0o644,
		); err != nil {
			return nil, fmt.Errorf("escribiendo %s: %w", docs.Constitucion, err)
		}
		r.Creados = append(r.Creados, docs.Constitucion)
	}

	// ③ el estado vacío
	//
	// Va ÚLTIMO a propósito: es lo que hace que `sf next` deje de decir "corré
	// sf init". Si algo falla antes, el proyecto sigue sin iniciar y volver a
	// correr el comando arranca de nuevo en vez de quedar a medias.
	//
	// El mapa se inicializa acá y no se deja en nil porque escribir en un mapa
	// nil es un panic en Go, y el primer `sf take` escribe uno.
	e := &estado.Estado{Features: map[string]*estado.Feature{}}
	if err := e.Guardar(raiz); err != nil {
		return nil, err
	}
	r.Creados = append(r.Creados, estado.Archivo)

	return &r, nil
}

// ────────────────────────────────────────────────────────────────────────────
// detectStack — lo único que se rescató del CLI viejo con nombre y apellido
// ────────────────────────────────────────────────────────────────────────────

// manifiestos es la tabla de detección, en orden de prioridad.
//
// Es una LISTA y no un mapa por una razón concreta: el orden importa. Un
// proyecto Python con un package.json para el front tiene los dos manifiestos, y
// el que gana tiene que ser el mismo siempre — un mapa en Go se recorre en orden
// aleatorio y la detección daría distinto en cada corrida.
var manifiestos = []struct {
	Archivo  string
	Lenguaje string
	TestCmd  string
}{
	{"go.mod", "go", "go test ./..."},
	{"Cargo.toml", "rust", "cargo test"},
	{"pyproject.toml", "python", "pytest -q"},
	{"pytest.ini", "python", "pytest -q"},
	{"setup.py", "python", "pytest -q"},
	{"requirements.txt", "python", "pytest -q"},
	{"package.json", "node", "npm test"},
	{"Gemfile", "ruby", "bundle exec rspec"},
	{"pom.xml", "java", "mvn -q test"},
	{"build.gradle", "java", "gradle test"},
}

// Detectar mira qué manifiesto hay y devuelve el stack.
//
// Si no reconoce nada devuelve el cero, y eso NO es un error: un repo vacío
// todavía no tiene manifiesto, y el ⑧ va a llenar el test_cmd a mano. Lo que sí
// importa es que quien llama lo diga, porque sin test_cmd la constitución no
// sella.
func Detectar(raiz string) Stack {
	for _, m := range manifiestos {
		if _, err := os.Stat(filepath.Join(raiz, m.Archivo)); err == nil {
			return Stack{Lenguaje: m.Lenguaje, Manifiesto: m.Archivo, TestCmd: m.TestCmd}
		}
	}
	return Stack{}
}

// esqueletoConstitucion arma el archivo con la cabecera llena y el cuerpo vacío.
//
// El cuerpo lleva los títulos de las secciones y nada más. Es divulgación
// progresiva aplicada a un artefacto: el skill del ⑧ sabe qué va en cada una, y
// repetirlo acá sería mantener el mismo texto en dos lugares.
func esqueletoConstitucion(s Stack) string {
	// Cuando no se detectó nada se deja el campo vacío CON el comentario que
	// dice qué falta. Un `test_cmd: ""` mudo obligaría a ir a buscar por qué la
	// constitución no sella; con el comentario, el que abre el archivo ya sabe.
	aviso := ""
	if !s.Reconocido() {
		aviso = "   # ⚠ no reconocí el proyecto: completá esto o el ⑧ no sella"
	}

	return fmt.Sprintf(`---
lenguaje: %s
manifiesto: %s
test_cmd: %s%s
mutacion: ""
dependencias_aprobadas: []
git:
  branch_por_feature: true
  patron_branch: "feat/{feature-id}-{slug}"
  commit: conventional
  merge: no-ff
  branch_base: main
---

# Constitución

<Esto lo escribe el ⑧. Corré: sf next>

## Arquitectura

## Stack y por qué

## Convenciones de código

## Estructura de carpetas

## Reglas de trabajo
`, s.Lenguaje, s.Manifiesto, s.TestCmd, aviso)
}
