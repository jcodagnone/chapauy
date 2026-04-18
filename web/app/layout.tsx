/**
 * Copyright 2025 The ChapaUY Authors
 * SPDX-License-Identifier: Apache-2.0
 */

import type React from "react"
import type { Metadata, Viewport } from "next"
import { Suspense } from "react"
import { StructuredData } from "@/components/structured-data"
import { BASE_URL } from "@/lib/config"
import "./globals.css"

export const viewport: Viewport = {
  themeColor: "#09090b", // Matches background
  width: "device-width",
  initialScale: 1,
}

export const metadata: Metadata = {
  metadataBase: new URL(BASE_URL),
  alternates: {
    canonical: "/",
  },
  title: "ChapaUY - Búsqueda de Infracciones de Tránsito",
  description: "Consultá y analizá infracciones de tránsito en Uruguay de forma transparente.",
  keywords: ["infracciones", "multas", "tránsito", "Uruguay", "sucive", "patente", "matrícula", "chapa", "transparencia", "datos abiertos"],
  robots: {
    index: true,
    follow: true,
  },
  openGraph: {
    title: "ChapaUY - Búsqueda de Infracciones de Tránsito",
    description: "Consultá y analizá infracciones de tránsito en Uruguay de forma transparente.",
    url: BASE_URL,
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
      <head>
        <StructuredData />
      </head>
      <body className="font-sans antialiased">
        <Suspense fallback={<div>Loading...</div>}>{children}</Suspense>
      </body>
    </html>
  )
}
