package doctor

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Texto dibuja el informe.
//
// El criterio de qué se muestra es el mismo que el de `compuerta.Texto`, con
// una diferencia deliberada: acá los ✓ SÍ se listan.
//
// En una compuerta enumerar lo que salió bien es ruido —al que trabajó sólo le
// importa qué le falta—. Acá la pregunta es otra: alguien acaba de instalar y
// quiere ver el inventario. Un doctor que no imprime nada cuando todo anda deja
// al que lo corrió sin saber si funcionó o si el comando no hizo nada.
func (i Informe) Texto() string {
	var b strings.Builder

	fmt.Fprintf(&b, "binario   %s\n", i.Binario.Version)
	fmt.Fprintf(&b, "          %s\n", acortar(i.Binario.Corriendo))
	if i.Binario.Sombra {
		// Se muestran los dos y en este orden porque el que importa es el
		// segundo: es el que va a correr el agente.
		fmt.Fprintf(&b, "       ✗  el del PATH es OTRO: %s\n", acortar(i.Binario.EnPath))
	}

	fmt.Fprintf(&b, "\nharness   %s\n", i.Harness)

	// Los perfiles van pegados al harness porque son la misma pregunta: el
	// catálogo está indexado por harness, así que "razonar" declarado no quiere
	// decir nada sin saber para cuál.
	for _, p := range i.Perfiles {
		switch {
		case !p.Declarado:
			fmt.Fprintf(&b, "       ✗  %-10s sin declarar\n", p.Nombre)
		case p.SinPortamodelo:
			fmt.Fprintf(&b, "       ✗  %-10s %s · le falta el archivo de agente\n", p.Nombre, p.Alias)
		default:
			fmt.Fprintf(&b, "       ✓  %-10s %s\n", p.Nombre, p.Alias)
		}
	}

	var faltan int
	for _, s := range i.Skills {
		if !s.Instalado() {
			faltan++
		}
	}
	fmt.Fprintf(&b, "\nskills    %d/%d de la máquina\n", len(i.Skills)-faltan, len(i.Skills))
	for _, s := range i.Skills {
		switch {
		case !s.Instalado():
			fmt.Fprintf(&b, "       ✗  %-18s no está instalado\n", s.Nombre)
		case len(s.Desconocidos) > 0:
			fmt.Fprintf(&b, "       ✗  %-18s nombra `sf %s`\n",
				s.Nombre, strings.Join(s.Desconocidos, "`, `sf "))
		default:
			fmt.Fprintf(&b, "       ✓  %-18s %s\n", s.Nombre, acortar(s.Ruta))
		}
	}

	// Las herramientas van entre los skills y el proyecto porque contestan la
	// misma pregunta que los skills —"¿el que va a trabajar tiene con qué?"—
	// pero del lado de afuera: el skill dice qué hacer y la herramienta es con
	// qué. El 2026-09-05 los nueve skills estaban instalados y la corrida
	// falló igual, porque la mitad de abajo no estaba.
	h := i.Herramientas
	b.WriteString("\nbuscar    ")
	if h.Nivel0() {
		fmt.Fprintf(&b, "nivel 0 disponible — alcanza sin ninguna llave\n")
		fmt.Fprintf(&b, "       ✓  %-18s %s\n", "curl", acortar(h.Curl))
	} else {
		fmt.Fprintf(&b, "SIN NIVEL 0\n")
		fmt.Fprintf(&b, "       ✗  %-18s no está en el PATH\n", "curl")
	}
	if h.MCP == "" {
		fmt.Fprintf(&b, "       ·  sin .mcp.json — no hay MCPs declarados\n")
	} else {
		fmt.Fprintf(&b, "       ·  %s\n", acortar(h.MCP))
		for _, s := range h.Servidores {
			switch {
			case s.Llave == "":
				fmt.Fprintf(&b, "       ✓  %-18s sin llave\n", s.Nombre)
			case s.Puesta:
				fmt.Fprintf(&b, "       ✓  %-18s %s puesta\n", s.Nombre, s.Llave)
			default:
				fmt.Fprintf(&b, "       ✗  %-18s falta %s\n", s.Nombre, s.Llave)
			}
		}
	}
	// La última línea del bloque es la más importante y por eso va última: sin
	// ella, cuatro ✓ se leen como "las herramientas andan", y lo que se
	// comprobó es que están DECLARADAS. Levantarlas es del arnés, y el arnés no
	// se ve desde un binario que se ejecuta y termina.
	fmt.Fprintf(&b, "       ?  si el arnés los levantó de verdad: no se sabe desde acá\n")

	fmt.Fprintf(&b, "\nproyecto  %s\n", acortar(i.Proyecto.Raiz))
	if i.Proyecto.Andamiado {
		fmt.Fprintf(&b, "          .docs/ está — `sf status` dice dónde va\n")
	} else {
		fmt.Fprintf(&b, "          sin .docs/ — `sf init` lo arma\n")
	}

	for _, a := range i.Avisos {
		fmt.Fprintf(&b, "\n⚠ %s\n", a)
	}
	if len(i.Fallas) > 0 {
		b.WriteString("\n")
		for _, f := range i.Fallas {
			fmt.Fprintf(&b, "✗ %s\n", f)
		}
		b.WriteString("\n")
		b.WriteString(comoArreglarlo(i))
	}
	return b.String()
}

// comoArreglarlo es la mitad que convierte un diagnóstico en algo accionable.
//
// Un doctor que dice "faltan 9 skills" y no dice cómo se instalan obliga a ir a
// buscar el README, y el que lo corre suele ser alguien que acaba de instalar y
// todavía no sabe dónde está nada.
func comoArreglarlo(i Informe) string {
	var b strings.Builder
	b.WriteString("cómo se arregla:\n")

	if i.Binario.Sombra || i.Binario.EnPath == "" {
		b.WriteString("  el binario   curl -fsSL https://raw.githubusercontent.com/jdbaigorria/specforge/main/install.sh | sh\n")
	}
	for _, s := range i.Skills {
		if !s.Instalado() {
			b.WriteString("  los skills   /plugin marketplace add jdbaigorria/specforge\n")
			b.WriteString("               /plugin install specforge\n")
			break
		}
	}
	for _, s := range i.Skills {
		if s.Instalado() && len(s.Desconocidos) > 0 {
			// Este caso es el desalineado, y la respuesta NO es reinstalar una
			// de las dos mitades al azar: hay que saber cuál quedó vieja.
			b.WriteString("  el desfase   los skills y el binario son de versiones distintas.\n")
			b.WriteString("               Reinstalá LOS DOS de la misma fuente.\n")
			break
		}
	}
	return b.String()
}

// acortar reemplaza el home por ~ para que las rutas entren en una línea.
func acortar(ruta string) string {
	if ruta == "" {
		return "(no está)"
	}
	h, err := os.UserHomeDir()
	if err != nil {
		return ruta
	}
	if r, err := filepath.Rel(h, ruta); err == nil && !strings.HasPrefix(r, "..") {
		return "~/" + r
	}
	return ruta
}
