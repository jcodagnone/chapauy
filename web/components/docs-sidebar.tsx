/**
 * Copyright 2025 The ChapaUY Authors
 * SPDX-License-Identifier: Apache-2.0
 */

"use client"

import Link from "next/link"
import { usePathname } from "next/navigation"
import { cn } from "@/lib/utils"
import type { DocNode } from "@/lib/docs"
import { Book, FileText, ChevronRight } from "lucide-react"
import { GlobalLinks } from "@/components/global-links"

interface DocsSidebarProps {
  items: DocNode[]
}

export function DocsSidebar({ items }: DocsSidebarProps) {
  const pathname = usePathname()

  return (
    <div className="bg-card border-border flex h-full w-full flex-col border-r">
      <div className="flex-1 overflow-y-auto p-6">
        <div className="mb-6">
          <Link href="/" className="block transition-opacity hover:opacity-80">
            <h1 className="text-foreground text-xl font-semibold">ChapaUY</h1>
            <p className="text-muted-foreground mt-1 text-sm">
              Infracciones de tránsito
            </p>
          </Link>
        </div>

        <div className="space-y-6">
          <SidebarItems items={items} pathname={pathname} />
        </div>
      </div>
      <div className="border-border border-t p-4">
        <GlobalLinks className="justify-center" includeDocsLink={false} />
      </div>
    </div>
  )
}

function SidebarItems({
  items,
  pathname,
}: {
  items: DocNode[]
  pathname: string
}) {
  return (
    <div className="flex flex-col gap-4">
      {items.map((item, index) => {
        const isActive = item.slug === pathname

        if (item.children) {
          return (
            <div key={index}>
              <div className="text-foreground mb-2 flex items-center gap-2 text-sm font-semibold">
                <Book className="h-4 w-4" />
                {item.title}
              </div>
              <div className="border-border/50 ml-2 space-y-1 border-l pl-4">
                <SidebarItems items={item.children} pathname={pathname} />
              </div>
            </div>
          )
        }

        return (
          <div
            key={index}
            className="border-border border-b pb-3 last:border-0"
          >
            <Link
              href={item.slug || "#"}
              className={cn(
                "group flex w-full items-center justify-between text-sm transition-colors",
                isActive
                  ? "text-foreground font-medium"
                  : "text-muted-foreground hover:text-foreground"
              )}
            >
              <span className="flex items-center gap-2">
                <FileText className="h-3.5 w-3.5" />
                {item.title}
              </span>
              {isActive && (
                <ChevronRight className="text-muted-foreground h-3.5 w-3.5" />
              )}
            </Link>
          </div>
        )
      })}
    </div>
  )
}
