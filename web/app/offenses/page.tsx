/**
 * Copyright 2025 The ChapaUY Authors
 * SPDX-License-Identifier: Apache-2.0
 */

import { OffensesSidebarClient } from "@/components/offenses-sidebar-client"
import { OffensesFeedClient } from "./offenses-feed-client"

// Force static generation - skeletons render at build time
// Force static generation - skeletons render at build time
// export const dynamic = "force-static"

export default function OffensesPage() {
  return (
    <div className="bg-background flex min-h-screen">
      <OffensesSidebarClient />

      <main className="flex-1 print:w-full">
        <OffensesFeedClient />
      </main>
    </div>
  )
}
