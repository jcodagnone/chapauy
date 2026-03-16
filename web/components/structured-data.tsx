/**
 * Copyright 2025 The ChapaUY Authors
 * SPDX-License-Identifier: Apache-2.0
 */

import React from "react"

export function StructuredData() {
  const jsonLd = {
    "@context": "https://schema.org",
    "@graph": [
      {
        "@type": "WebSite",
        "@id": "https://chapa.uy/#website",
        "url": "https://chapa.uy",
        "name": "ChapaUY",
        "description": "Consultá y analizá infracciones de tránsito en Uruguay de forma transparente.",
        "publisher": {
          "@id": "https://chapa.uy/#organization"
        },
        "potentialAction": [
          {
            "@type": "SearchAction",
            "target": {
              "@type": "EntryPoint",
              "urlTemplate": "https://chapa.uy/offenses?vehicle={search_term_string}"
            },
            "query-input": "required name=search_term_string"
          }
        ],
        "inLanguage": "es-UY"
      },
      {
        "@type": "Organization",
        "@id": "https://chapa.uy/#organization",
        "name": "ChapaUY",
        "url": "https://chapa.uy",
        "logo": {
          "@type": "ImageObject",
          "url": "https://chapa.uy/logo.webp",
          "width": "1200",
          "height": "630"
        },
        "sameAs": [
          "https://github.com/jcodagnone/chapauy"
        ]
      },
      {
        "@type": "WebApplication",
        "@id": "https://chapa.uy/#application",
        "name": "ChapaUY",
        "url": "https://chapa.uy",
        "description": "Herramienta para la visualización y análisis de infracciones de tránsito en Uruguay.",
        "applicationCategory": "PublicInformation",
        "operatingSystem": "All",
        "author": {
          "@id": "https://chapa.uy/#organization"
        }
      }
    ]
  }

  return (
    <script
      type="application/ld+json"
      dangerouslySetInnerHTML={{ __html: JSON.stringify(jsonLd) }}
    />
  )
}
