/**
 * Copyright 2025 The ChapaUY Authors
 * SPDX-License-Identifier: Apache-2.0
 */

import Image from "next/image"
import Link from "next/link"
import { Button } from "@/components/ui/button"
import { GlobalLinks } from "@/components/global-links"
import { HomeSearch } from "@/components/home-search"


export default function Home() {
  return (
    <div className="bg-background flex min-h-screen flex-col">
      {/* Header */}
      <header className="absolute top-0 right-0 left-0 z-10">
        <div className="container mx-auto flex items-center justify-end px-6 py-4">
          <GlobalLinks includeDocsLink={false} />
        </div>
      </header>

      {/* Main Content */}
      <main className="flex flex-1 flex-col items-center justify-center px-4">
        <div className="flex w-full flex-col items-center gap-8">
          {/* Logo */}
          <div className="flex flex-col items-center gap-4 text-center">
            <div className="relative h-24 w-full max-w-[300px]">
              <Image
                src="/logo.webp"
                alt="ChapaUY Logo"
                fill
                className="object-contain"
                priority
              />
            </div>
            <p className="text-muted-foreground text-lg">
              Consultá infracciones de tránsito en Uruguay.
            </p>
          </div>

          {/* Search Component */}
          <HomeSearch />

          {/* Action Links */}
          <div className="flex flex-wrap items-center justify-center gap-4 pt-4">
            <Button variant="secondary" asChild>
              <Link href="/offenses">
                Explorar todas las infracciones
              </Link>
            </Button>
            <Button variant="ghost" className="text-muted-foreground" asChild>
              <Link href="/docs">
                Ver Documentación
              </Link>
            </Button>
          </div>
        </div>
      </main>

      {/* Footer */}
      <footer className="bg-background py-4">
        <div className="text-muted-foreground container mx-auto px-6 text-center text-sm">
          <p>ChapaUY - Infracciones de Tránsito en Uruguay</p>
        </div>
      </footer>
    </div>
  )
}
