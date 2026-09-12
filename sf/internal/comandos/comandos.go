// Package comandos es el inventario de la superficie de sf: qué comandos hay.
//
// ────────────────────────────────────────────────────────────────────────────
// POR QUÉ UNA LISTA Y NO EL `switch` DE main.go
// ────────────────────────────────────────────────────────────────────────────
//
// Porque la lista hacía falta en tres lugares y estaba escrita en dos:
//
//	el switch de main.go        el que despacha de verdad
//	el mensaje de "no existe"   una copia a mano, con los nombres tipeados
//	sf doctor                   el tercero, que es el que trajo el problema
//
// Dos copias ya eran una de más —agregar un comando y olvidarse del mensaje no
// rompe nada, sólo hace mentir a la ayuda—, y una tercera lo volvía inevitable.
//
// El `switch` no se puede eliminar sin convertir el despacho en un mapa de
// funciones, que es peor: se pierde el orden de lectura y la firma de cada
// comando. Lo que sí se puede es COMPROBAR que los dos coinciden, y eso lo hace
// un test punta a punta que le pasa cada nombre de esta lista al binario y mira
// que no conteste "todavía no está construido".
//
// ────────────────────────────────────────────────────────────────────────────
// PARA QUÉ LO QUIERE EL DOCTOR
// ────────────────────────────────────────────────────────────────────────────
//
// Un skill y el binario se instalan por caminos distintos —el plugin trae los
// skills, `install.sh` trae el binario— y por lo tanto pueden quedar
// desalineados. Un skill de una versión más nueva nombra `sf loquesea` y el
// binario contesta que no existe: el bucle se traba y el que lo lee es un
// agente que no sabe que hay dos mitades.
//
// `sf doctor` lo detecta preguntando un HECHO —¿los comandos que nombran los
// skills instalados existen en ESTE binario?— en vez de comparar dos números de
// versión, que es lo que se rompe cuando alguien instala desde el repo.
package comandos

import (
	"slices"
	"strings"
)

// Todos son los comandos que el binario acepta, en el orden de la ayuda.
//
// `lote start` está con su espacio adentro a propósito: es el único de dos
// palabras del inventario, y quien compara contra esta lista tiene que poder
// distinguirlo de un `sf lote` suelto, que NO existe.
var Todos = []string{
	"init",
	"install",
	"uninstall",
	"models",
	"next",
	"lanzar",
	"context",
	"done",
	"lote start",
	"new",
	"status",
	"log",
	"audit",
	"doctor",
	"version",
	"approve",
	"reject",
	"take",
	"model",
	"dismiss",
	"ampliar",
	"help",
}

// Existe dice si sf sabe atender ese comando.
//
// Recibe el comando ya armado —"lote start", no "lote"— porque el que llama es
// quien sabe cuántas palabras leyó.
func Existe(c string) bool { return slices.Contains(Todos, c) }

// Lista arma la línea que se le muestra a alguien que erró el nombre.
func Lista() string { return strings.Join(Todos, " · ") }
