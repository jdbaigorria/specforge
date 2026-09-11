package revision

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// ────────────────────────────────────────────────────────────────────────────
// La escalera de evidencia — y la compatibilidad va PRIMERO
// ────────────────────────────────────────────────────────────────────────────
//
// `criterios` pasó de `map[string]string` a `map[string]Criterio` para que un
// `cumple` diga TAMBIÉN cómo se sabe. Eso deja un riesgo que no se ve hasta
// mucho después: `.docs/archivado/*/revision.json` tiene revisiones escritas
// con el formato viejo, y el auditor las lee.
//
// Si el Unmarshal no acepta las dos formas, `sf audit` se rompe EN SILENCIO
// sobre features cerradas, y nadie se entera hasta la próxima auditoría. Por eso
// éste es el primer test del paquete y no el último.

// escribir deja un revision.json en un temp dir y devuelve su ruta.
func escribir(t *testing.T, contenido string) string {
	t.Helper()
	ruta := filepath.Join(t.TempDir(), "revision.json")
	if err := os.WriteFile(ruta, []byte(contenido), 0o644); err != nil {
		t.Fatal(err)
	}
	return ruta
}

// EL FORMATO VIEJO SIGUE LEYÉNDOSE. Es el test que sostiene todo lo demás.
func TestLeeElFormatoViejoDeCriterios(t *testing.T) {
	ruta := escribir(t, `{
	  "feature": "f-1", "vuelta": 1, "veredicto": "limpio",
	  "criterios": {"us-1/CA-1": "cumple", "us-1/CA-2": "no-cumple"},
	  "hallazgos": []
	}`)

	r, err := Leer(ruta)
	if err != nil {
		t.Fatalf("una revisión archivada tiene que seguir leyéndose: %v", err)
	}
	if got := r.Criterios["us-1/CA-1"].Veredicto; got != "cumple" {
		t.Errorf("us-1/CA-1: veredicto %q, quería \"cumple\"", got)
	}
	if got := r.Criterios["us-1/CA-2"].Veredicto; got != "no-cumple" {
		t.Errorf("us-1/CA-2: veredicto %q, quería \"no-cumple\"", got)
	}

	// Y el escalón queda en 0, que NO es "escalón bajo": es "esta revisión es
	// anterior a la escalera". La compuerta distingue las dos cosas; si acá
	// saliera 1, una revisión vieja parecería una afirmación sin respaldo.
	if got := r.Criterios["us-1/CA-1"].Escalon; got != SinEscalera {
		t.Errorf("el formato viejo declara escalón %d, quería %d (sin escalera)", got, SinEscalera)
	}
}

// EL FORMATO NUEVO, que es el que escribe el ㉑ de acá en adelante.
func TestLeeElFormatoNuevoDeCriterios(t *testing.T) {
	ruta := escribir(t, `{
	  "feature": "f-1", "vuelta": 1, "veredicto": "limpio",
	  "criterios": {
	    "us-1/CA-1": {"veredicto": "cumple", "escalon": 4,
	                  "prueba": "internal/suite/suite_test.go::TestCierra"}
	  },
	  "hallazgos": []
	}`)

	r, err := Leer(ruta)
	if err != nil {
		t.Fatal(err)
	}
	c := r.Criterios["us-1/CA-1"]
	if c.Veredicto != "cumple" || c.Escalon != 4 {
		t.Errorf("leyó %+v, quería cumple en escalón 4", c)
	}
	if c.Prueba != "internal/suite/suite_test.go::TestCierra" {
		t.Errorf("prueba %q", c.Prueba)
	}
}

// LOS DOS FORMATOS EN EL MISMO ARCHIVO. No es hipotético: una vuelta 2 escrita
// con sf nuevo sobre una feature cuya vuelta 1 quedó archivada mezcla las dos
// mientras alguien edita a mano.
func TestLeeLosDosFormatosMezclados(t *testing.T) {
	ruta := escribir(t, `{
	  "feature": "f-1", "vuelta": 2, "veredicto": "limpio",
	  "criterios": {
	    "us-1/CA-1": "cumple",
	    "us-1/CA-2": {"veredicto": "cumple", "escalon": 5, "prueba": ".docs/verificar/pruebas/f-1.txt"}
	  },
	  "hallazgos": []
	}`)

	r, err := Leer(ruta)
	if err != nil {
		t.Fatal(err)
	}
	if r.Criterios["us-1/CA-1"].Escalon != SinEscalera {
		t.Error("el viejo tiene que quedar sin escalera")
	}
	if r.Criterios["us-1/CA-2"].Escalon != 5 {
		t.Error("el nuevo tiene que conservar su escalón")
	}
}

// UN CRITERIO QUE NO ES NI UNA COSA NI LA OTRA es un archivo corrupto, y tiene
// que decirlo. Un error acá es barato; un veredicto vacío que pasa la compuerta
// no lo es.
func TestRechazaUnCriterioQueNoEsNiStringNiObjeto(t *testing.T) {
	ruta := escribir(t, `{"feature":"f-1","criterios":{"us-1/CA-1": 4},"hallazgos":[]}`)

	if _, err := Leer(ruta); err == nil {
		t.Error("un criterio numérico tiene que ser un error, no un veredicto vacío")
	}
}

// GUARDAR ESCRIBE EL FORMATO NUEVO, y de paso migra lo que leyó viejo.
//
// Lo usa `sf dismiss`, que es el único comando que toca el archivo después del
// ㉑. Que al guardar quede en el formato nuevo es deliberado: el archivo se
// actualiza solo la primera vez que alguien lo toca, sin una migración aparte.
func TestGuardarEscribeElFormatoNuevo(t *testing.T) {
	ruta := escribir(t, `{"feature":"f-1","criterios":{"us-1/CA-1":"cumple"},"hallazgos":[]}`)

	r, err := Leer(ruta)
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Guardar(ruta); err != nil {
		t.Fatal(err)
	}

	b, err := os.ReadFile(ruta)
	if err != nil {
		t.Fatal(err)
	}
	var crudo struct {
		Criterios map[string]map[string]any `json:"criterios"`
	}
	if err := json.Unmarshal(b, &crudo); err != nil {
		t.Fatalf("lo guardado ya no es el formato nuevo: %v", err)
	}
	if crudo.Criterios["us-1/CA-1"]["veredicto"] != "cumple" {
		t.Errorf("guardó %+v", crudo.Criterios["us-1/CA-1"])
	}
}

// ────────────────────────────────────────────────────────────────────────────
// Cumple() — la pregunta que hacen el auditor y la compuerta
// ────────────────────────────────────────────────────────────────────────────

func TestCumple(t *testing.T) {
	for _, c := range []struct {
		nombre   string
		criterio Criterio
		quiere   bool
	}{
		{"cumple pelado", Criterio{Veredicto: "cumple"}, true},
		{"no-cumple", Criterio{Veredicto: "no-cumple"}, false},
		{"vacío", Criterio{}, false},
	} {
		if got := c.criterio.Cumple(); got != c.quiere {
			t.Errorf("%s: Cumple()=%v, quería %v", c.nombre, got, c.quiere)
		}
	}
}
