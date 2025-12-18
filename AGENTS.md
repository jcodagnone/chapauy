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

## ⚠️ MANDATORY TASK COMPLETION CHECKLIST ⚠️

**🔴 BEFORE MARKING ANY TASK/SUBTASK AS COMPLETE:**

□ **Integration Tests Written**: Write integration tests for new functionality.
□ **All Tests Pass**: Run `make test` (backend) AND `pnpm test` (frontend).
□ **All Linting Passes**: Run `make` (backend) AND `pnpm lint` (frontend).
□ **Documentation Updated**: If adding new features, update the relevant files in `web/docs/`.
□ **Rioplatense Formal**: Verify that all user-facing documentation follows the "Rioplatense formal" style.

## PERMANENT INSTRUCTIONS

**CRITICAL DEVELOPMENT RULES - NEVER OVERRIDE:**

1. **Always Write Integration Tests**: You MUST write or update integration tests to validate changes.
2. **Never Claim Done with Failing Tests**: A task is NOT complete if any tests are failing.
3. **No Code Duplication**: Search the codebase for existing utilities before implementing new ones. Re-use logic from `impo` or `curation` packages.
4. **Consistency**: Follow the established architectural patterns (Repository pattern in frontend, CLI structure in backend).
5. **Rioplatense style**: Use "nosotros", "usted", and local terminology
