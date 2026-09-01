package lanzar

import (
	"fmt"
	"strings"
	"time"
)

// Contar arma el parte de una corrida, para que lo lea el que la lanzó.
//
// El que lee esto es el arnés principal —o Javier mirándolo— justo después de
// que el hijo terminó. Tiene que contestar tres cosas y nada más: cómo salió,
// cuánto costó, y QUÉ FALTA.
//
// Y lo que falta es siempre lo mismo y es lo más importante del texto: `sf done`.
// La ficha cuenta qué pasó; NO decide si se avanza. Command Code devolvió
// `exit=0` y `subtype: success` sin haber hecho nada (headless.md §7), así que un
// parte que dijera "listo" sin nombrar la compuerta estaría enseñando lo
// contrario de lo que hay que aprender.
func Contar(f Ficha) string {
	var b strings.Builder

	switch f.Fin {
	case FinTimeout:
		fmt.Fprintf(&b, "⏱ se agotó la espera a los %s. Lo corté.\n", duracion(f.DuroMs))
	case FinError:
		fmt.Fprintf(&b, "🛑 no se pudo lanzar.\n")
	default:
		if f.Salida == 0 {
			fmt.Fprintf(&b, "✓ terminó en %s\n", duracion(f.DuroMs))
		} else {
			fmt.Fprintf(&b, "✗ salió con %d, a los %s\n", f.Salida, duracion(f.DuroMs))
		}
	}

	// Lo que el arnés no dio no se imprime. Un "US$ 0" donde el arnés no dijo
	// nada es una cifra inventada, y ésta es la pantalla donde Javier decide con
	// qué modelo sigue.
	var datos []string
	if f.CostoUSD != nil {
		datos = append(datos, fmt.Sprintf("US$ %.4f", *f.CostoUSD))
	}
	if f.Tokens != nil {
		datos = append(datos, fmt.Sprintf("%d↓ %d↑ tokens", f.Tokens.Entrada, f.Tokens.Salida))
	}
	if len(f.Herramientas) > 0 {
		datos = append(datos, strings.Join(f.Herramientas, ", "))
	}
	if len(datos) > 0 {
		fmt.Fprintf(&b, "  %s\n", strings.Join(datos, " · "))
	}

	// Que la skill NO se haya cargado es el aviso más caro de la ficha: quiere
	// decir que el modelo hizo algo, pero no lo que la máquina pidió. `nil` es
	// "no sé" y no se dice nada — afirmar sobre lo que no se pudo mirar sería
	// exactamente el error que este proyecto viene evitando.
	if f.CargoLaSkill != nil && !*f.CargoLaSkill && f.Skill != "" {
		fmt.Fprintf(&b, "  ⚠ no cargó la skill %s. Puede haber hecho otra cosa.\n", f.Skill)
	}
	if f.Dijo != "" {
		fmt.Fprintf(&b, "  dijo: %s\n", primeraLinea(f.Dijo))
	}
	// El stderr se muestra si algo salió mal, Y UN EXIT != 0 CUENTA COMO MAL.
	//
	// La primera versión sólo lo mostraba cuando `fin` no era "termino", y en la
	// primera corrida de verdad eso escondió el motivo: opencode salió con 1
	// diciendo por stderr que el portamodelo era inválido, y el parte mostró un
	// "✗ salió con 1" pelado. El motivo estaba en la ficha y había que ir a
	// abrirla — que es justo lo que este texto existe para evitar.
	if f.Error != "" && (f.Fin != FinTermino || f.Salida != 0) {
		fmt.Fprintf(&b, "  %s\n", primeraLinea(f.Error))
	}
	fmt.Fprintf(&b, "  registro: %s/%s\n", Carpeta, f.Registro)

	// `sf done` sólo se ofrece cuando hubo una corrida que juzgar. Ofrecerlo
	// después de un timeout sería mandar a la compuerta a mirar un artefacto a
	// medio escribir.
	if f.Fin == FinTermino {
		b.WriteString("\nAhora la compuerta: `sf done`. La ficha cuenta qué pasó, no decide si vale.\n")
	}
	return b.String()
}

func duracion(ms int64) string {
	d := time.Duration(ms) * time.Millisecond
	if d < time.Minute {
		return d.Round(100 * time.Millisecond).String()
	}
	return d.Round(time.Second).String()
}

func primeraLinea(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		s = s[:i] + " …"
	}
	if len(s) > 160 {
		s = s[:160] + " …"
	}
	return s
}
