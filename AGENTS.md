## ChapaUY Development Guidelines

This document provides guidelines for developers working on the ChapaUY project.

### Project Structure

The project is organized into the following packages:

- `cmd`: Contains the command-line interface for the application.
- `dashboard`: Contains the web-based dashboard for visualizing the data.
- `impo`: Contains the logic for importing and processing data from the official sources.
- `utils`: Contains utility functions used throughout the application.
- `web`: Contains the Next.js frontend application.

### Development

#### Backend
The project uses Go for the backend and the Cobra library for command-line interfaces.

#### Useful Commands
- `make`: Runs linting to check for code quality issues. **Run this periodically during development to catch errors early.**
- `make test`: Runs all the tests in the project
- `go run main.go seed`: Seeds the database with test data from `cmd/testdata/seed.json`. This is useful for development and testing.
- `go run main.go impo update <db_id>`: Updates a specific database by its ID.
- `go run main.go serve`: Runs the web dashboard server. The server will be available at `http://localhost:8080`.
- `go run main.go curation`: Runs the interactive geocoding workflow server on `http://localhost:8081` (local only, not exposed to internet). Includes database filtering for department-specific geocoding.
- `go run main.go curation description`: Runs description curation. Use with `--interactive` for interactive mode.
- `go run main.go curation description --multi`: Filters to show only multi-article descriptions.
- `go run main.go curation load`: Load description judgments from a file.
- `go run main.go curation store`: Store description judgments to a file.

### Dashboard Development

The `dashboard` package follows the repository pattern to separate the web handlers from the database logic.

- `dashboard.go`: Contains the Gin web handlers for the dashboard.

### Geocoding Development

The `geocode` package implements the interactive location geocoding workflow with database filtering capabilities.

- `server.go`: Contains the Gin web handlers for the geocoding interface
- Key endpoints:
  - `GET /api/databases`: Lists available databases for filtering
  - `GET /api/locations/queue?db_id=X`: Location queue with optional database filtering
  - `GET /api/locations/progress?db_id=X`: Progress statistics with optional database filtering
- Frontend: `templates/geocode.html` with database selector dropdown and URL state management

### Description Curation Development

The description curation workflow supports both single and multi-article offense descriptions.

- **Detection**: `DetectMultiArticle()` analyzes comma-separated parts to identify multi-article descriptions
- **Display**: Multi-article descriptions show with `# MULTI | ` marker and `## PART` labels for each part
- **Filtering**: Use `--multi` flag to show only multi-article descriptions or omit flag for all descriptions
- **Ingestion**: Supports both legacy `# DESC` format and multi-article `# MULTI | DESC` format with `## PART` sub-sections
- **Storage**: Each part is saved as an independent entry for reusability across descriptions
- **Key methods**:
  - `Suggest()`: Returns combined suggestions from full description and parts
  - `DetectMultiArticle()`: Detects if description has multiple distinct articles
  - `SuggestWithBreakdown()`: Returns suggestions grouped by comma-separated parts
  - `SuggestionBreakdown`: Type representing part-based suggestion groups

### Testing

Tests are located in the same package as the code they are testing, with the `_test.go` suffix.

- To run all tests, use the `go test ./...` command from the root of the project.
- When adding new features, please add corresponding tests to ensure that the code is working correctly and to prevent regressions.

### SQL Queries

All SQL queries should be formatted using multi-line backticked strings for better readability. For example:

```go
rows, err := db.Query(`
    SELECT
        db_id,
        COUNT(*) as count,
        COUNT(DISTINCT doc_source) as doc_count,
        MAX(time) as last_doc_date
    FROM offenses
    GROUP BY db_id
`)
```

### Dependency Management

This project uses Go modules for dependency management. To add a new dependency, use the `go get` command.

### Infrastructure & Troubleshooting

#### Developer Connect "Permission Denied"
When setting up Developer Connect (specifically the `github-repo1` connection), you might encounter a "Permission denied" error even if you have the correct roles. This is often due to limitations in the Go SDK's interaction with ADC.
*   **Resolution:** The infrastructure code (`.dagger/infra`) implements a fallback. It attempts to verify the connection's existence using the `gcloud` CLI. If `gcloud` confirms the connection exists, the error is suppressed, and the setup proceeds.
*   **Manual Verification:** You can manually verify the connection using: `gcloud developer-connect connections describe github-repo1 --location=southamerica-east1 --project=chapauy-202512`

#### Running Dagger Locally
To run Dagger pipelines locally that interact with GCP (like `data-bootstrap-and-publish`), you need to provide a valid Access Token.
*   **Command:** `export GCP_ACCESS_TOKEN=$(gcloud auth print-access-token) && dagger call data-bootstrap-and-publish --token=env:GCP_ACCESS_TOKEN`


## ⚠️ MANDATORY TASK COMPLETION CHECKLIST ⚠️

**🔴 BEFORE MARKING ANY TASK/SUBTASK AS COMPLETE:**

□ **Integration Tests Written**: Write integration tests for new functionality
□ **All Tests Pass**: Run `make test` - ALL tests must pass
□ **All linting/security scann Pass**: Run `make` - ALL linting must pass
□ **No Test Failures**: Fix any failing tests before proceeding
□ **AGENTS.md Updated**: Update this file if new features/commands/structure added

**❌ TASK IS NOT COMPLETE IF:**
- Any integration tests are failing
- New code lacks integration test coverage
- You haven't run `make` to verify

## PERMANENT INSTRUCTIONS

**CRITICAL DEVELOPMENT RULES - NEVER OVERRIDE:**

1. **Always Write Integration Tests**: When making code changes, you MUST write or update integration tests to validate the changes
2. **Always Run All Tests**: Before marking any task as complete, run `npm run test:integration` to ensure all tests pass
3. **Never Claim Done with Failing Tests**: A task is NOT complete if any tests are failing - fix all test failures first
4. **Test-Driven Validation**: Changes are only considered successful when validated by passing integration tests
5. **Always Check for Reusability**: Before implementing new functionality, ALWAYS search the codebase for existing similar functions or utilities that can be reused or extended. Never duplicate code - create shared utilities instead.
6. **AGENTS.md Updates**: Only update AGENTS.md for fundamental changes to development workflow, new architectural patterns, or structural changes that future developers need to know. Do NOT add recent updates, change logs, or temporary information - use git commits and PR descriptions for those.

---
================================================================================
# FRONTEND (Next.js) - from chapauy-frontend
================================================================================

# AGENTS.md - Guía para Agentes de IA

Este documento contiene toda la información contextual necesaria para que un agente de IA (como v0) pueda trabajar efectivamente en el proyecto ChapaUY.

## Índice

1. [Visión General del Proyecto](#visión-general-del-proyecto)
2. [Stack Tecnológico](#stack-tecnológico)
3. [Arquitectura del Proyecto](#arquitectura-del-proyecto)
4. [Tipos y Modelos de Datos](#tipos-y-modelos-de-datos)
5. [API Routes](#api-routes)
6. [Patrón Repository](#patrón-repository)
7. [Componentes Principales](#componentes-principales)
8. [Configuración de Visualización](#configuración-de-visualización)
9. [Variables de Entorno](#variables-de-entorno)
10. [Datos Mock](#datos-mock)
11. [Estilos y Tema](#estilos-y-tema)
12. [Estándares de Código](#estándares-de-código)
13. [Comandos de Desarrollo](#comandos-de-desarrollo)
14. [Rutas de la Aplicación](#rutas-de-la-aplicación)
15. [Automación y CI/CD](#automación-y-ci-cd)
16. [Puntos de Integración Clave](#puntos-de-integración-clave)

---

## Visión General del Proyecto

**ChapaUY** es una aplicación Next.js para búsqueda de infracciones de tránsito en Uruguay y Argentina. Proporciona capacidades avanzadas de búsqueda y filtrado de violaciones de tránsito a través de múltiples bases de datos.

### Características principales:

- Búsqueda y filtrado multi-dimensional
- Visualización de infracciones con detalles completos
- Filtros facetados con contadores
- Paginación de resultados
- Sincronización de estado con URL
- Modo mock para desarrollo sin backend

---

## Stack Tecnológico

- **Framework**: Next.js 15 (App Router)
- **Lenguaje**: TypeScript (modo estricto)
- **Estilos**: Tailwind CSS v4
- **Componentes UI**: shadcn/ui + Radix UI
- **Formularios**: React Hook Form + Zod
- **Data Fetching**: Native fetch API
- **Gestión de paquetes**: pnpm 10.18.3+
- **Node.js**: 18+
- **Fuentes**: Geist Sans + Geist Mono

---

## Arquitectura del Proyecto

### Estructura de Directorios

```
/
├── app/ # Next.js App Router
│ ├── page.tsx # Página de inicio (hero)
│ ├── layout.tsx # Layout raíz con metadata
│ ├── globals.css # Estilos globales con variables CSS
│ ├── loading.tsx # Loading state global
│ ├── api/
│ │ └── v1/
│ │ ├── offenses/route.ts # GET /api/v1/offenses
│ │ └── suggest/route.ts # GET /api/v1/suggest
│ └── offenses/
│ ├── page.tsx # Página de interfaz de búsqueda
│ └── loading.tsx # Loading state de búsqueda
├── components/
│ ├── search-interface.tsx # Componente principal de búsqueda (client)
│ ├── offense-card.tsx # Tarjeta individual de infracción
│ ├── facet-filter.tsx # Componente de filtro por dimensión
│ ├── filter-search.tsx # Búsqueda con sugerencias
│ ├── pagination.tsx # Controles de paginación
│ ├── active-filters.tsx # Visualización de filtros activos
│ ├── theme-provider.tsx # Contexto de tema
│ └── ui/ # Primitivos de shadcn/ui
├── lib/
│ ├── types.ts # Interfaces TypeScript principales
│ ├── display-config.ts # Configuración UI (iconos, labels)
│ ├── utils.ts # Funciones helper (cn())
│ └── client.ts # Funciones API del cliente
├── hooks/ # Custom React hooks
├── public/ # Assets estáticos
└── user_read_only_context/ # Documentación y ejemplos
```

### Flujo de Datos

1. **Usuario interactúa** → `SearchInterface` (componente cliente)
2. **SearchInterface** → Lee parámetros de URL
3. **SearchInterface** → Llama a `fetchOffenses()` de `lib/api/client.ts`
4. **client.ts** → Hace fetch a `/api/v1/offenses`
5. **route.ts** → Usa `getOffenses()` de `lib/repository.ts`
6. **Repository** → Consulta DuckDB y retorna datos
7. **SearchInterface** → Renderiza resultados y filtros

---

## Tipos y Modelos de Datos

### Enums

```typescript
enum Dimension {
Database = "database",
Year = "year",
Country = "country",
VehicleType = "vehicle_type",
Vehicle = "vehicle",
DocSource = "doc_source",
Location = "location",
Description = "description",
DocSrc = "doc_src",
}

enum SidebarMode {
Offenses = "offenses",
Documents = "documents",
}
```

### Interfaces Principales

```typescript
interface OffenseDocument {
db_id: number
doc_id: string
doc_date: string
doc_source: string
records: number // count(*)
ur: number // sum(ur)
errors: number // sum("error" IS NOT NULL)
}

interface Offense {
doc_src: string // URL al documento fuente
doc_id: string // ID de referencia del documento
doc_date: string // Fecha ISO string
country: string // "UY", "AR", etc.
adm_division: string // División administrativa
vehicle_type: string // "Auto", "Moto", "Camión"
mercosur_format: boolean // Formato de matrícula Mercosur
repo_id: number // ID de base de datos
record_id: number // Número de registro en repo
vehicle: string // Matrícula del vehículo
time: string // Datetime ISO
location: string // Ubicación de la infracción
id: string // ID del registro
description: string // Descripción de la violación
ur: number // Unidades de Ruina (unidades de penalización * 100)
error?: string // Problemas de calidad de datos
}

interface FacetValue {
value: string
count: number
selected: boolean
label?: string // Nombre para mostrar cuando value es ID
}

interface Facet {
dimension: Dimension
values: FacetValue[]
hasMore: boolean // Indica si hay más valores disponibles
}

interface InPredicate {
dimension: Dimension
values: string[]
}

interface OffensesParams {
predicates: InPredicate[]
page?: number
per_page?: number
}

interface OffensesResponse {
offenses: Offense[]
pagination: {
current_page: number
total_pages: number
}
repos: Record<string, Repo>
summary: {
avg_ur: number
facets: Facet[]
record_count: number
total_ur: number
}
active_filters: Record<string, string> // Labels para valores de filtros
}

interface Repo {
name: string
}
```

### Notas sobre Tipos

- **UR (Unidades de Ruina)**: Se almacenan multiplicadas por 100 (ej: 250 = 2.50 UR)
- **Fechas**: Siempre en formato ISO string
- **Dimensiones**: Usadas tanto para filtros como para facetas
- **hasMore**: Flag que indica si se debe mostrar búsqueda en el filtro

---

## API Routes

### 1. `/api/v1/offenses` (GET)

**Propósito**: Obtener infracciones con filtrado y paginación

**Parámetros** (query string):

- Dimensiones de filtro: `database`, `country`, `vehicle_type`, `vehicle`, `doc_src`, `location`, `description` (repetibles)
- Paginación: `page` (indexado desde 1)

**Ejemplo de URL**:
```
/api/v1/offenses?vehicle=AAO3197&country=UY&page=2
```

**Respuesta**: `OffensesResponse`

**Implementación**: `app/api/v1/offenses/route.ts`

- Usa `offensesParamsFromQueryParams()` para parsear parámetros
- Delega a `REPO.getOffenses()`
- Maneja errores con status 500

### 2. `/api/v1/suggest` (GET)

**Propósito**: Obtener sugerencias filtradas para una dimensión con query de búsqueda

**Parámetros** (query string):

- `dimension`: Dimensión a buscar (requerido)
- `q`: Query de búsqueda
- Dimensiones de filtro: `database`, `country`, etc. (repetibles)

**Ejemplo de URL**:
```
/api/v1/suggest?dimension=vehicle&q=AAO&country=UY
```

**Respuesta**: `Facet`

**Implementación**: `app/api/v1/suggest/route.ts`

- Valida que `dimension` esté presente
- Usa `REPO.getFacet()` con filtros actuales
- Retorna valores que coinciden con el query

---

## Capa de Datos

### Repository (`lib/repository.ts`)

El proyecto utiliza una implementación directa de acceso a datos usando DuckDB (WASM/Node) en lugar del patrón Repository con interfaz abstracta anterior.

**Funciones principales**:

```typescript
export async function getOffenses(params: OffensesParams, ...): Promise<any[]>
export async function getOffensesSummary(params: OffensesParams, ...): Promise<any[]>
export async function getDimensionResults(params: OffensesParams, ...): Promise<Facet[]>
export async function getDocuments(predicates: InPredicate[], ...): Promise<{ documents: OffenseDocument[]; total: number }>
export async function getDocumentFacets(predicates: InPredicate[], ...): Promise<Facet[]>
```

## La base de datos se inicializa automáticamente en `lib/duckdb.ts` y `lib/init-db.ts`

## Componentes Principales

### SearchInterface (`components/search-interface.tsx`)

**Tipo**: Client Component

**Responsabilidades**:

- Obtener infracciones basadas en parámetros de URL
- Gestionar estado de filtros y sincronización con URL
- Layout: Sidebar (filtros) + Main (resultados)
- Manejar clicks en filtros, paginación, limpiar todo

**Características clave**:

- Gestión de parámetros URL en tiempo real
- Auto-scroll al top en cambio de filtro
- Soporte para múltiples valores por dimensión
- Filtros activos con capacidad de remover individualmente

**Estado**:
```typescript
const [data, setData] = useState<OffensesResponse | null>(null)
const [loading, setLoading] = useState(false)
```

**Hooks importantes**:

- `useSearchParams()` - Lee parámetros de URL
- `useRouter()` - Navegación programática
- `usePathname()` - Path actual
- `useEffect()` - Carga datos cuando cambian parámetros

### OffenseCard (`components/offense-card.tsx`)

**Propósito**: Mostrar registro individual de infracción

**Props**:
```typescript
interface OffenseCardProps {
offense: Offense
repos: Record<string, Repo>
onFilterClick: (filterType: string, value: string) => void
}
```

**Características**:

- Campos filtrables (vehículo, ubicación, país, etc.)
- Muestra valor UR, fecha, link a documento externo
- Visualización de estado de error
- Indicación de tipo de vehículo con iconos
- Formato de fecha inteligente (solo fecha si es medianoche)

**Elementos filtrables**:

- Base de datos (badge)
- País (flag emoji)
- Vehículo (matrícula)
- Tipo de vehículo (icono)
- Ubicación (con icono MapPin)
- Ubicación (con icono MapPin)
- Fuente del documento

### DocumentCard (`components/document-card.tsx`)

**Propósito**: Mostrar detalles de un documento (resolución) agrupado.

**Características**:

- Muestra ID del documento, fecha y estadísticas (total registros, total UR, errores).
- Link "Ver Infracciones" para filtrar `/offenses` por ese documento.
- Link a fuente original.
- Reuse de `FilterableItem` para filtros rápidos.

### DocumentsFeed (`app/documents/documents-feed.tsx`)

**Propósito**: Lista con scroll infinito para documentos.

- Usa **Server Actions** (`loadMoreDocuments`) para paginación.

### FacetFilter (`components/facet-filter.tsx`)

**Propósito**: Sección de filtro colapsable por dimensión

**Props**:
```typescript
interface FacetFilterProps {
title?: string
dimension: string
facets: FacetValue[]
selectedValues: string[]
onSelect: (value: string | null) => void
currentFilters: URLSearchParams
defaultOpen?: boolean
hasMore?: boolean
}
```

**Características**:

- Muestra contadores de valores
- Búsqueda disponible cuando `hasMore` es true
- Muestra labels para bases de datos y países
- Soporte para valores vacíos con estilo especial
- Icono de check para valores seleccionados

### FilterSearch (`components/filter-search.tsx`)

**Propósito**: Búsqueda en tiempo real con sugerencias

**Características**:

- Llamadas API con debounce (300ms)
- Muestra valores coincidentes con contadores
- Loading state durante búsqueda
- Mensaje cuando no hay resultados

**Flujo**:

1. Usuario escribe en input
2. Debounce de 300ms
3. Fetch a `/api/v1/suggest?dimension={dim}&q={query}&{filters}`
4. Muestra sugerencias con contadores
5. Click en sugerencia → agrega filtro

### Pagination (`components/pagination.tsx`)

**Características**:

- Muestra máximo 5 páginas visibles
- Elipsis para conteos grandes de páginas
- Botones prev/next deshabilitados en límites
- Siempre muestra primera y última página

### ActiveFilters (`components/active-filters.tsx`)

**Propósito**: Muestra todos los filtros activos agrupados por dimensión

**Características**:

- Botones individuales de remover
- Acción de limpiar todo
- Muestra iconos de dimensión
- Estilo especial para valores vacíos (muted + italic)
- Truncado de texto largo con tooltip

---

## Configuración de Visualización

### Configuración de Dimensiones (`lib/display-config.ts`)

```typescript
interface DimensionConfig {
icon: LucideIcon
label: string
empty: string // Texto para valores vacíos
}
```

**Dimensiones configuradas**:

- `Database`: Database icon, "Base de datos", "Sin base de datos"
- `Country`: Globe icon, "País de la matrícula", "Sin información"
- `Vehicle`: Car icon, "Matrícula", "Sin matrícula"
- `VehicleType`: Truck icon, "Tipo de vehículo", "Sin información"
- `Location`: MapPin icon, "Ubicación", "Sin ubicación"
- `Description`: AlignLeft icon, "Descripción", "Sin descripción"
- `DocSrc`: File icon, "Fuente", "Sin fuente"

**Función helper**:

- `getDimensionConfig(dimension: string): DimensionConfig`

**Nota**: Los labels para valores de filtros (como nombres de países y bases de datos) ahora son provistos por el backend en el campo `active_filters` de la respuesta de la API.

### Configuración de Países

```typescript
interface CountryConfig {
name: string
flag: string // Emoji de bandera
}
```

**Países soportados**:

- `AR`: Argentina 🇦🇷
- `UY`: Uruguay 🇺🇾
- `BR`: Brasil 🇧🇷
- `PY`: Paraguay 🇵🇾
- `CL`: Chile 🇨🇱

**Funciones helper**:

- `getCountryName(countryCode: string): string`
- `getCountryFlag(countryCode: string): string`
- `getCountryConfig(countryCode: string): CountryConfig | undefined`

---

## Variables de Entorno

## Actualmente el proyecto no requiere variables de entorno específicas para la conexión a datos, ya que utiliza DuckDB embebido.

## Datos Mock

### Características de Datos Mock

**Infracciones de ejemplo**:

- Fechas variadas (2023-2024)
- Países: UY, AR
- Tipos de vehículo: Auto, Moto, Camión
- Matrículas realistas (formato Mercosur y tradicional)
- Violaciones: Velocidad, estacionamiento, casco, deuda de patente
- Algunos registros con campo `error` para simular problemas de calidad

**Ejemplo de registro mock**:
```typescript
{
doc_src: "https://example.com/doc/123",
doc_id: "DOC-123",
doc_date: "2024-01-15",
country: "UY",
adm_division: "Montevideo",
vehicle_type: "Auto",
mercosur_format: true,
repo_id: 6,
record_id: 1,
vehicle: "AAO3197",
time: "2024-01-15T14:30:00Z",
location: "Av. 18 de Julio",
id: "6-1",
description: "Exceso de velocidad",
ur: 250, // 2.50 UR
}
```

---

## Estilos y Tema

### Sistema de Diseño

- **Tema**: Dark minimalista (inspirado en grep.app y Vercel)
- **Colores**: Basados en oklch() (variables CSS nativas)
- **Componentes**: Suite completa de shadcn/ui disponible
- **Tailwind Config**: CSS utility-first con colores base personalizados

### Variables CSS (`app/globals.css`)

**Colores principales**:

- `--background`, `--foreground`
- `--primary`, `--primary-foreground`
- `--accent`, `--accent-foreground`
- `--muted`, `--muted-foreground`
- `--border`, `--input`, `--ring`
- `--card`, `--card-foreground`

**Colores de gráficos**:

- `--chart-1` a `--chart-5`

**Sidebar**:

- Variables separadas `--sidebar-*` para componentes de sidebar

### Configuración de Fuentes

```typescript
// app/layout.tsx
import { GeistSans } from "geist/font/sans"
import { GeistMono } from "geist/font/mono"

// Aplicadas vía className en body
className={`font-sans ${GeistSans.variable} ${GeistMono.variable}`}
```

```css
/_ app/globals.css _/
@theme inline {
--font-sans: 'Geist', 'Geist Fallback';
--font-mono: 'Geist Mono', 'Geist Mono Fallback';
}
```

---

## Estándares de Código

### Convenciones de Nombres

**Archivos**:

- Componentes: `kebab-case.tsx` (ej: `offense-card.tsx`)
- Utilidades: `kebab-case.ts` (ej: `format-date.ts`)
- Tipos: `kebab-case.ts` (ej: `offense-types.ts`)

**Componentes**:
```typescript
// ✅ Correcto
export function OffenseCard() { }
export default function OffensesPage() { }

// ❌ Incorrecto
export function offenseCard() { }
export default function offenses_page() { }
```

### Orden de Imports

```typescript
// 1. React/Next
import { useState } from 'react'
import Link from 'next/link'

// 2. Librerías externas
import { format } from 'date-fns'
import { z } from 'zod'

// 3. Componentes
import { Button } from '@/components/ui/button'
import { OffenseCard } from '@/components/offense-card'

// 4. Utilidades y tipos
import { cn } from '@/lib/utils'
import type { Offense } from '@/types/offense'
```

### Estructura de Componentes

```typescript
interface ComponentProps {
// Props con tipos explícitos
}

export function Component({ prop1, prop2 }: ComponentProps) {
// 1. Hooks primero
const [state, setState] = useState()

// 2. Funciones auxiliares
const handleClick = () => { }

// 3. Render
return <div>...</div>
}
```

### TypeScript

- **Modo estricto** habilitado
- Tipos explícitos en parámetros requeridos
- Props basadas en interfaces para componentes
- No usar `any` (enforced)
- Usar `type` para unions, `interface` para objetos

### Formateo

**Prettier** (`.prettierrc`):

- Sin punto y coma al final
- Comillas simples
- Indentación de 2 espacios
- Ancho máximo de línea: 80 caracteres
- Ordenamiento automático de clases de Tailwind

**ESLint** (`.eslintrc.json`):

- Reglas de Next.js
- Checks estrictos de TypeScript
- Integración con Prettier

**EditorConfig** (`.editorconfig`):

- Indentación de 2 espacios
- UTF-8
- LF line endings
- Trim trailing whitespace

- Trim trailing whitespace

---

## Testing

El proyecto utiliza **Vitest** como framework de pruebas.

### Estructura de Tests

- **Tests Unitarios**: Ubicados junto al código fuente o en archivos `*.test.ts`.
  - `lib/types.test.ts`: Tests de modelos y enums.
- **Tests de Integración**:
  - `lib/repository.test.ts`: Tests del repositorio contra una base de datos DuckDB en memoria.
  - `app/api/v1/offenses/route.test.ts`: Tests de endpoints de API mockeando el repositorio.

### Ejecución de Tests

```bash

# Ejecutar todos los tests

pnpm test

# Ejecutar tests en modo watch

pnpm test --watch

# Ejecutar un archivo específico

npx vitest run lib/repository.test.ts
```

---

## Comandos de Desarrollo

```bash

# Desarrollo

pnpm dev # Servidor de desarrollo (puerto 3000)

# Build

pnpm build # Build de producción
pnpm start # Ejecutar build de producción

# Calidad de código

pnpm lint # Ejecutar ESLint
pnpm lint:fix # Auto-corregir problemas de ESLint
pnpm format # Formatear con Prettier
pnpm format:check # Verificar formato sin modificar
pnpm typecheck # Verificación de tipos TypeScript
pnpm test # Ejecutar tests con Vitest
```

### Workflow recomendado antes de commit

```bash
pnpm format && pnpm lint:fix && pnpm typecheck && pnpm test
```

---

## Rutas de la Aplicación

### 1. `/` - Página de Inicio

**Archivo**: `app/page.tsx`

**Contenido**:

- Hero section con título y descripción
- 3 tarjetas de características (Búsqueda Avanzada, Datos Oficiales, Detalles Completos)
- CTA button que lleva a `/offenses`
- Header y footer simples

**Tipo**: Server Component

### 2. `/offenses` - Interfaz de Búsqueda

**Archivo**: `app/offenses/page.tsx`

**Contenido**:

- Renderiza `<SearchInterface />` dentro de `<Suspense>`
- Toda la lógica de búsqueda y filtrado

**Tipo**: Server Component que renderiza Client Component

### 3. `/documents` - Navegador de Documentos

**Archivo**: `app/documents/page.tsx`

**Contenido**:

- Lista de documentos agrupados por ID y base de datos.
- Sidebar simplificado (solo Database y Year) usando `SidebarMode.Documents`.
- Scroll infinito.

**Tipo**: Server Component con `use server` actions para paginación.

---

## Puntos de Integración Clave

### Gestión de Parámetros URL

**Patrón usado**:

- Todos los filtros se reflejan en parámetros de query URL
- Múltiples valores por dimensión: `?vehicle=AAO3197&vehicle=BBX4521`
- Parámetro de página: `?page=2`
- Sincronización URL previene pérdida de estado en refresh

**Implementación**:
```typescript
const updateURL = useCallback((params: Record<string, string | string[] | null>) => {
const current = new URLSearchParams(window.location.search)
// Lógica de actualización...
router.push(`${pathname}${query}`, { scroll: false })
}, [pathname, router])
```

### Manejo de Errores

- Registros mock con campo `error` mostrados con icono de warning
- Errores de API logueados a consola con fallback graceful
- Loading states durante llamadas API
- Mensajes de "No se encontraron infracciones" cuando no hay resultados

### Performance

- Debounce de 300ms en sugerencias de búsqueda
- Delay simulado de 200ms en respuestas mock
- Paginación (20 registros por página)
- Scroll-to-top en cambios de filtro
- `scroll: false` en navegación para prevenir scroll no deseado

---

## Automación y CI/CD

El proyecto cuenta con un pipeline de CI/CD automatizado y serverless para la actualización de datos y despliegue.

- **Documentación detallada**: [web/docs/automacion.md](web/docs/automacion.md)
- **Pipeline Principal**: `.dagger/automation.go` (`DataRefresh`, `Deploy`)
- **Configuración Serverless**: `cloudbuild-daily.yaml`
- **Ejecución**: Cloud Build (Manual o Trigger) + Cloud Scheduler (Planificado)

