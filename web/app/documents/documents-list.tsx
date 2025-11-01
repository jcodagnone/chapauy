/**
 * Copyright 2025 The ChapaUY Authors
 * SPDX-License-Identifier: Apache-2.0
 */

import { offensesParamsFromQueryParams } from "@/lib/url-utils"
import { getDocuments } from "@/lib/repository"
import { DocumentsFeed } from "./documents-feed"

interface DocumentsListProps {
  searchParams: Promise<Record<string, string | string[] | undefined>>
}

export async function DocumentsList({ searchParams }: DocumentsListProps) {
  const rawParams = await searchParams
  const params = offensesParamsFromQueryParams(rawParams)

  const page = params.page || 1
  const limit = 50

  const { documents, total } = await getDocuments(
    params.predicates,
    page,
    limit
  )

  const totalPages = Math.ceil(total / limit)

  const initialPagination = {
    current_page: page,
    total_pages: totalPages,
  }

  return (
    <DocumentsFeed
      initialDocuments={documents}
      initialPagination={initialPagination}
    />
  )
}
