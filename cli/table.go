package main

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// renderTable imprime una tabla de texto con columnas alineadas a la izquierda.
// Los anchos se miden en RUNAS (utf8.RuneCountInString), no en bytes, para que
// caracteres multibyte como "—" no desalineen. Lo comparten sf status y
// sf gate status.
func renderTable(cols []string, rows [][]string) {
	// Ancho de cada columna = máximo entre su encabezado y todas sus celdas.
	widths := make([]int, len(cols))
	for i, c := range cols {
		widths[i] = utf8.RuneCountInString(c)
	}
	for _, r := range rows {
		for i, cell := range r {
			if w := utf8.RuneCountInString(cell); w > widths[i] {
				widths[i] = w
			}
		}
	}

	// pad alinea una fila a los anchos calculados. En fmt el ancho de "%-*s" se
	// mide en runas, así que combina bien con widths (también en runas).
	pad := func(cells []string) string {
		parts := make([]string, len(cells))
		for i, cell := range cells {
			parts[i] = fmt.Sprintf("%-*s", widths[i], cell)
		}
		return strings.Join(parts, "  ")
	}

	fmt.Println(pad(cols))

	seps := make([]string, len(cols))
	for i := range cols {
		seps[i] = strings.Repeat("-", widths[i])
	}
	fmt.Println(strings.Join(seps, "  "))

	for _, r := range rows {
		fmt.Println(pad(r))
	}
}
