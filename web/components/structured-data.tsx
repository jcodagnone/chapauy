/**
 * Copyright 2025 The ChapaUY Authors
 * SPDX-License-Identifier: Apache-2.0
 */

import React from "react"
import { BASE_URL } from "@/lib/config"

export function StructuredData() {
  const jsonLd = {
    "@context": "https://schema.org",
    "@graph": [
      {
        "@type": "WebSite",
        "@id": `${BASE_URL}/#website`,
        "url": BASE_URL,
        "name": "ChapaUY",
        "description": "Consultá y analizá infracciones de tránsito en Uruguay de forma transparente.",
        "publisher": {
          "@id": `${BASE_URL}/#organization`
        },
        "potentialAction": [
          {
            "@type": "SearchAction",
            "target": {
              "@type": "EntryPoint",
              "urlTemplate": `${BASE_URL}/offenses?vehicle={search_term_string}`
            },
            "query-input": "required name=search_term_string"
          }
        ],
        "inLanguage": "es-UY"
      },
      {
        "@type": "Organization",
        "@id": `${BASE_URL}/#organization`,
        "name": "ChapaUY",
        "url": BASE_URL,
        "logo": {
          "@type": "ImageObject",
          "url": `${BASE_URL}/logo.webp`,
          "width": "1200",
          "height": "630"
        },
        "sameAs": [
          "https://github.com/jcodagnone/chapauy"
        ]
      },
      {
        "@type": "WebApplication",
        "@id": `${BASE_URL}/#application`,
        "name": "ChapaUY",
        "url": BASE_URL,
        "description": "Herramienta para la visualización y análisis de infracciones de tránsito en Uruguay.",
        "applicationCategory": "PublicInformation",
        "operatingSystem": "All",
        "author": {
          "@id": `${BASE_URL}/#organization`
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
