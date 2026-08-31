// Package andamio es `sf install`: lo que hace que SpecForge se pueda usar.
//
// ────────────────────────────────────────────────────────────────────────────
// LA DIFERENCIA ENTRE "FUNCIONA" Y "SE PUEDE INSTALAR"
// ────────────────────────────────────────────────────────────────────────────
//
// El bucle ya cerraba de punta a punta antes de este paquete, pero sólo si
// alguien armaba las dos cosas a mano:
//
//	el orquestador   CLAUDE.md y AGENTS.md en el proyecto
//	~/.specforge/    qué modelos hay y cómo se invoca cada uno
//
// `sf install` es lo que las pone. Y no es de la máquina —no mira el estado ni
// lo mueve—: es andamio, la categoría que `que-sobrevive.md` §1 tuvo que
// declarar aparte porque la vara "¿lo consume alguno de los 9 estados?" la
// dejaba afuera sin ser inútil.
//
// ────────────────────────────────────────────────────────────────────────────
// EL ORQUESTADOR VIAJA EMBEBIDO EN EL BINARIO
// ────────────────────────────────────────────────────────────────────────────
//
// `plantillas/CLAUDE.md` se compila ADENTRO de `sf` con go:embed. La
// alternativa —buscar el archivo en disco relativo al binario— se cae sola: un
// `sf` instalado con `go install` vive en ~/go/bin y no tiene el repo al lado.
//
// Un archivo, dos destinos: se escribe UNA vez y se copia DOS veces, porque el
// diseño dice que AGENTS.md es el mismo texto y no una traducción
// (superficie-sf.md §6). Mantenerlo como dos archivos en el repo sería el mismo
// contenido en dos lugares — justo lo que el proyecto evita en todo lo demás.
package andamio

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jdbaigorria/specforge/sf/internal/global"
)

// orquestador es plantillas/CLAUDE.md, embebido en tiempo de compilación.
//
// La ruta sube dos niveles porque este paquete vive en sf/internal/andamio/ y
// la plantilla en la raíz del repo. Si alguien mueve una de las dos, esto NO
// compila — que es exactamente lo que uno quiere: un puntero roto a la
// plantilla del orquestador tiene que romper el build, no descubrirse en la
// máquina de un usuario.
//
//go:embed plantilla/CLAUDE.md
var orquestador []byte

// Los dos nombres del mismo texto.
//
// Son dos harness distintos leyendo lo mismo: lo único que cambia entre ellos
// es cómo se lanza un subagente, y eso el harness ya lo sabe hacer.
var destinos = []string{"CLAUDE.md", "AGENTS.md"}

// Resultado es qué hizo la instalación.
type Resultado struct {
	// Escritos son los archivos que se crearon o actualizaron.
	Escritos []string

	// Salteados son los que ya existían y NO se pisaron, con su motivo.
	Salteados []string

	// Viejos son los orquestadores instalados que YA NO son la plantilla.
	//
	// Separado de Salteados porque no es lo mismo: saltear un archivo al día no
	// cuesta nada, y saltear uno viejo deja al harness sin instrucciones que sf
	// da por sentadas.
	Viejos []string

	// Harness es dónde se detectó (o lo que se pidió con --harness).
	Harness string

	// Modelos es cuántos quedaron declarados.
	Modelos int
}

// Opciones son las del comando.
type Opciones struct {
	// Harness pisa la detección. Vacío = detectar.
	Harness string

	// Forzar pisa los archivos del proyecto que ya existan.
	//
	// Existe porque el caso normal es NO pisar —el CLAUDE.md de un proyecto
	// suele tener cosas de Javier— pero actualizar el orquestador cuando cambia
	// tiene que ser posible sin borrar a mano.
	Forzar bool
}

// ErrSinProyecto es que no se puede instalar el orquestador acá.
var ErrSinProyecto = errors.New("no encuentro el proyecto")

// marca es cómo sf reconoce un archivo que escribió él.
//
// NO SE CAMBIA sin cambiar también a mano todo lo que ya está instalado en el
// mundo: es la única forma de distinguir un orquestador viejo —que hay que
// actualizar— del CLAUDE.md propio de Javier —que no se toca jamás—. Sin esta
// distinción, `sf doctor` fallaría en cualquier repo con un CLAUDE.md a mano.
//
// `TestLaPlantillaLlevaLaMarca` rompe si alguien le cambia el título a la
// plantilla, que es justo el descuido que dejaría el chequeo mudo para siempre.
var marca = []byte("# SpecForge\n")

// comoEsta mira un orquestador instalado: si está, si es nuestro, y si está al día.
func comoEsta(ruta string) (hay, nuestro, igual bool) {
	b, err := os.ReadFile(ruta)
	if err != nil {
		return false, false, false
	}
	return true, bytes.HasPrefix(b, marca), bytes.Equal(b, orquestador)
}

// OrquestadoresViejos son los CLAUDE.md/AGENTS.md que quedaron atrás.
//
// Lo usa `sf doctor`, y es FALLA y no aviso: el orquestador es un archivo que
// genera sf, no de Javier, y el contrato entre `sf next` y él está versionado
// por su contenido. Uno viejo no se rompe ruidosamente —sigue corriendo el
// bucle— sino que ignora en silencio lo que no entiende, que es la forma más
// cara de fallar. Un archivo que no existe no se cuenta acá —eso es un proyecto
// sin `sf install`, y de eso avisa otro chequeo— y uno SIN LA MARCA tampoco: ése
// es el CLAUDE.md propio de Javier, y reclamarle que se actualice sería
// reclamarle por un archivo que sf nunca escribió.
func OrquestadoresViejos(raiz string) []string {
	var viejos []string
	for _, nombre := range destinos {
		if hay, nuestro, igual := comoEsta(filepath.Join(raiz, nombre)); hay && nuestro && !igual {
			viejos = append(viejos, nombre)
		}
	}
	return viejos
}

// harnessDeLaInstalacion es para cuál harness se está instalando.
//
// `EnUso()` y no `g.Harness` por el mismo motivo que en Regenerar: el puntero
// escrito es el último que se instaló, no dónde estás parado ahora. Con el
// escrito, un `sf install` desde Claude Code sobre un catálogo que decía
// "opencode" armaba el andamio de opencode —.opencode/, sus portamodelo— y no
// escribía nada de lo que el harness de verdad necesitaba.
//
// El flag sigue ganando: `sf install --harness=opencode` corrido desde otro
// lado es un acto deliberado —"preparame el otro"— y es la única forma de
// corregir una detección que salió mal.
//
// Bajar a "desconocido" no es un riesgo: EnUso() cae en el escrito antes de
// devolverlo, así que un catálogo con harness y sin detección se queda como está.
func harnessDeLaInstalacion(o Opciones, g *global.Config) string {
	if o.Harness != "" {
		return o.Harness
	}
	return g.EnUso()
}

// Instalar pone el orquestador en el proyecto y arma ~/.specforge/.
//
// Las dos mitades son independientes a propósito: se puede correr en una
// máquina nueva sobre un proyecto que ya tenía CLAUDE.md, o en un proyecto
// nuevo con el ~/.specforge/ ya armado. Ninguna pisa lo que la otra hizo.
func Instalar(raiz string, o Opciones) (*Resultado, error) {
	r := &Resultado{}

	// ① el orquestador, en los dos nombres
	for _, nombre := range destinos {
		ruta := filepath.Join(raiz, nombre)

		if hay, nuestro, igual := comoEsta(ruta); hay && !o.Forzar {
			// NO se pisa, y no es timidez: el CLAUDE.md de un proyecto suele
			// tener instrucciones propias de Javier. Pisarlo sin avisar sería
			// borrarle trabajo para poner cuatro líneas.
			//
			// Pero "existe" y "está al día" son dos cosas distintas, y decir
			// sólo la primera es lo que le costó una hora a Javier: reinstaló
			// todo en un proyecto viejo, el renglón dijo "ya existe" —que se lee
			// como "no hay nada que hacer"— y se quedó con un orquestador que no
			// sabía invocar el `agente:`. opencode hizo el PRD con el modelo
			// principal y nadie tenía cómo enterarse.
			switch {
			case igual:
				r.Salteados = append(r.Salteados, nombre+" (ya está al día)")
			case nuestro:
				r.Viejos = append(r.Viejos, nombre)
			default:
				// El de Javier. Ni se pisa ni se le reclama nada.
				r.Salteados = append(r.Salteados, nombre+" (ya existe — `--forzar` lo pisa)")
			}
			continue
		}

		if err := os.WriteFile(ruta, orquestador, 0o644); err != nil {
			return nil, fmt.Errorf("escribiendo %s: %w", nombre, err)
		}
		r.Escritos = append(r.Escritos, nombre)
	}

	// ② ~/.specforge/
	//
	// Si ya existe NO se toca: ahí viven los modelos que Javier fue aprobando
	// una aprobación a la vez, y re-sembrarlo sería tirar esa lista.
	g, err := global.Leer()
	switch {
	case errors.Is(err, global.ErrNoHay):
		g = global.Semilla(harnessDeLaInstalacion(o, nil))
		if err := g.Guardar(); err != nil {
			return nil, err
		}
		r.Escritos = append(r.Escritos, "~/.specforge/"+global.Archivo)

	case err != nil:
		return nil, err

	default:
		// Ya existía. Lo único que se actualiza es el harness, y se actualiza
		// solo: instalar ES decir "dejámelo listo ACÁ", así que el puntero
		// escrito queda apuntando adonde se instaló.
		if h := harnessDeLaInstalacion(o, g); h != g.Harness {
			g.Harness = h
			if err := g.Guardar(); err != nil {
				return nil, err
			}
			r.Escritos = append(r.Escritos, "~/.specforge/"+global.Archivo+" (harness → "+h+")")
		} else {
			r.Salteados = append(r.Salteados, "~/.specforge/"+global.Archivo+" (ya existe — tus modelos quedan)")
		}
	}

	// Después del switch g.Harness YA es el de esta instalación en las tres
	// ramas, y de ahí en adelante hay uno solo. El conteo va por él y no por
	// `Alias()` —que mira el EN USO— porque con `--harness` los dos difieren, y
	// mezclarlos imprimía un encabezado que hablaba de dos harness a la vez:
	// "harness: opencode · 0 modelos declarados", donde el cero era el de otro.
	r.Harness = g.Harness
	r.Modelos = len(g.AliasDe(g.Harness))

	// ③ lo que necesita EL HARNESS: permisos, dónde están los skills, y los
	// portamodelo. Las tres salieron de medir, no de leer — ver harness.go.
	if err := escribirPermisos(raiz, g.Harness, o.Forzar, r); err != nil {
		return nil, err
	}
	if err := escribirPortamodelos(raiz, g, g.Harness, r); err != nil {
		return nil, err
	}
	return r, nil
}

// Regenerar reescribe los portamodelo sin tocar nada más.
//
// Lo llama `sf model` cuando declara un alias nuevo: el archivo tiene que
// aparecer en el momento, aunque la sesión viva no lo vaya a ver hasta que se
// reinicie —que es justo lo que el aviso de `sf model` dice—.
func Regenerar(raiz string, g *global.Config) (*Resultado, error) {
	// Acá va el que está EN USO y no el escrito, y la distinción es la misma
	// que separa instalar de resolver: `sf model` declara donde estás PARADO,
	// así que el portamodelo tiene que aparecer en ese mismo arnés. Con el
	// escrito, declarar en Command Code mientras `sf install` había sido para
	// Claude Code dejaba el modelo en el catálogo y el archivo sin escribir —
	// y el próximo `sf next` nombraba un agente que no existía.
	h := g.EnUso()
	r := &Resultado{Harness: h, Modelos: len(g.AliasDe(h))}
	if err := escribirPortamodelos(raiz, g, h, r); err != nil {
		return nil, err
	}
	return r, nil
}

// Desinstalar saca el orquestador del proyecto.
//
// NO toca ~/.specforge/, y es a propósito: los modelos son de la máquina, no
// del proyecto. Sacar SpecForge de un repo no puede borrarle a Javier la lista
// que fue construyendo en los otros quince.
func Desinstalar(raiz string) (*Resultado, error) {
	r := &Resultado{}
	for _, nombre := range destinos {
		ruta := filepath.Join(raiz, nombre)

		// Sólo se borra lo que escribimos nosotros. Un CLAUDE.md que alguien
		// editó tiene cosas que no pusimos, y borrarlo sería tirar eso.
		b, err := os.ReadFile(ruta)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, err
		}
		if string(b) != string(orquestador) {
			r.Salteados = append(r.Salteados, nombre+" (lo editaste — no lo toco)")
			continue
		}
		if err := os.Remove(ruta); err != nil {
			return nil, err
		}
		r.Escritos = append(r.Escritos, nombre+" (borrado)")
	}
	return r, nil
}
