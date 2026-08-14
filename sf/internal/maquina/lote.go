package maquina

import (
	"fmt"

	"github.com/jdbaigorria/specforge/sf/internal/constitucion"
	"github.com/jdbaigorria/specforge/sf/internal/estado"
	"github.com/jdbaigorria/specforge/sf/internal/git"
	"github.com/jdbaigorria/specforge/sf/internal/roadmap"
	"github.com/jdbaigorria/specforge/sf/internal/suite"
	"github.com/jdbaigorria/specforge/sf/internal/tareas"
)

// EmpezarLote es `sf lote start`: la compuerta del medio, y la joya del diseño.
//
// ────────────────────────────────────────────────────────────────────────────
// TRES COSAS, EN ESTE ORDEN
// ────────────────────────────────────────────────────────────────────────────
//
//  1. la branch    sf la CREA, no la verifica            → mata el #4
//  2. ¿existen?    los N tests planificados del lote     → mata el #8
//  3. ¿fallan?     corre la suite y exige rojo           → mata el #5 y el #8
//
// Y guarda el hash de los archivos de test, que es lo que después impide
// ablandar el assert para llegar a verde.
//
// ────────────────────────────────────────────────────────────────────────────
// POR QUÉ ESTE ES EL COMANDO QUE MÁS PAGA
// ────────────────────────────────────────────────────────────────────────────
//
// Los otros nueve exponen, sirven o cuentan. Éste es el único que ve pasar un
// hecho que NO DEJA HUELLA EN NINGÚN LADO: el momento en que los tests fallaban.
// Si sf no lo anota cuando lo ve, se pierde para siempre — y sin eso, "todo
// verde" no significa nada.
func EmpezarLote(raiz string, e *estado.Estado, r *roadmap.Roadmap) Efecto {
	var ef Efecto

	if r == nil {
		ef.falla("todavía no hay roadmap: falta el ⑩")
		return ef
	}
	f, hay := e.Actual()
	if !hay {
		ef.falla("no hay feature en curso: corré `sf take <feature>` primero")
		return ef
	}
	if f.Estado != estado.Implementar {
		ef.falla("%s está en %q, y `sf lote start` es de implementar", e.FeatureActual, f.Estado)
		return ef
	}
	fr, enRoadmap := r.Buscar(e.FeatureActual)
	if !enRoadmap {
		ef.falla("%s no está en el roadmap", e.FeatureActual)
		return ef
	}

	c, err := constitucion.Leer(raiz)
	if err != nil {
		ef.falla("%v", err)
		return ef
	}

	// Los lotes salen de tareas.json, y el estado sólo guarda por cuál va. Si
	// todavía no hay ninguno anotado, se siembran acá: es la primera vez que
	// alguien mira el plan con intención de trabajarlo.
	p, err := tareas.Leer(raiz, fr.Carpeta())
	if err != nil {
		ef.falla("%v", err)
		return ef
	}
	if len(f.Lotes) == 0 {
		for _, n := range p.Lotes() {
			f.Lotes = append(f.Lotes, estado.Lote{Lote: n})
		}
	}

	l, hayLote := f.LoteActual()
	if !hayLote {
		ef.falla("todos los lotes de %s ya están commiteados", fr.ID)
		return ef
	}
	if l.Rojo {
		ef.falla("el rojo del lote %d ya está confirmado. Implementá y cerrá con `sf done --msg \"…\"`", l.Lote)
		return ef
	}

	var pasos []string

	// ① La branch. Se crea, no se verifica.
	if c.Git.BranchPorFeature {
		branch := c.Branch(fr.ID, fr.Slug)
		if err := git.CrearBranch(raiz, branch); err != nil {
			ef.falla("no pude crear la branch %s: %v", branch, err)
			return ef
		}
		pasos = append(pasos, "branch "+branch)
	}

	// ② ¿Existen los tests planificados?
	//
	// Esta lista es lo que separa "escribí tests" de "escribí LOS tests que el
	// plan pedía". Sin ella, el #8 se cuela por el costado: la suite falla por
	// cualquier otra cosa y el lote pasa igual.
	planificados := p.TestsDelLote(l.Lote)
	if len(planificados) == 0 {
		ef.falla("el lote %d no tiene tests planificados. Eso tendría que haberlo frenado el ⑰.", l.Lote)
		return ef
	}
	if faltan := suite.Faltantes(raiz, planificados); len(faltan) > 0 {
		ef.falla("planificados %d tests para el lote %d. Faltan %d:", len(planificados), l.Lote, len(faltan))
		for _, t := range faltan {
			ef.falla("    %s", t)
		}
		return ef
	}
	pasos = append(pasos, fmt.Sprintf("los %d tests planificados existen", len(planificados)))

	// ③ El rojo. Le alcanza un exit code.
	res, err := suite.Correr(raiz, c.TestCmd)
	if err != nil {
		ef.falla("no pude correr los tests: %v", err)
		return ef
	}
	if res.Verde {
		ef.falla("la suite YA PASA, y todavía no se escribió el código del lote %d.", l.Lote)
		ef.falla("    Un test que pasa antes de que exista el código es un test de mentira.")
		return ef
	}
	pasos = append(pasos, "rojo confirmado")

	// ④ El hash, que es la memoria de CÓMO eran los tests cuando estaban en
	// rojo. Sin esto, el subagente que no logra implementar puede ablandar el
	// assert y llegar a verde sin que nadie se entere.
	hash, err := git.HashDe(raiz, suite.Archivos(planificados))
	if err != nil {
		// Sin git no hay hash. No frena: la compuerta del rojo ya se cumplió, y
		// el hash protege de un caso más fino.
		hash = ""
	}

	l.Rojo = true
	l.HashTests = hash

	ef.Mensaje = fmt.Sprintf("lote %d de %d listo — %s.\n   Podés implementar.",
		l.Lote, len(p.Lotes()), joinConY(pasos))
	return ef
}

// joinConY arma "a, b y c". Es sólo para que el mensaje se lea como una frase.
func joinConY(xs []string) string {
	switch len(xs) {
	case 0:
		return ""
	case 1:
		return xs[0]
	}
	s := ""
	for i, x := range xs[:len(xs)-1] {
		if i > 0 {
			s += ", "
		}
		s += x
	}
	return s + " y " + xs[len(xs)-1]
}
