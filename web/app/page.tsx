/**
 * Copyright 2025 The ChapaUY Authors
 * SPDX-License-Identifier: Apache-2.0
 */

import Link from "next/link"
import { Button } from "@/components/ui/button"
import { Search, Shield, FileText } from "lucide-react"
import { GlobalLinks } from "@/components/global-links"


export default function Home() {
  return (
    <div className="bg-background flex min-h-screen flex-col">
      {/* Header */}
      <header className="border-border bg-card border-b">
        <div className="container mx-auto flex items-center justify-between px-6 py-4">
          <h1 className="text-foreground text-xl font-semibold">ChapaUY</h1>
          <GlobalLinks />
        </div>
      </header>

      {/* Hero Section */}
      <main className="flex flex-1 items-center justify-center px-6">
        <div className="mx-auto max-w-3xl space-y-8 text-center">
          <div className="space-y-4">
            <h2 className="text-foreground text-4xl font-bold text-balance md:text-5xl">
              Buscador de Infracciones de Tránsito
            </h2>
            <p className="text-muted-foreground text-lg text-pretty md:text-xl">
              Consultá infracciones de tránsito en Uruguay. Buscá por
              descripción, vehículo, ubicación y más.
            </p>
          </div>

          {/* CTA */}
          <div className="flex flex-wrap items-center justify-center gap-4 pt-4">
            <Link href="/offenses">
              <Button size="lg" className="px-8 py-6 text-lg">
                Consultar Infracciones
              </Button>
            </Link>
            <Link href="/docs">
              <Button size="lg" variant="outline" className="px-8 py-6 text-lg">
                Ver Documentación
              </Button>
            </Link>
          </div>

          {/* Features */}
          <div className="mt-12 grid gap-6 md:grid-cols-3">
            <div className="bg-card border-border flex flex-col items-center gap-3 rounded-lg border p-6">
              <Search className="text-primary h-8 w-8" />
              <h3 className="text-foreground font-semibold">
                Búsqueda Avanzada
              </h3>
              <p className="text-muted-foreground text-sm text-pretty">
                Filtrá por múltiples criterios para encontrar exactamente lo que
                necesitás
              </p>
            </div>

            <div className="bg-card border-border flex flex-col items-center gap-3 rounded-lg border p-6">
              <Shield className="text-primary h-8 w-8" />
              <h3 className="text-foreground font-semibold">Datos Oficiales</h3>
              <p className="text-muted-foreground text-sm text-pretty">
                Información de bases de datos oficiales de tránsito
              </p>
            </div>

            <div className="bg-card border-border flex flex-col items-center gap-3 rounded-lg border p-6">
              <FileText className="text-primary h-8 w-8" />
              <h3 className="text-foreground font-semibold">
                Detalles Completos
              </h3>
              <p className="text-muted-foreground text-sm text-pretty">
                Accedé a toda la información disponible de cada infracción
              </p>
            </div>
          </div>
        </div>
      </main>

      {/* Footer */}
      <footer className="border-border bg-card border-t py-6">
        <div className="text-muted-foreground container mx-auto px-6 text-center text-sm">
          <p>ChapaUY - Infracciones de Tránsito</p>
        </div>
      </footer>
    </div>
  )
}
