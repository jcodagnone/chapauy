/**
 * Copyright 2025 The ChapaUY Authors
 * SPDX-License-Identifier: Apache-2.0
 */

"use server"

import { getDocumentFacets, getDocuments } from "@/lib/repository"
import { OffensesParams, InPredicate, OffenseDocument } from "@/lib/types"

export interface DocumentsListResponse {
  documents: OffenseDocument[]
  pagination: {
    current_page: number
    total_pages: number
  }
}

export async function loadMoreDocuments(
  params: OffensesParams
): Promise<DocumentsListResponse> {
  try {
    const page = params.page || 1
    const limit = 50 // Use a reasonable limit for documents list

    // Predicates should be passed from the client, constructed from URL params
    const predicates: InPredicate[] = params.predicates

    const { documents, total } = await getDocuments(predicates, page, limit)

    const totalPages = Math.ceil(total / limit)

    return {
      documents,
      pagination: {
        current_page: page,
        total_pages: totalPages,
      },
    }
  } catch (error) {
    console.error("[ServerAction] Error in loadMoreDocuments:", error)
    throw new Error(
      `Failed to load more documents: ${error instanceof Error ? error.message : String(error)}`
    )
  }
}
