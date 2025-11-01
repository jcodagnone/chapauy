/**
 * Copyright 2025 The ChapaUY Authors
 * SPDX-License-Identifier: Apache-2.0
 */

import { Card } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"

export function OffenseCardSkeleton() {
  return (
    <Card className="border-border p-3 print:hidden">
      <div className="space-y-1">
        <div className="flex items-start justify-between gap-3">
          <div className="flex flex-1 items-center gap-2">
            {/* Title skeleton */}
            <Skeleton className="h-5 flex-1" />
            {/* Repo badge skeleton */}
            <Skeleton className="h-5 w-16 flex-shrink-0" />
          </div>
          {/* UR value skeleton */}
          <Skeleton className="h-5 w-16 flex-shrink-0" />
        </div>

        {/* Metadata line skeleton */}
        <div className="flex flex-wrap items-center gap-1.5">
          <Skeleton className="h-3 w-24" />
          <Skeleton className="h-3 w-1" />
          <Skeleton className="h-3 w-20" />
          <Skeleton className="h-3 w-1" />
          <Skeleton className="h-3 w-16" />
          <Skeleton className="h-3 w-12" />
          <Skeleton className="h-3 w-1" />
          <Skeleton className="h-3 w-20" />
        </div>
      </div>
    </Card>
  )
}
