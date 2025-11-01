/**
 * Copyright 2025 The ChapaUY Authors
 * SPDX-License-Identifier: Apache-2.0
 */

"use client"

import { useSearchParams, useRouter, usePathname } from "next/navigation"
import { useCallback, useMemo } from "react"
import { type ActiveFilter, type OffensesSummaryResponse } from "@/lib/types"
import { offensesParamsFromQueryParams } from "@/lib/url-utils"

export function useOffenseSearchParams() {
  const router = useRouter()
  const pathname = usePathname()
  const searchParams = useSearchParams()

  const updateURL = useCallback(
    (params: Record<string, string | string[] | null>) => {
      const current = new URLSearchParams(window.location.search)

      Object.entries(params).forEach(([key, value]) => {
        if (value === null) {
          current.delete(key)
        } else if (value === "") {
          const existingValues = current.getAll(key)
          if (existingValues.includes("")) {
            current.delete(key)
            existingValues
              .filter((v) => v !== "")
              .forEach((v) => current.append(key, v))
          } else if (existingValues.length > 0) {
            current.append(key, "")
          } else {
            current.set(key, "")
          }
        } else if (Array.isArray(value)) {
          current.delete(key)
          value.forEach((v) => current.append(key, v))
        } else {
          const existingValues = current.getAll(key)
          if (existingValues.includes(value)) {
            current.delete(key)
            existingValues
              .filter((v) => v !== value)
              .forEach((v) => current.append(key, v))
          } else if (existingValues.length > 0) {
            current.append(key, value)
          } else {
            current.set(key, value)
          }
        }
      })

      const search = current.toString()
      const query = search ? `?${search}` : ""

      router.push(`${pathname}${query}`, { scroll: false })
    },
    [pathname, router]
  )

  const addFilter = useCallback(
    (dimension: string, value: string) => {
      const current = new URLSearchParams(window.location.search)
      const existingValues = current.getAll(dimension)

      if (!existingValues.includes(value)) {
        current.append(dimension, value)
        const search = current.toString()
        const query = search ? `?${search}` : ""
        router.push(`${pathname}${query}`, { scroll: false })
      }
    },
    [pathname, router]
  )

  const removeFilter = useCallback(
    (dimension: string, value: string) => {
      const currentValues = searchParams.getAll(dimension)
      const newValues = currentValues.filter((v) => v !== value)

      if (newValues.length === 0) {
        updateURL({ [dimension]: null })
      } else {
        updateURL({ [dimension]: newValues })
      }
    },
    [searchParams, updateURL]
  )

  const clearAllFilters = useCallback(() => {
    const newParams = new URLSearchParams()

    searchParams.getAll("facet").forEach((facet) => {
      newParams.append("facet", facet)
    })

    const query = newParams.toString() ? `?${newParams.toString()}` : ""
    router.push(`${pathname}${query}`, { scroll: false })
  }, [searchParams, pathname, router])

  const setViewMode = useCallback(
    (mode: "list" | "charts" | "map") => {
      const current = new URLSearchParams(window.location.search)
      if (mode === "charts") {
        current.set("view", "charts")
      } else if (mode === "map") {
        current.set("view", "map")
      } else {
        current.delete("view")
      }
      const search = current.toString()
      const query = search ? `?${search}` : ""
      router.push(`${pathname}${query}`, { scroll: false })
    },
    [pathname, router]
  )

  const setGroupBy = useCallback(
    (value: string) => {
      const current = new URLSearchParams(window.location.search)
      if (value) {
        current.set("group_by", value)
      } else {
        current.delete("group_by")
      }
      const search = current.toString()
      const query = search ? `?${search}` : ""
      router.push(`${pathname}${query}`, { scroll: false })
    },
    [pathname, router]
  )

  const getActiveFilters = useCallback(
    (summaryData: OffensesSummaryResponse | null): ActiveFilter[] => {
      if (!summaryData) return []

      const params = offensesParamsFromQueryParams(searchParams)

      return params.predicates
        .filter((predicate) => predicate.values.length > 0)
        .map((predicate) => {
          const facet = summaryData.facets.find(
            (f) => f.dimension === predicate.dimension
          )
          return {
            dimension: predicate.dimension,
            values: predicate.values.map((value) => {
              const facetValue = facet?.values.find((fv) => fv.value === value)
              return {
                value,
                label: facetValue?.label,
              }
            }),
          }
        })
    },
    [searchParams]
  )

  const viewMode = useMemo(() => {
    const view = searchParams.get("view")
    if (view === "charts") return "charts"
    if (view === "map") return "map"
    return "list"
  }, [searchParams])

  const groupBy = useMemo(() => {
    return searchParams.get("group_by") || ""
  }, [searchParams])

  const applyUpdates = useCallback(
    (updates: (params: URLSearchParams) => void) => {
      const current = new URLSearchParams(window.location.search)
      updates(current)
      const search = current.toString()
      const query = search ? `?${search}` : ""
      router.push(`${pathname}${query}`, { scroll: false })
    },
    [pathname, router]
  )

  return {
    searchParams,
    updateURL,
    addFilter,
    removeFilter,
    clearAllFilters,
    setViewMode,
    setGroupBy,
    getActiveFilters,
    viewMode,
    groupBy,
    applyUpdates,
  }
}
