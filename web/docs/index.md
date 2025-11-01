---
title: Introducción
description:
author: Juan F. Codagnone. Punta del Este, 17 de diciembre de 2025.
---
ChapaUY recopila las infracciones de tránsito de Uruguay que han sido notificadas públicamente, permitiendo entenderlas y analizarlas desde diferentes dimensiones.

En Uruguay, diez departamentos y dos ministerios[^1] publican en el Diario Oficial las notificaciones de infracciones (ver [Directorio de Bases de Datos Institucionales](https://www.impo.com.uy/directorio-bases-institucionales/#seccion-bases-multas)). Un ejemplo de esta publicación es la [Notificación del Departamento de Movilidad de la Intendencia de Maldonado N° 486/025](https://www.impo.com.uy/bases/notificaciones-transito-movilidad-maldonado/486-2025).

Si bien esta información es pública por definición y se puede rastrear mediante buscadores clásico, como ser[`SMO579 site:www.impo.com.uy`](https://www.google.com/search?q=SMO579+site%3Awww.impo.com.uy), su formato disperso dificulta transformar estos datos en información, y ya es fuente de OSINT. Aunque el SUCIVE ofrece un servicio de [avisos por correo electrónico](https://www.sucive.gub.uy/suscripcion_notificaciones?1), no provee información pública accionable de forma agregada. Estos datos a veces se filtran en notas periodísticas —como [«Argentinos sorprendidos por recibir multas del verano pasado en Punta del Este...»](https://www.elpais.com.uy/mundo/argentina/argentinos-sorprendidos-por-recibir-multas-del-verano-pasado-en-punta-del-este-con-deudas-de-hasta-us-10-000)—, pero los datos crudos no suelen estar disponibles para el análisis masivo. El catálogo Datos abierto presenta un dataset [Multas de tránsito, Intendencia de Montevideo](https://catalogodatos.gub.uy/dataset/multas-de-transito), pero además de ser únicamente montevideanos, no se actualizan desde el año 2023.

Este proyecto *indie*, sin fines de lucro, surge de una experiencia personal: al recibir una notificación por una infracción —un descuido en una vía con cambio de velocidad máxima— y constatar que la información era pública, comencé a cuestionarme sobre la recaudación teórica de cada radar, los puntos calientes... Fue la excusa perfecta para experimentar con tecnologías que tenía pendientes desde hace 15 años (`golang`) y otras más recientes (`v0`, `Copilot`, `Gemini`, `dagger`, etc.). Quien me conoce sabe que estos temas siempre me resultan entretenidos: hace 23 años hacía el primer *scraper* online [iolsucker](http://www.leak.com.ar/~juan/code/iolsucker/), uno offline [Páginas Blancas](http://www.leak.com.ar/~juan/papers/guia/), y una primera incursión en sistemas GIS y web semántica con [flof.com.ar](https://web.archive.org/web/20061129014000/http://flof.com.ar/bin/home/;jsessionid=32544D8A531871795D4F87C5DCD6517B).

Más allá de la extracción de datos de los[ documentos originales](/docs/010-acquire), la propuesta de valor de este sitio radica en el trabajo de [curación](/docs/020-curate) y enriquecimiento de la información. Hemos trabajado en inferir:

*   **El tipo de vehículo:** Distinguiendo, por ejemplo, si es un birrodado (moto).
*   **El país de la matrícula:** *El 25% de las infracciones de 2025 en Maldonado corresponden a matrículas extranjeras* ([Fuente](/offenses?database=45&facet=database&facet=year&facet=country&year=2025#facet-country)).
*   **La geolocalización:** Permitiendo visualizar las infracciones en el mapa, con la posibilidad futura de correlacionarlas con bases de siniestralidad vial.
*   **La normalización de ubicaciones:** Unificando la nomenclatura de lugares. Por ejemplo, los siguientes textos reales refieren al mismo punto:
	*   *Av. Roosevelt y Av. Los Alpes*
	*   *Ruta Av. Roosevelt y Av. Los Alpes*
	*   *ROOSEVELT FRANKLIN D. AVDA. Y LOS ALPES AVDA.*
*   **La codificación de las descripciones:** Toda infracción debe tener un fundamento normativo, usualmente basado en la [Ley N° 18.191](https://www.impo.com.uy/bases/leyes/18191-2007) o en el [Texto Ordenado del SUCIVE 2024](https://www.gub.uy/congreso-intendentes/comunicacion/publicaciones/texto-ordenado-del-sucive-2024). Aunque cada departamento posee sus propias normas, trabajamos para alinearlas a estos estándares. Esto permite agrupar infracciones por conceptos generales (ej. *13 - De las velocidades*) o artículos específicos (ej. *13.3.B - Exceso de velocidad hasta 20 km/h*), independientemente de la [diversidad de redacciones](/offenses?facet=description#facet-description) utilizadas por el cuerpo inspectivo.

En la pantalla principal, encontrará a la izquierda las diferentes dimensiones para filtrar las infracciones (*Base de datos*, *Año*, *País de matrícula*, *Tipo de vehículo*). A la derecha, podrá alternar la visualización: desde un listado similar al documento original, hasta gráficos estadísticos y un mapa interactivo. También es posible navegar directamente sobre los documentos fuente para ver la infracción en su contexto original.

Estoy seguro de que encontrará hallazgos interesantes entre los [principales reincidentes](/offenses?facet=features&facet=vehicle#facet-vehicle). Espero que el análisis escape a una mirada meramente moralista; *dura lex, sed lex*, sí, pero la fiscalización automatizada no siempre es sinónimo de mayor seguridad vial. Recomiendo las lecturas de Manuel da Fonte, como [«Radares y cómo empeorar la seguridad vial»](https://contraviento.uy/2025/12/08/radares-y-como-empeorar-la-seguridad-vial/) y [«¿Qué incidencia han tenido los semáforos en los recientes accidentes en 8 de octubre?»](https://contraviento.uy/2023/02/26/semaforos-y-seguridad-vial-mito-o-realidad/).

Un de estos rincidentes tiene una  [historia](/offenses?date=2025-01-13&vehicle=AG050HW&facet=vehicle) *cliché* que no resultaba evidente en los documentos originales,  una cálida noche de verano en la primera quincena de enero, un trayecto desde José Ignacio hacia la Península donde saltan todas las cámaras,  una alcoholemia  y más  exceso de velocidad a la vuelta.

Los invito a explorar los datos, sacar sus propias conclusiones y sumarse a este ejercicio de transparencia ciudadana. Quedo a disposición por dudas, sugerencias o aportes a través de [X](https://x.com/juam) o por [correo electrónico](mailto:jcodagnone-AT-gmail.com).

[^1]: Intendencia de Canelones, Intendencia de Colonia, Intendencia de Lavalleja, Intendencia de Maldonado, Intendencia de Montevideo, Intendencia de Paysandú, Intendencia de Río Negro, Intendencia de Soriano,  Intendencia de Tacuarembó, Intendencia de Treinta y Tres,  Ministerio de Transporte (Vialidad), Ministerio del Interior (Policía Caminera)
	
