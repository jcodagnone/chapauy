/**
 * Copyright 2025 The ChapaUY Authors
 * SPDX-License-Identifier: Apache-2.0
 */

import { Skeleton } from "@/components/ui/skeleton"
import { OffenseCardSkeleton } from "@/components/offense-card-skeleton"

export default function DocumentsLoading() {
  return (
    <div className="bg-background flex min-h-screen">
      {/* Sidebar Skeleton */}
      <aside className="border-border bg-card hidden w-64 border-r p-6 lg:block">
        <div className="mb-8 space-y-2">
          <Skeleton className="h-8 w-32" />
          <Skeleton className="h-4 w-24" />
        </div>
        <div className="space-y-6">
          {Array.from({ length: 3 }).map((_, i) => (
            <div key={i} className="space-y-2">
              <Skeleton className="h-5 w-20" />
              <div className="space-y-1">
                <Skeleton className="h-6 w-full" />
                <Skeleton className="h-6 w-3/4" />
              </div>
            </div>
          ))}
        </div>
      </aside>

      <main className="flex-1">
        <div className="border-card h-full p-8 print:p-2">
          <div className="mb-6 space-y-2">
            <Skeleton className="h-5 w-48" />
          </div>

          <div className="space-y-4">
            {Array.from({ length: 6 }).map((_, i) => (
              <OffenseCardSkeleton key={i} />
            ))}
          </div>
        </div>
      </main>
    </div>
  )
}
