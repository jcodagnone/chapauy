---
title: Enriquecimiento
description: Estrategias de normalización, geolocalización y clasificación.
---

La hidratación/enriquecimiento de los datos extraídos es uno de los valores agregados por el sistema y permite reducir cardinalidades en algunas dimensiones. Para que ocurra la hidratación primero tuvo que haber una curación de la información. Trabajamos sobre:
* geolocalización de ubicaciones - Ubicar `Ruta Interbalnearia y Milton Lussich` [aquí](https://www.google.com/maps/place/34%C2%B052'59.2%22S+55%C2%B002'40.4%22W)
* unificar la nomenclatura. `Ruta Interbalnearia y Milton Lussich` es igual a decir `Interbalnearia y Mil ton Lussich` o `Ruta Ruta Interbalnearia y Milton Lussich`.
* Extraer artículos normativos de las descripciones

## Inferencia basada en matriculas

A partir de la matrícula (`vehicle`) tenemos una heurística que no depende de anotación previa de información para inferir datos (implementada en [`impo/vehicle.go`](https://github.com/jcodagnone/chapauy/blob/master/impo/vehicle.go)). Este proceso se dispara automáticamente durante la etapa de [Extracción](/docs/010-acquire#extraccion). La posición de letras y números permite reconocer el país de origen, si es formato Mercosur, y en algunos casos si es un birrodado o no.

El sistema soporta la detección de matrículas de varios países de la región cuando no se indica la procedencia:

| Matrícula | País      | Formato  | Tipo      | Categoría / Departamento          |
| :-------- | :-------- | :------- | :-------- | :-------------------------------- |
| `BEC0000` | Uruguay   | Mercosur | Auto/Moto | Maldonado (B)                     |
| `SOF0000` | Uruguay   | Mercosur | Auto/Moto | Oficial (Montevideo)              |
| `LTX0000` | Uruguay   | Mercosur | Auto/Moto | Taxi (Colonia)                    |
| `123456`  | Uruguay   | Antiguo  | Auto/Moto | Montevideo                        |
| `B000000` | Uruguay   | Antiguo  | Auto/Moto | Maldonado                         |
| `SFA1234` | Uruguay   | Antiguo  | Moto      | Montevideo (Estrategia de letras) |
| `BFM2698` | Uruguay   | Mercosur | Moto      | Maldonado (Estrategia de bloques) |
| `AA000AA` | Argentina | Mercosur | Auto      | -                                 |
| `A000AAA` | Argentina | Mercosur | Moto      | -                                 |
| `AAA000`  | Argentina | Antiguo  | Auto      | -                                 |
| `AAA0A00` | Brasil    | Mercosur | Auto/Moto | -                                 |
| `AAAA000` | Paraguay  | Mercosur | Auto      | -                                 |
| `000AAAA` | Paraguay  | Mercosur | Moto      | -                                 |
| `AA0000`  | Chile     |          | Auto/Moto | -                                 |

El artículo del *Congreso de Intendentes* [Multas de tránsito a vehículos con matrícula “extranjera”](https://www.gub.uy/congreso-intendentes/comunicacion/noticias/multas-transito-vehiculos-matricula-extranjera) hace eco de algunas ambigüedades que existen con las matrículas previas al mercosur:
>Las intendencias, al igual que los restantes organismos asociados al Sucive (MTOP y Ministerio del Interior/Policía Caminera), que gestionan sus infracciones y peajes en el Sucive, deberán indicar en el formulario de alta al sistema, que las infracciones o cargos corresponden efectivamente a un vehículo extranjero.
>
>Esta instrucción se imparte porque el sistema informático no distingue matrículas nacionales de extranjeras. Por ese motivo el dato de la procedencia debe ser preciso por constituir un factor central para su correcta visualización. A título informativo, por ejemplo, las motos de Uruguay y los autos de origen argentino –con matrículas anteriores a la del Mercosur-, comparten la misma estructura de “3 letras + 3 números”, por lo que, si al anotarse la infracción se la marca como “vehículo nacional”, la misma irá directamente al Sucive, y si lo marcan como “vehículo extranjero” irá al nuevo departamento “extranjeros”. De la forma en que se haga esta anotación en el sistema, dependerá la correcta visualización como vehículo extranjero desde las plataformas del Sucive.

De hecho desde la [Resolución Policía Caminera N° 1000/025](https://impo.com.uy/bases/resoluciones-policia-caminera/1000-2025) ese organismo incorporó el país de la matrícula, lo que permite reducir la ambigüedad. Un caso es la patente de motos uruguayas previas al Mercosur `SFA1234` con las patentes de autos argentinas previas al Mercosur. En toda ambigüedad se presupone procedencia uruguaya.

En la detección de tipo de vehículo también existe ambigüedad en las patentes Mercosur uruguaya.
A diferencia de Argentina (donde `AA000AA` es auto y `A000AAA` es moto), el formato Mercosur de Uruguay (`AAA1234`) no distingue inherentemente entre tipos de vehículo.

Para resolver esto, el sistema utiliza una **estrategia de bloques** basada en los prefijos asignados por cada intendencia. Mediante la observación de los documentos originales, hemos identificado rangos específicos destinados a motos. Por ejemplo:
*   **Canelones:** Utiliza el bloque `AV` para motos.
*   **Colonia:** Utiliza el bloque `LS`.
*   **Maldonado:** Utiliza sub-bloques dentro de la serie de Garzón `BF` (como `BFM`, `BFN`, `BFO`, `BFP`).
*   **Lavalleja (Minas):** Utiliza el bloque `PAV`.
Esta heurística nos permite asignar la categoría `Moto` correctamente incluso cuando el documento de IMPO no lo especifica explícitamente, mejorando la precisión de las estadísticas y filtros de la aplicación. No he encontrado información oficial sobre estos bloques y seguramente faltan montones.

Para Uruguay, la matrícula permite inferencias más granulares:
*   **Departamento:** La primera letra de las matrículas (salvo formato Mercosur global) indica el departamento (ej. `S` = Montevideo, `B` = Maldonado).
*   **Motos:** Se detectan patrones específicos de motos, incluyendo la **estrategia de bloques** para series Mercosur.
*   **Categorías Especiales:** Se identifican vehículos oficiales, diplomáticos, taxis, remises, médicos, etc., mediante combinaciones reservadas (ej. `OF` para Oficial, `TX` para Taxi).

Pero esto no ha sido expuesto en la web.

El resto de las hidrataciones requiere anotar los datos. Para eso disponemos de una aplicación web secundaria que opera únicamente localmente, que presenta diferentes interfaces para anotar los datos. Todas las anotaciones se terminan persistiendo en [`judgments.json`](https://github.com/jcodagnone/chapauy/blob/master/judgments.json).

El flujo de trabajo es cargar las anotaciones:
```
$ go run main.go curation load
2025-12-18 15:20:15 ℹ️  New location judgments available (7097 > 0). Reloading...
2025-12-18 15:20:15 ♻️  Reloading curation data...
2025-12-18 15:20:20 ✅ Imported 7,097 location judgments from judgments.json
2025-12-18 15:20:20 ✅ Imported 220 articles from judgments.json
2025-12-18 15:20:26 ✅ Imported 3,529 description judgments from judgments.json
2025-12-18 15:20:26 ✅ Backfilled 0 offenses with geocoding data (23,364 pending offenses, 9,268 unique locations)
2025-12-18 15:20:26 ✅ Backfilled 0 offenses with description articles (0 pending offenses, 0 unique descriptions)
```

correr la interface
```
$ go run main.go curation serve
2025-12-18 15:21:48 GOOGLE_MAPS_API_KEY is not set. Attempting to retrieve via ADC...
2025-12-18 15:21:48 ⚠️ No Project ID found in credentials. Using fallback: chapauy-20251216
2025-12-18 15:21:49 Found key resource 'projects/722960972580/locations/global/keys/1d6f85b0-580e-4401-84cd-37b24ec23e71', retrieving secret...
2025-12-18 15:21:50 ✅ Successfully retrieved Google Maps API Key via ADC
📍 Geocoding: Google Maps (primary)
🗺️  Geocoding workflow server starting...
📍 Open http://localhost:8080 in your browser
🔒 Local only - not exposed to internet
```

Se proveen 3 endpoints que trabajan de la misma forma. Van desencolando items que requieren revision. Por defecto intenta proveer un valor, por ejemplo para una ubicación una busqueda hecha en Google maps.  En todos los casos `CTRL+ENTER` permite aceptar la sugerencia, y `ESC` saltear el item.
* http://localhost:8080/?view=queue - permite geolocalizar ubicaciones
* http://localhost:8080/?view=cluster - permite normalizar los nombres de ubicaciones `AV 8 DE OCTUBRE y AV CENTENARIO` vs `AV CENTENARIO y AV 8 DE OCTUBRE`
* http://localhost:8080/descriptions - permite curar descripciones contra los artículos

Toda la información se almacena [online en la base DuckDB](/docs/000-arquitectura#base-de-datos-sql), pero se recomienda que, terminada la sesión de curación, se almacene la información de vuelta en `judgments.json`. Esto permite mantener diferentes bases o arrancar desde cero.

```
$ go run main.go curation store
✅ Exported 7,097 location judgments, 3,529 description judgments, and 220 articles to judgments.json
```

### Geocoding

Por defecto utilizamos la [Geocoding API](https://developers.google.com/maps/documentation/geocoding/overview) de Google Maps Platform. Es rápida, tiene buenos resultados, y para el volumen que debemos manejar no es costoso. La inferencia se hace una vez por cada texto nuevo y almacenamos el resultado para siempre.

Se priorizan el resultado con mayor nivel de confianza (`ROOFTOP` o `RANGE_INTERPOLATED`), y se contextualiza con la base de datos. Si se busca por ejemplo `AV 8 DE OCTUBRE y AV CENTENARIO` que viene de una multa de la base de Montevideo, se busca `AV 8 DE OCTUBRE y AV CENTENARIO, Montevideo`.

En Montevideo funciona muy bien, tiene en general problemas con algunas calles que no siguen el damero, como `L A DE HERRERA`.

Google Maps no funciona bien para las multas en Rutas `RUTA NACIONAL 3 y km 383`. Es el caso de  las infracciones provienen de radares fijos de rutas manejadas por el  Ministerio de Transporte y Obras Públicas (MTOP). Supo existir el recurso 


Se accede al geoportal mediante un link ubicado en https://web.archive.org/web/20241106094300/https://www.gub.uy/ministerio-transporte-obras-publicas/comunicacion/publicaciones/ubicacion-radares que apuntaba al [geoportal del MTOP](https://geoportal.mtop.gub.uy/visualizador/?call=capas&c=radares_rutas) que tiiene una capa con los radares y los nombres que se utilizan en el portal.

El layer de radares se puede descargar directamente desde
https://geoservicios.mtop.gub.uy/geoserver/inf_tte_ttelog_terrestre/ows?service=WFS&version=1.0.0&request=GetFeature&typeName=inf_tte_ttelog_terrestre:radares_rutas&outputFormat=application/json 
```
{
  "type": "FeatureCollection",
  "features": [
    {
      "type": "Feature",
      "id": "radares_rutas.34",
      "geometry": {
        "type": "Point",
        "coordinates": [
          -56.2629,
          -34.5912
        ]
      },
      "geometry_name": "geom",
      "properties": {
        "ruta": 5,
        "progresiva": "38K131",
        "gestion": "MTOP",
        "descrip": "Juanicó"
      }
    },
    …
```

En las notificaciones esto suele escribirse como `Ruta 005 y 038K131_D`. Hay toda una heurística para intentar usar estos nombres.

### Descripciones

Las descripciones de las infracciones también son texto libre y varían enormemente ("Exceso vel.", "Art 13 vel.", "Velocidad excesiva"). El proceso de curación asigna a cada descripción única:
1.  **ID de Artículo:** (Ej. `13.3.A`) Referencia a normas estandarizadas (Reglamento Nacional de Circulación Vial, SUCIVE).
2.  **Código de Grupo:** (Ej. `13` para Velocidad, `18` para Estacionamiento).

Para asistir en la curación, el sistema implementa un clasificador automático basado en similitud (ver [`impo/description_classifier.go`](https://github.com/jcodagnone/chapauy/blob/master/curation/description_classifier.go)):
*   **Vectorización (Bag-of-words):** El texto se limpia, normaliza a minúsculas y se divide en tokens.
*   **Similitud de Coseno:** Se calcula la similitud entre el vector de la descripción y los vectores de los artículos reglamentarios.
*   **Sugerencias:** Se presentan los artículos con mayor puntaje (0 a 1), donde 1.0 indica una coincidencia exacta.

Muchas descripciones contienen múltiples infracciones separadas por comas (ej. `EXCESO DE VELOCIDAD, SIN CINTURON`). El sistema detecta estos casos inteligentemente:
*   **Detección:** Si el análisis por partes arroja artículos diferentes, se activa el modo multi-artículo.
*   **Desglose:** La interfaz (y el comando `--multi`) desglosan la descripción para clasificar cada fragmento de forma independiente.
*   **Efecto Acumulativo:** Cada fragmento clasificado se guarda por separado. Al encontrarlo nuevamente en otra descripción, el sistema lo reconoce con puntaje 1.0, permitiendo saltar el trabajo repetitivo y mejorando la eficiencia en un 60%.
