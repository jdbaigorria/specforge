package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/jdbaigorria/specforge/sf/internal/andamio"
	"github.com/jdbaigorria/specforge/sf/internal/git"
	"github.com/jdbaigorria/specforge/sf/internal/global"
	"github.com/jdbaigorria/specforge/sf/internal/maquina"
	"github.com/jdbaigorria/specforge/sf/internal/registro"
)

// parada corre los cinco comandos con los que Javier le contesta a una parada.
//
// Van juntos en una función porque comparten el esqueleto entero —cargar,
// ejecutar, guardar el estado, imprimir— y lo único que cambia es qué llaman.
// Separarlos sería copiar cinco veces las mismas veinte líneas.
//
// Y comparten algo más importante: LOS CINCO LOS CORRE EL ORQUESTADOR, nunca el
// subagente. Son la mitad del reparto que la ronda de la superficie descubrió.
func parada(cmd string, args []string) int {
	raiz, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, "sf:", err)
		return SalidaError
	}

	e, r, err := cargar(raiz)
	if err != nil {
		fmt.Fprintln(os.Stderr, "sf:", err)
		return SalidaError
	}

	arg := func(i int) string {
		if i < len(args) {
			return args[i]
		}
		return ""
	}

	// El catálogo sólo lo toca `sf model`. Se lee igual para todos porque leerlo
	// es barato y el error de "no está" ya está contemplado: g queda nil y
	// Modelo() lo maneja.
	//
	// LeerPara y no Leer: tiene que escribir en el MISMO archivo que después va
	// a leer `sf next`, y el del proyecto le gana al global.
	g, _ := global.LeerPara(raiz)

	var ef maquina.Efecto
	switch cmd {
	case "approve":
		ef = maquina.Aprobar(raiz, e, r)
	case "reject":
		ef = maquina.Rechazar(e, arg(0))
	case "take":
		ef = maquina.Tomar(e, r, arg(0))
	case "model":
		fm, err := flagsDeModelo(args)
		if err != nil {
			fmt.Fprintln(os.Stderr, "sf model:", err)
			return SalidaError
		}
		ef = maquina.Modelo(e, g, maquina.Declaracion(fm))
	case "dismiss":
		ef = maquina.Descartar(raiz, e, r, arg(0), arg(1))
	case "lote start":
		ef = maquina.EmpezarLote(raiz, e, r)
	case "new":
		// Todo lo que venga después del comando es el texto, no flags: `sf new
		// que sf soporte brownfield` tiene que funcionar sin comillas.
		ef = maquina.Nueva(raiz, e, strings.Join(args, " "))
	}

	// Las cinco respuestas de Javier y el `lote start` pasan por acá, así que
	// una sola llamada cubre los seis. `Movio` no existe en `Efecto`: lo
	// equivalente es que no haya fallas, o sea que el sello se dio.
	registro.AnotarVeredicto(ef.Estado, ef.Fallas, ef.Avisos, ef.Pasa())

	// Un alias nuevo necesita su portamodelo AHORA. La sesión viva no lo va a
	// ver —los agentes se leen al arrancar, y por eso `sf model` avisa que hay
	// que reiniciar— pero el archivo tiene que quedar escrito igual: si se
	// dejara para el próximo `sf install`, reiniciar no alcanzaría.
	if ef.Portamodelo && g != nil {
		if _, err := andamio.Regenerar(raiz, g); err != nil {
			fmt.Fprintln(os.Stderr, "sf model: no pude escribir el portamodelo:", err)
			return SalidaError
		}
	}

	// El estado se guarda si el comando funcionó. Un `sf take f-99` que falla no
	// tiene que dejar rastro.
	//
	// Y `Cambio` es la excepción, que existe por un caso concreto: `archivar`
	// toca el disco —mueve la carpeta, mergea, borra la branch— y si algo falla
	// DESPUÉS de eso, no guardar deja al estado.json describiendo un repo que ya
	// no existe. Cuando el comando falló y el mundo cambió igual, lo correcto es
	// anotar lo que sí pasó.
	if ef.Pasa() || ef.Cambio {
		if err := e.Guardar(raiz); err != nil {
			fmt.Fprintln(os.Stderr, "sf: no pude guardar el estado:", err)
			return SalidaError
		}
		// Y el mapa global sólo cuando algo lo cambió: escribirlo en cada
		// `sf approve` sería tocar el home de Javier para no cambiar nada.
		if ef.Global && g != nil {
			if err := g.Guardar(); err != nil {
				fmt.Fprintln(os.Stderr, "sf: no pude guardar ~/.specforge/:", err)
				return SalidaError
			}
		}
		// El commit va acá y no adentro del comando, porque recién ahora el
		// estado.json está escrito y puede entrar en el mismo commit. Hoy sólo
		// lo pide `archivar`: es el único lugar donde sf mueve archivos sin
		// estar cerrando un lote.
		//
		// Que falle NO invalida lo que ya pasó —la carpeta se movió, la branch
		// se mergeó, el estado se guardó—, así que avisa y sigue. Frenar acá
		// sería reportar como error un archivado que salió bien.
		if ef.Commit != "" {
			if _, err := git.Commit(raiz, ef.Commit); err != nil {
				fmt.Fprintln(os.Stderr, "sf: quedó sin commitear:", err)
			}
		}
	}

	fmt.Print(ef.Texto())
	if ef.Pasa() {
		return SalidaTrabajo
	}
	return SalidaError
}
