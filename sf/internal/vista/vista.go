// Package vista genera lo que `sf status` le muestra a Javier.
//
// ────────────────────────────────────────────────────────────────────────────
// EL ÚNICO COMANDO DE sf CUYO CONSUMIDOR ES UN HUMANO
// ────────────────────────────────────────────────────────────────────────────
//
// Los otros nueve los lee un agente. Éste no aparece en ningún trazado del
// bucle —el bucle no lo necesita nunca— y por un rato pareció que se caía. No:
// tiene un consumidor, y es Javier cuando quiere mirar dónde está todo sin
// preguntarle a nadie.
//
// ────────────────────────────────────────────────────────────────────────────
// Y ES LA REGLA 1.5 CON FORMA DE COMANDO
// ────────────────────────────────────────────────────────────────────────────
//
//	Lo que se puede generar, se genera.
//
// El backlog, el roadmap y el índice NO EXISTEN COMO ARCHIVO. Un índice
// mantenido a mano siempre queda viejo; los archivos guardan los hechos y sf
// arma las vistas. Por eso este paquete no escribe nada: lee tres fuentes y
// devuelve un string.
package vista

import (
	"fmt"
	"strings"

	"github.com/jdbaigorria/specforge/sf/internal/estado"
	"github.com/jdbaigorria/specforge/sf/internal/historia"
	"github.com/jdbaigorria/specforge/sf/internal/roadmap"
)

// marca es el símbolo con el que se lee de un vistazo en qué anda cada feature.
//
// Los tres primeros son un estado real; `▸` cubre los cuatro estados con
// trabajo, porque para mirar de reojo no importa cuál de los cuatro es: importa
// que ésa es la que se está tocando.
func marca(est string) string {
	switch est {
	case estado.Cerrada:
		return "✓"
	case estado.Planificada:
		return "◷"
	case "":
		return " "
	default:
		return "▸"
	}
}

// Estado arma la vista entera.
func Estado(raiz string, e *estado.Estado, r *roadmap.Roadmap) string {
	var b strings.Builder

	b.WriteString(producto(e))

	if r == nil || len(r.Features) == 0 {
		b.WriteString("\nTodavía no hay roadmap.\n")
		return b.String()
	}

	// Los títulos de las historias salen de los us-#, no del roadmap: el
	// roadmap guarda SÓLO IDS para no poder contradecir al us-# (artefactos.md
	// §8). Ese diseño se paga acá, leyendo un archivo por historia — y es
	// barato comparado con tener el título en dos lugares.
	titulos := map[string]string{}
	for _, id := range historia.Ids(raiz) {
		if h, err := historia.Leer(raiz, id); err == nil {
			titulos[id] = h.Titulo
		}
	}

	b.WriteString("\n")
	for i, f := range r.Features {
		fe := e.Features[f.ID]
		est := ""
		if fe != nil {
			est = fe.Estado
		}

		fmt.Fprintf(&b, "%s %d. %s", marca(est), i+1, f.Nombre)
		if detalle := detalleDe(f.ID, e, fe); detalle != "" {
			fmt.Fprintf(&b, "   %s", detalle)
		}
		b.WriteString("\n")

		for _, id := range f.Historias {
			t := titulos[id]
			if t == "" {
				t = "—"
			}
			fmt.Fprintf(&b, "     %-6s %s\n", id, t)
		}
	}
	return b.String()
}

// detalleDe es lo que se agrega a la derecha del nombre de la feature.
//
// Sólo la que se está tocando lo lleva: para el resto, la marca ya dice todo lo
// que hay que saber, y repetir "pendiente" en cinco líneas es ruido.
func detalleDe(id string, e *estado.Estado, f *estado.Feature) string {
	if f == nil {
		return "pendiente"
	}
	switch f.Estado {
	case estado.Cerrada:
		return ""
	case estado.Planificada:
		return "plan aprobado, esperando"
	}

	partes := []string{f.Estado}
	if f.Estado == estado.Implementar && len(f.Lotes) > 0 {
		partes = append(partes, fmt.Sprintf("lote %d de %d", f.Cerrados()+1, len(f.Lotes)))
	}
	if f.Modelo != "" {
		partes = append(partes, f.Modelo)
	}
	// El contador sólo se muestra cuando ya hay fracasos: en cero es ruido, y
	// en dos es la señal de que ME TRABÉ está cerca.
	if f.IntentosFallidos > 0 {
		partes = append(partes, fmt.Sprintf("⚠ %d intentos fallidos", f.IntentosFallidos))
	}
	// Mismo criterio que el de arriba, sobre el otro bucle: en cero es ruido, y
	// en dos avisa que la próxima vuelta del ㉑ levanta la parada.
	if f.RondasRevision > 0 {
		partes = append(partes, fmt.Sprintf("⚠ %d rondas de revisión", f.RondasRevision))
	}
	// Ampliada se muestra siempre que esté, y no es un contador: es por qué
	// esta feature está planificando algo que se había declarado chico. Sin
	// esto, `sf status` la muestra igual que cualquier otra y el salto de
	// camino queda sólo en el estado.json.
	if f.Ampliada {
		partes = append(partes, "ampliada")
	}
	if id != e.FeatureActual {
		partes = append(partes, "(no es la actual)")
	}
	return "· " + strings.Join(partes, " · ")
}

// producto son las tres líneas de arriba: en qué anda el producto mismo.
//
// Se muestran sólo mientras faltan. Un producto ya armado no necesita que le
// recuerden que el brief está sellado — eso ya pasó hace semanas.
func producto(e *estado.Estado) string {
	if e.Producto.BriefSellado == "hacelo" && e.Producto.ConstitucionSellada {
		return ""
	}

	var b strings.Builder
	b.WriteString("producto\n")

	linea := func(nombre string, listo bool, extra string) {
		s := "  "
		if listo {
			s = "✓ "
		}
		fmt.Fprintf(&b, "  %s%s%s\n", s, nombre, extra)
	}
	sello := ""
	if e.Producto.BriefSellado != "" {
		sello = "   (" + e.Producto.BriefSellado + ")"
	}
	linea("brief", e.Producto.BriefSellado != "", sello)
	linea("prd", e.Producto.PrdHash != "", "")
	linea("constitución", e.Producto.ConstitucionSellada, "")

	return b.String()
}
