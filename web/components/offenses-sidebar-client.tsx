/**
 * Copyright 2025 The ChapaUY Authors
 * SPDX-License-Identifier: Apache-2.0
 */

"use client"

import { useEffect, useState, useMemo } from "react"
import { useSearchParams } from "next/navigation"
import Link from "next/link"
import { Dimension, SidebarMode, Facet } from "@/lib/types"
import { SidebarItem } from "@/components/sidebar-item"
import { NavSwitcher } from "@/components/nav-switcher"
import { GlobalLinks } from "@/components/global-links"
import { OffensesSidebarSkeleton } from "@/components/offenses-sidebar-skeleton"
import { buildApiUrl } from "@/lib/api-utils"

interface OffensesSidebarProps {
    visibleDimensions?: Dimension[]
    mode?: SidebarMode
    className?: string
    onClose?: () => void
}

const allDimensions: Dimension[] = [
    Dimension.Database,
    Dimension.Year,
    Dimension.Country,
    Dimension.VehicleType,
    Dimension.ArticleCode,
    Dimension.ArticleID,
    Dimension.Description,
    Dimension.Location,
    Dimension.Vehicle,
    Dimension.Features,
]

export function OffensesSidebarClient({
    visibleDimensions = allDimensions,
    mode = SidebarMode.Offenses,
    className,
    onClose,
}: OffensesSidebarProps) {
    const searchParams = useSearchParams()
    const [facets, setFacets] = useState<Facet[]>([])
    const [loading, setLoading] = useState(true)
    const [isUpdating, setIsUpdating] = useState(false)
    const [error, setError] = useState<string | null>(null)

    // Create a stable dependency that ignores map viewport parameters
    const sidebarQueryString = useMemo(() => {
        const params = new URLSearchParams()
        searchParams.forEach((value, key) => {
            if (key !== "lat" && key !== "lng" && key !== "zoom" && key !== "h3_index") {
                params.append(key, value)
            }
        })
        params.sort()
        return params.toString()
    }, [searchParams])

    useEffect(() => {
        const fetchFacets = async () => {
            // If we already have data, we are updating (not initial load)
            if (facets.length > 0) {
                setIsUpdating(true)
            }
            // Don't set loading(true) here to avoid full sidebar skeleton flash on updates
            // We rely on stale-while-revalidate UI + local optimistic feedback in SidebarItem
            try {
                // Build params from searchParams
                const params: Record<string, string | string[]> = {
                    mode: mode,
                }

                searchParams.forEach((value, key) => {
                    if (params[key]) {
                        if (Array.isArray(params[key])) {
                            ; (params[key] as string[]).push(value)
                        } else {
                            params[key] = [params[key] as string, value]
                        }
                    } else {
                        params[key] = value
                    }
                })

                // Also ensure facets parameter is passed if needed, but the API 
                // usually calculates facets based on predicates. 
                // The component logic in original server component checked `params.facets` 
                // but `OffensesSidebar` implementation seemed to fetch all dimensions 
                // requested by `visibleDimensions` or those in `facets` param?
                // Actually, the server component logic was:
                // const params = offensesParamsFromQueryParams(rawParams)
                // const facetsToFetch = params.facets || []  <-- This comes from URL param 'facets'
                // ...
                // facetResults = await getDimensionResults(params.predicates, facetsToFetch)
                // Wait, `facetsToFetch` being `params.facets` implies that if no facets param is present,
                // it fetches EMPTY list?
                // Let's re-read the original component:
                // Only explicitly request facets if they differ from the default set
                // This keeps the URL clean and relies on API defaults
                const defaultCountForMode = mode === SidebarMode.Documents ? 3 : 10
                if (visibleDimensions.length !== defaultCountForMode) {
                    params["facets"] = visibleDimensions
                }

                // Detect active dimensions from searchParams and ensure their data is fetched
                // This auto-expands the facet in the UI since the API will return data for it
                const activeDimensions = visibleDimensions.filter(dim => searchParams.has(dim))

                const facetParams = new Set<string>()

                // Add explicit 'facet' params from URL
                const updates = searchParams.getAll("facet")
                updates.forEach(f => facetParams.add(f))

                // Add active dimensions
                activeDimensions.forEach(dim => facetParams.add(dim))

                // Update params
                if (facetParams.size > 0) {
                    params["facet"] = Array.from(facetParams)
                }

                const url = buildApiUrl("/api/v1/sidebar", params)
                const res = await fetch(url)

                if (!res.ok) {
                    throw new Error(`HTTP error! status: ${res.status}`)
                }

                const data = await res.json()
                setFacets(data)
                setError(null)
            } catch (err) {
                console.error("Failed to load sidebar facets:", err)
                setError(err instanceof Error ? err.message : String(err))
            } finally {
                setLoading(false)
                setIsUpdating(false)
            }
        }

        fetchFacets()
    }, [sidebarQueryString, mode, visibleDimensions])

    // Helper to extract selected values for a dimension
    const getSelectedValues = (dimension: Dimension) => {
        const val = searchParams.getAll(dimension)
        return val || []
    }

    // Reconstruct rawParams for SidebarItem (it expects Record<string, ...>)
    const rawParams: Record<string, string | string[] | undefined> = {}
    searchParams.forEach((value, key) => {
        if (rawParams[key]) {
            if (Array.isArray(rawParams[key])) {
                (rawParams[key] as string[]).push(value)
            } else {
                rawParams[key] = [rawParams[key] as string, value]
            }
        } else {
            rawParams[key] = value
        }
    })

    // Helper to handle interaction for closing mobile sidebar
    const handleInteraction = (e: React.MouseEvent) => {
        const target = (e.target as HTMLElement).closest('a, button') as HTMLElement
        if (!onClose || !target) return

        // Check if the element (or any parent) explicitly requests NOT to close the sidebar
        if (target.closest('[data-no-close="true"]')) return

        setTimeout(onClose, 150)
    }

    // We render the skeleton structure but with client content logic
    if (loading) {
        return (
            <div className={className}>
                <OffensesSidebarSkeleton />
            </div>
        )
    }

    if (error) {
        return (
            <aside className={`border-border bg-card flex h-full w-64 flex-col border-r p-6 ${className || ''}`}>
                <div className="text-red-500 text-sm">Error loading filters</div>
            </aside>
        )
    }

    return (
        <aside
            className={`border-border bg-card flex h-full w-64 flex-col border-r print:hidden ${className || ''}`}
            onClickCapture={handleInteraction}
        >
            <div className={`flex-1 overflow-y-auto p-6 transition-opacity duration-200 ${isUpdating ? "opacity-60" : "opacity-100"}`}>
                <div className="mb-6">
                    <Link href="/" className="block transition-opacity hover:opacity-80">
                        <h1 className="text-foreground text-xl font-semibold">ChapaUY</h1>
                        <p className="text-muted-foreground mt-1 text-sm">
                            Infracciones de tránsito
                        </p>
                    </Link>
                </div>

                <NavSwitcher />

                <div>
                    {visibleDimensions.map((dimension) => {
                        const facet = facets.find((f) => f.dimension === dimension)

                        return (
                            <SidebarItem
                                key={dimension}
                                dimension={dimension}
                                facet={facet}
                                selectedValues={getSelectedValues(dimension)}
                                rawParams={rawParams}
                            />
                        )
                    })}
                </div>
            </div>
            <div className="border-border border-t p-4">
                <GlobalLinks className="justify-center" />
            </div>
        </aside>
    )
}
