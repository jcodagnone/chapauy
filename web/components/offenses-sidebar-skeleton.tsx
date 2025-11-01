/**
 * Copyright 2025 The ChapaUY Authors
 * SPDX-License-Identifier: Apache-2.0
 */

import { Skeleton } from "@/components/ui/skeleton"
import Link from "next/link"
import { FacetFilterSkeleton } from "@/components/facet-filter-skeleton"

export function OffensesSidebarSkeleton() {
  return (
    <aside className="border-border bg-card h-full w-64 overflow-y-auto border-r p-6 print:hidden">
      <div className="mb-6">
        <Link href="/" className="block transition-opacity hover:opacity-80">
          <h1 className="text-foreground text-xl font-semibold">ChapaUY</h1>
          <p className="text-muted-foreground mt-1 text-sm">
            Infracciones de tránsito
          </p>
        </Link>
      </div>

      <div>
        {Array.from({ length: 4 }).map((_, i) => (
          <FacetFilterSkeleton key={i} />
        ))}
      </div>
    </aside>
  )
}
