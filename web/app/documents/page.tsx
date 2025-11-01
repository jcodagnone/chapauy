/**
 * Copyright 2025 The ChapaUY Authors
 * SPDX-License-Identifier: Apache-2.0
 */

import { Suspense } from "react"
import { OffensesSidebarClient } from "@/components/offenses-sidebar-client"
import { OffensesSidebarSkeleton } from "@/components/offenses-sidebar-skeleton"
import { DocumentsList } from "./documents-list"
import DocumentsLoading from "./loading"
import { Dimension, SidebarMode } from "@/lib/types"

interface PageProps {
  searchParams: Promise<Record<string, string | string[] | undefined>>
}

export default function DocumentsPage({ searchParams }: PageProps) {
  return (
    <div className="bg-background flex min-h-screen">
      <OffensesSidebarClient
        visibleDimensions={[
          Dimension.Database,
          Dimension.Year,
          Dimension.Features,
        ]}
        mode={SidebarMode.Documents}
      />

      <main className="flex-1 print:w-full">
        <Suspense fallback={<DocumentsLoading />}>
          <DocumentsList searchParams={searchParams} />
        </Suspense>
      </main>
    </div>
  )
}
