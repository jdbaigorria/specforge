package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
)

// ----------------------------------------------------------------------------
// `sf domain terms` — el mismo concepto con nombres distintos entre features
// (DL-14b).
//
// EN CONSULTORÍA ES LA FUENTE NÚMERO UNO DE RETRABAJO, y es un defecto que no se
// ve leyendo una feature: cada spec es internamente coherente. Aparece recién
// cuando dos features que nadie leyó juntas llegan a integrarse y resulta que
// `Order` y `Pedido` eran la misma cosa — o peor, que no lo eran y nadie se dio
// cuenta hasta que los datos no cerraron.
//
// POR QUÉ ESTO SÍ SE PUEDE HACER DETERMINISTA. El resto de `DL-14` —leer actas y
// mails y extraer hallazgos con cita literal— lo tiene que hacer un modelo. Esto
// no: `domain.json` YA declara el lenguaje ubicuo con sus alias, y los requisitos
// YA están escritos. La pregunta *"¿qué forma superficial usa cada feature para
// este mismo término?"* se contesta con una expresión regular, sin inferir nada.
//
// Por eso se separó del ingestor: es la mitad que paga sola y no depende de las
// decisiones que el ingestor todavía necesita.
//
// LOS DOS HALLAZGOS, y sólo uno mueve el exit code:
//
//	inconsistente  dos features nombran distinto al MISMO término → exit 1
//	sin usar       un término del glosario que ninguna feature menciona → informativo
//
// La asimetría es deliberada y es la misma lección de `DL-4`: si un glosario con
// términos sin usar pusiera el comando en rojo, saldría en rojo en todo proyecto
// real desde el primer día, y un chequeo que siempre grita se deja de correr.
// Un término sin usar puede ser vocabulario que todavía no llegó a una spec —
// eso es normal. Dos nombres para una cosa nunca es normal.
// ----------------------------------------------------------------------------

// termFinding es un término del glosario nombrado de más de una forma.
type termFinding struct {
	Term string `json:"term"`
	// Forms mapea la forma superficial encontrada → las features que la usan.
	// Se reporta el mapa entero y no sólo "hay conflicto" porque la acción
	// depende de quién dice qué: si nueve features dicen `Order` y una dice
	// `Pedido`, se corrige la una; si está tres a tres, hay que decidir cuál es
	// el nombre y eso es una conversación, no un fix.
	Forms map[string][]string `json:"forms"`
}

type termsReport struct {
	Inconsistent []termFinding `json:"inconsistent"`
	Unused       []string      `json:"unused"`
}

// termFormRe arma el matcher de UNA forma superficial.
//
// Los límites de palabra no son un detalle: sin ellos `Order` matchearía dentro
// de `Reorder` y de `Ordering`, y el reporte se llenaría de falsos que lo
// volverían inservible. Case-insensitive porque `Order` y `order` son la misma
// palabra — una diferencia de mayúscula no es "otro nombre para el concepto", y
// reportarla sería ruido con forma de hallazgo.
func termFormRe(form string) (*regexp.Regexp, bool) {
	f := strings.TrimSpace(form)
	if f == "" {
		return nil, false
	}
	return regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(f) + `\b`), true
}

// requirementText junta el texto de un requisito donde puede aparecer un término
// del dominio. Se mira lo que DESCRIBE el comportamiento, no los ids ni las
// rutas: un `path:symbol` que contenga la palabra no es alguien nombrando el
// concepto, es un archivo.
func requirementText(r requirement) string {
	var b strings.Builder
	b.WriteString(r.Trigger)
	b.WriteByte(' ')
	b.WriteString(r.Behavior)
	for _, c := range r.Acceptance {
		b.WriteByte(' ')
		b.WriteString(c.Text)
		b.WriteByte(' ')
		b.WriteString(c.Given)
		b.WriteByte(' ')
		b.WriteString(c.When)
		b.WriteByte(' ')
		b.WriteString(c.Then)
	}
	return b.String()
}

// computeTermsReport recorre glosario × features y arma los dos hallazgos.
func computeTermsReport(projectDir string) termsReport {
	var rep termsReport

	df, ok := loadDomainQuiet(projectDir)
	if !ok || len(df.Glossary) == 0 {
		return rep
	}

	// El texto de cada feature VIVA, una sola vez. Las cerradas quedan afuera por
	// el mismo motivo que en coverage y drift (DL-5 F3): una spec retirada a
	// propósito no es terminología que haya que unificar.
	texts := map[string]string{}
	ff, err := readFeaturesFile(projectDir)
	if err != nil {
		return rep
	}
	for _, f := range ff.Features {
		if terminalStatuses[f.Status] {
			continue
		}
		var b strings.Builder
		for _, r := range requirementsOf(projectDir, f.Name) {
			b.WriteString(requirementText(r))
			b.WriteByte(' ')
		}
		if s := strings.TrimSpace(b.String()); s != "" {
			texts[f.Name] = s
		}
	}

	features := make([]string, 0, len(texts))
	for name := range texts {
		features = append(features, name)
	}
	sort.Strings(features)

	for _, g := range df.Glossary {
		forms := map[string][]string{}
		// El término canónico y sus alias son la MISMA entrada del glosario: son
		// las formas superficiales del mismo concepto, que es justo lo que hay
		// que contrastar entre features.
		for _, form := range append([]string{g.Term}, g.Aliases...) {
			re, ok := termFormRe(form)
			if !ok {
				continue
			}
			for _, name := range features {
				if re.MatchString(texts[name]) {
					forms[form] = append(forms[form], name)
				}
			}
		}
		switch {
		case len(forms) == 0:
			rep.Unused = append(rep.Unused, g.Term)
		case len(forms) > 1:
			rep.Inconsistent = append(rep.Inconsistent, termFinding{Term: g.Term, Forms: forms})
		}
	}
	sort.Strings(rep.Unused)
	sort.Slice(rep.Inconsistent, func(i, j int) bool {
		return rep.Inconsistent[i].Term < rep.Inconsistent[j].Term
	})
	return rep
}

// runDomainTerms es el subcomando.
func runDomainTerms(args []string) int {
	projectDir := "."
	asJSON := false
	for _, a := range args {
		switch {
		case a == "--json":
			asJSON = true
		case strings.HasPrefix(a, "-"):
			fmt.Fprintf(os.Stderr, "sf domain terms: unknown flag %q\n", a)
			return 2
		default:
			projectDir = a
		}
	}

	rep := computeTermsReport(projectDir)

	if asJSON {
		if rep.Inconsistent == nil {
			rep.Inconsistent = []termFinding{}
		}
		if rep.Unused == nil {
			rep.Unused = []string{}
		}
		out, _ := json.MarshalIndent(rep, "", "  ")
		fmt.Println(string(out))
		if len(rep.Inconsistent) > 0 {
			return 1
		}
		return 0
	}

	if len(rep.Inconsistent) == 0 && len(rep.Unused) == 0 {
		fmt.Println("Domain terminology is consistent across live features.")
		return 0
	}

	for _, f := range rep.Inconsistent {
		fmt.Printf("%s is named %d different ways:\n", f.Term, len(f.Forms))
		names := make([]string, 0, len(f.Forms))
		for form := range f.Forms {
			names = append(names, form)
		}
		sort.Strings(names)
		for _, form := range names {
			fmt.Printf("  %-20s %s\n", form, strings.Join(f.Forms[form], ", "))
		}
		fmt.Println()
	}

	if len(rep.Unused) > 0 {
		// Informativo a propósito: no mueve el exit code. Ver el encabezado.
		fmt.Printf("%d glossary term(s) no live requirement mentions: %s\n",
			len(rep.Unused), strings.Join(rep.Unused, ", "))
		fmt.Println("That may just be vocabulary that hasn't reached a spec yet — not an error.")
	}

	if len(rep.Inconsistent) > 0 {
		fmt.Printf("\n%d concept(s) go by more than one name. Pick the one the glossary declares and\n",
			len(rep.Inconsistent))
		fmt.Println("amend the features that use another, or add the missing alias if both are correct.")
		fmt.Println("This is the defect that stays invisible until two features have to integrate.")
		return 1
	}
	return 0
}
