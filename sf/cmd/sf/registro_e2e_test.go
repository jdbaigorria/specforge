//go:build e2e

// El registro, visto desde afuera del binario.
//
// ────────────────────────────────────────────────────────────────────────────
// POR QUÉ ESTOS TRES NO SON TESTS DE PAQUETE
// ────────────────────────────────────────────────────────────────────────────
//
// Porque lo que prueban es el ENGANCHE, no el paquete: que `main` abra y cierre
// la entrada en todas las ramas del switch, que el refactor de `os.Exit(x)` a
// `return x` no haya dejado ninguna afuera, y que el conteo de vueltas cuente lo
// que la máquina de verdad hizo pasar.
//
// Es la misma razón por la que existe e2e_test.go entero: los diez defectos de
// arreglos.md vivían en la costura entre comandos, y ninguno se veía desde
// adentro de un paquete.
package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jdbaigorria/specforge/sf/internal/comandos"
	"github.com/jdbaigorria/specforge/sf/internal/global"
	"github.com/jdbaigorria/specforge/sf/internal/registro"
)

// registroDe lee las líneas escritas por el binario bajo prueba.
func registroDe(t *testing.T, raiz string) []registro.Entrada {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(raiz, registro.Archivo))
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		t.Fatal(err)
	}
	var es []registro.Entrada
	for _, l := range strings.Split(strings.TrimRight(string(b), "\n"), "\n") {
		var e registro.Entrada
		if err := json.Unmarshal([]byte(l), &e); err != nil {
			t.Fatalf("el binario escribió una línea que no es JSON:\n%s\n%v", l, err)
		}
		es = append(es, e)
	}
	return es
}

// ────────────────────────────────────────────────────────────────────────────
// ① Una línea por invocación — para CADA comando del inventario
// ────────────────────────────────────────────────────────────────────────────

// El agujero que cubre es el del refactor: `main` se partió en `despachar()` y
// cada `os.Exit(x)` pasó a ser `return x`. Si quedó UNA rama con `os.Exit`, esa
// invocación no escribe nada y el hueco es invisible — que es exactamente lo que
// este archivo existe para que no pase.
//
// No importa si el comando salió bien o mal: lo que se afirma es que dejó
// rastro. Un `sf done` que falla es tan parte de la película como uno que pasa.
func TestCadaComandoDejaSuLineaEnElRegistro(t *testing.T) {
	p := nuevoProyecto(t)

	// `sf init` primero: sin `.docs/` ni `.specforge/` el registro no escribe, y
	// eso es correcto (que un `sf version` suelto no deje basura).
	p.sf("init")

	for _, c := range comandos.Todos {
		antes := len(registroDe(t, p.raiz))
		p.sf(strings.Fields(c)...) // el exit code no importa acá
		despues := registroDe(t, p.raiz)

		if len(despues) != antes+1 {
			t.Errorf("`sf %s` no dejó línea en el registro (había %d, hay %d)",
				c, antes, len(despues))
			continue
		}
		if got := despues[len(despues)-1].Cmd; got != strings.Fields(c)[0] {
			t.Errorf("`sf %s` quedó registrado como %q", c, got)
		}
	}
}

// Y el espejo: fuera de un proyecto no se escribe NADA.
func TestFueraDeUnProyectoNoDejaRastro(t *testing.T) {
	p := nuevoProyecto(t)
	// nuevoProyecto no corre `sf init`: no hay .docs/ ni .specforge/ todavía.

	p.sf("version")
	p.sf("next")

	if es := registroDe(t, p.raiz); len(es) != 0 {
		t.Errorf("escribió %d líneas en una carpeta sin proyecto", len(es))
	}
}

// ────────────────────────────────────────────────────────────────────────────
// ② EL QUE MÁS IMPORTA: el registro no puede romper un comando
// ────────────────────────────────────────────────────────────────────────────

// Un tablero que puede apagar el motor es peor que no tener tablero.
//
// Se corre `sf next` dos veces —una con el registro pudiendo escribir y otra con
// la carpeta trabada— y se exige que la salida y el exit code sean IDÉNTICOS.
// No "parecidos": iguales, porque el que lo lee es un agente que compara texto.
func TestConElRegistroTrabadoElComandoContestaExactamenteLoMismo(t *testing.T) {
	p := nuevoProyecto(t)
	p.sf("init")

	codigoOK, salidaOK := p.sf("next")

	dir := filepath.Join(p.raiz, global.Carpeta)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(dir, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })

	// Si el filesystem ignora el 0555 —pasa corriendo como root— el test no
	// prueba nada y decirlo es mejor que dar un verde falso.
	if f, err := os.Create(filepath.Join(dir, ".prueba")); err == nil {
		f.Close()
		t.Skip("el filesystem ignoró el 0555: acá no se puede probar el disco trabado")
	}

	codigoTrabado, salidaTrabada := p.sf("next")

	if codigoTrabado != codigoOK {
		t.Errorf("con el registro trabado el exit code cambió: %d → %d", codigoOK, codigoTrabado)
	}
	if salidaTrabada != salidaOK {
		t.Errorf("con el registro trabado la salida cambió:\n--- con registro ---\n%s\n--- trabado ---\n%s",
			salidaOK, salidaTrabada)
	}
}

// ────────────────────────────────────────────────────────────────────────────
// ③ El conteo cuenta lo que la máquina de verdad hizo pasar
// ────────────────────────────────────────────────────────────────────────────

// El test se compara contra un RECORRIDO CONOCIDO y no contra sí mismo: se
// mueve el estado a mano por los pasos de producto, y después se le pide a
// `sf log --vueltas` que los haya visto.
//
// Es lo que hace que el número sirva: hoy "¿cuántas vueltas dio la revisión?" lo
// contesta el campo `vuelta` que escribe el propio revisor, y sf nunca lo
// comprueba (verificado el 2026-09-03: nada en Go lo incrementa).
func TestElConteoDeVueltasVeElRecorridoDeVerdad(t *testing.T) {
	p := nuevoProyecto(t)

	// El mismo arranque que TestVueltaCompletaDeUnaHistoria, y no otro: si el
	// andamio de este test se desviara del guion principal estaría probando una
	// máquina que nadie usa.
	p.paso("el andamio", hayTrabajo, "install")
	p.paso("el estado vacío", hayTrabajo, "init")
	p.conPerfiles()

	// ⑥ el brief, y su sello.
	p.escribir(".docs/brief.md",
		"---\nveredicto: hacelo\n---\n# Brief\nUn sumador.\n\n"+
			"| suma-cli | lo mismo | no exporta | `retrieved` https://github.com/x/suma-cli |\n")
	p.paso("el brief está y lo sella Javier", hayTrabajo, "done")
	p.paso("approve copia el veredicto", hayTrabajo, "approve")

	p.paso("y el next del ⑦ nombra su skill", hayTrabajo, "next")

	// ⑦ el PRD: no tiene parada, así que `sf done` mueve solo.
	p.escribir(".docs/prd.md", "# PRD\nSumar dos números.\n")
	p.paso("del ⑦ se pasa derecho al ⑧", hayTrabajo, "done")

	es := registroDe(t, p.raiz)
	if len(es) < 6 {
		t.Fatalf("el recorrido dejó %d líneas, esperaba al menos 6", len(es))
	}

	// ── El sello del ⑥ quedó con su veredicto ───────────────────────────────
	//
	// Es la prueba de que el antes y el después son dos lecturas distintas del
	// estado: el `approve` entró con el brief sin sellar y salió con "hacelo".
	var sello *registro.Entrada
	for i := range es {
		if es[i].Cmd == "approve" {
			sello = &es[i]
		}
	}
	if sello == nil {
		t.Fatal("el `sf approve` del ⑥ no quedó en el registro")
	}
	if sello.Antes.Producto.Brief != "" {
		t.Errorf("antes del sello el brief ya decía %q", sello.Antes.Producto.Brief)
	}
	if sello.Despues.Producto.Brief != "hacelo" {
		t.Errorf("después del sello el brief dice %q, quería hacelo", sello.Despues.Producto.Brief)
	}

	// ── El ⑦ quedó con su hash ──────────────────────────────────────────────
	//
	// El `prd_hash` se anota en el `sf done` del ⑦, y verlo aparecer entre el
	// antes y el después es la transición de producto que el registro captura.
	var vioElHash bool
	for _, e := range es {
		if e.Cmd == "done" && e.Antes.Producto.Prd == "" && e.Despues.Producto.Prd != "" {
			vioElHash = true
		}
	}
	if !vioElHash {
		t.Error("el prd_hash del ⑦ no se ve aparecer en ninguna línea")
	}

	// ── Y `sf next` anotó lo que sf CONTESTÓ ────────────────────────────────
	//
	// Es la mitad que la foto del estado.json no puede dar: en los pasos de
	// producto no hay feature, y "brief" no es un campo de ningún archivo.
	var vioElPaso bool
	for _, e := range es {
		if e.Cmd == "next" && e.Skill != "" && e.Estado != "" {
			vioElPaso = true
		}
	}
	if !vioElPaso {
		t.Error("ningún `sf next` anotó el estado y la skill que contestó")
	}

	// ── Y el conteo ve el recorrido ─────────────────────────────────────────
	_, out := p.sf("log", "--vueltas")
	if strings.Contains(out, "todavía no hay ninguna transición") {
		t.Fatalf("el conteo no vio nada después de recorrer el ⑥ y el ⑦:\n%s", out)
	}
	// Los dos sellos que de verdad se movieron en este recorrido.
	for _, quiero := range []string{"brief", "prd"} {
		if !strings.Contains(out, quiero) {
			t.Errorf("el conteo no nombra %q:\n%s", quiero, out)
		}
	}
}

// ────────────────────────────────────────────────────────────────────────────
// ④ `sf log` es un lector: sale con 0 SIEMPRE
// ────────────────────────────────────────────────────────────────────────────

// Un 2 significaría "parada, es de Javier" y confundiría al orquestador que lo
// corriera adentro del bucle; un 1 haría que un registro vacío pareciera roto.
func TestLogSaleConCeroInclusoSinRegistro(t *testing.T) {
	p := nuevoProyecto(t)
	p.sf("init")

	for _, args := range [][]string{
		{"log"},
		{"log", "--vueltas"},
		{"log", "--json"},
		{"log", "--feature", "f-99"},
		{"log", "--ultimas", "3"},
	} {
		if c, out := p.sf(args...); c != hayTrabajo {
			t.Errorf("`sf %s` → exit %d, quería 0\n%s", strings.Join(args, " "), c, out)
		}
	}

	// Y un flag que no existe SÍ es un error, que es distinto.
	if c, _ := p.sf("log", "--inventado"); c != esError {
		t.Errorf("`sf log --inventado` → exit %d, quería 1", c)
	}
}
