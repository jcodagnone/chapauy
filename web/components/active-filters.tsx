/**
 * Copyright 2025 The ChapaUY Authors
 * SPDX-License-Identifier: Apache-2.0
 */

"use client"
import { Button } from "@/components/ui/button"
import { Trash2, X } from "lucide-react"
import { getDimensionConfig } from "@/lib/display-config"
import type { ActiveFilter as ActiveFilterType } from "@/lib/types"

interface ActiveFiltersProps {
  filters: ActiveFilterType[]
  onRemoveAction: (dimension: string, value: string) => void
  onClearAllAction: () => void
}

export function ActiveFilters({
  filters,
  onRemoveAction,
  onClearAllAction,
}: ActiveFiltersProps) {
  if (filters.length === 0) return null

  return (
    <div className="border-border mb-4 border-b pb-4">
      <div className="mb-2 flex items-center justify-between">
        <span className="text-muted-foreground text-sm font-medium">
          Filtros activos
        </span>
        <Button
          variant="ghost"
          size="sm"
          onClick={onClearAllAction}
          className="hover:bg-accent h-6 px-2 text-xs"
          title="Limpiar todos los filtros"
        >
          <Trash2 className="mr-1 h-3 w-3" />
          Limpiar
        </Button>
      </div>

      <div className="space-y-2">
        {filters.map((filter) => {
          const config = getDimensionConfig(filter.dimension)
          const IconComponent = config.icon

          return (
            <div key={filter.dimension} className="flex flex-col gap-1.5">
              {filter.values.map((valueLabel, index) => {
                const isEmpty = valueLabel.value === ""
                const displayText = isEmpty
                  ? config.empty
                  : valueLabel.label || valueLabel.value
                const pillClassName = isEmpty
                  ? "inline-flex items-center justify-between bg-muted/50 text-muted-foreground px-2.5 py-1 rounded-full text-xs font-medium max-w-full gap-1.5 italic"
                  : "inline-flex items-center justify-between bg-primary/10 text-primary px-2.5 py-1 rounded-full text-xs font-medium max-w-full gap-1.5"

                return (
                  <span
                    key={`${filter.dimension}-${valueLabel.value}-${index}`}
                    className={pillClassName}
                    title={`${config.label}: ${displayText}`}
                  >
                    <div className="flex min-w-0 items-center gap-1.5">
                      <IconComponent className="h-3.5 w-3.5 flex-shrink-0" />
                      <span className="truncate">{displayText}</span>
                    </div>
                    <button
                      onClick={() =>
                        onRemoveAction(filter.dimension, valueLabel.value)
                      }
                      className="hover:bg-primary/20 flex-shrink-0 rounded-full p-0.5 transition-colors"
                      aria-label={`Eliminar filtro ${displayText}`}
                    >
                      <X className="h-3 w-3" />
                    </button>
                  </span>
                )
              })}
            </div>
          )
        })}
      </div>
    </div>
  )
}
