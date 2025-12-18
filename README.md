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

El desarrollo del programa tiene fines educativos. Por un lado, probar Golang —14 años después de la charla de Rob Pike en OSCON 2011—, pero también como un proyecto de referencia para estudiantes de informática y ciencia de datos.

Para una documentación detallada sobre el proyecto, consulte:
*   [**Arquitectura**](web/docs/000-arquitectura.md): Visión general del sistema y base de datos DuckDB.
*   [**Adquisición de datos**](web/docs/010-acquire.md): Detalles sobre las fases de búsqueda, descarga y extracción.
*   [**Enriquecimiento / Curación**](web/docs/020-curate.md): Estrategias de geolocalización, normalización y análisis de matrículas.

La obtención de nueva información se realiza utilizando el comando `./chapa impo update`. Este proceso descubre nuevos documentos, los descarga localmente y extrae la información estructurada.

Por defecto actualiza todas las bases de datos, pero se puede
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

Para más detalles sobre parámetros y funcionamiento interno, consulte la documentación de [Adquisición de datos](web/docs/010-acquire.md).

Es posible verificar la extracción con documentos individuales de la siguiente forma:
```
$ curl https://www.impo.com.uy/bases/notificaciones-transito-lavalleja/11-2025 |
iconv --from-code=iso-8859-1 --to-code=utf8 |
go run main.go  debug  document
```

# Persistencia y Dashboard Web

Además de los archivos JSON/HTML, el sistema exporta toda la información procesada a una base de datos [DuckDB](web/docs/000-arquitectura.md#base-de-datos-sql).

Para ejecutar el dashboard:
1. Asegúrese de tener el archivo `db/chapauy.duckdb` (generado por `impo update`).
2. Vincule la base a la web: `ln -sf db/chapauy.duckdb web/chapauy.duckdb`.
3. Inicie el entorno: `cd web && pnpm install && pnpm dev`.

# Curación de datos

ChapaUY incluye interfaces para la geocodificación interactiva y la clasificación de infracciones.

```bash
# Iniciar el servidor de curación
./chapa curation serve
```

Consulte el manual de [Enriquecimiento](web/docs/020-curate.md) para más detalles sobre la estrategia de bloques de matrículas y flujos de trabajo de curación.
