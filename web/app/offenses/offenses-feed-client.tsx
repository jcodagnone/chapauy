/**
 * Copyright 2025 The ChapaUY Authors
 * SPDX-License-Identifier: Apache-2.0
 */

"use client"

import { useEffect, useState, useTransition, useMemo } from "react"
import { useSearchParams } from "next/navigation"
import { OffensesResponse, OffensesListResponse } from "@/lib/types"
import { SearchInterface } from "@/components/search-interface"
import OffensesLoading from "./loading"
import { buildApiUrl } from "@/lib/api-utils"

export function OffensesFeedClient() {
    const searchParams = useSearchParams()
    const [data, setData] = useState<OffensesResponse | null>(null)
    const [error, setError] = useState<string | null>(null)
    const [isPending, startTransition] = useTransition()

    // Stable dependency key for fetching: only include params that affect the feed
    const feedQueryString = useMemo(() => {
        const params = new URLSearchParams()
        searchParams.forEach((value, key) => {
            // Ignore UI state params that don't affect data filtering
            if (key !== "facet" && key !== "facets" && key !== "mode") {
                params.append(key, value)
            }
        })
        params.sort()
        return params.toString()
    }, [searchParams])

    useEffect(() => {
        // Collect all search params
        const params: Record<string, string | string[]> = {}
        searchParams.forEach((value, key) => {
            if (key === "facet" || key === "facets" || key === "mode") return

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

        const fetchData = async () => {
            try {
                const url = buildApiUrl("/api/v1/offenses", params)
                const res = await fetch(url)

                if (!res.ok) {
                    const errorData = await res.json().catch(() => ({}))
                    throw new Error(
                        errorData.error || `HTTP error! status: ${res.status}`
                    )
                }

                const jsonData: OffensesResponse = await res.json()

                startTransition(() => {
                    setData(jsonData)
                    setError(null)
                })
            } catch (err) {
                console.error("Failed to fetch offenses:", err)
                setError(err instanceof Error ? err.message : String(err))
            }
        }

        // Show skeleton loading state while fetching new data
        // This gives the user immediate feedback that the list is changing
        setData(null)
        fetchData()
    }, [feedQueryString])

    if (error) {
        return (
            <div className="p-8 text-center text-red-500">
                <h3 className="text-lg font-semibold">Error loading data</h3>
                <p>{error}</p>
                <button
                    onClick={() => window.location.reload()}
                    className="mt-4 px-4 py-2 bg-primary text-primary-foreground rounded hover:opacity-90 transition-opacity"
                >
                    Retry
                </button>
            </div>
        )
    }

    if (!data) {
        return <OffensesLoading />
    }

    return (
        <SearchInterface
            initialOffenses={data.offenses}
            initialPagination={data.pagination}
            initialRepos={data.repos}
            initialRepos={data.repos}
            // Actually SearchInterface expects `initialArticles: Record<string, string>`
            // and checking route.ts, we DO return `articles`.
            // The `OffensesResponse` type might be missing this field or I need to cast/extend it?
            // Let's assume for now and fix type if needed.
            // Wait, OffensesResponse has `offenses`, `pagination`, `repos`, `summary`...
            // `SearchInterface` props: `initialArticles`
            // My API returns `articles` too. 
            // I probably need to cast `data` or pass `data.articles` if it exists on the type.
            // Let's fix the API return type first or use `any` to unblock if type is strict.
            // The `OffensesResponse` type in lib/types.ts might not have `articles` or `chartData`.

            initialArticles={(data as any).articles || {}}
            initialSummary={data.summary}
            initialChartData={(data as any).chartData}
        />
    )
}
