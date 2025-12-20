/**
 * Copyright 2025 The ChapaUY Authors
 * SPDX-License-Identifier: Apache-2.0
 */

"use client"

import { useState, useTransition } from "react"
import Link from "next/link"
import { useRouter } from "next/navigation"
import { Plus } from "lucide-react"
import { Dimension, Facet } from "@/lib/types"
import { getDimensionConfig } from "@/lib/display-config"
import { FacetFilter } from "@/components/facet-filter"
import { FacetFilterSkeleton } from "@/components/facet-filter-skeleton"

interface SidebarItemProps {
  dimension: Dimension
  facet?: Facet
  selectedValues: string[]
  rawParams: Record<string, string | string[] | undefined>
}

export function SidebarItem({
  dimension,
  facet,
  selectedValues,
  rawParams,
}: SidebarItemProps) {
  const router = useRouter()
  const [isPending, startTransition] = useTransition()
  const [isOptimisticLoading, setIsOptimisticLoading] = useState(false)

  // Reset optimistic loading when facet data arrives
  if (facet && isOptimisticLoading) {
    setIsOptimisticLoading(false)
  }

  // Helper to build URL for adding a facet
  const getAddFacetUrl = () => {
    const p = new URLSearchParams()
    Object.entries(rawParams).forEach(([key, value]) => {
      if (Array.isArray(value)) {
        value.forEach((v) => p.append(key, v))
      } else if (value) {
        p.set(key, value)
      }
    })
    p.append("facet", dimension)
    p.sort()
    return `?${p.toString()}`
  }

  const handleAddClick = (e: React.MouseEvent) => {
    e.preventDefault()
    setIsOptimisticLoading(true)
    startTransition(() => {
      router.push(getAddFacetUrl())
    })
  }

  // If we have the facet data, render the filter
  if (facet) {
    // Reconstruct params for the filter component usage
    const currentFilters = new URLSearchParams()
    Object.entries(rawParams).forEach(([key, value]) => {
      if (Array.isArray(value)) {
        value.forEach((v) => currentFilters.append(key, v))
      } else if (value) {
        currentFilters.set(key, value)
      }
    })

    return (
      <FacetFilter
        dimension={facet.dimension}
        facets={facet.values}
        selectedValues={selectedValues}
        currentFilters={currentFilters}
        totalValues={facet.total_values}
      />
    )
  }

  // If we are loading (optimistic or transition), show skeleton
  if (isOptimisticLoading || isPending) {
    return <FacetFilterSkeleton />
  }

  // Otherwise, render the "Add" button
  return (
    <div className="border-border border-b pb-3">
      <Link
        href={getAddFacetUrl()}
        prefetch={false}
        onClick={handleAddClick}
        data-no-close="true"
        className="text-muted-foreground hover:text-foreground flex w-full items-center justify-between transition-colors"
      >
        <span className="flex items-center gap-1.5 text-sm">
          <FacetIcon dimension={dimension} />
          {getDimensionConfig(dimension).label}
        </span>
        <Plus className="text-primary h-3 w-3" />
      </Link>
    </div>
  )
}

function FacetIcon({ dimension }: { dimension: Dimension }) {
  const config = getDimensionConfig(dimension)
  const Icon = config.icon
  return <Icon className="h-3.5 w-3.5" />
}
