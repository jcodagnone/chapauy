/**
 * Copyright 2025 The ChapaUY Authors
 * SPDX-License-Identifier: Apache-2.0
 */

import { NextRequest, NextResponse } from "next/server"
import { getMapClusters } from "@/lib/repository"
import { Dimension, InPredicate } from "@/lib/types"
import { checkETag } from "@/lib/etag"

export async function GET(
  request: NextRequest,
  props: { params: Promise<{ h3_index: string }> }
) {
  try {
    const etagCheck = await checkETag(request)
    if (etagCheck.response) {
      return etagCheck.response
    }
    const { headers } = etagCheck.options!

    const params = await props.params
    const h3Index = params.h3_index

    // Basic validation (length and hex chars)
    if (!/^[0-9a-fA-F]{15,16}$/.test(h3Index)) {
      return NextResponse.json(
        { error: `Invalid h3 index: ${h3Index}` },
        { status: 400 }
      )
    }

    const searchParams = request.nextUrl.searchParams
    const predicates: InPredicate[] = []

    // Parse predicates
    Object.values(Dimension).forEach((dim) => {
      const values = searchParams.getAll(dim)
      if (values.length > 0) {
        predicates.push({ dimension: dim, values })
      }
    })

    const data = await getMapClusters(predicates, h3Index)

    return NextResponse.json(data, { headers })
  } catch (error) {
    console.error(`[API] Error in /api/v1/map/[h3_index]:`, error)
    return NextResponse.json(
      {
        error: `Internal Server Error: ${error instanceof Error ? error.message : String(error)}`,
      },
      { status: 500 }
    )
  }
}
