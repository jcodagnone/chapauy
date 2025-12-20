/**
 * Copyright 2025 The ChapaUY Authors
 * SPDX-License-Identifier: Apache-2.0
 */

import { OffensesSidebarClient } from "@/components/offenses-sidebar-client"
import { MobileNav } from "@/components/mobile-nav"
import { OffensesFeedClient } from "./offenses-feed-client"

// Force static generation - skeletons render at build time
// Force static generation - skeletons render at build time
// export const dynamic = "force-static"

export default function OffensesPage() {
  return (
    <div className="bg-background flex flex-col min-h-screen md:flex-row">
      <MobileNav />
      <OffensesSidebarClient className="hidden md:flex" />

      <main className="flex-1 w-full print:w-full min-w-0">
        <OffensesFeedClient />
      </main>
    </div>
  )
}
