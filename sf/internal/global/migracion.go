// Este archivo es el ÚNICO lugar del repo donde viven los nombres de modelo del
// formato viejo, y está separado a propósito.
//
// `TestElPaqueteNoNombraNingunModelo` prohíbe que `opus`, `sonnet` o `haiku`
// aparezcan en `global.go`, porque un nombre de proveedor en el código es
// exactamente lo que produjo el "siempre opus" (H1 y H2). La migración necesita
// nombrarlos —tiene que reconocer lo que había— así que vive acá, con la etiqueta
// puesta:
//
//	esto es código de museo. Se borra el día que ningún ~/.specforge/ del mundo
//	tenga el formato viejo, y ese día no hay que pensar nada: se borra el archivo.
package global

// migrar sube un catálogo viejo al formato nuevo, al leerlo.
//
// El formato viejo era `modelos:` con nombre → {via}, sin ids y sin perfiles. Se
// mueven los tres nativos al bloque de CLAUDE-CODE, cada uno como ÚNICO elemento
// de la lista de su perfil, con alias e id iguales al nombre viejo. Los demás
// quedan donde estaban: son los sueltos.
//
// Van a claude-code y NO al harness que el archivo declara, y esto se pagó
// caro. El puntero escrito es memoria de la última instalación, no de dónde
// salieron esos nombres: un catálogo que decía `harness: opencode` terminaba
// afirmando que opencode sabe correr `opus`, y de ahí `sf next` le entregaba al
// orquestador un id que ese harness no tiene. Era H1 —"siempre opus"— entrando
// por la ventana de la migración.
//
// Que sea claude-code no es sf opinando sobre modelos (R3): en el formato viejo
// el nombre ERA el id, y esos tres nombres son vocabulario de Claude Code
// porque el sf viejo los tenía hardcodeados. Eso es de dónde vienen, un hecho
// del pasado — no un juicio sobre a qué se parece cada uno.
//
// Se hace al leer y se persiste en el primer Guardar(). No hay comando de
// migración y no debería haberlo: un archivo que se arregla solo la primera vez
// que se lo toca es mejor que uno que exige acordarse de un comando.
//
// Es de UNA SOLA VEZ: con `harnesses:` ya poblado no vuelve a correr. Un archivo
// que quedó mal por la versión anterior de esta función se arregla a mano.
func (c *Config) migrar() {
	if len(c.Harnesses) > 0 || len(c.Modelos) == 0 || c.Harness == "" {
		return
	}
	// El harness al que pertenecían esos tres nombres, por construcción.
	const nativo = "claude-code"

	viejos := []struct{ nombre, perfil string }{
		{"opus", Razonar}, {"sonnet", Construir}, {"haiku", Mecanico},
	}
	cat := Catalogo{}
	for _, v := range viejos {
		m, hay := c.Modelos[v.nombre]
		if !hay {
			continue
		}
		m.Alias, m.ID = v.nombre, v.nombre
		cat[v.perfil] = []Modelo{m}
		delete(c.Modelos, v.nombre)
	}
	if len(cat) > 0 {
		c.Harnesses[nativo] = cat
	}
}
