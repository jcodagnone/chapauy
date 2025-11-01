/**
* Copyright 2025 The ChapaUY Authors
* SPDX-License-Identifier: Apache-2.0
*/

import { NextRequest, NextResponse } from "next/server"
import {
    getOffenses,
    getOffensesSummary,
    getArticles,
    getChartDataByDayOfWeek,
    getChartDataByDayOfYear,
    getChartDataByTimeOfDay,
    determineSortBy,
} from "@/lib/repository"
import { offensesParamsFromQueryParams } from "@/lib/url-utils"
import {
    OffensesResponse,
    Repo,
    Dimension,
    Facet,
    FacetValue,
    SortBy,
} from "@/lib/types"
import { getDBName, countryDisplay } from "@/lib/db-refs"
import { checkETag } from "@/lib/etag"

const ERROR_CACHE_HEADERS = {
    "Cache-Control": "public, max-age=60, s-maxage=3600",
}

// Allowed query parameters for strict validation
const ALLOWED_PARAMS = new Set([
    "page",
    "per_page",
    "view",
    "group_by",
    "facets", // Sometimes passed, though usually specific to sidebar
    "facet",
    "q",
    ...Object.values(Dimension),
])

// Hotlinking protection
import { isAllowedOrigin } from "@/lib/security"

export async function GET(request: NextRequest) {
    try {
        // 1. Hotlinking Protection
        if (!isAllowedOrigin(request)) {
            return NextResponse.json(
                { error: "Forbidden" },
                { status: 403, headers: ERROR_CACHE_HEADERS }
            )
        }

        // 2. ETag / Conditional GET Support
        const etagCheck = await checkETag(request)
        if (etagCheck.response) {
            return etagCheck.response
        }
        const { headers } = etagCheck.options!

        const searchParams = request.nextUrl.searchParams

        // 3. Strict Parameter Validation
        for (const key of searchParams.keys()) {
            if (!ALLOWED_PARAMS.has(key)) {
                return NextResponse.json(
                    { error: `Unknown parameter: ${key}` },
                    { status: 400, headers: ERROR_CACHE_HEADERS }
                )
            }
        }

        // 4. Parse Parameters
        const rawParams: Record<string, string | string[] | undefined> = {}
        searchParams.forEach((value, key) => {
            if (rawParams[key]) {
                if (Array.isArray(rawParams[key])) {
                    ; (rawParams[key] as string[]).push(value)
                } else {
                    rawParams[key] = [rawParams[key] as string, value]
                }
            } else {
                rawParams[key] = value
            }
        })

        const params = offensesParamsFromQueryParams(rawParams)

        const page = params.page || 1
        const sortBy = determineSortBy(params.predicates)
        const limit = params.per_page || (sortBy === SortBy.Document ? 1500 : 20)

        // 5. Fetch Data
        const [offenses, summaryStats, allArticles] = await Promise.all([
            getOffenses(params.predicates, sortBy, page, limit),
            getOffensesSummary(params.predicates, null),
            getArticles(),
        ])

        const stats = summaryStats[0] || { count: 0 }

        // Hydrate active facets
        const hydratedFacets: Facet[] = params.predicates.map((predicate) => {
            const values: FacetValue[] = predicate.values.map((val) => {
                let label = val
                switch (predicate.dimension) {
                    case Dimension.Database:
                        label = getDBName(Number(val))
                        break
                    case Dimension.Country:
                        label = countryDisplay[val] || val
                        break
                    case Dimension.ArticleID:
                        label = allArticles.byId[val] || val
                        break
                    case Dimension.ArticleCode:
                        label = allArticles.byCode[val] || val
                        break
                    case Dimension.Features:
                        switch (val) {
                            case "with_error":
                                label = "Con Errores"
                                break
                            case "no_error":
                                label = "Sin Errores"
                                break
                            case "with_ur":
                                label = "Con UR"
                                break
                            case "no_ur":
                                label = "Sin UR"
                                break
                        }
                        break
                }

                return {
                    value: val,
                    label: label,
                    count: 0,
                    selected: true,
                }
            })

            return {
                dimension: predicate.dimension,
                values: values,
                total_values: values.length,
            }
        })

        // Charts
        const viewMode = rawParams.view as string | undefined
        let chartData = undefined

        if (viewMode === "charts") {
            const groupBy = (rawParams.group_by as Dimension) || undefined
            const [dayOfWeek, dayOfYear, timeOfDay] = await Promise.all([
                getChartDataByDayOfWeek(params.predicates, groupBy),
                getChartDataByDayOfYear(params.predicates, groupBy),
                getChartDataByTimeOfDay(params.predicates, groupBy),
            ])
            chartData = { dayOfWeek, dayOfYear, timeOfDay }
        }

        const totalCount = Number(stats.count)
        const totalPages = Math.ceil(totalCount / limit)

        // Build repos map
        const repos: Record<string, Repo> = {}
        offenses.forEach((offense) => {
            if (!repos[offense.repo_id]) {
                repos[offense.repo_id] = { name: getDBName(offense.repo_id) }
            }
        })

        // Build articles map
        const articles: Record<string, string> = {}
        offenses.forEach((offense) => {
            if (offense.article_id && Array.isArray(offense.article_id)) {
                offense.article_id.forEach((id: string) => {
                    if (allArticles.byId[id]) {
                        articles[id] = allArticles.byId[id]
                    }
                })
            }
        })

        const responseData: OffensesResponse = {
            offenses,
            pagination: {
                current_page: page,
                total_pages: totalPages,
            },
            repos,
            summary: {
                avg_ur: Number(stats.ur_avg),
                facets: hydratedFacets,
                record_count: totalCount,
                total_ur: Number(stats.ur_total),
                viewport_h3_index: stats.viewport_h3_index,
            },
            chartData,
        }

        // Return with Cache Headers
        return NextResponse.json(responseData, {
            status: 200,
            headers: headers,
        })
    } catch (error) {
        console.error("[API] Error in /api/v1/offenses:", error)
        return NextResponse.json(
            {
                error: `Internal Server Error: ${error instanceof Error ? error.message : String(error)}`,
            },
            { status: 500, headers: ERROR_CACHE_HEADERS }
        )
    }
}
