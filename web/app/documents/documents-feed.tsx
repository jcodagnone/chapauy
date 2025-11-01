/**
 * Copyright 2025 The ChapaUY Authors
 * SPDX-License-Identifier: Apache-2.0
 */

"use client"

import { useEffect, useState, useRef, useTransition } from "react"
import { useOffenseSearchParams } from "@/lib/search-params"
import { offensesParamsFromQueryParams } from "@/lib/url-utils"
import { OffenseDocument, OffensesParams } from "@/lib/types"
import { DocumentCard } from "@/components/document-card"
import { OffenseCardSkeleton } from "@/components/offense-card-skeleton"
import { Button } from "@/components/ui/button"
import { AlertCircle } from "lucide-react"
import { loadMoreDocuments } from "./actions"

interface DocumentsFeedProps {
  initialDocuments: OffenseDocument[]
  initialPagination: {
    current_page: number
    total_pages: number
  }
}

export function DocumentsFeed({
  initialDocuments,
  initialPagination,
}: DocumentsFeedProps) {
  const { searchParams } = useOffenseSearchParams()
  const [documents, setDocuments] =
    useState<OffenseDocument[]>(initialDocuments)
  const [pagination, setPagination] = useState(initialPagination)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(false)
  const sentinelRef = useRef<HTMLDivElement>(null)

  // Reset state when initial props change (e.g. navigation / filtering)
  useEffect(() => {
    setDocuments(initialDocuments)
    setPagination(initialPagination)
    setLoading(false)
    setError(false)
  }, [initialDocuments, initialPagination])

  const loadMore = async () => {
    if (loading || pagination.current_page >= pagination.total_pages) return

    setLoading(true)
    setError(false)
    try {
      const params = offensesParamsFromQueryParams(searchParams)
      const nextPage = pagination.current_page + 1
      const res = await loadMoreDocuments({ ...params, page: nextPage })

      setDocuments((prev) => [...prev, ...res.documents])
      setPagination(res.pagination)
    } catch (err) {
      console.error("Error loading more documents:", err)
      setError(true)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    if (!sentinelRef.current || error) return

    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0].isIntersecting) {
          loadMore()
        }
      },
      { rootMargin: "400px" }
    )

    observer.observe(sentinelRef.current)
    return () => observer.disconnect()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [
    pagination.current_page,
    pagination.total_pages,
    loading,
    error,
    searchParams,
  ])

  const params = offensesParamsFromQueryParams(searchParams)

  return (
    <div className="border-card p-8 print:p-2">
      <div className="space-y-4 print:space-y-0">
        {documents.map((doc, i) => (
          <DocumentCard
            key={`${doc.db_id}-${doc.doc_id}-${i}`}
            document={doc}
            params={params}
          />
        ))}
      </div>

      {pagination.current_page < pagination.total_pages && (
        <div ref={sentinelRef} className="py-8">
          {loading ? (
            <div className="space-y-4">
              <OffenseCardSkeleton />
              <OffenseCardSkeleton />
            </div>
          ) : error ? (
            <div className="flex flex-col items-center gap-2">
              <div className="text-destructive flex items-center gap-2">
                <AlertCircle className="h-4 w-4" />
                <span>Error al cargar más documentos</span>
              </div>
              <Button onClick={() => loadMore()} variant="outline">
                Reintentar
              </Button>
            </div>
          ) : null}
        </div>
      )}

      {documents.length === 0 && (
        <div className="text-muted-foreground py-12 text-center">
          No se encontraron documentos
        </div>
      )}
    </div>
  )
}
