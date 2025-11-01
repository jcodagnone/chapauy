---
title: Adquisición de datos
description: Adquisición, y sistematización de la información
---

La fuente de datos de todas las infracciones son las bases institucionales disponibles en la sección *Consultar bases de infracciones y multas de tránsito publicadas en el Diario Oficial* del [directorio de Base de Datos Institucional](https://www.impo.com.uy/directorio-bases-institucionales/). Diez departamentos y dos ministerios publican las multas en el diario oficial.

```
$ build/chapa impo list
Base de datos disponibles:
╭────┬────────────────┬─────────────────────────────────────────────────────────────╮
│ Id │ Nombre         │ Ubicación                                                   │
├────┼────────────────┼─────────────────────────────────────────────────────────────┤
│ 65 │ Caminera       │ https://www.impo.com.uy/base-institucional/multascaminera   │
│ 40 │ Canelones      │ https://www.impo.com.uy/base-institucional/multascanelones  │
│ 48 │ Colonia        │ https://www.impo.com.uy/base-institucional/multascolonia    │
│ 26 │ Lavalleja      │ https://impo.com.uy/base-institucional/multaslavalleja      │
│ 45 │ Maldonado      │ https://impo.com.uy/base-institucional/multasmaldonado      │
│  6 │ Montevideo     │ https://www.impo.com.uy/base-institucional/cgm              │
│ 43 │ Paysandu       │ https://impo.com.uy/base-institucional/multaspaysandu       │
│ 55 │ Rio Negro      │ https://impo.com.uy/base-institucional/multasrionegro       │
│ 49 │ Soriano        │ https://www.impo.com.uy/base-institucional/multassoriano    │
│ 56 │ Tacuarembó     │ https://www.impo.com.uy/base-institucional/multastacuarembo │
│ 52 │ Treinta y Tres │ https://impo.com.uy/base-institucional/multastreintaytres   │
│ 68 │ Vialidad       │ https://www.impo.com.uy/base-institucional/multasmtop       │
╰────┴────────────────┴─────────────────────────────────────────────────────────────╯
```

Los puntos de entrada y parámetros de cada base se encuentran definidas en [impo/dbrefs.go](https://github.com/jcodagnone/chapauy/blob/master/impo/dbrefs.go).

Cada artículo del diario oficial (PDF) tiene una versión HTML. Un ejemplo es [Notificación Departamento de Movilidad Intendencia de Maldonado N° 486/025](https://www.impo.com.uy/bases/notificaciones-transito-movilidad-maldonado/486-2025). Todos los documentos usan la misma estructura de tabla, pero las columnas y los formatos varían.

El módulo [`impo`](https://github.com/jcodagnone/chapauy/tree/master/impo) se ocupa de descubrir incrementalmente los documentos de una base, descargarlos para tener una copia local, extraer la información y almacenarla de forma sistematizada. La información obtenida se enriquece con datos curados (clasificación de descripciones y de ubicaciones) para su geolocalización y sistematización (ver [Enriquecimiento](/docs/020-curate)).

El proceso se ve de esta forma.
``` 
$ build/chapa impo update
2025-12-13 09:24:12 ℹ️ Location judgments table is not empty, skipping import.
2025-12-13 09:24:12 ℹ️ Articles table is not empty, skipping import.
2025-12-13 09:24:12 ℹ️ Description judgments table is not empty, skipping import.
2025-12-13 09:24:12 Updating database 65 - Caminera
2025-12-13 09:24:12 Search - Retrieving first page <https://www.impo.com.uy/cgi-bin/bases/consultaBasesBS.cgi?tipoServicio=65>
2025-12-13 09:24:15 Search - Page 1 stats - 0 new records from a total of 50 records
2025-12-13 09:24:15 Total stats - 0 new records from a total of 50 records across 1 pages
2025-12-13 09:24:15 Nothing to download
…
2025-12-13 09:24:35 Updating database 45 - Maldonado
2025-12-13 09:24:35 Search - Retrieving first page <https://impo.com.uy/cgi-bin/bases/consultaBasesBS.cgi?tipoServicio=45>
2025-12-13 09:24:38 Search - Page 1 stats - 3 new records from a total of 50 records
2025-12-13 09:24:38 Search - Retrieving next page tipoServicio=45&realizarconsulta=SI&idconsulta=OTzEZzlJ&nrodocdesdehasta=51-100
2025-12-13 09:24:39 Search - Page 1 stats - 0 new records from a total of 50 records
2025-12-13 09:24:39 Total stats - 3 new records from a total of 100 records across 2 pages
2025-12-13 09:24:39 [1/3] Downloading https://www.impo.com.uy/bases/notificaciones-transito-movilidad-maldonado/484-2025
2025-12-13 09:24:40 [2/3] Downloading https://www.impo.com.uy/bases/notificaciones-transito-movilidad-maldonado/485-2025
2025-12-13 09:24:40 [3/3] Downloading https://www.impo.com.uy/bases/notificaciones-transito-movilidad-maldonado/486-2025
2025-12-13 09:24:41 Download phase completed - 3 successful, 0 failed
2025-12-13 09:24:41 Extraction phase complete - 1898 new records, 0 errors from 3 documents, 3 successful and 0 failed.
2025-12-13 09:24:41 Updating database 6 - Montevideo
…
2025-12-13 09:25:33 Total search phase metrics - 21 new records from a total of 1044 records across 21 pages
2025-12-13 09:25:33 Total download phase metrics - 21 successful, 0 failed
2025-12-13 09:25:33 Total extraction phase metrics - 9424 new records, 3 errors from 60 documents, 48 successful and 12 failed.
2025-12-13 09:25:33 ✅ Backfilled 7,235 offenses with geocoding data (29,958 pending offenses, 10,293 unique locations)
2025-12-13 09:25:33 ✅ Backfilled 0 offenses with description articles (0 pending offenses, 0 unique descriptions)
```

## Descubrimiento

El objetivo de esta etapa es descubrir nuevos documentos. Esta etapa puede ser salteada con el argumento `--skip-search`.
Los documentos se pueden clasificar entre _notificaciones_ y _resoluciones_.

Para descubrir documentos usamos el buscador de la base en el sitio. Para poder usarlo, se debe seguir una serie de pasos precisa: para  acceder correctamente, es fundamental utilizar la **SeedURL** (URL semilla).
Cada base tiene un link de entrada específico (por ejemplo, para Maldonado es [https://impo.com.uy/base-institucional/multasmaldonado](https://impo.com.uy/base-institucional/multasmaldonado)). Este link redirige al buscador de la base correspondiente, pero lo importante es ingresar por el link inicial, ya que es en ese momento donde se setean las cookies de sesión y de usuario necesarias para la navegación.

El cliente HTTP ([impo/client.go](https://github.com/jcodagnone/chapauy/blob/master/impo/client.go)) mantiene las cookies necesarias para mantener la sesión, se identifica responsablemente, minimiza el tráfico utilizando compresión. Los argumentos `--trace-http`  y `--trace-http-body` permiten observar el tráfico en la salida:
```
$ go run main.go impo update --trace-http
2025-12-18 12:28:23 ✅ Curation data is up to date. Skipping import.
2025-12-18 12:28:23 Updating database 65 - Caminera
> HEAD /base-institucional/multascaminera HTTP/1.1
> Host: www.impo.com.uy
> User-Agent: chapauy/development (+https://github.com/jcodagnone/chapauy)
> Accept: */*
> 
> 
< RESPONSE: [965.035425ms]
< HTTP/1.1 302 Found
< Connection: close
< Content-Type: text/html; charset=iso-8859-1
< Date: Thu, 18 Dec 2025 15:28:24 GMT
< Location: /cgi-bin/bases/principalBases.cgi?tipoServicio=65
< Server: Apache/2.2.15 (Red Hat)
< Set-Cookie: usrts_65=anonimompcam; path=/
< Set-Cookie: idsesionanonimompcam=w3TMuELc; path=/
< 
< 
> HEAD /cgi-bin/bases/principalBases.cgi?tipoServicio=65 HTTP/1.1
> Host: www.impo.com.uy
> User-Agent: chapauy/development (+https://github.com/jcodagnone/chapauy)
> Accept: */*
> Cookie: usrts_65=anonimompcam; idsesionanonimompcam=w3TMuELc
> Referer: https://www.impo.com.uy/base-institucional/multascaminera
> …
```

Ya con la cookie podemos realizar la búsqueda:
```
2025-12-13 09:49:31 Search - Retrieving first page <https://impo.com.uy/cgi-bin/bases/consultaBasesBS.cgi?tipoServicio=45>
> POST /cgi-bin/bases/consultaBasesBS.cgi?tipoServicio=45 HTTP/1.1
> Host: impo.com.uy
> User-Agent: chapauy/development (+https://github.com/jcodagnone/chapauy)
> Content-Length: 304
> Accept: */*
> Content-Type: application/x-www-form-urlencoded
> Cookie: usrts_45=anonimommal; idsesionanonimommal=2rcA9Wup
> Accept-Encoding: gzip
> 
> articulos=&campotexto1=TODOS&campotexto2=TODOS&campotexto3=TODOS&combo1=802&fechadiar1=&fechadiar2=&fechapro1=&fechapro2=&indexcombobasetematica=-1&ntema=&nuevaconsulta=SI&numeros=&optexto1=Y&optexto2=Y&parlistabases=&realizarconsulta=SI&refinar=&tema=&texto1=&texto2=&texto3=&textolibre=&tipoServicio=45
< RESPONSE: [2.333881221s]
< HTTP/1.1 200 OK
< Transfer-Encoding: chunked
< Content-Type: text/html; charset=ISO-8859-1
< Date: Sat, 13 Dec 2025 12:49:31 GMT
< Server: Apache/2.2.15 (Red Hat)
< 
< 
2025-12-13 09:49:34 Search - Page 1 stats - 0 new records from a total of 50 records
…

< 					<tr class="sitem">
< 						<td width="60px">3.</td>
< 						<td>
< 							<a class="visitado" href="/bases/notificaciones-transito-movilidad-maldonado/488-2025">
< 								<strong>Notificaci�n Departamento de Movilidad Intendencia de Maldonado 488/025    </strong>
< 								 <font color="#014895">(Documento original)</font> 
< 							</a>
< 							<font size="2"><br>NOTIFICACION POR CONTRAVENCION A NORMAS DE TRANSITO</font>
< 						</td>
< 					</tr>
```

El tiempo de respuesta de cada búsqueda es de varios segundos. En la medida que una página contenga algún elemento nuevo iremos a la siguiente, siempre con un límite máximo para evitar ciclos que puedan ser inducidos para gastar recursos (`--search-max-depth`).

La información de cada paso se almacena localmente en una [base de datos sobre el filesystem](/docs/000-arquitectura#chapa-cli).

## Descarga

Una vez identificados los documentos, el sistema procede a descargar aquellos que aún no han sido almacenados localmente. Esta etapa puede ser salteada con el argumento `--skip-download`.

La función `downloadMissing` en [impo/client.go](https://github.com/jcodagnone/chapauy/blob/master/impo/client.go) se encarga de esta tarea:

1.  **Verificación:** Compara los documentos descubiertos contra los ya existentes en el `FileStore`.
2.  **Descarga:** Itera sobre los documentos faltantes, descargando el HTML original de las resoluciones o notificaciones.
3.  **Manejo de fallos:** Los errores de conexión, la validación de status codes, y headers,  o descarga son registrados para no interrumpir el proceso general, permitiendo reintentar en futuras ejecuciones.

## Extracción

En esta etapa se procesan las copias locales de los documentos y se transforma el HTML no estructurado en datos útiles [impo/extract.go](https://github.com/jcodagnone/chapauy/blob/master/impo/extract.go). Esta etapa puede ser salteada con el argumento `--skip-extract`. Solo se realiza la extracción de los documentos que se encuentren en filesystem pero que no se encuentren en la base datos. Este comportamiento puede cambiarse con el argumento `--extract-full`.

El proceso implica:
*   **Parsing:** Se procesa el árbol DOM del documento HTML.
*   **Identificación de Datos:** Se busca la tabla principal (clase `tabla_en_texto`) que contiene los detalles de las infracciones.
*   **Normalización de Columnas:** Dado que los encabezados varían entre intendencias (ej. "Matrícula", "Dominio", "Matrícula y padrón"), se utiliza una lógica de mapeo (`documentPropertyFromString`) para unificar estos campos.
*   **Sanitización:**
    *   **Fechas:** Se normalizan diversos formatos de fecha y hora.
    *   **Valores Monetarios:** Las Unidades Reajustables (UR) se almacenan como enteros escalados para preservar la precisión.
    *   **Matrículas:** Se eliminan espacios y caracteres extraños para estandarizar los identificadores vehiculares.

Los documentos originales en IMPO están codificados `ISO-8859-1`, y algunos documentos ya contienen problemas de codificación - seguramente del documento origen que enviaron las intendencias a IMPO. Para mitigar estos errores la función [`Node2string`](https://github.com/jcodagnone/chapauy/blob/master/utils/htmlutils/htmlutils.go) implementa una lógica de detección y corrección: 
* Detección del carácter de reemplazo de Unicode (`U+FFFD`) para detectar problemas en el manejo de *charset*. 
* Corregir secuencias mal codificadas comunes (como `Ã³` por `ó`) y se validan los caracteres resultantes contra expresiones regulares de seguridad para evitar la inyección de contenido inesperado.

En ocasiones, la tabla de infracciones carece de una columna de descripción explícita. Sin embargo, el cuerpo del documento puede contener referencias normativas, como "se constató la contravención a lo dispuesto en el art. 9" - un clásico de Montevideo. El extractor analiza el texto circundante (`<p>`, `<div>`) para inferir y completar estos datos faltantes.

El proceso de extracción usa muchos ciclos de CPU y procesa en paralelo - esto permite ahorrar tiempo cuando se arranca desde una base vacía. Se puede manejar el paralelismo con `--extract-max-procs`, y se puede evitar almacenar los resultados de documentos que tengan al menos un error con `--skip-extract-errors`. Esto permite revisar detalladamente estos errores. Hay errores legítimos, por ejemplo en la [Notificación Dirección de Tránsito Intendencia de Lavalleja N° 14/024](https://www.impo.com.uy/bases/notificaciones-transito-lavalleja/14-2024) para el dominio `PAV 1450` hay un error que permite suponer que el documento se armó con una planilla de cálculo y al arrastrar las fechas se generaron fechas del futuro:
* 30/03/2025
* 30/03/2026
* 30/03/2027
* 30/03/2028
* 30/03/2029
* 30/03/2030
Hay otros errores que pueden surgir por cambios en el formato de los documentos. Por ejemplo Colonia desde la [Notificación Dirección de Tránsito y Transporte Intendencia de Colonia N° 76/025](https://www.impo.com.uy/bases/notificaciones-transito-colonia/76-2025) incorporó la Cédula de Identidad como columna - seguramente preparando el terreno para la quita de puntos. O por ejemplo desde la
[Resolución Policía Caminera N° 1000/025](https://impo.com.uy/bases/resoluciones-policia-caminera/1000-2025) se incorporó el país de la matrícula -seguramente a pedido de SUCIVE, ver [Enriquecimiento](/docs/020-curate).

Como mecanismo de seguridad adicional, el sistema cuenta con un *failsafe* que impide el almacenamiento de documentos si la proporción de errores supera el 5%. Esto permite detectar de forma temprana cambios en la estructura de IMPO que requieran ajustes en la extracción. Aquellos documentos que superan este umbral por errores legítimos (como la citada [Notificación Dirección de Tránsito Intendencia de Lavalleja N° 14/024](https://www.impo.com.uy/bases/notificaciones-transito-lavalleja/14-2024)) son revisados manualmente e incorporados a una lista de excepciones en el código.

Esta fase aplica algunos de los enriquecimientos como ser la inferencia de información en base a la matrícula, geocoding, y la detección de norma en base a la descripción (ver detalles en el proceso de [Enriquecimiento](/docs/020-curate)).
