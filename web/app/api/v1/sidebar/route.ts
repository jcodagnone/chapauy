/**
 * Copyright 2025 The ChapaUY Authors
 * SPDX-License-Identifier: Apache-2.0
 */

import { NextRequest, NextResponse } from "next/server"
import { getDimensionResults, getDocumentFacets } from "@/lib/repository"
import { offensesParamsFromQueryParams } from "@/lib/url-utils"
import { Dimension, SidebarMode } from "@/lib/types"
import { checkETag } from "@/lib/etag"

const ERROR_CACHE_HEADERS = {
    "Cache-Control": "public, max-age=60, s-maxage=3600",
}

// Allowed query parameters for strict validation
const ALLOWED_PARAMS = new Set([
    "mode",
    "facets",
    "facet",
    "view",
    "group_by",
    "page",
    "per_page",
    ...Object.values(Dimension),
])

// Hotlinking protection
import { isAllowedOrigin } from "@/lib/security"

export async function GET(request: NextRequest) {
    try {
        if (!isAllowedOrigin(request)) {
            return NextResponse.json(
                { error: "Forbidden" },
                { status: 403, headers: ERROR_CACHE_HEADERS }
            )
        }

        const etagCheck = await checkETag(request)
        if (etagCheck.response) {
            return etagCheck.response
        }
        const { headers } = etagCheck.options!

        const searchParams = request.nextUrl.searchParams

        // Strict Parameter Validation
        for (const key of searchParams.keys()) {
            if (!ALLOWED_PARAMS.has(key)) {
                return NextResponse.json(
                    { error: `Unknown parameter: ${key}` },
                    { status: 400, headers: ERROR_CACHE_HEADERS }
                )
            }
        }

        // Parse params
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
        const mode = rawParams.mode as SidebarMode

        if (!mode) {
            return NextResponse.json(
                { error: "Missing mode parameter" },
                { status: 400, headers: ERROR_CACHE_HEADERS }
            )
        }

        // Determine facets to fetch based on 'facet' (singular) and 'facets' (plural) params
        // 'facet' is used by the UI for expanded accordions
        const facetParam = rawParams.facet
        const facetsParam = params.facets

        const facetsToFetch: Dimension[] = []

        if (facetsParam) {
            facetsToFetch.push(...facetsParam)
        }

        if (facetParam) {
            if (Array.isArray(facetParam)) {
                facetsToFetch.push(...(facetParam as Dimension[]))
            } else {
                facetsToFetch.push(facetParam as Dimension)
            }
        }

        // Remove duplicates
        const uniqueFacetsToFetch = Array.from(new Set(facetsToFetch))

        let facetResults
        if (mode === SidebarMode.Documents) {
            facetResults = await getDocumentFacets(params.predicates, uniqueFacetsToFetch)
        } else {
            facetResults = await getDimensionResults(params.predicates, uniqueFacetsToFetch)
        }

        return NextResponse.json(facetResults, {
            status: 200,
            headers: headers,
        })
    } catch (error) {
        console.error("[API] Error in /api/v1/sidebar:", error)
        return NextResponse.json(
            {
                error: `Internal Server Error: ${error instanceof Error ? error.message : String(error)}`,
            },
            { status: 500, headers: ERROR_CACHE_HEADERS }
        )
    }
}
