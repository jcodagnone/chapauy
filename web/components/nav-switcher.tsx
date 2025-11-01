/**
 * Copyright 2025 The ChapaUY Authors
 * SPDX-License-Identifier: Apache-2.0
 */

"use client"

import Link from "next/link"
import { usePathname, useSearchParams } from "next/navigation"
import { cn } from "@/lib/utils"
import { Dimension } from "@/lib/types"

export function NavSwitcher() {
  const pathname = usePathname()
  const searchParams = useSearchParams()
  const isDocuments = pathname.startsWith("/documents")

  const getPreservedParams = () => {
    const newParams = new URLSearchParams()

    // Common dimensions to preserve
    const commonDimensions = [
      Dimension.Year,
      Dimension.Database,
      Dimension.Features,
    ]

    commonDimensions.forEach((dim) => {
      // Preserve filters
      const values = searchParams.getAll(dim)
      values.forEach((val) => newParams.append(dim, val))
    })

    // Preserve open facets state for common dimensions
    const openFacets = searchParams.getAll("facet")
    openFacets.forEach((facet) => {
      if (commonDimensions.includes(facet as Dimension)) {
        newParams.append("facet", facet)
      }
    })

    const str = newParams.toString()
    return str ? `?${str}` : ""
  }

  const queryString = getPreservedParams()

  return (
    <div className="bg-muted mb-4 grid grid-cols-2 gap-1 rounded-lg p-1">
      <Link
        href={`/offenses${queryString}`}
        className={cn(
          "flex items-center justify-center rounded-md px-3 py-1.5 text-sm font-medium transition-all",
          !isDocuments
            ? "bg-background text-foreground shadow-sm"
            : "text-muted-foreground hover:text-foreground hover:bg-background/50"
        )}
      >
        Infracciones
      </Link>
      <Link
        href={`/documents${queryString}`}
        className={cn(
          "flex items-center justify-center rounded-md px-3 py-1.5 text-sm font-medium transition-all",
          isDocuments
            ? "bg-background text-foreground shadow-sm"
            : "text-muted-foreground hover:text-foreground hover:bg-background/50"
        )}
      >
        Documentos
      </Link>
    </div>
  )
}
