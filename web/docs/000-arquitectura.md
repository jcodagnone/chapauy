---
title: Arquitectura
---

El sistema se compone de tres aplicaciones principales:

*   `./chapa`: una interfaz de línea de comandos (CLI) encargada del descubrimiento, adquisición, curación, extracción y almacenamiento de documentos. Persiste los datos en una base de datos DuckDB.
*   **Aplicación web**: expone la información a los usuarios.
*   `./infra`: una CLI dedicada a la provisión de recursos mediante *infrastructure as code*.

## ./chapa CLI

Desarrollada en Go, su punto de entrada es el archivo [`main.go`](https://github.com/jcodagnone/chapauy/blob/master/main.go). Ofrece tres subcomandos:

*   `impo`: gestiona el *pipeline* completo de descubrimiento, adquisición, extracción y almacenamiento (ver [Adquisición](/docs/010-acquire)).
*   `curation`: permite la curación de ubicaciones y descripciones (webapp) y el almacenamiento duradero de esta información (ver [Enriquecimiento](/docs/020-curate)). Con esta información se enriquecen las infracciones.
*   `debug`: provee herramientas para *troubleshooting* y pruebas unitarias de componentes.

Se puede compilar directamente con `go run main.go`, mediante `Makefile`, o con `call build-cli-base`.

La aplicación, mediante su subcomando `impo`, realiza conexiones salientes únicamente a `https://impo.com.uy/` y `https://www.impo.com.uy`, leyendo y escribiendo archivos en el directorio `db/`.

```text
$ tree db/
db
├── 45
│   ├── documents.json
│   ├── notificaciones
│   │   └── 2025
│   │       ├── 488.html.gz
│   …       …
└── chapauy.duckdb

108 directories, 9205 files
```

El proceso inicia buscando documentos en la [base institucional](https://impo.com.uy/base-institucional/multasmaldonado). Nótese que Maldonado corresponde a la base 45. Si `db/:id/documents.json` no existe, se crea. Mientras existan páginas con documentos no vistos, se itera - hay una cantidad máxima de seguridad - y almacenan:

```json
$ grep -A 3 -B 1  488/025  db/45/documents.json
  "https://www.impo.com.uy/bases/notificaciones-transito-movilidad-maldonado/488-2025": {
    "title": "Notificación Departamento de Movilidad Intendencia de Maldonado 488/025",
    "href": "https://www.impo.com.uy/bases/notificaciones-transito-movilidad-maldonado/488-2025",
    "subtitle": ""
  },
```

Posteriormente, comienza la etapa de descarga: se verifica en el sistema de archivos si existe una copia local del documento. De no ser así, se descarga, transformando la URL en una ruta del sistema de archivos y almacenándola comprimida. Este diseño permite trabajar escalonadamente y recrear la base de datos de manera económica ante cambios en los algoritmos. Finalmente, se procesan aquellos documentos que aún no se encuentran en la base de datos `chapauy.duckdb`. El proceso de extracción de datos se puede visualizar así:

```json
$ zcat db/45/notificaciones/2025/488.html.gz | ./build/chapa debug document
[
  {
    "doc_id": "488/025",
    "doc_date": "2025-12-16T00:00:00-03:00",
    "repo_id": 45,
    "record_id": 1,
    "vehicle": "241AFF",
    "time": "2025-12-10T17:24:00-03:00",
    "location": "Puente Leonel Viera y Aparicio Saravia",
    "id": "IDM 0001882047",
    "description": "Exceso de velocidad hasta 20 km/h",
    "ur": 500,
    …
```

Estos registros se persisten en la base de datos SQL.

## Base de datos (SQL)

DuckDB opera como una base de datos *serverless*, ya que no requiere un proceso independiente. Las infracciones de tránsito se almacenan en la tabla `offenses`. Esta tabla se encuentra totalmente desnormalizada, optimizada para flujos de trabajo analíticos.

Para analizar sus columnas, tomemos como referencia la [Notificación del Departamento de Movilidad de la Intendencia de Maldonado N° 488/025](https://www.impo.com.uy/bases/notificaciones-transito-movilidad-maldonado/488-2025).

```sql
D select * from offenses where db_id=45 order by time DESC LIMIT 1;

           db_id = 45
          doc_id = 488/025
        doc_date = 2025-12-16
      doc_source = https://www.impo.com.uy/bases/notificaciones-transito-movilidad-maldonado/488-2025
```

Las primeras columnas contienen información del documento. Esta es la sección desnormalizada, dado que un documento suele contener múltiples registros. `db_id` mantiene su correspondencia en la [base institucional](https://impo.com.uy/base-institucional/multasmaldonado).

A continuación, figuran las columnas extraídas del documento. No todos los documentos presentan la totalidad de estos datos.

```text
       record_id = 343
      offense_id = IDM 0001883465
         vehicle = BED6747
 vehicle_country = UY
    vehicle_type = NULL
            time = 2025-12-11 09:54:00-03
       time_year = 2025
        location = Ruta Interbalnearia y Milton Lussich
display_location = NULL
     description = Exceso de velocidad hasta 20 km/h
              ur = 500
           error = NULL
```

Aquí, `record_id` es el número de registro en la tabla (otorgando direccionabilidad), y `display_location` se vincula al proceso de unificación de nomenclatura de ubicaciones (ver [Normalización de Ubicaciones](/docs/020-curate#geocoding)): si determinamos que para agregaciones el nombre canónico es otro, preservamos el nombre original (`location`) para la visualización del registro individual.

Posteriormente, encontramos la información enriquecida. Las coordenadas `point` surgen de un proceso de geolocalización (ver [Geocoding](/docs/020-curate#geocoding)). A partir de ellas, se sintetizan diferentes resoluciones de [índices H3](https://h3geo.org/). Estos índices permiten resolver consultas espaciales para el mapa sin necesidad de operadores GIS especializados.

```text
           point = {'x': -55.044561624526985, 'y': -34.8831234433184}
         h3_res1 = 584394828209455103
         h3_res2 = 588897878081011711
         h3_res3 = 593401202830475263
         h3_res4 = 597904793867911167
         h3_res5 = 602408390274056191
         h3_res6 = 606911989230338047
         h3_res7 = 611415588790599679
         h3_res8 = 615919188407484415
     article_ids = [13.3.A]
   article_codes = [13]
```

Por último, `article_ids` representa la codificación del articulado de la descripción. En este ejemplo, la descripción posee un único código (exceso de velocidad), pero existen casos con múltiples códigos. Esto depende de cada base de datos y, fundamentalmente, de si la infracción fue labrada manualmente. Por ejemplo, para el texto *ESTACIONAR A MAYOR DISTANCIA DEL CORDON QUE LA PERMITIDA, NO POSEER LICENCIA DE CONDUCIR, NO PORTAR DOCUMENTACION DEL VEHICULO*, correspondería:
*   `article_ids = [18.1.2, 3.1.1, 4.1.2]`
*   `article_codes = [18, 3, 4]`

Ver más detalles en [Descripciones](/docs/020-curate#descripciones).

La tabla no cuenta con un ID único global, ya que si se remueve un documento, se eliminan todos sus registros asociados (por ejemplo, en un reprocesamiento).

Por otro lado, existe una serie de tablas satélites que soportan el proceso de curación (geolocalización, extracción de artículos) e impactan al momento de almacenar la información curada.

La primera describe los artículos normativos (utilizada en `offenses#article_ids`):

```sql
D SELECT * FROM articles WHERE id='13.3.A';
   id = 13.3.A
 text = Superar las velocidades máximas permitidas: hasta 20 km
 code = 13
title = De las velocidades
```

Sus valores provienen básicamente del PDF adjunto en [Se aprueban "Valores Unificados de Multas de Tránsito"](https://www.gub.uy/congreso-intendentes/comunicacion/noticias/se-aprueban-valores-unificados-multas-transito-acuerdo-propuesta-realizada), con la adición de casos del reglamento SUCIVE.

```sql
D SELECT * FROM descriptions WHERE article_ids = ['13.3.A'];
           id = 4355
  description = 5 ART. 13.3.A: Superar las velocidades máximas permitidas (hasta 20 km)
  article_ids = [13.3.A]
article_codes = [13]
```

Luego siguen las anotaciones de descripciones vistas en `offenses` y su mapeo a artículos y códigos:

```text
           id = 4542
  description = AR T. 13.3.A: Exceso de velocidad hasta 20 km/h
  article_ids = [13.3.A]
article_codes = [13]

           id = 4547
  description = ART . 13.3.A: Superar las velocidades máximas permitidas (hasta 20 km)
  article_ids = [13.3.A]
article_codes = [13]

           id = 4558
  description = ART 13: Exceso de velocidad hasta 20 km/h
  article_ids = [13.3.A]
article_codes = [13]
```

Para finalizar, almacenamos también la curación de la geolocalización:

```sql
D SELECT * FROM locations
                id = 20228
             db_id = 45
          location = Ruta Interbalnearia y Milton Lussich
canonical_location = NULL
             point = {'x': -55.044561624526985, 'y': -34.8831234433184}
     is_electronic = false
  geocoding_method = manual_click
        confidence = medium
             notes = Milton Lussich, 20003 Punta Ballena, Departamento de Maldonado, Uruguay
           h3_res1 = 584394828209455103
           h3_res2 = 588897878081011711
       h3_res3 = 593401202830475263
           h3_res4 = 597904793867911167
           h3_res5 = 602408390274056191
           h3_res6 = 606911989230338047
           h3_res7 = 611415588790599679
           h3_res8 = 615919188407484415
```
## Aplicación web

La aplicación web es la cara visible del proyecto, diseñada para explorar los datos. Si bien en un principio la idea era no requerir JavaScript en el navegador, incluso antes del comentario de [Pablo Sabattela](https://x.com/PabloSabbatella/status/1997413381901267233)

> Most of the websites hacked these days because of the React exploit (react2shell) had no reason to be running code on the server, when they could just have been HTML

la funcionalidad de mapas motivó un cambio de dirección total, resultando en la utilización de [Next.js 16](https://nextjs.org/) con React 19, aprovechando las últimas características como *Server Components* y *App Router* sobre Node 24. Esto también dio pie a probar [v0](https://v0.app/) (consume más nafta que una Ford F150) y otros agentes. La incorporación de DuckDB limitó la utilización de v0 y el despliegue en Vercel, dado que DuckDB utiliza *bindings* nativos cuya compilación resulta dificultosa en esos entornos.

La aplicación en producción corre en un contenedor minimalista *distroless*; salvo por el directorio de caché interno, el resto del sistema de archivos es de solo lectura (*read-only*). Node corre con un set de [permisos reducidos](https://nodejs.org/api/permissions.html), aunque queda pendiente aplicar políticas más granulares, como bloquear *system calls* innecesarias. Por ejemplo, la aplicación no realiza conexiones TCP/UDP salientes. La base de datos se embebe en el contenedor y se abre también en modo solo lectura. Durante el ciclo de vida de la aplicación ningún dato cambiará; al día siguiente, se generará un nuevo contenedor con la imagen web y los últimos datos procesados. Por ello, se implementa un *caching* agresivo, tanto en el renderizado interno como en las directivas de caché externas.

## ./infra - Provisión de infraestructura

Uno de los objetivos secundarios del proyecto era poder recrear la infraestructura automáticamente. La hipótesis es que esto por un lado fuerza a que esté documentado (en código) toda la configuración, y por otro facilita recrear/replicar el entorno. Se evitó los grandes jugadores (Pulumi, Terraform) y fuimos por usar los SDK de forma directa con un modelo a la Kubernetes: hay diferentes tipos de recurso, se declara el estado deseado, se detectan drifts, y se aplican los cambios para llegar al estado deseado.

La [ubicación del código fuente](https://github.com/jcodagnone/chapauy/tree/master/.dagger/infra) es un tanto extraña, pero facilita que se pueda usar directamente sin binario intermediario desde `dagger` como la generación del binario para ejecución local. Elegimos programar en Go ya que va muy bien para estas tareas, para reemplazar command line de bash, tiene excelente fit con dagger, y con Google Cloud - todo lo que se puede hacer desde el binario `gcloud` también se puede hacer con el SDK.

Desplegamos en Google Cloud Platform (GCP). Ya estábamos en GCP debido a que el proceso de geolocalización depende en parte de [Google Maps Geocoding API](https://developers.google.com/maps/documentation/geocoding/overview), y por otro lado aprovechamos los créditos de prueba. Es fácilmente desplegable en otros entornos: requerimos únicamente la ejecución de un contenedor sin estado (base de datos embebida y de solo lectura). El resto de la complejidad reside en automatizar la construcción de imágenes y el despliegue.

Los recursos principales provisionados son:
*  **[Google Artifact Registry](https://cloud.google.com/artifact-registry/docs)**: Repositorio privado para almacenar las imágenes Docker generadas por los procesos de construcción. Contamos con las siguientes imágenes:
		* `cli` y `web`: las dos aplicaciones principales *distroless*.
		* `data`: mantiene el directorio `db`, las copias de los documentos y la base de datos DuckDB.
		* `web-data`: la composición de `web` con únicamente la base de datos DuckDB de `data`. Es el artefacto que se ejecuta.
*   **[Google Cloud Run](https://cloud.google.com/run/docs)**: Aloja la aplicación web (`web-data`). Se configura como un servicio *serverless* (escalado a cero).
*   **[Google Cloud Build](https://cloud.google.com/build/docs)**: Motor de CI/CD que orquesta las tareas de construcción y despliegue. Se definen varios *triggers*:
    *   `build-master`: Se activa con cada *push* a la rama `master` para construir las aplicaciones `cli` y `web`.
    *   `daily-data-refresh`: Tarea programada que actualiza la base de datos DuckDB con la información más reciente (`./chapa impo update`).
    *   `deploy-web`: Despliega la combinación más reciente de aplicación web y datos, para ejecución manual.
*   **[Google Cloud Scheduler](https://cloud.google.com/scheduler/docs)**: Ejecuta periódicamente la actualización de datos mediante el trabajo `daily-data-refresh-job` diariamente a las 10:00 AM (hora de Uruguay), desencadenando la actualización de datos `daily-data-refresh`, la construcción de una nueva imagen `web-data` y su despliegue a producción.
*   **[IAM & Service Accounts](https://cloud.google.com/iam/docs/service-accounts)**: Se utiliza una cuenta de servicio dedicada (`deploy`) con permisos granulares (mínimo privilegio) para realizar las operaciones de despliegue, evitando el uso de credenciales personales o permisos excesivos.
*   **[Developer Connect](https://cloud.google.com/developer-connect/docs)**: Gestiona la conexión segura con el repositorio de GitHub (`jcodagnone/chapauy`), permitiendo a Cloud Build acceder al código fuente.
*   **[Service Usage](https://cloud.google.com/service-usage/docs)**: Habilita automáticamente las APIs necesarias en el proyecto de Google Cloud (Run, Build, Scheduler, IAM, etc.).

El objetivo era desplegar en **São Paulo** (`southamerica-east1`), pero el soporte para mapear dominios personalizados [es limitado](https://docs.cloud.google.com/run/docs/mapping-custom-domains#run) fuera de ciertas regiones, por lo que de momento corre en `us-east4` (N. Virginia).

La ejecución se puede hacer directamente con el binario o mediante `dagger call infra-setup`. Por defecto, la ejecución muestra las diferencias con el estado deseado:
```
$ go run infra/cmd/main.go setup
2025-12-18 08:56:07 Bootstrapping: Ensuring critical APIs are enabled for chapauy-20251216...
2025-12-18 08:56:12 Reconciling...
2025-12-18 08:56:13 ✅ Service Usage
2025-12-18 08:56:14 ✅ Artifact Registry (prod)
2025-12-18 08:56:15 ✅ Service Account (deploy)
2025-12-18 08:56:16 ✅ IAM Policy Binding: deploy
2025-12-18 08:56:17 ✅ IAM Policy Binding: Service Agent developer-connect
2025-12-18 08:56:17 ✅ IAM Policy Binding: Service Agent cloud-build
2025-12-18 08:56:18 ✅ IAM Policy Binding: Service Agent cloud-build-legacy
2025-12-18 08:56:23 ✅ Developer Connect Connection: github-repo1
2025-12-18 08:56:24 ✅ Cloud Build Trigger: build-master
2025-12-18 08:56:25 ✅ Cloud Build Trigger: daily-data-refresh
2025-12-18 08:56:26 ✅ Cloud Build Trigger: deploy-web
2025-12-18 08:56:27 ✅ Cloud Scheduler Job: daily-data-refresh-job
```


Si hubiese alguna diferencia, se puede ejecutar el comando para un único recurso con `--target` y forzar los cambios con `--apply`.

```
$ go run infra/cmd/main.go setup
2025-12-18 11:15:42 Bootstrapping: Ensuring critical APIs are enabled for chapauy-20251216...
2025-12-18 11:15:47 Reconciling...
2025-12-18 11:15:48 ✅ Service Usage
…
2025-12-18 11:16:00 ⚠️  Cloud Build Trigger: build-master: Drift detected! (Run with --target=trigger-build-master --apply to apply)
   diff: Description: Build images when push to masterX -> Build images when push to master; SourceToBuild missing; 
…
```
La primera vez que se corre, con un proyecto nuevo, es necesario correr varias veces el setup y esperar entre ejecuciones a que se propaguen los cambios (sobre todo de *Service Usage*).

El subcomando `list` permite descubrir los recursos:

```
$ go run infra/cmd/main.go  list
KEY                              NAME
-------------------------------- ----
services                         Service Usage
registry                         Artifact Registry (prod)
sa                               Service Account (deploy)
iam-deploy                       IAM Policy Binding: deploy
iam-sa-developer-connect         IAM Policy Binding: Service Agent developer-connect
iam-sa-cloud-build               IAM Policy Binding: Service Agent cloud-build
iam-sa-cloud-build-legacy        IAM Policy Binding: Service Agent cloud-build-legacy
iam-sa-cloud-scheduler           IAM Policy Binding: Service Agent cloud-scheduler
devconnect-github-repo1          Developer Connect Connection: github-repo1
trigger-build-master             Cloud Build Trigger: build-master
trigger-daily-data-refresh       Cloud Build Trigger: daily-data-refresh
trigger-deploy-web               Cloud Build Trigger: deploy-web
scheduler-daily-data-refresh-job Cloud Scheduler Job: daily-data-refresh-job
maps-key                         Google Maps API Key
```

El subcomando `maps` se encarga de configurar las APIs para que el proceso de curación pueda geolocalizar ubicaciones
``` 
$ go run infra/cmd/main.go maps
2025-12-18 11:19:07 Bootstrapping: Ensuring critical APIs are enabled for chapauy-20251216...
2025-12-18 11:19:11 Reconciling...
2025-12-18 11:19:11 ✅ Service Usage
2025-12-18 11:19:13 ✅ Google Maps API Key
```

El subcomando `deploy` dispara el despliegue de la última imagen `web-data`:
```
$ go run infra/cmd/main.go deploy
2025-12-18 11:21:15 Bootstrapping: Ensuring critical APIs are enabled for chapauy-20251216...
2025-12-18 11:21:21 🔍 Resolving latest image digest...
2025-12-18 11:21:22    Resolved: us-east4-docker.pkg.dev/chapauy-20251216/prod/web-data@sha256:b8b449d66c93afd6df0c20d178f39c100ec3146bfc29fb9630909c68f86e04ab
2025-12-18 11:21:22 🚀 Deploying to Cloud Run...
2025-12-18 11:21:25 ✅ Service updated successfully
2025-12-18 11:21:25 🔓 Setting IAM policy to allow unauthenticated access...
2025-12-18 11:21:25 ✅ IAM policy updated (allUsers -> roles/run.invoker)
```

Todos los comandos utilizan las [credenciales ADC](https://docs.cloud.google.com/docs/authentication/provide-credentials-adc), pero puede apuntarse a un archivo específico. Al correr en Dagger se utilizan los tokens de `gcloud auth print-access-token`, que son más portables.
## Pipeline (Dagger)

```text
       [ Código Fuente ]
               │
               ▼
    ╭───────────────────────╮
    │  build-and-publish    │
    ╰──────────┬────────────╯
               │
       ╭───────┴───────╮
       │               │
       ▼               ▼
   [ cli ]          [ web ]
    (base)           (base)
       │               │
       │               │
       │     ╭─────────┴──────────╮
       │     │     data-refresh   │
       │     ╰─────────┬──────────╯
       │               │
       │               │
       │               ▼
       │           [ data ]
       │         (DuckDB + DB)
       │               │
       ╰───────┬───────╯
               │
               ▼
    ╭───────────────────────╮
    │     build-web-data    │
    ╰──────────┬────────────╯
               │
               ▼
         [ web-data ]
        (App + Datos)
               │
               ▼
    ╭───────────────────────╮
    │     Cloud Run         │
    ╰───────────────────────╯
```
[^1]
Dagger.io entre sus promesas tiene la de proveer build portables, al ejecutar los procesos en contenedores aislados, el resultado es siempre consistente, independientemente del sistema operativo o la configuración local. Conocí la tecnología en la charla [Taming Your Application’s Environments - Marcos Lilljedahl, Dagger & Mauricio "Salaboy" Salatino, Diagrid](https://kccncna2024.sched.com/event/1i7kY/taming-your-applications-environments-marcos-lilljedahl-dagger-mauricio-salaboy-salatino-diagrid) y me pareció muy interesante. Sobre todo la posibilidad de describir el pipeline en lenguajes de programación en vez de nuevas sintaxis, sea Dockerfiles, YAMLs, etc.

Las funcionalidades principales expuestas en [`.dagger/main.go`](https://github.com/jcodagnone/chapauy/blob/master/.dagger/main.go) son:
*   **`infra-setup`**: Gestiona el aprovisionamiento de la nube detallado en la sección anterior.
*   **`build-and-publish`**: Construye las imágenes base de la CLI y la web desde el código fuente, publicándolas en el Artifact Registry.
*   **`data-refresh`**: Ejecuta la actualización diaria de datos. Levanta la imagen de la CLI, monta el volumen de datos actual, ejecuta `impo update` y genera una nueva imagen de datos actualizada.
*   **`build-web-data`**: Realiza la composición final. Inyecta la base de datos DuckDB más reciente (desde la imagen de datos) en la imagen de la aplicación web, produciendo el artefacto `web-data`.
*   **`deploy`**: Activa el despliegue del servicio en Cloud Run utilizando la última imagen `web-data` generada.

Estas son las funciones utilizadas por las tareas en **Cloud Build**.

Como desventaja, encontré un "impuesto" de dos minutos en cada construcción debido al *setup* de Dagger (no hallé una imagen de [cloud-builder](https://github.com/GoogleCloudPlatform/cloud-builders) oficial), y no logré un objetivo secundario: la generación de *provenance* [SLSA 3.0](https://slsa.dev/). Sin embargo, la portabilidad resulta sumamente conveniente.

[^1]: En mi defensa, uso ASCII ART [antes del advenimiento](https://github.com/jcodagnone/itba-ieee-computer-curso-git-adv/tree/master/diapositivas
	) de Genie
