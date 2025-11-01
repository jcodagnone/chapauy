/**
 * Copyright 2025 The ChapaUY Authors
 * SPDX-License-Identifier: Apache-2.0
 */

"use client"

import { useState, useEffect } from "react"
import { Search, Loader2, X } from "lucide-react"
import { cn } from "@/lib/utils"
import type { Facet, FacetValue } from "@/lib/types"
import { getDimensionConfig } from "@/lib/display-config"
import { FacetItem } from "./facet-item"

interface FilterSearchProps {
  dimension: string
  currentFilters: URLSearchParams
  selectedValues?: string[]
  onSelect: (value: string) => void
  placeholder?: string
  initialValues?: FacetValue[]
  onTotalValuesChange?: (totalValues: number) => void
  onSuggestionsChange?: (suggestions: FacetValue[]) => void
}

export function FilterSearch({
  dimension,
  currentFilters,
  selectedValues = [],
  onSelect,
  placeholder,
  initialValues = [],
  onTotalValuesChange,
  onSuggestionsChange,
}: FilterSearchProps) {
  const [query, setQuery] = useState("")
  const [suggestions, setSuggestions] = useState<FacetValue[]>(initialValues)
  const [isLoading, setIsLoading] = useState(false)

  useEffect(() => {
    const fetchSuggestions = async () => {
      if (query.length === 0) {
        setSuggestions(initialValues)
        onTotalValuesChange?.(0)
        onSuggestionsChange?.(initialValues)
        return
      }

      setIsLoading(true)
      try {
        const params = new URLSearchParams(currentFilters)
        params.set("dimension", dimension)
        params.set("q", query)

        const response = await fetch(`/api/v1/suggest?${params.toString()}`)
        const data: Facet = await response.json()
        const newSuggestions = data.values || []
        setSuggestions(newSuggestions)
        onTotalValuesChange?.(data.total_values || 0)
        onSuggestionsChange?.(newSuggestions)
      } catch (error) {
        console.error("[v0] Error fetching suggestions:", error)
        setSuggestions([])
        onTotalValuesChange?.(0)
        onSuggestionsChange?.([])
      } finally {
        setIsLoading(false)
      }
    }

    const timeoutId = setTimeout(fetchSuggestions, 300)
    return () => clearTimeout(timeoutId)
  }, [
    query,
    dimension,
    currentFilters,
    initialValues,
    onTotalValuesChange,
    onSuggestionsChange,
  ])

  return (
    <div className="space-y-2">
      <div className="relative">
        <Search className="text-muted-foreground absolute top-2.5 left-2 h-4 w-4" />
        <input
          type="text"
          placeholder={placeholder || "Buscar..."}
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          className={cn(
            "file:text-foreground placeholder:text-muted-foreground selection:bg-primary selection:text-primary-foreground dark:bg-input/30 border-input h-9 w-full min-w-0 rounded-md border bg-transparent px-3 py-1 text-base shadow-xs transition-[color,box-shadow] outline-none file:inline-flex file:h-7 file:border-0 file:bg-transparent file:text-sm file:font-medium disabled:pointer-events-none disabled:cursor-not-allowed disabled:opacity-50 md:text-sm",
            "focus-visible:border-ring focus-visible:ring-ring/50 focus-visible:ring-[3px]",
            "aria-invalid:ring-destructive/20 dark:aria-invalid:ring-destructive/40 aria-invalid:border-destructive",
            query.length > 0 ? "h-9 pr-8 pl-8 text-sm" : "h-9 pl-8 text-sm"
          )}
        />
        {query.length > 0 && !isLoading && (
          <button
            onClick={() => setQuery("")}
            className="text-muted-foreground hover:text-foreground absolute top-2.5 right-2 h-4 w-4 transition-colors"
            title="Limpiar búsqueda"
          >
            <X className="h-4 w-4" />
          </button>
        )}
        {isLoading && (
          <Loader2 className="text-muted-foreground absolute top-2.5 right-2 h-4 w-4 animate-spin" />
        )}
      </div>

      {suggestions.length > 0 && (
        <div className="max-h-64 space-y-1 overflow-y-auto">
          {suggestions.map((suggestion) => {
            const isSelected = selectedValues.includes(suggestion.value)
            const totalCount = suggestions.reduce((sum, s) => sum + s.count, 0)
            const showBar = suggestions.length > 1

            return (
              <FacetItem
                key={suggestion.value}
                facet={suggestion}
                selected={isSelected}
                onSelect={onSelect}
                dimension={dimension}
                totalCount={totalCount}
                showBar={showBar}
              />
            )
          })}
        </div>
      )}

      {query.length > 0 && !isLoading && suggestions.length === 0 && (
        <div className="text-muted-foreground px-2 py-1.5 text-sm">
          No se encontraron resultados
        </div>
      )}
    </div>
  )
}
