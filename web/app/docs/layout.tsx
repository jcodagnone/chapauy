/**
 * Copyright 2025 The ChapaUY Authors
 * SPDX-License-Identifier: Apache-2.0
 */

import { DocsSidebar } from "@/components/docs-sidebar"
import { getDocPages } from "@/lib/docs"
import { Sheet, SheetContent, SheetTrigger } from "@/components/ui/sheet"
import { Button } from "@/components/ui/button"
import { Menu } from "lucide-react"
import { ScrollArea } from "@/components/ui/scroll-area"

export default async function DocsLayout({
  children,
}: {
  children: React.ReactNode
}) {
  const items = await getDocPages()

  return (
    <div className="flex min-h-screen flex-col">
      <header className="bg-background/95 supports-[backdrop-filter]:bg-background/60 sticky top-0 z-40 w-full border-b backdrop-blur lg:hidden">
        <div className="container flex h-14 items-center">
          <Sheet>
            <SheetTrigger asChild>
              <Button
                variant="ghost"
                size="icon"
                className="mr-2 px-0 text-base hover:bg-transparent focus-visible:bg-transparent focus-visible:ring-0 focus-visible:ring-offset-0 lg:hidden"
              >
                <Menu className="h-6 w-6" />
                <span className="sr-only">Toggle Menu</span>
              </Button>
            </SheetTrigger>
            <SheetContent side="left" className="pr-0">
              <DocsSidebar items={items} />
            </SheetContent>
          </Sheet>
          <span className="font-bold">ChapaUY Documentación</span>
        </div>
      </header>

      <div className="flex h-[calc(100vh-3.5rem)]">
        <aside className="hidden h-full w-64 shrink-0 lg:block">
          <DocsSidebar items={items} />
        </aside>
        <main className="bg-background flex-1 overflow-y-auto pl-8">
          <div className="container max-w-5xl py-6 lg:py-10">{children}</div>
        </main>
      </div>
    </div>
  )
}
