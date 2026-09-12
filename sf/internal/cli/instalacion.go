package cli

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/jdbaigorria/specforge/sf/internal/andamio"
	"github.com/jdbaigorria/specforge/sf/internal/global"
	"github.com/jdbaigorria/specforge/sf/internal/modelos"
	"github.com/jdbaigorria/specforge/sf/internal/pregunta"
	"github.com/jdbaigorria/specforge/sf/internal/registro"
)

// instalar es `sf install`: el andamio.
//
// No es de la máquina —no mira el estado ni lo mueve— y por eso no aparece en
// ningún trazado del bucle. Es lo que hace que SpecForge se pueda USAR: pone el
// orquestador en el proyecto y arma ~/.specforge/.
func instalar(args []string) int {
	var o andamio.Opciones
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--forzar" || a == "--force":
			o.Forzar = true
		// --harness acepta coma y además es repetible. Las dos formas porque las
		// dos aparecen solas: `--harness=a,b,c` es lo que uno tipea, y
		// `--harness a --harness b` es lo que sale de un script que arma la lista
		// en un bucle. Sostener las dos cuesta una línea.
		case a == "--harness" && i+1 < len(args):
			i++
			o.Harness = append(o.Harness, separarPorComa(args[i])...)
		case strings.HasPrefix(a, "--harness="):
			o.Harness = append(o.Harness, separarPorComa(strings.TrimPrefix(a, "--harness="))...)
		default:
			fmt.Fprintf(os.Stderr, "sf install: no conozco %q. Sólo --harness y --forzar.\n", a)
			return SalidaError
		}
	}

	raiz, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, "sf:", err)
		return SalidaError
	}

	// LA ÚNICA VEZ QUE SF CONVERSA, y sólo si se dan las dos condiciones.
	//
	// Con flags no se pregunta, haya persona o no: pasar `--harness=` ES la
	// respuesta, y volver a preguntarla sería no escuchar. Y sin TTY tampoco,
	// que es el caso que no puede fallar nunca — un `sf install` corrido por el
	// orquestador o por `install.sh` colgaría esperando algo que nadie va a
	// escribir.
	var (
		declaraciones []pregunta.Declaracion
		permisos      map[string]string
	)
	if len(o.Harness) == 0 && registro.HayPersona() {
		resp, err := pregunta.Preguntar(os.Stdout, os.Stdin, andamio.Instalados(), modelos.Listar)
		if err != nil {
			fmt.Fprintln(os.Stderr, "sf:", err)
			return SalidaError
		}
		o.Harness = resp.Harnesses
		declaraciones = resp.Declaraciones
		permisos = resp.Permisos
		fmt.Println()
	}

	r, err := andamio.Instalar(raiz, o)
	if err != nil {
		fmt.Fprintln(os.Stderr, "sf:", err)
		return SalidaError
	}

	// Lo elegido se declara DESPUÉS de instalar, porque instalar es lo que
	// siembra el catálogo — antes puede no haber ninguno donde escribir. Y por
	// eso hay que regenerar los portamodelo: el andamio se armó cuando estos
	// modelos todavía no existían.
	if len(declaraciones) > 0 || len(permisos) > 0 {
		if err := declararYRegenerar(raiz, declaraciones, permisos, r); err != nil {
			fmt.Fprintln(os.Stderr, "sf:", err)
			return SalidaError
		}
	}

	for _, e := range r.Escritos {
		fmt.Println("+", e)
	}
	for _, s := range r.Salteados {
		fmt.Println("·", s)
	}
	for _, v := range r.Viejos {
		fmt.Printf("⚠ %s quedó de una versión anterior — `sf install --forzar` lo actualiza\n", v)
	}

	fmt.Println()
	for _, h := range r.Para {
		fmt.Printf("%-12s %d modelos declarados\n", h, r.Modelos[h])
	}
	// El puntero sólo se nombra cuando hay más de uno, que es cuando la pregunta
	// "¿y cuál usa si no sabe dónde está?" se le puede ocurrir a alguien. Con uno
	// solo, decirlo sería ruido: es obvio.
	if len(r.Para) > 1 {
		fmt.Printf("\npuntero: %s — el que resuelve modelos si no se detecta ninguno\n", r.Puntero)
	}

	if r.Puntero == "desconocido" {
		// Se avisa fuerte porque el harness es la mitad de H1b: sin él, sf no
		// puede decidir si un modelo va por subagente o por consola.
		fmt.Println("⚠ No reconocí el harness. Corregilo con `sf install --harness=<nombre>`.")
		return SalidaTrabajo
	}

	// Y si hay OTROS arneses en esta máquina, se dicen. No se instalan: sólo se
	// nombran, porque instalar para un arnés que no pediste sería decidir por
	// Javier. Es el único uso de la detección hasta que exista la pregunta
	// interactiva, y ya paga: hoy uno se entera de que le falta preparar un arnés
	// cuando lo abre y el bucle se traba.
	//
	// Sólo con `sf install` pelado: si vino --harness, ya dijiste qué querías.
	if len(o.Harness) == 0 {
		if faltan := sinPreparar(r.Para); len(faltan) > 0 {
			fmt.Println()
			fmt.Printf("También tenés %s en esta máquina.\n", strings.Join(faltan, " y "))
			fmt.Printf("    sf install --harness=%s\n", strings.Join(append(r.Para, faltan...), ","))
		}
	}
	return SalidaTrabajo
}

// desinstalar es `sf uninstall`: saca el orquestador y NO toca ~/.specforge/.
func desinstalar() int {
	raiz, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, "sf:", err)
		return SalidaError
	}

	r, err := andamio.Desinstalar(raiz)
	if err != nil {
		fmt.Fprintln(os.Stderr, "sf:", err)
		return SalidaError
	}

	for _, e := range r.Escritos {
		fmt.Println("-", e)
	}
	for _, s := range r.Salteados {
		fmt.Println("·", s)
	}
	if len(r.Escritos) == 0 && len(r.Salteados) == 0 {
		fmt.Println("No había nada que sacar.")
	}
	fmt.Println()
	fmt.Println("~/.specforge/ no se toca: tus modelos son de la máquina, no de este proyecto.")
	return SalidaTrabajo
}

// flagsDeModelo parte `sf model <perfil|alias> [--alias …] [--id …] [--via …] [--comando "…"]`.
//
// Los flags son cómo se DECLARA un modelo, y por eso van acá y no en un
// `sf model add` aparte: aprobar es declarar (ver maquina.Modelo). Sin flags, el
// mismo comando es el ⑳ — "usá este de acá en adelante".
//
// Los cuatro se parsean juntos y el que decide qué significan es maquina.Modelo:
// este parser no sabe de perfiles ni de alias, sólo parte la línea. Meterle la
// lógica acá la pondría fuera del alcance de los tests de la máquina, que es
// donde tiene que estar.
func flagsDeModelo(args []string) (f flagsModelo, err error) {
	fallo := func(s string, a ...any) (flagsModelo, error) {
		return flagsModelo{}, fmt.Errorf(s, a...)
	}
	destinos := map[string]*string{
		"--alias": &f.Alias, "--id": &f.ID, "--via": &f.Via,
		"--comando": &f.Comando, "--esfuerzo": &f.Esfuerzo,
	}

	for i := 0; i < len(args); i++ {
		a := args[i]

		if d, hay := destinos[a]; hay {
			if i+1 >= len(args) {
				return fallo("`%s` sin valor", a)
			}
			i++
			*d = args[i]
			continue
		}
		if k, v, corta := strings.Cut(a, "="); corta {
			if d, hay := destinos[k]; hay {
				*d = v
				continue
			}
		}
		if strings.HasPrefix(a, "-") {
			return fallo("no conozco %q. Son --alias, --id, --via, --comando y --esfuerzo", a)
		}
		if f.Nombre != "" {
			return fallo("dos nombres: %q y %q", f.Nombre, a)
		}
		f.Nombre = a
	}
	return f, nil
}

// flagsModelo son los cinco de `sf model`, juntos.
//
// Se agrupan en un struct y no en cinco retornos porque cinco strings en fila
// son cinco oportunidades de invertir dos al llamar, y el compilador no ayuda:
// son todos del mismo tipo.
type flagsModelo struct {
	Nombre, Alias, ID, Via, Comando, Esfuerzo string
}

// separarPorComa parte "a,b,c" y descarta los vacíos de un "a,,b" o un "a,".
func separarPorComa(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// sinPreparar son los arneses instalados en la máquina para los que NO se acaba
// de armar el andamio.
//
// Cuesta ejecutar un binario por cada uno que falta, así que se pregunta sólo
// por los que no están en `ya` — como mucho dos, y sólo en el `sf install` pelado.
func sinPreparar(ya []string) []string {
	var faltan []string
	for _, h := range global.Harness {
		if !slices.Contains(ya, h) && andamio.Hay(h) {
			faltan = append(faltan, h)
		}
	}
	return faltan
}

// declararYRegenerar mete en el catálogo lo que Javier eligió y reescribe los
// portamodelo de los arneses tocados.
func declararYRegenerar(raiz string, ds []pregunta.Declaracion, permisos map[string]string, r *andamio.Resultado) error {
	// El del proyecto si existe, y si no el global: es el que rige acá, y es
	// donde `sf model` escribiría si Javier lo declarara a mano.
	g, err := global.LeerPara(raiz)
	if err != nil {
		return err
	}

	// El permiso va al mismo archivo que los modelos y por el mismo motivo: es
	// una decisión de Javier que sf sólo transporta.
	for h, permiso := range permisos {
		g.DeclararPermiso(h, permiso)
	}

	tocados := map[string]bool{}
	for _, d := range ds {
		if _, err := g.DeclararEn(d.Harness, d.Perfil, global.Modelo{
			Alias: d.Alias, ID: d.ID, Via: global.Subagente, Esfuerzo: d.Esfuerzo,
		}); err != nil {
			return err
		}
		tocados[d.Harness] = true
	}
	if err := g.Guardar(); err != nil {
		return err
	}

	var cuales []string
	for _, h := range r.Para {
		if tocados[h] {
			cuales = append(cuales, h)
		}
	}
	rr, err := andamio.RegenerarPara(raiz, g, cuales)
	if err != nil {
		return err
	}
	r.Escritos = append(r.Escritos, rr.Escritos...)
	for h, n := range rr.Modelos {
		r.Modelos[h] = n
	}
	return nil
}
