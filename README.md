# Introducción

ChapaUY permite acceder a un subconjunto de infracciones de tránsito ocurridas
en Uruguay notificadas por 10 departamentos y dos ministerios
en el [Diario Oficial](https://www.impo.com.uy/directorio-bases-institucionales/#seccion-bases-multas]).

Nótese que quien publica las bases de datos también provee un acceso puntual por
mátricula en https://www.impo.com.uy/consultamultas/, y que también se pueden
recibir notificaciones de eventos cuando se publica una infracción con https://www.sucive.gub.uy/suscripcion_notificaciones?1. Es decir, no busca
reemplazar esos usos, sino acceder a los datos crudos.

El programa permite descubrir nuevas resoluciones, descargar su versión HTML,
extraer la información contenida en ellas y exportarla. Está diseñado para ser
amigable con quien publica la información - se minimizan las búsquedas, se
descargan únicamente los datos nuevos, se realiza una descarga por vez
comprimiendo las representaciones.

El desarrollo del programa tiene fines educativos. Por un lado probar golang,
14 años tarde después de que vimos la charla de Rob Pike en OSCON2011, pero un
viejo perro siempre puede aprender nuevos trucos, pero tambien es un lindo
proyecto de referencia, para estudiantes de informática, y de ciencias de datos.

El archivo [dbrefs.go](blob/main/impo/dbrefs.go) describe las diferentes bases
de datos disponibles, seguimos por lo menos para la adquicisión, los mismos
conceptos que el publisher.

# Fases

La obtención de nueva información se realiza utilizando el comando
`./chapa impo update`. Por defecto actualiza todas las bases de datos, pero se puede
trabajar puntualmente sobre una única base `./chapa impo update maldonado`.
Puede usarse parte del nombre o el id de la base. Se pueden listar las bases
disponibles con…
```
$ ./build/chapa impo list
chapa impo list
Base de datos disponibles
╭────┬────────────────┬──────────────────────╮
│ Id │ Nombre         │ Ubicación            │
├────┼────────────────┼──────────────────────┤
│ 65 │ Caminera       │ …/multascaminera     │
│ 40 │ Canelones      │ …/multascanelones    │
│ 48 │ Colonia        │ …/multascolonia      │
│ 26 │ Lavalleja      │ …/multaslavalleja    │
│ 45 │ Maldonado      │ …/multasmaldonado    │
│  6 │ Montevideo     │ …/cgm                │
│ 43 │ Paysandu       │ …/multaspaysandu     │
│ 55 │ Rio Negro      │ …/multasrionegro     │
│ 49 │ Soriano        │ …/multassoriano      │
│ 56 │ Tacuarembó     │ …/multastacuarembo   │
│ 52 │ Treinta y Tres │ …/multastreintaytres │
│ 68 │ Vialidad       │ …/multasmtop         │
╰────┴────────────────┴──────────────────────╯
```

La actualización está compuestas por varias fases. Se puede evitar la
ejecución de fases específicas con la familia de argumentos `--skip-…`.

La opción `--dry-run` evita persistir cualquier cambio permanente local.

## Búsqueda

El proceso de descubrimiento de nuevos documentos (resoluciones y
notificaciones) comienza en la raíz de la base de datos. Por ejemplo para el
[departamento de Maldonado](https://impo.com.uy/base-institucional/multasmaldonado).
Esta fase requiere de mantener una serie de HTTP Cookies. La búsquedas parece un
proceso costoso (a veces se tarda segundos en retornar una página), pero los
documentos nuevos

Mientras existan nuevos resultados (documentos que no se conocían previamente)
se van iterando las diferentes páginas. Esto permite que mantener actualizada
la base sea un proceso corto (una o dos páginas).

Hay una cantidad máximas de páginas que se iteran (algo así como un
_Dead man's switch_), pero se puede modificar  utilizando
`--search-max-depth 25`. También es posible forzar recorrer todas las páginas
en vez de quedarse hasta donde se vean cosas nuevas, con `--search-full`.

La fase puede ser salteada con `--skip-search`.  Esta fase se encuentra implementada en el archivo [search.go](blob/main/impo/search.go).

Los documentos que se encuentra son persistidos en un archivo.
Se identifican con su URL, y la  ubicación de la base de datos puede
modificarse con el argumento `./chapa impo --db-path /path/to/db …`.

```
$ head db/45/documents.json
{
  "https://www.impo.com.uy/bases/notificaciones-transito-maldonado/1-2023": {
    "title": "Notificación Dirección General de Tránsito y Transporte Intendencia de Maldonado 1/023",
    "href": "https://www.impo.com.uy/bases/notificaciones-transito-maldonado/1-2023",
    "subtitle": "NOTIFICACION POR CONTRAVENCION A NORMAS DE TRANSITO"
  },
  "https://www.impo.com.uy/bases/notificaciones-transito-maldonado/1-2024": {
    "title": "Notificación Dirección General de Tránsito y Transporte Intendencia de Maldonado 1/024",
    "href": "https://www.impo.com.uy/bases/notificaciones-transito-maldonado/1-2024",
    "subtitle": "NOTIFICACION POR CONTRAVENCION A NORMAS DE TRANSITO"
    …
```

## Descarga de documentos

Esta fase se encarga únicamente de descargar los documentos que son referidos
en `documents.json` que no tienen una copia local. Las descargas se hacen
negociando compresión.

Cada base de datos tiene una función que mapea el id/URL a un path en el
filesystem. Los archivos quedan comprimidos en disco.
```
db/45
├── documents.json
├── notificaciones
│   └── 2023
│       ├── 100.html.gz
│       ├── …
│       └── 9.html.gz
…
```
La fase puede ser salteada con `--skip-download`.  Esta fase se encuentra implementada en el archivo [extract.go](blob/main/impo/extract.go).

## Extracción

Esta estructura la información contenida en los documentos HTML.
Almacena los resultados junto a los documentos HTML.

```
db/45
├── documents.json
├── notificaciones
│   └── 2023
│       ├── 100.html.gz
│       ├── 100.json.gz
│       ├── …
│       ├── 9.html.gz
│       └── 9.json.gz
…
```

Se busca la tabla con infracciones, y hay diferentes estrategias para
reconocer las columnas (que varían según los documentos) y manejar errores.
La opción `--skip-extract-errors` permite no almacenar documentos donde se
detectaron al menos un error. Las UR se almacenan sin decimales
(escalado por cien).

```
{
  "offenses": [
    {
      "vehicle": "██████",
      "time": "2025-04-18T17:04:00-03:00",
      "location": "Ruta 10 y Calle 19 (El Apero)",
      "id": "IDM 0001515196",
      "description": "Exceso de velocidad de entre 31km/h y el doble - 1",
      "ur": 1200
    },
    …
  ]
  "date": "2025-04-30T00:00:00-03:00",
  "id": "327/025",
  "errors": []
}
```

La fase puede ser salteada con `--skip-extract`.  Esta fase se encuentra implementada en el archivo [extract.go](blob/main/impo/extract.go).

Es posible verificar la extracción con documentos individuales de la siguiente forma:
```
$ curl https://www.impo.com.uy/bases/notificaciones-transito-lavalleja/11-2025 | 
iconv --from-code=iso-8859-1 --to-code=utf8 | 
go run main.go  debug  document
```

# Persistencia y Dashboard Web

Además de los archivos JSON/HTML, el comando `impo update` exporta toda la información procesada a una base de datos DuckDB.

- **Archivo generado**: `db/chapauy.duckdb` (o el path especificado con `--db-path`)
- **Contenido**: Tablas estructuradas con infracciones, documentos, y metadatos geoespaciales.

Esta base de datos es utilizada por la aplicación web para visualización y consultas.

## Ejecutar el Dashboard Web

La aplicación web (en el directorio `web/`) requiere acceso al archivo de base de datos generado.

1. **Generar la base de datos**:
   Asegúrate de haber corrido `impo update` al menos una vez.
   ```bash
   ./chapa impo update
   ```

2. **Copiar/Enlazar la base de datos**:
   La aplicación web busca el archivo `chapauy.duckdb` en su directorio raíz o en el directorio desde donde se ejecuta.
   ```bash
   # Opción A: Copiar
   cp db/chapauy.duckdb web/

   # Opción B: Ejecutar desde la raíz (recomendado para desarrollo)
   # (La aplicación web intentará buscar ./chapauy.duckdb)
   ln -sf db/chapauy.duckdb web/chapauy.duckdb
   ```

3. **Iniciar la Web App**:
   ```bash
   cd web
   pnpm install
   pnpm dev
   ```
   El dashboard estará disponible en [http://localhost:3000](http://localhost:3000).

   > **Nota**: Si la aplicación web no encuentra el archivo `.duckdb`, iniciará en modo "demo" utilizando una base de datos en memoria con datos de prueba.

# Geocodificación Interactiva

ChapaUY incluye una interfaz web para geocodificar las ubicaciones de las infracciones. Esto permite mapear las infracciones y realizar análisis espaciales.

## Inicio Rápido

```bash
# Iniciar el servidor de geocodificación
./build/chapa curation

# Abrir en el navegador
open http://localhost:8081
```

## Proveedores de Geocodificación

El sistema utiliza una estrategia de múltiples proveedores con fallback automático:

1. **Base de Datos de Radares RUTA** (instantáneo, local)
   - Reconoce patrones como "RUTA 005 Y 038K131_D"
   - Usa datos GIS pre-cargados de radares de rutas
   - Alta confianza para ubicaciones de vigilancia electrónica

2. **Google Maps Geocoding API** (200ms, de pago, alta calidad)
   - Maneja intersecciones nativamente en una sola llamada
   - Mejor precisión para direcciones uruguayas
   - Excelente para formato "AV 8 DE OCTUBRE Y AV CENTENARIO"
   - Requiere `GOOGLE_MAPS_API_KEY` (ver más abajo)

## Uso con Google Maps (Recomendado)

Para geocodificación 10× más rápida (especialmente para intersecciones):

```bash
# Configurar clave de API
export GOOGLE_MAPS_API_KEY="tu-clave-aqui"

# Ejecutar servidor
./build/chapa curation
```

Deberías ver:
```
📍 Geocoding: Google Maps (primary)
```

### Obtener Clave de Google Maps

Ver [GOOGLE_MAPS_SETUP.md](GOOGLE_MAPS_SETUP.md) para instrucciones detalladas.

**Costo aproximado:**
- Geocodificación inicial (~16,000 ubicaciones): $25-80 una vez
- Mantenimiento (~100 ubicaciones/mes): <$1/mes
- Primeros $200/mes son gratis con Google Cloud



## Características

- **Priorización por Impacto**: Las ubicaciones se presentan ordenadas por cantidad de infracciones
- **🆕 Filtrado por Base de Datos**: Selector para trabajar en departamentos específicos (ej. Maldonado)
- **Geocodificación Automática**: Sugerencias de coordenadas basadas en múltiples estrategias
- **Verificación Humana**: Mapa interactivo para validar/corregir sugerencias
- **Validación de Datos**: Verifica coordenadas dentro de límites de Uruguay
- **Seguimiento de Progreso**: Visualiza cobertura de geocodificación en tiempo real
- **Exportación**: Descarga juicios en formato JSON/CSV
- **Persistencia**: Las decisiones se almacenan en la base de datos con respaldo JSON

### Filtrado por Base de Datos

La interfaz incluye un selector de base de datos que permite a los usuarios enfocarse en ubicaciones de un departamento específico:

```
Todas las Bases: 16,314 ubicaciones
Solo Maldonado: 2,164 ubicaciones
```

**Beneficios del Filtrado:**
- **Conocimiento Local**: Los usuarios pueden trabajar en áreas que conocen bien
- **Mayor Precisión**: Familiaridad con calles y puntos de referencia locales
- **Progreso Enfocado**: Estadísticas específicas por departamento
- **URL Persistente**: `?db_id=45` para marcar/compartir vistas específicas

**Uso:**
1. Abrir http://localhost:8081
2. Seleccionar "Maldonado" del dropdown "Database"
3. La cola de ubicaciones se filtra automáticamente
4. El progreso muestra estadísticas específicas de Maldonado

## Guía Completa

Para instrucciones detalladas, solución de problemas y referencia de API, ver [GEOCODING.md](GEOCODING.md).

Para más detalles técnicos, ver [prds/LOCAL-001-interactive-location-geocoding.md](prds/LOCAL-001-interactive-location-geocoding.md).
