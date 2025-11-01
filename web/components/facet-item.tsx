/**
 * Copyright 2025 The ChapaUY Authors
 * SPDX-License-Identifier: Apache-2.0
 */

import Link from "next/link"
import { Check } from "lucide-react"
import type { FacetValue } from "@/lib/types"
import { getDimensionConfig } from "@/lib/display-config"

interface FacetItemProps {
  facet: FacetValue
  selected: boolean
  onSelect?: (value: string) => void
  href?: string
  dimension: string
  totalCount: number
  showBar?: boolean
}

export function FacetItem({
  facet,
  selected,
  onSelect,
  href,
  dimension,
  totalCount,
  showBar = true,
}: FacetItemProps) {
  const displayText = facet.label || facet.value
  const isEmpty = facet.value === ""
  const percentage = totalCount > 0 ? (facet.count / totalCount) * 100 : 0

  const content = (
    <>
      {showBar && (
        <div
          className="bg-primary/10 absolute inset-y-0 right-0 rounded transition-all"
          style={{ width: `${percentage}%` }}
        />
      )}
      <div className="relative z-10 flex flex-1 items-center gap-2 truncate">
        <div className="flex h-4 w-4 flex-shrink-0 items-center justify-center">
          {selected && <Check className="h-3 w-3" />}
        </div>
        {isEmpty ? (
          <span className="text-muted-foreground bg-muted rounded px-2 py-0.5 text-xs italic">
            {getDimensionConfig(dimension).empty}
          </span>
        ) : (
          <span className="truncate">{displayText}</span>
        )}
      </div>
      <div className="relative z-10 flex flex-shrink-0 items-center gap-2">
        <span className="text-xs">
          {selected && facet.count === 0 ? "—" : facet.count.toLocaleString()}
        </span>
      </div>
    </>
  )

  const className = `group relative flex w-full items-center justify-between rounded px-2 py-1.5 text-left text-sm transition-colors ${
    selected
      ? "bg-accent text-accent-foreground"
      : "text-muted-foreground hover:text-foreground hover:bg-accent/50"
  }`

  const title = `${isEmpty ? getDimensionConfig(dimension).empty : displayText} - ${percentage.toFixed(2)}%`

  if (href) {
    return (
      <Link href={href} className={className} title={title}>
        {content}
      </Link>
    )
  }

  return (
    <button
      onClick={() => onSelect?.(facet.value)}
      className={className}
      title={title}
    >
      {content}
    </button>
  )
}
