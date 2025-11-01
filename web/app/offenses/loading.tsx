/**
 * Copyright 2025 The ChapaUY Authors
 * SPDX-License-Identifier: Apache-2.0
 */

import { Skeleton } from "@/components/ui/skeleton"
import { OffenseCardSkeleton } from "@/components/offense-card-skeleton"

export default function OffensesLoading() {
  return (
    <div className="border-card h-full p-8 print:p-2">
      {/* Top Bar Skeleton */}
      <div className="mb-4">
        <div className="flex items-start justify-between gap-3">
          <Skeleton className="h-5 w-48" />
          <div className="flex items-center gap-4">
            <Skeleton className="h-8 w-64 rounded-lg" /> {/* View Toggles */}
            <div className="flex flex-col items-end gap-1">
              <Skeleton className="h-4 w-24" />
              <Skeleton className="h-3 w-20" />
            </div>
          </div>
        </div>
      </div>

      {/* List/Grid Skeleton */}
      <div className="space-y-4">
        {Array.from({ length: 6 }).map((_, i) => (
          <OffenseCardSkeleton key={i} />
        ))}
      </div>
    </div>
  )
}
