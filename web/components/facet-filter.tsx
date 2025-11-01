/**
 * Copyright 2025 The ChapaUY Authors
 * SPDX-License-Identifier: Apache-2.0
 */

"use client"
import Link from "next/link"
import { Trash2 } from "lucide-react"
import type { FacetValue } from "@/lib/types"
import { FilterSearch } from "./filter-search"
import { getDimensionConfig } from "@/lib/display-config"
import { FacetItem } from "./facet-item"
import { useState, useEffect } from "react"
import { useRouter } from "next/navigation"
import { buildUrlToggleFilter } from "@/lib/url-utils"

interface FacetFilterProps {
  title?: string
  dimension: string
  facets: FacetValue[]
  selectedValues: string[]
  onSelect?: (value: string | null) => void
  currentFilters: URLSearchParams
  totalValues?: number
}

export function FacetFilter({
  title,
  dimension,
  facets,
  selectedValues,
  onSelect,
  currentFilters,
  totalValues = 0,
}: FacetFilterProps) {
  const router = useRouter()
  const [displayTotalValues, setDisplayTotalValues] = useState(totalValues)
  const [displayShownCount, setDisplayShownCount] = useState(facets.length)

  const handleSelect = (value: string | null) => {
    if (onSelect) {
      onSelect(value)
    } else if (value) {
      // Default navigation logic
      const newUrl = buildUrlToggleFilter(currentFilters, dimension, value)
      router.push(newUrl)
    }
  }

  useEffect(() => {
    setDisplayTotalValues(totalValues)
    setDisplayShownCount(facets.length)
  }, [totalValues, facets.length])

  const config = getDimensionConfig(dimension)
  const IconComponent = config.icon
  const displayTitle = title || config.label

  const shownCount = displayShownCount
  // const hiddenCount = displayTotalValues - shownCount; // Unused
  const hasMore = totalValues > 10

  // Calculate URL for removing the entire dimension
  const getRemoveUrl = () => {
    const params = new URLSearchParams(currentFilters.toString())
    params.delete(dimension)
    const currentFacets = params.getAll("facet")
    const newFacets = currentFacets.filter((f) => f !== dimension)
    params.delete("facet")
    newFacets.forEach((f) => params.append("facet", f))
    // Reset page is generally good practice when filters change significantly
    params.delete("page")
    const search = params.toString()
    return (search ? `?${search}` : "?") + `#facet-${dimension}`
  }

  return (
    <div id={`facet-${dimension}`} className="border-border mb-4 border-b pb-4">
      <Link
        href={getRemoveUrl()}
        className="text-foreground hover:text-destructive mb-3 flex w-full items-center justify-between text-sm font-medium transition-colors"
        title="Eliminar dimensión"
      >
        <div className="flex items-center gap-1.5">
          <IconComponent className="h-4 w-4" />
          <span>{displayTitle}</span>
        </div>
        <Trash2 className="h-3.5 w-3.5" />
      </Link>

      <div className="space-y-2">
        {hasMore && (
          <FilterSearch
            dimension={dimension}
            currentFilters={currentFilters}
            selectedValues={selectedValues}
            onSelect={(value) => {
              handleSelect(value)
            }}
            placeholder={`Buscar ${title?.toLowerCase() || config.label.toLowerCase()}...`}
            initialValues={facets}
            onTotalValuesChange={(newTotal) => {
              setDisplayTotalValues(
                newTotal > 0 ? newTotal : displayTotalValues
              )
            }}
            onSuggestionsChange={(suggestions) => {
              setDisplayShownCount(suggestions.length)
            }}
          />
        )}

        {!hasMore && (
          <div className="space-y-1">
            {facets.map((facet) => {
              const isSelected = selectedValues.includes(facet.value)
              const totalCount = facets.reduce((sum, f) => sum + f.count, 0)
              const showBar = facets.length > 1
              const href =
                buildUrlToggleFilter(currentFilters, dimension, facet.value) +
                `#facet-${dimension}`

              return (
                <FacetItem
                  key={facet.value}
                  facet={facet}
                  selected={isSelected}
                  href={href}
                  dimension={dimension}
                  totalCount={totalCount}
                  showBar={showBar}
                />
              )
            })}
          </div>
        )}

        {hasMore && (
          <div className="text-muted-foreground mt-2 text-xs">
            Mostrando {shownCount} de {displayTotalValues.toLocaleString()}{" "}
            valores
          </div>
        )}
      </div>
    </div>
  )
}
