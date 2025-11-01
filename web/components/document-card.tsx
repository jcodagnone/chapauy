/**
 * Copyright 2025 The ChapaUY Authors
 * SPDX-License-Identifier: Apache-2.0
 */

import type React from "react"
import Link from "next/link"
import { ExternalLink, AlertTriangle, FileText } from "lucide-react"
import { Card } from "@/components/ui/card"
import {
  type OffenseDocument,
  type OffensesParams,
  Dimension,
} from "@/lib/types"
import { formatUR } from "@/lib/utils"
import { getDBName } from "@/lib/db-refs"
import {
  buildUrlToggleFilter,
  isFilterActiveInParams,
  toSearchParams,
} from "@/lib/url-utils"
import { FilterableItem } from "@/components/ui/filterable-item"

interface DocumentCardProps {
  document: OffenseDocument
  params: OffensesParams
}

export function DocumentCard({ document, params }: DocumentCardProps) {
  const dateTime = new Date(document.doc_date)
  const searchParams = toSearchParams(params)
  const urValue = formatUR(document.ur)
  const repoName = getDBName(document.db_id)

  const isFilterActive = (dimension: string, value: string): boolean => {
    return isFilterActiveInParams(params, dimension, value)
  }

  const getToggleUrl = (dimension: string, value: string) => {
    return buildUrlToggleFilter(searchParams, dimension, value)
  }

  // Check if we should render localized date? existing code did it.
  // The document.doc_date is "YYYY-MM-DD".
  // new Date("YYYY-MM-DD") creates UTC date usually, or local?
  // Let's assume standard behavior.
  // We can just format it nicely.

  // existing OffenseCard logic for date:
  // dateTime.toLocaleDateString("es-UY", { ... })

  // Since time is T00:00:00 probably (because we split by T previously), it's safe.

  const offensesUrl = `/offenses/?${Dimension.DocSource}=${encodeURIComponent(document.doc_source)}`

  return (
    <Card
      className={`hover:bg-accent/50 p-3 transition-colors print:rounded-none print:border-0 print:border-b print:p-1 print:shadow-none ${
        document.errors > 0
          ? "border-accent bg-accent/20 print:bg-transparent"
          : "border-border"
      }`}
    >
      <div className="space-y-1 print:space-y-0.5">
        <div className="flex items-start justify-between gap-3">
          <div className="flex flex-1 flex-wrap items-start gap-1.5">
            <h3 className="text-foreground flex items-center gap-2 text-sm leading-tight font-medium">
              <FileText className="text-muted-foreground h-4 w-4" />
              <Link
                href={offensesUrl}
                className="hover:underline"
                title="Ver infracciones de este documento"
                prefetch={false}
              >
                {document.doc_id}
              </Link>
            </h3>
          </div>

          <div className="flex flex-shrink-0 items-center gap-2">
            {document.errors > 0 && (
              <div
                className="flex items-center gap-1 text-yellow-600 dark:text-yellow-500"
                title={`${document.errors} errores`}
              >
                <AlertTriangle className="h-4 w-4 flex-shrink-0" />
                <span className="font-mono text-xs font-bold">
                  {document.errors}
                </span>
              </div>
            )}

            {!isFilterActive(Dimension.Database, String(document.db_id)) && (
              <FilterableItem
                href={getToggleUrl(Dimension.Database, String(document.db_id))}
                title="Filtrar por esta base de datos"
              >
                <span className="bg-muted text-muted-foreground hover:bg-muted/80 print:hover:bg-muted flex-shrink-0 cursor-pointer rounded px-1.5 py-0.5 font-mono text-[10px] transition-colors print:cursor-default">
                  {repoName}
                </span>
              </FilterableItem>
            )}
            {isFilterActive(Dimension.Database, String(document.db_id)) && (
              <span className="bg-muted text-muted-foreground flex-shrink-0 rounded px-1.5 py-0.5 font-mono text-[10px]">
                {repoName}
              </span>
            )}

            {urValue && (
              <div className="text-foreground flex-shrink-0 text-sm font-semibold">
                {urValue} UR
              </div>
            )}
          </div>
        </div>

        <div className="text-muted-foreground flex flex-wrap items-center gap-1.5 text-xs leading-tight">
          <span>
            {dateTime.toLocaleDateString("es-UY", {
              year: "numeric",
              month: "short",
              day: "numeric",
              timeZone: "UTC", // Important since we stored as YYYY-MM-DD
            })}
          </span>
          <span>•</span>
          <span>{document.records.toLocaleString()} registros</span>

          <>
            <span>•</span>
            <Link
              href={offensesUrl}
              className="text-primary hover:text-primary/80 inline-flex items-center gap-1 hover:underline"
              title="Ver infracciones de este documento"
              prefetch={false}
            >
              Ver Infracciones
            </Link>
          </>

          {document.doc_source && (
            <>
              <span>•</span>
              <a
                href={document.doc_source}
                target="_blank"
                rel="noopener noreferrer"
                className="text-primary hover:text-primary/80 inline-flex items-center gap-1"
                title="Ver documento original"
              >
                <span>Fuente</span>
                <ExternalLink className="h-3 w-3" />
              </a>
            </>
          )}
        </div>
      </div>
    </Card>
  )
}
