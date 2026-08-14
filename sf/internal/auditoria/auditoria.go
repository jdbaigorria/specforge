// Package auditoria es `sf audit`: el punta a punta, y el único que mira VARIAS
// features a la vez.
//
// ────────────────────────────────────────────────────────────────────────────
// QUÉ VE ESTO QUE EL ㉑ NO PUEDE VER
// ────────────────────────────────────────────────────────────────────────────
//
//	sf-check (㉑)   UNA feature · contra SUS criterios · justo al terminarla
//	sf audit        VARIAS features · punta a punta · cuando Javier quiere
//
// La diferencia no es de rigor, es de DÓNDE ESTÁ PARADO CADA UNO. El ㉑ revisa
// f-2 el día que f-2 termina, y después nadie vuelve a mirarla nunca. Entonces
// hay tres cosas que estructuralmente no puede ver:
//
//	· f-4 rompió un criterio de f-1        — nadie re-revisa una feature cerrada
//	· f-2 y f-3 pasaron solas y no se integran
//	· un criterio que se marcó `cumple` y hoy ya no es cierto
//
// ────────────────────────────────────────────────────────────────────────────
// Y "EL IMPLEMENTADOR NO MINTIÓ" SE PUEDE CONTAR
// ────────────────────────────────────────────────────────────────────────────
//
// Es lo que hace que esto no sea sólo "leé todo con un modelo grande". Cuatro
// de las cinco preguntas del auditor tienen UNA SOLA RESPUESTA CORRECTA, así
// que las contesta sf (R1) y el modelo queda libre para la única que es juicio:
// ¿el código hace de verdad lo que la historia pedía?
//
//	① ¿qué features entran?           roadmap.json + estado.json
//	② ¿qué criterios tienen?           los us-#
//	③ ¿qué dijo la revisión?           los revision.json
//	④ ¿los tests que lo probaban        tareas.json + buscarlos en el repo
//	   SIGUEN EXISTIENDO?
//	⑤ ¿la suite pasa hoy?              test_cmd
//
// La ④ es la que atrapa la mentira, y no es una opinión:
//
//	"la revisión de f-1 dio us-3/CA-2 por cumplido, y su test ya no existe"
//
// ────────────────────────────────────────────────────────────────────────────
// SIGUE SIN ROMPER NINGUNA REGLA
// ────────────────────────────────────────────────────────────────────────────
//
// Contar veredictos, buscar nombres de test y correr una suite no es pensar. sf
// no dice si la auditoría pasó: dice qué encontró contando. El veredicto lo
// escribe el modelo grande, en un subagente fresco, con el sobre que esto arma.
package auditoria

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/jdbaigorria/specforge/sf/internal/constitucion"
	"github.com/jdbaigorria/specforge/sf/internal/docs"
	"github.com/jdbaigorria/specforge/sf/internal/estado"
	"github.com/jdbaigorria/specforge/sf/internal/git"
	"github.com/jdbaigorria/specforge/sf/internal/historia"
	"github.com/jdbaigorria/specforge/sf/internal/revision"
	"github.com/jdbaigorria/specforge/sf/internal/roadmap"
	"github.com/jdbaigorria/specforge/sf/internal/sobre"
	"github.com/jdbaigorria/specforge/sf/internal/suite"
	"github.com/jdbaigorria/specforge/sf/internal/tareas"
)

// Informe es lo que `sf audit` devuelve: el material y los hechos.
//
// Van juntos y no en dos comandos porque se leen juntos: un hecho sin el
// artefacto al lado no se puede investigar, y el material sin los hechos es una
// pila de archivos sin señalar dónde mirar.
type Informe struct {
	// Features son los ids en alcance, en orden de roadmap.
	Features []string

	// Sobre es el material a leer. Se reusa el tipo del ⑱ a propósito: el que
	// lo lee es un modelo, y dos formatos de sobre distintos para el mismo
	// lector serían dos formatos que mantener.
	Sobre *sobre.Sobre

	// Hechos es lo que sf comprobó por su cuenta.
	Hechos []Hecho
}

// Nivel es qué tan grave es un hecho.
type Nivel int

const (
	// Ok es algo que se comprobó y está bien. Acá SÍ se listan los ✓, al revés
	// que en las compuertas: una compuerta contesta "¿podés avanzar?" y ahí lo
	// que salió bien es ruido. Un informe contesta "¿qué encontraste?", y "la
	// suite pasa" es parte de la respuesta.
	Ok Nivel = iota
	// Aviso es algo que hay que mirar y no es una falla.
	Aviso
	// Falla es una contradicción entre lo que se declaró y lo que hay.
	Falla
)

// Hecho es una comprobación con su resultado.
type Hecho struct {
	Nivel Nivel
	Texto string
}

func (i *Informe) hecho(n Nivel, formato string, args ...any) {
	i.Hechos = append(i.Hechos, Hecho{Nivel: n, Texto: fmt.Sprintf(formato, args...)})
}

// ────────────────────────────────────────────────────────────────────────────
// El alcance
// ────────────────────────────────────────────────────────────────────────────

// Alcance resuelve qué features se auditan.
//
// El módulo NO es un concepto de la máquina, y no hace falta que lo sea: el ⑩
// ya agrupa por "comparten solución técnica", así que UN MÓDULO ES UN CONJUNTO
// DE FEATURES — nombrarlas es nombrarlo. Un campo `modulo` en el roadmap sería
// un concepto nuevo que mantener y que ningún otro estado consume.
//
// Con `ids` vacío se auditan TODAS las que tengan trabajo hecho, que es el caso
// que se corre seguido: "auditá lo que llevamos".
func Alcance(e *estado.Estado, r *roadmap.Roadmap, ids []string) ([]string, error) {
	if r == nil {
		return nil, fmt.Errorf("no hay roadmap.json: no hay nada construido que auditar")
	}

	if len(ids) > 0 {
		for _, id := range ids {
			if _, hay := r.Buscar(id); !hay {
				return nil, fmt.Errorf("%s no está en el roadmap", id)
			}
		}
		// Se devuelven en orden de roadmap y no en el que los tipearon: el
		// orden es información —f-1 se construyó antes que f-3— y el auditor la
		// necesita para saber quién pudo romper a quién.
		return enOrden(r, ids), nil
	}

	var todas []string
	for _, f := range r.Features {
		// Una feature que nunca arrancó no aparece en el mapa (R6), y auditar
		// algo que no se construyó no tiene sentido: no hay código que mirar.
		if _, hay := e.Features[f.ID]; hay {
			todas = append(todas, f.ID)
		}
	}
	if len(todas) == 0 {
		return nil, fmt.Errorf("todavía no se construyó ninguna feature")
	}
	return todas, nil
}

func enOrden(r *roadmap.Roadmap, ids []string) []string {
	var ordenadas []string
	for _, f := range r.Features {
		if slices.Contains(ids, f.ID) {
			ordenadas = append(ordenadas, f.ID)
		}
	}
	return ordenadas
}

// ────────────────────────────────────────────────────────────────────────────
// La auditoría
// ────────────────────────────────────────────────────────────────────────────

// Auditar arma el material y corre las comprobaciones.
func Auditar(raiz string, e *estado.Estado, r *roadmap.Roadmap, ids []string) (*Informe, error) {
	i := &Informe{Features: ids}

	s := &sobre.Sobre{Estado: "audit"}
	s.Partes = append(s.Partes, sobre.Parte{
		Titulo: "Las reglas que todo esto tenía que respetar",
		Rutas:  []string{docs.Constitucion},
	})

	// El alcance va PRIMERO en la lista de hechos, antes que cualquier hallazgo:
	// es el marco de todo lo que sigue. Un ✗ leído sin saber sobre cuántas
	// features se buscó no se puede dimensionar.
	//
	// Se calcula al final —hay que contar los criterios— así que se reserva el
	// lugar acá y se rellena después. Es más simple que dar dos vueltas.
	i.Hechos = append(i.Hechos, Hecho{})

	// El material se junta feature por feature y las comprobaciones se corren
	// en la misma pasada: los dos necesitan lo mismo leído del disco, y
	// separarlos sería leer cada tareas.json dos veces.
	var criterios []string
	for _, id := range ids {
		fr, _ := r.Buscar(id)
		carpeta := carpetaDe(raiz, fr)

		s.Partes = append(s.Partes, sobre.Parte{
			Titulo: fmt.Sprintf("%s — %s", fr.ID, fr.Nombre),
			Rutas: append(
				rutasDeHistorias(fr.Historias),
				filepath.Join(carpeta, docs.Spec),
				filepath.Join(carpeta, docs.Revision),
			),
		})

		delaFeature, err := historia.Criterios(raiz, fr.Historias)
		if err != nil {
			i.hecho(Falla, "%s: %v", id, err)
			continue
		}
		criterios = append(criterios, delaFeature...)

		revisar(raiz, i, fr, carpeta, delaFeature)
	}

	// El código entero del rango, en una línea por archivo. El diff completo de
	// varias features no entra en ningún contexto; el resumen dice QUÉ MIRAR y
	// el auditor abre lo que necesita.
	s.Partes = append(s.Partes, codigo(raiz, e, ids))

	i.Sobre = s
	i.Hechos[0] = Hecho{Nivel: Ok, Texto: fmt.Sprintf("alcance: %s · %s",
		plural(len(ids), "feature", "features"), plural(len(criterios), "criterio", "criterios"))}
	correrSuite(raiz, i)

	return i, nil
}

// plural evita el "1 features", que en un informe que Javier lee canta.
func plural(n int, uno, varios string) string {
	if n == 1 {
		return "1 " + uno
	}
	return fmt.Sprintf("%d %s", n, varios)
}

// carpetaDe encuentra dónde viven los artefactos de una feature.
//
// Mira primero en archivado y después en features, y el orden importa: una
// feature cerrada YA SE MOVIÓ, y si quedó una carpeta vieja con el mismo nombre
// —un archivado a medias, un merge raro— la que vale es la archivada.
func carpetaDe(raiz string, fr roadmap.Feature) string {
	base := filepath.Base(fr.Carpeta())
	archivada := filepath.Join(docs.Archivado, base)
	if _, err := os.Stat(filepath.Join(raiz, archivada)); err == nil {
		return archivada
	}
	return fr.Carpeta()
}

// revisar corre las comprobaciones ③ y ④ sobre una feature.
func revisar(raiz string, i *Informe, fr roadmap.Feature, carpeta string, criterios []string) {
	rev, err := revision.Leer(filepath.Join(raiz, carpeta, docs.Revision))
	if err != nil {
		// Sin revisión no hay nada que contrastar. Es un aviso y no una falla
		// porque tiene un caso legítimo: un bug saltea `revision` entero.
		i.hecho(Aviso, "%s no tiene revision.json — ¿fue un bug (saltea el ㉑) o quedó sin revisar?", fr.ID)
		return
	}

	// ③ cada criterio tiene veredicto
	for _, c := range criterios {
		if _, hay := rev.Criterios[c]; !hay {
			i.hecho(Falla, "%s no tiene veredicto en la revisión de %s", c, fr.ID)
		}
	}

	// Lo que Javier descartó vuelve a la superficie, y a propósito: el ㉑ lo
	// dejó pasar porque él lo decidió, pero el auditor mira otra pregunta —"¿la
	// solución entera se sostiene?"— y ahí un hallazgo descartado tres veces es
	// un patrón, no una decisión.
	for _, h := range rev.Hallazgos {
		if h.Estado == revision.Descartado {
			i.hecho(Aviso, "%s/%s se descartó: %s", fr.ID, h.ID, h.Motivo)
		}
	}

	// ④ los tests que probaban cada criterio siguen existiendo
	//
	// ACÁ MUERE "EL IMPLEMENTADOR MINTIÓ", y es un conteo. La revisión dijo que
	// el criterio se cumple; el plan dice con qué tests se probaba. Si esos
	// tests ya no están, la afirmación perdió su respaldo — y eso no lo detecta
	// nadie más: el ㉑ de esta feature ya pasó, y el de la siguiente mira otros
	// criterios.
	p, err := tareas.Leer(raiz, carpeta)
	if err != nil {
		i.hecho(Aviso, "%s: no pude leer el plan (%v) — no puedo comprobar sus tests", fr.ID, err)
		return
	}

	for _, c := range criterios {
		if rev.Criterios[c] != "cumple" {
			continue
		}
		tests := testsDe(p, c)
		if len(tests) == 0 {
			i.hecho(Falla, "%s se dio por cumplido y ninguna tarea declaró un test para él", c)
			continue
		}
		if faltan := suite.Faltantes(raiz, tests); len(faltan) > 0 {
			i.hecho(Falla, "%s se dio por cumplido en %s y su test ya no existe: %s",
				c, fr.ID, strings.Join(faltan, " · "))
		}
	}
}

// testsDe junta los tests de todas las tareas que dicen satisfacer un criterio.
func testsDe(p *tareas.Plan, criterio string) []string {
	var tests []string
	for _, t := range p.Tareas {
		if slices.Contains(t.Satisface, criterio) {
			tests = append(tests, t.Tests...)
		}
	}
	return tests
}

// correrSuite es la ⑤, y es la más barata de las cinco.
func correrSuite(raiz string, i *Informe) {
	c, err := constitucion.Leer(raiz)
	if err != nil {
		i.hecho(Aviso, "no pude leer la constitución: %v", err)
		return
	}
	r, err := suite.Correr(raiz, c.TestCmd)
	if err != nil {
		i.hecho(Aviso, "no pude correr los tests: %v", err)
		return
	}
	if r.Verde {
		i.hecho(Ok, "la suite pasa (%s)", c.TestCmd)
		return
	}
	// La salida NO se embebe: puede ser enorme, y el auditor tiene el comando
	// para correrla él si le interesa el detalle. Lo que importa acá es el
	// hecho, y el hecho es que no pasa.
	i.hecho(Falla, "la suite NO pasa (%s). Todo lo que sigue se juzga sobre un árbol roto.", c.TestCmd)
}

// codigo es el resumen del diff de todo el rango auditado.
func codigo(raiz string, e *estado.Estado, ids []string) sobre.Parte {
	const titulo = "El código que quedó"

	if !git.EsRepo(raiz) {
		return sobre.Parte{Titulo: titulo, Falta: "el proyecto no está bajo git"}
	}

	// El punto de partida es el base_commit de la PRIMERA feature del rango:
	// es el estado del repo justo antes de que se construyera nada de lo que se
	// está auditando.
	base := ""
	for _, id := range ids {
		if f, hay := e.Features[id]; hay && f.BaseCommit != "" {
			base = f.BaseCommit
			break
		}
	}
	if base == "" {
		return sobre.Parte{Titulo: titulo, Falta: "ninguna feature del rango guardó base_commit"}
	}

	resumen, err := git.DiffResumen(raiz, base)
	if err != nil {
		return sobre.Parte{Titulo: titulo, Falta: err.Error()}
	}
	if resumen == "" {
		return sobre.Parte{Titulo: titulo, Falta: "no hay cambios desde " + base}
	}
	return sobre.Parte{Titulo: titulo + " (desde " + base + ")", Contenido: resumen}
}

func rutasDeHistorias(ids []string) []string {
	r := make([]string, 0, len(ids))
	for _, id := range ids {
		r = append(r, docs.Historia(id))
	}
	return r
}

// ────────────────────────────────────────────────────────────────────────────
// Render
// ────────────────────────────────────────────────────────────────────────────

// Texto arma el informe para que lo lea el auditor.
//
// Los hechos van ARRIBA del material, al revés que en un informe humano: el que
// lee es un modelo que va a gastar contexto leyendo archivos, y saber qué
// buscar antes de abrirlos cambia qué abre.
func (i *Informe) Texto(raiz string, completo bool) string {
	var b strings.Builder

	fmt.Fprintf(&b, "# Auditoría — %s\n\n", strings.Join(i.Features, " · "))

	b.WriteString("## Lo que sf comprobó\n")
	for _, h := range i.Hechos {
		fmt.Fprintf(&b, "%s %s\n", marca(h.Nivel), h.Texto)
	}

	if i.Fallas() == 0 {
		b.WriteString("\nNada de lo comprobable falla. Lo que sigue es juicio, y es tuyo.\n")
	}

	b.WriteString("\n")
	b.WriteString(i.Sobre.Texto(raiz, completo))
	return b.String()
}

func marca(n Nivel) string {
	switch n {
	case Falla:
		return "✗"
	case Aviso:
		return "⚠"
	default:
		return "✓"
	}
}

// Fallas cuenta las contradicciones encontradas.
//
// No decide nada —el auditor igual tiene que juzgar— pero sirve para el código
// de salida: un informe con fallas no es lo mismo que uno limpio.
func (i *Informe) Fallas() int {
	n := 0
	for _, h := range i.Hechos {
		if h.Nivel == Falla {
			n++
		}
	}
	return n
}
