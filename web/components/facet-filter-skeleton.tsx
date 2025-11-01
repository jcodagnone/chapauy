/**
 * Copyright 2025 The ChapaUY Authors
 * SPDX-License-Identifier: Apache-2.0
 */

import { Skeleton } from "@/components/ui/skeleton"

export function FacetFilterSkeleton() {
  return (
    <div className="border-border mb-4 border-b pb-4">
      {/* Header */}
      <div className="mb-3 flex items-center justify-between">
        <div className="flex items-center gap-1.5">
          <Skeleton className="h-4 w-4" />
          <Skeleton className="h-4 w-24" />
        </div>
        <Skeleton className="h-3.5 w-3.5" />
      </div>

      {/* Facet items */}
      <div className="space-y-1">
        {Array.from({ length: 5 }).map((_, i) => (
          <div
            key={i}
            className="flex items-center justify-between rounded px-2 py-1.5"
          >
            <div className="flex flex-1 items-center gap-2">
              <Skeleton className="h-4 w-4 flex-shrink-0" />
              <Skeleton
                className="h-4 flex-1"
                style={{ width: `${60 + ((i * 13) % 30)}%` }}
              />
            </div>
            <Skeleton className="h-3 w-8" />
          </div>
        ))}
      </div>
    </div>
  )
}
