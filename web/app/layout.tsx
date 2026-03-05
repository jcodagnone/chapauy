/**
 * Copyright 2025 The ChapaUY Authors
 * SPDX-License-Identifier: Apache-2.0
 */

import type React from "react"
import type { Metadata } from "next"
import { Suspense } from "react"
import "./globals.css"

export const metadata: Metadata = {
  metadataBase: new URL("https://chapa.uy"),
  title: "ChapaUY - Búsqueda de Infracciones de Tránsito",
  description: "Consultá y analizá infracciones de tránsito en Uruguay de forma transparente.",
  openGraph: {
    title: "ChapaUY - Búsqueda de Infracciones de Tránsito",
    description: "Consultá y analizá infracciones de tránsito en Uruguay de forma transparente.",
    url: "https://chapa.uy",
    siteName: "ChapaUY",
    images: [
      {
        url: "/logo.webp",
        width: 1200,
        height: 630,
        alt: "ChapaUY Logo",
      },
    ],
    locale: "es_UY",
    type: "website",
  },
  twitter: {
    card: "summary_large_image",
    title: "ChapaUY - Búsqueda de Infracciones de Tránsito",
    description: "Consultá y analizá infracciones de tránsito en Uruguay de forma transparente.",
    images: ["/logo.webp"],
  },
}

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode
}>) {
  return (
    <html lang="es" className="dark">
      <body className="font-sans antialiased">
        <Suspense fallback={<div>Loading...</div>}>{children}</Suspense>
      </body>
    </html>
  )
}
