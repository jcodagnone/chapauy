# ChapaUY: Guía para Agentes de IA

Este documento sirve de guía para agentes de IA (como Antigravity, Gemini, Copilot, v0) que trabajen en ChapaUY. Provee el contexto esencial y las reglas de desarrollo, delegando la documentación técnica detallada a los archivos en `web/docs/`.

## Estructura del Proyecto

El proyecto está organizado en los siguientes paquetes principales:

- `cmd/`: Punto de entrada de la CLI (`main.go`).
- `impo/`: Lógica de adquisición, descubrimiento y extracción de documentos (ver [Adquisición](web/docs/010-acquire.md)).
- `curation/`: Servidor de curación para geocodificación y normalización de descripciones (ver [Enriquecimiento](web/docs/020-curate.md)).
- `web/`: Aplicación frontend Next.js 15+ (ver [Arquitectura](web/docs/000-arquitectura.md)).
- `infra/`: Provisión de infraestructura mediante código (ver [Arquitectura](web/docs/000-arquitectura.md)).

## Documentación de Referencia

La fuente de verdad para la lógica del sistema y su arquitectura se encuentra en:
- [**Arquitectura**](web/docs/000-arquitectura.md): Visión general, esquema de DuckDB y despliegue.
- [**Adquisición de datos**](web/docs/010-acquire.md): Funcionamiento del pipeline `impo`.
- [**Enriquecimiento / Curación**](web/docs/020-curate.md): Lógica de análisis de matrículas (estrategia de bloques) y herramientas de curación.

## Desarrollo y Comandos Útiles

### Backend (Go)
- `go run main.go impo update`: Ejecuta el pipeline de actualización.
- `go run main.go curation serve`: Inicia el servidor de curación local.
- `make test`: Ejecuta todos los tests de Go.
- `make`: Ejecuta el linter.

### Frontend (Next.js)
- `cd web && pnpm dev`: Inicia el servidor de desarrollo frontend.
- `pnpm test`: Ejecuta tests de frontend con Vitest.

---

## Uso de Git

### Filosofía de Commit
La historia de commits debe contar una historia clara y lineal. Preferimos múltiples commits pequeños y atómicos en lugar de uno solo monolítico.

- **Commits Atómicos**: Si una funcionalidad requiere una refactorización previa, realice el commit de la refactorización **primero** y la funcionalidad **después**. No los mezcle.
- **Staging Selectivo**: Evite `git add .`. Use staging manual o `git add -p` para revisar y confirmar cada cambio.
- **Historia Lineal (Rebase)**: Mantenga una historia limpia. Al actualizar su rama, use siempre `git pull --rebase`.
- **Sin Firma**: Los agentes no tienen acceso a claves privadas. Use explícitamente el flag `--no-gpg-sign`.
- **Mensajes de Commit**: 
  - **Asunto**: Máximo 50 caracteres.
  - **Cuerpo**: Explique el *porqué* del cambio, no solo el *qué*. Ajuste el texto a 72 columnas.
  - **Trailer**: Incluya `Co-authored-by:` cuando la participación del agente sea relevante.

## Calidad de Código y Dependencias

- **KISS & YAGNI**: Escriba código simple y no implemente soluciones para problemas hipotéticos futuros.
- **DRY (Don't Repeat Yourself)**: Si una lógica o constante se usa más de una vez, **extráigala** inmediatamente a una función o constante compartida. No duplique código.
- **Dependencias Mínimas**: Agote las capacidades de la biblioteca estándar antes de añadir paquetes externos. Cada dependencia nueva requiere justificación.

## Manejo de Errores y Testing

- **Fail Loudly**: Nunca ignore ni oculte errores. Si una función devuelve un error, debe ser manejado o propagado con contexto: `fmt.Errorf("falló al realizar X: %w", err)`.
- **Existence Proof**: Si una funcionalidad no está probada, no existe. Entienda cómo probará el cambio antes de escribir el código.

## Documentación del Producto

- **Docs as Code**: El repositorio es la única fuente de verdad. No se debe escribir código de funcionalidad sin una definición de producto o documento de descubrimiento correspondiente en `web/docs/` o `prds/`.

---

## ⚠️ MANDATORY TASK COMPLETION CHECKLIST ⚠️

**🔴 BEFORE MARKING ANY TASK/SUBTASK AS COMPLETE:**

□ **Integration Tests Written**: Write integration tests for new functionality.
□ **All Tests Pass**: Run `make test` (backend) AND `pnpm test` (frontend).
□ **All Linting Passes**: Run `make` (backend) AND `pnpm lint` (frontend).
□ **Documentation Updated**: If adding new features, update the relevant files in `web/docs/`.
□ **Rioplatense Formal**: Verify that all user-facing documentation follows the "Rioplatense formal" style.

## PERMANENT INSTRUCTIONS

**CRITICAL DEVELOPMENT RULES - NEVER OVERRIDE:**

1. **Existence Proof**: You MUST write or update integration tests to validate changes. A feature without tests does not exist.
2. **Never Claim Done with Failing Tests**: A task is NOT complete if any tests are failing.
3. **DRY (No Code Duplication)**: Search the codebase for existing utilities before implementing new ones. Re-use logic from `impo` or `curation` packages. If logic or constants are used more than once, extract them.
4. **Consistency**: Follow the established architectural patterns (Repository pattern in frontend, CLI structure in backend).
5. **Rioplatense style**: Use "nosotros", "usted", and local terminology.
6. **Error Wrapping (Go)**: Always wrap errors with context using `%w` for preservation of the error chain.
7. **Entity Links**: Ensure all entity references in the UI are linked to their detail views.
