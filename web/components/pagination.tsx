/**
 * Copyright 2025 The ChapaUY Authors
 * SPDX-License-Identifier: Apache-2.0
 */

"use client"

import { ChevronLeft, ChevronRight } from "lucide-react"
import { Button } from "@/components/ui/button"

interface PaginationProps {
  currentPage: number
  totalPages: number
  onPageChangeAction: (page: number) => void
}

export function Pagination({
  currentPage,
  totalPages,
  onPageChangeAction,
}: PaginationProps) {
  const pages = []
  const maxVisible = 5

  let startPage = Math.max(1, currentPage - Math.floor(maxVisible / 2))
  const endPage = Math.min(totalPages, startPage + maxVisible - 1)

  if (endPage - startPage + 1 < maxVisible) {
    startPage = Math.max(1, endPage - maxVisible + 1)
  }

  for (let i = startPage; i <= endPage; i++) {
    pages.push(i)
  }

  if (totalPages <= 1) return null

  return (
    <div className="mt-8 flex items-center justify-center gap-2">
      <Button
        variant="outline"
        size="sm"
        onClick={() => onPageChangeAction(currentPage - 1)}
        disabled={currentPage === 1}
        className="border-border"
      >
        <ChevronLeft className="h-4 w-4" />
      </Button>

      {startPage > 1 && (
        <>
          <Button
            variant="outline"
            size="sm"
            onClick={() => onPageChangeAction(1)}
            className="border-border"
          >
            1
          </Button>
          {startPage > 2 && (
            <span className="text-muted-foreground px-2">...</span>
          )}
        </>
      )}

      {pages.map((page) => (
        <Button
          key={page}
          variant={page === currentPage ? "default" : "outline"}
          size="sm"
          onClick={() => onPageChangeAction(page)}
          className={
            page === currentPage
              ? "bg-primary text-primary-foreground"
              : "border-border"
          }
        >
          {page}
        </Button>
      ))}

      {endPage < totalPages && (
        <>
          {endPage < totalPages - 1 && (
            <span className="text-muted-foreground px-2">...</span>
          )}
          <Button
            variant="outline"
            size="sm"
            onClick={() => onPageChangeAction(totalPages)}
            className="border-border"
          >
            {totalPages}
          </Button>
        </>
      )}

      <Button
        variant="outline"
        size="sm"
        onClick={() => onPageChangeAction(currentPage + 1)}
        disabled={currentPage === totalPages}
        className="border-border"
      >
        <ChevronRight className="h-4 w-4" />
      </Button>
    </div>
  )
}
