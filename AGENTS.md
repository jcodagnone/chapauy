# ChapaUY: Guía para Agentes de IA

## Estructura del Proyecto

- `cmd/`: Punto de entrada de la CLI (`main.go`).
- `impo/`: Adquisición y extracción de documentos — ver [Adquisición](web/docs/010-acquire.md).
- `curation/`: Geocodificación y normalización de descripciones — ver [Enriquecimiento](web/docs/020-curate.md).
- `web/`: Frontend Next.js 15+ — ver [Arquitectura](web/docs/000-arquitectura.md).
- `infra/`: Infraestructura como código.

Documentación de referencia: [Arquitectura](web/docs/000-arquitectura.md) · [Adquisición](web/docs/010-acquire.md) · [Curación](web/docs/020-curate.md).

## Comandos

Backend: `make test` · `make` (linter) · `go run main.go impo update` · `go run main.go curation serve`

Frontend: `cd web && pnpm dev` · `pnpm test` · `pnpm lint`

## Principios de Desarrollo

- Código mínimo que resuelve el problema. Sin abstracciones para un solo uso ni flexibilidad no pedida.
- Si una lógica o constante se usa más de una vez, extráigala. No duplique código.
- Agote la biblioteca estándar antes de añadir dependencias externas.
- Nunca ignore errores; propague siempre con contexto: `fmt.Errorf("falló al realizar X: %w", err)`.
- Si una funcionalidad no está probada, no existe. Defina cómo probarla antes de escribir el código.
- No escriba código de funcionalidad sin su definición en `web/docs/` o `prds/`.
- Estilo Rioplatense formal: use "nosotros", "usted" y terminología local.

## Uso de Git y Disciplina de Historia

- Commits atómicos: primero el refactor, luego la funcionalidad, nunca juntos.
- Stage intencional — nunca `git add .` a ciegas.
- NUNCA incluir el trailer `Co-authored-by:`.
- Las ramas temáticas se mantienen lineales (`git pull --rebase`). Sin burbujas de merge dentro de una rama.
- No enmendar commits ya publicados en ramas compartidas.

**Los mensajes de commit explican el PORQUÉ, no solo el QUÉ.**
`config: add tsconfig/library.json` no dice nada. ¿Qué desbloquea? ¿Por qué ahora?
Un buen cuerpo responde: qué problema resuelve, qué habilita, qué compromiso se tomó.

**Merge a master: siempre `--no-ff`.** El commit de merge es la carta de presentación:
la intención global, qué logra la serie de commits, qué desbloquea. Los lectores futuros
deben entender el cambio leyendo `git log --merges` sin necesidad de revisar cada commit.

Verificación antes del commit: `git status --short` muestra solo los archivos esperados;
`git diff --staged` coincide con la intención. Asunto ≤ 50 caracteres, cuerpo a 72.
