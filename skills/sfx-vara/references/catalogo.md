# El catálogo — las tres varas que no se reescriben

Tres materiales llegan siempre con la misma forma, así que su vara ya está escrita. Se lee, se le
agregan a lo sumo uno o dos criterios propios del caso, y se dice cuáles se agregaron.

Cada criterio se contesta con un hecho. Al lado de cada uno está **cómo se contesta**, y si eso
necesita a Javier o no.

---

## ① Una feature — *¿esto entra, y cuánto proceso pide?*

```
V-1  ¿Cae adentro de lo que este producto dice que es?
     cómo: la constitución, o .docs/repo/repo.md          corre solo

V-2  ¿Hay un flujo parecido en el repo para ir a leer?
     cómo: .docs/repo/flujos.md, o buscarlo               corre solo

V-3  ¿Se puede escribir el test antes que el código?
     cómo: nombrá el test. Si no podés, no es un criterio  corre solo

V-4  ¿Alguien lo pidió por una situación concreta?
     cómo: la situación, con nombre                        te necesita

V-5  Si no se hace ahora, ¿qué se rompe?
     cómo: lo que se rompe, o "nada"                       te necesita
```

> **La V-2 decide sola cuánto proceso pide.** Si hay un flujo parecido para leer, el cambio es
> acotado; si no hay nada, necesita diseño. Es un hecho sobre el repo, no una impresión sobre vos
> mismo — y eso es lo que `sfx-mapa` ya dice de `chico`: *"mide el repo, no tu confianza"*.
>
> Y el trinquete corre para un solo lado: de acotado se sube a largo cuando la máquina comprueba
> que no era acotado, y no se baja nunca.

---

## ② Un bug — *¿cuál de estos arreglos?*

Sale casi entera de `superpowers/systematic-debugging`, leído el 2026-09-15.

```
V-1  ¿Podés nombrar la causa? ¿Este arreglo la ataca, o tapa el síntoma?
     cómo: la causa, en una línea, con file:line           corre solo

V-2  ¿Hay un test que falla ahora y pasa después?
     cómo: nombrá el test y correlo en rojo                corre solo

V-3  ¿Cambia una sola cosa?
     cómo: contá los archivos y los motivos                corre solo

V-4  ¿Rompe algo que hoy anda?
     cómo: corré la suite                                  corre solo

V-5  ¿Arregla donde NACE el problema, o donde se VE el error?
     cómo: rastreá el valor malo hacia atrás               corre solo
```

> **La V-5 es de ellos y es la más valiosa.** El error aparece lejos de donde se originó, y
> arreglarlo donde aparece es tapar el síntoma con otro nombre.
>
> **Las cinco corren solas.** Elegir un arreglo no te necesita hasta el veredicto.

---

## ③ Un hallazgo de revisión — *¿el revisor tiene razón?*

Traducida de `superpowers/receiving-code-review`, leído el 2026-09-15. Ellos las tienen como
"razones para discutir el hallazgo"; acá son criterios.

```
V-1  ¿Rompe algo que ya funciona?
     cómo: corré la suite con el cambio propuesto          corre solo
     sí → el hallazgo pierde fuerza

V-2  ¿El que revisó tenía todo el contexto?
     cómo: ¿le llegó el sobre entero, o miró un diff suelto?  corre solo

V-3  ¿Alguien usa eso realmente?
     cómo: grep. Si no lo llama nadie, la respuesta no es    corre solo
           implementarlo mejor — es borrarlo

V-4  ¿Es correcto para ESTE stack y esta versión?
     cómo: el manifiesto y la versión mínima soportada     corre solo

V-5  ¿Hay un motivo viejo que explica por qué está así?
     cómo: git log / git blame de esas líneas              corre solo

V-6  ¿Choca con una decisión que ya se tomó?
     cómo: .docs/archivado/ y .docs/repo/decidido.md       corre solo
```

> **Las seis corren solas.** Decidir si un hallazgo es portante no necesita a nadie hasta el
> veredicto — y hoy esa decisión existe como comando (`sf dismiss <h-#> "motivo"`) sin ningún
> criterio detrás. Es el hueco que abrió el `f828ec2` ③.

---

## La cuarta no está acá

La vara de **una idea** es a medida y crece durante la entrevista. No se puede poner en un
catálogo, porque los criterios salen de la conversación. Ver `SKILL.md`, *"Bespoke"*.
