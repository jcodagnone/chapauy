/**
 * Copyright 2025 The ChapaUY Authors
 * SPDX-License-Identifier: Apache-2.0
 */

import { NextRequest, NextResponse } from "next/server"
import { getDimensionResults } from "@/lib/repository"
import { Dimension, InPredicate } from "@/lib/types"
import { checkETag } from "@/lib/etag"
import { normalizeVehicleId } from "@/lib/utils"

export async function GET(request: NextRequest) {
  try {
    const etagCheck = await checkETag(request)
    if (etagCheck.response) {
      return etagCheck.response
    }
    const { headers } = etagCheck.options!

    const searchParams = request.nextUrl.searchParams
    const predicates: InPredicate[] = []
    const dimensionParam = searchParams.get("dimension")
    const q = searchParams.get("q")

    if (!dimensionParam) {
      return NextResponse.json(
        { error: 'missing query param "dimension"' },
        { status: 400 }
      )
    }

    const dimension = dimensionParam as Dimension
    if (!Object.values(Dimension).includes(dimension)) {
      return NextResponse.json(
        { error: `unknown dimension "${dimensionParam}"` },
        { status: 400 }
      )
    }

    // Parse predicates
    Object.values(Dimension).forEach((dim) => {
      const values = searchParams.getAll(dim)
      if (values.length > 0) {
        predicates.push({ dimension: dim, values })
      }
    })

    const queries: Record<string, string> = {}
    if (q) {
      queries[dimension] =
        dimension === Dimension.Vehicle ? normalizeVehicleId(q) : q
    }

    const results = await getDimensionResults(predicates, [dimension], queries)

    if (results.length === 0) {
      return NextResponse.json({ dimension, values: [], total_values: 0 }, { headers })
    }

    return NextResponse.json(results[0], { headers })
  } catch (error) {
    console.error("[API] Error in /api/v1/suggest:", error)
    return NextResponse.json(
      {
        error: `Internal Server Error: ${error instanceof Error ? error.message : String(error)}`,
      },
      { status: 500 }
    )
  }
}
