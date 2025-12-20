/**
 * Copyright 2025 The ChapaUY Authors
 * SPDX-License-Identifier: Apache-2.0
 */

import type React from "react"
import Link from "next/link"
import { ExternalLink, MapPin, Car, Bike, AlertTriangle } from "lucide-react"
import { Card } from "@/components/ui/card"
import {
  type Offense,
  type Repo,
  type OffensesParams,
  Dimension,
} from "@/lib/types"
import { FilterableItem } from "@/components/ui/filterable-item"
import { formatUR } from "@/lib/utils"
import { getCountryFlag } from "@/lib/display-config"
import {
  buildUrlToggleFilter,
  isFilterActiveInParams,
  toSearchParams,
} from "@/lib/url-utils"

interface OffenseCardProps {
  offense: Offense
  repos: Record<string, Repo>
  articles?: Record<string, string>
  params: OffensesParams
}

const getVehicleIcon = (vehicleType: string) => {
  var ret
  if (vehicleType === "Moto") {
    ret = <Bike className="h-3 w-3" />
  } else if (vehicleType === "Auto") {
    ret = <Car className="h-3 w-3" />
  } else {
    ret = undefined
  }
  return ret
}

export function OffenseCard({
  offense,
  repos,
  articles,
  params,
}: OffenseCardProps) {
  const dateTime = new Date(offense.time)
  const searchParams = toSearchParams(params)

  const urValue = formatUR(offense.ur)
  const repoName = repos[offense.repo_id]?.name || "Desconocido"
  const vehicleIcon = getVehicleIcon(offense.vehicle_type)

  const displayLocation = offense.display_location || offense.location

  const isFilterActive = (dimension: string, value: string): boolean => {
    return isFilterActiveInParams(params, dimension, value)
  }

  const getToggleUrl = (dimension: string, value: string) => {
    return buildUrlToggleFilter(searchParams, dimension, value)
  }

  return (
    <Card
      className={`hover:bg-accent/50 p-3 transition-colors print:rounded-none print:border-0 print:border-b print:p-1 print:shadow-none overflow-hidden ${offense.error ? "border-accent bg-accent/20 print:bg-transparent" : "border-border"}`}
    >
      <div className="space-y-1 print:space-y-0.5">
        <div className="flex items-start justify-between gap-3">
          <div className="flex flex-1 flex-wrap items-start gap-1.5">
            {offense.article_id && offense.article_id.length > 0 && (
              <>
                {offense.article_id.map((id, index) => {
                  const description = articles?.[id]
                  const title = description
                    ? description
                    : "Filtrar por artículo"
                  const isActive = isFilterActive(Dimension.ArticleID, id)
                  return (
                    <span key={id} className="inline-flex items-center gap-1">
                      {isActive ? (
                        <span
                          className="bg-primary/10 text-primary rounded px-1.5 py-0.5 font-mono text-xs font-medium"
                          title={description}
                        >
                          {id}
                        </span>
                      ) : (
                        <FilterableItem
                          href={getToggleUrl(Dimension.ArticleID, id)}
                          title={title}
                        >
                          <span className="bg-primary/10 text-primary hover:bg-primary/20 rounded px-1.5 py-0.5 font-mono text-xs font-medium transition-colors">
                            {id}
                          </span>
                        </FilterableItem>
                      )}
                      {index < offense.article_id!.length - 1 && (
                        <span className="text-muted-foreground text-xs">,</span>
                      )}
                    </span>
                  )
                })}
              </>
            )}
            <h3 className="text-foreground text-sm leading-tight font-medium break-words min-w-0">
              {isFilterActive(Dimension.Description, offense.description) ? (
                <span>{offense.description}</span>
              ) : (
                <FilterableItem
                  href={getToggleUrl(
                    Dimension.Description,
                    offense.description
                  )}
                  title="Filtrar por descripción"
                >
                  {offense.description}
                </FilterableItem>
              )}
            </h3>
          </div>

          <div className="flex flex-shrink-0 items-center gap-2">
            {offense.error && (
              <AlertTriangle className="text-muted-foreground h-4 w-4 flex-shrink-0 print:hidden" />
            )}
            {!isFilterActive(Dimension.Database, String(offense.repo_id)) && (
              <Link
                href={getToggleUrl(Dimension.Database, String(offense.repo_id))}
                className="bg-muted text-muted-foreground hover:bg-muted/80 print:hover:bg-muted flex-shrink-0 cursor-pointer rounded px-1.5 py-0.5 font-mono text-[10px] transition-colors print:cursor-default"
                title="Filtrar por esta base de datos"
                prefetch={false}
              >
                {repoName}
              </Link>
            )}
            {isFilterActive(Dimension.Database, String(offense.repo_id)) && (
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

        <div className="text-muted-foreground flex flex-wrap items-center gap-1.5 text-xs leading-tight break-words">
          {(() => {
            const dateValue = offense.time.substring(0, 10)
            const label =
              dateTime.getHours() === 0 &&
                dateTime.getMinutes() === 0 &&
                dateTime.getSeconds() === 0
                ? dateTime.toLocaleDateString("es-UY", {
                  year: "numeric",
                  month: "short",
                  day: "numeric",
                })
                : dateTime.toLocaleString("es-UY", {
                  year: "numeric",
                  month: "short",
                  day: "numeric",
                  hour: "2-digit",
                  minute: "2-digit",
                  hour12: false,
                })

            if (isFilterActive(Dimension.Date, dateValue)) {
              return <span>{label}</span>
            }

            return (
              <FilterableItem
                href={getToggleUrl(Dimension.Date, dateValue)}
                title="Filtrar por fecha"
              >
                {label}
              </FilterableItem>
            )
          })()}
          <span>•</span>
          {displayLocation &&
            !isFilterActive(Dimension.Location, offense.location) && (
              <>
                <FilterableItem
                  href={getToggleUrl(Dimension.Location, offense.location)}
                  title="Filtrar por ubicación"
                >
                  <MapPin className="h-3 w-3" />
                  {displayLocation}
                </FilterableItem>
                <span>•</span>
              </>
            )}
          {displayLocation &&
            isFilterActive(Dimension.Location, offense.location) && (
              <>
                <span className="inline-flex items-center gap-1">
                  <MapPin className="h-3 w-3" />
                  {displayLocation}
                </span>
                <span>•</span>
              </>
            )}

          {isFilterActive(Dimension.Country, offense.country) ? (
            <span>{getCountryFlag(offense.country)}</span>
          ) : (
            <FilterableItem
              href={getToggleUrl(Dimension.Country, offense.country)}
              title="Filtrar por país"
            >
              {getCountryFlag(offense.country)}
            </FilterableItem>
          )}

          {isFilterActive(Dimension.Vehicle, offense.vehicle) ? (
            <span>{offense.vehicle}</span>
          ) : (
            <FilterableItem
              href={getToggleUrl(Dimension.Vehicle, offense.vehicle)}
              title="Filtrar por vehículo"
            >
              {offense.vehicle}
            </FilterableItem>
          )}

          {vehicleIcon &&
            !isFilterActive(Dimension.VehicleType, offense.vehicle_type) && (
              <FilterableItem
                href={getToggleUrl(Dimension.VehicleType, offense.vehicle_type)}
                title="Filtrar por tipo de vehículo"
              >
                {vehicleIcon}
              </FilterableItem>
            )}
          {vehicleIcon &&
            isFilterActive(Dimension.VehicleType, offense.vehicle_type) && (
              <span>{vehicleIcon}</span>
            )}

          {offense.doc_source && offense.doc_id && (
            <>
              <span>•</span>
              {isFilterActive(Dimension.DocSource, offense.doc_source) ? (
                <span>
                  {offense.doc_id} #{offense.record_id}
                </span>
              ) : (
                <FilterableItem
                  href={getToggleUrl(Dimension.DocSource, offense.doc_source)}
                  title="Filtrar por fuente"
                >
                  {offense.doc_id} #{offense.record_id}
                </FilterableItem>
              )}
              <a
                href={offense.doc_source}
                target="_blank"
                rel="noopener noreferrer"
                className="text-primary hover:text-primary/80 inline-flex items-center"
                title="Ver documento"
              >
                <ExternalLink className="h-3 w-3" />
              </a>
            </>
          )}
          {offense.id && (
            <>
              <span>•</span>
              <span title="ID de la infracción">{offense.id}</span>
            </>
          )}
        </div>

        {offense.error && (
          <div className="bg-muted/50 border-border mt-2 rounded border p-2 print:hidden">
            <div className="flex items-start gap-2">
              <div className="text-muted-foreground text-xs">
                <span className="font-medium">⚠</span> {offense.error}
              </div>
            </div>
          </div>
        )}
      </div>
    </Card>
  )
}
