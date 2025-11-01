/**
 * Copyright 2025 The ChapaUY Authors
 * SPDX-License-Identifier: Apache-2.0
 */

"use client"
import { Plus, X, AlertCircle, BarChart3, List, MapPin } from "lucide-react"
import {
  useCallback,
  useEffect,
  useState,
  useRef,
  useMemo,
  useTransition,
} from "react"
import { } from "@/lib/api/client"
import { buildApiUrl } from "@/lib/api-utils"
import type {
  OffensesListResponse,
  OffensesSummaryResponse,
  ActiveFilter,
} from "@/lib/types"
import { Dimension } from "@/lib/types"
import { FacetFilter } from "@/components/facet-filter"
import { OffenseCard } from "@/components/offense-card"
import { OffenseCardSkeleton } from "@/components/offense-card-skeleton"
import { FacetFilterSkeleton } from "@/components/facet-filter-skeleton"
import dynamic from "next/dynamic"
import Link from "next/link"
import { ReadonlyURLSearchParams, useRouter } from "next/navigation"
import { formatUR } from "@/lib/utils"
import { useOffenseSearchParams } from "@/lib/search-params"
import {
  offensesParamsFromQueryParams,
  buildUrlWithoutFilter,
} from "@/lib/url-utils"
import { getDimensionConfig } from "@/lib/display-config"
import { Card } from "@/components/ui/card"
import { Button } from "@/components/ui/button"

// Dynamically import heavy visualization components
const OffenseMap = dynamic(
  () => import("@/components/offense-map").then((mod) => mod.OffenseMap),
  {
    ssr: false,
    loading: () => (
      <div className="bg-muted flex h-[calc(100vh-12rem)] w-full items-center justify-center rounded-lg border">
        <p className="text-muted-foreground">Cargando mapa...</p>
      </div>
    ),
  }
)

const OffenseCharts = dynamic(
  () => import("@/components/offense-charts").then((mod) => mod.OffenseCharts),
  {
    loading: () => (
      <div className="flex h-96 w-full items-center justify-center">
        <p className="text-muted-foreground">Cargando gráficos...</p>
      </div>
    ),
  }
)

const allDimensions: Dimension[] = [
  Dimension.Database,
  Dimension.Year,
  Dimension.Country,
  Dimension.VehicleType,
  Dimension.ArticleCode,
  Dimension.ArticleID,
  Dimension.Description,
  Dimension.Location,
  Dimension.Vehicle,
]

interface SearchInterfaceProps {
  initialOffenses: OffensesListResponse["offenses"]
  initialPagination: OffensesListResponse["pagination"]
  initialRepos: OffensesListResponse["repos"]
  initialArticles: Record<string, string>
  initialSummary: OffensesSummaryResponse
  initialChartData: {
    dayOfWeek: Record<string, Record<string, number>> | null
    dayOfYear: Record<string, Record<string, number>> | null
    timeOfDay: Record<string, Record<string, number>> | null
  }
}

export function SearchInterface({
  initialOffenses,
  initialPagination,
  initialRepos,
  initialArticles,
  initialSummary,
  initialChartData,
}: SearchInterfaceProps) {
  const {
    searchParams,
    updateURL,
    addFilter,
    removeFilter,
    clearAllFilters,
    setViewMode,
    setGroupBy,
    getActiveFilters,
    viewMode,
    groupBy,
  } = useOffenseSearchParams()

  const router = useRouter()
  const [isPending, startTransition] = useTransition()

  const handleViewChange = (newView: "list" | "charts" | "map") => {
    startTransition(() => {
      const params = new URLSearchParams(searchParams.toString())
      params.set("view", newView)
      params.sort()
      router.push(`?${params.toString()}`)
    })
  }

  const getViewUrl = (view: string) => {
    const params = new URLSearchParams(searchParams.toString())
    params.set("view", view)
    params.sort()
    return `?${params.toString()}`
  }

  const handleLinkClick = (
    e: React.MouseEvent<HTMLAnchorElement>,
    url: string
  ) => {
    if (e.metaKey || e.ctrlKey || e.shiftKey || e.altKey) return
    e.preventDefault()
    startTransition(() => {
      router.push(url)
    })
  }

  const [accumulatedOffenses, setAccumulatedOffenses] =
    useState<OffensesListResponse["offenses"]>(initialOffenses)
  const [pagination, setPagination] = useState<
    OffensesListResponse["pagination"] | null
  >(initialPagination)
  const [repos, setRepos] =
    useState<OffensesListResponse["repos"]>(initialRepos)
  const [articles, setArticles] =
    useState<Record<string, string>>(initialArticles)
  const [summaryData, setSummaryData] =
    useState<OffensesSummaryResponse | null>(initialSummary)
  const [loading, setLoading] = useState(false)
  const [isLoadingMore, setIsLoadingMore] = useState(false)
  const [loadMoreError, setLoadMoreError] = useState(false)
  const sentinelRef = useRef<HTMLDivElement>(null)
  const [currentPage, setCurrentPage] = useState(1)

  const [chartData, setChartData] = useState<{
    dayOfWeek: Record<string, Record<string, number>> | null
    dayOfYear: Record<string, Record<string, number>> | null
    timeOfDay: Record<string, Record<string, number>> | null
  }>(initialChartData)
  const [chartsLoading, setChartsLoading] = useState(false)

  const lastLoadedListKeyRef = useRef<string>("")
  const lastLoadedSummaryKeyRef = useRef<string>("")
  const lastLoadedChartsKeyRef = useRef<string>("")

  // Sync state with props when they change (e.g. on navigation)
  useEffect(() => {
    setAccumulatedOffenses(initialOffenses)
    setPagination(initialPagination)
    setRepos(initialRepos)
    setArticles(initialArticles)
    setSummaryData(initialSummary)
    setChartData(initialChartData)
    setCurrentPage(initialPagination.current_page)
    setChartsLoading(false) // Reset loading state on nav
    setLoading(false)
  }, [
    initialOffenses,
    initialPagination,
    initialRepos,
    initialArticles,
    initialSummary,
    initialChartData,
  ])

  const searchParamsString = searchParams.toString()

  useEffect(() => {
    if (!sentinelRef.current || !pagination || loadMoreError) return

    const observer = new IntersectionObserver(
      async (entries) => {
        const [entry] = entries
        if (
          entry.isIntersecting &&
          !isLoadingMore &&
          !loading &&
          currentPage < pagination.total_pages
        ) {
          setIsLoadingMore(true)
          try {
            const params = offensesParamsFromQueryParams(searchParams)
            const nextPage = currentPage + 1

            // Convert params to Record<string, string | string[]>
            const apiParams: Record<string, string | string[]> = {}
            searchParams.forEach((val, key) => {
              if (apiParams[key]) {
                if (Array.isArray(apiParams[key])) {
                  (apiParams[key] as string[]).push(val)
                } else {
                  apiParams[key] = [apiParams[key] as string, val]
                }
              } else {
                apiParams[key] = val
              }
            })
            // Force page param
            apiParams["page"] = nextPage.toString()

            const url = buildApiUrl("/api/v1/offenses", apiParams)
            const res = await fetch(url)
            if (!res.ok) throw new Error("API Error")
            const list: import("@/lib/types").OffensesResponse = await res.json()

            setAccumulatedOffenses((prev) => [...prev, ...list.offenses])
            setPagination(list.pagination)
            setRepos((prev) => ({ ...prev, ...(list.repos || {}) }))
            // @ts-ignore
            setArticles((prev) => ({ ...prev, ...(list.articles || {}) }))
            setCurrentPage(nextPage)
          } catch (error) {
            console.error("[v0] Error loading more:", error)
            setLoadMoreError(true)
          } finally {
            setIsLoadingMore(false)
          }
        }
      },
      {
        rootMargin: "400px",
      }
    )

    observer.observe(sentinelRef.current)

    return () => observer.disconnect()
  }, [
    currentPage,
    pagination,
    isLoadingMore,
    loading,
    searchParams,
    searchParamsString,
    loadMoreError,
  ])

  const toggleViewMode = useCallback(
    (mode: "list" | "charts" | "map") => {
      setViewMode(mode)
    },
    [setViewMode]
  )

  const handleGroupByChange = useCallback(
    (value: string) => {
      setGroupBy(value)
    },
    [setGroupBy]
  )

  const handleRetryLoadMore = useCallback(async () => {
    if (!pagination || isLoadingMore) return

    setLoadMoreError(false)
    setIsLoadingMore(true)
    try {
      const nextPage = currentPage + 1

      const apiParams: Record<string, string | string[]> = {}
      searchParams.forEach((val, key) => {
        if (apiParams[key]) {
          if (Array.isArray(apiParams[key])) {
            (apiParams[key] as string[]).push(val)
          } else {
            apiParams[key] = [apiParams[key] as string, val]
          }
        } else {
          apiParams[key] = val
        }
      })
      apiParams["page"] = nextPage.toString()

      const url = buildApiUrl("/api/v1/offenses", apiParams)
      const res = await fetch(url)
      if (!res.ok) throw new Error("API Error")
      const list: import("@/lib/types").OffensesResponse = await res.json()

      setAccumulatedOffenses((prev) => [...prev, ...list.offenses])
      setPagination(list.pagination)
      setRepos((prev) => ({ ...prev, ...(list.repos || {}) }))
      // @ts-ignore
      setArticles((prev) => ({ ...prev, ...(list.articles || {}) }))
      setCurrentPage(nextPage)
    } catch (error) {
      console.error("[v0] Error loading more:", error)
      setLoadMoreError(true)
    } finally {
      setIsLoadingMore(false)
    }
  }, [pagination, isLoadingMore, searchParams, currentPage])

  const handleFilterClick = useCallback(
    async (filterType: string, value: string) => {
      addFilter(filterType, value)
      window.scrollTo({ top: 0, behavior: "smooth" })
    },
    [addFilter]
  )

  const hasActiveFilters = Object.keys(searchParams).some(
    (key) => key !== "facet" && searchParams.getAll(key).length > 0
  )

  const handleRemoveFilter = (dimension: string, value: string) => {
    removeFilter(dimension, value)
  }

  const activeFilters = useMemo<ActiveFilter[]>(() => {
    return getActiveFilters(summaryData)
  }, [summaryData, getActiveFilters])

  // Removed sidebar rendering from here.
  // This component now represents only the content area.

  return (
    <div className={viewMode === "map" ? "p-4" : "border-card p-8 print:p-2"}>
      {loading && (viewMode !== "map" || !summaryData) ? (
        <div className="space-y-4">
          {Array.from({ length: 6 }).map((_, i) => (
            <OffenseCardSkeleton key={i} />
          ))}
        </div>
      ) : summaryData ? (
        <div>
          <Card className="mb-4 border-0 p-4 print:mb-2 print:p-2">
            <div className="flex flex-col gap-3">
              <div className="flex items-start justify-between gap-3">
                <div className="text-muted-foreground text-sm">
                  {summaryData.record_count.toLocaleString()} infracciones
                  encontradas
                </div>
                <div className="flex items-center gap-4">
                  <div className="bg-muted/50 flex items-center gap-1 rounded-lg border p-1">
                    <Link
                      href={getViewUrl("list")}
                      prefetch={false}
                      onClick={(e) => handleLinkClick(e, getViewUrl("list"))}
                      className={`flex items-center gap-1.5 rounded px-3 py-1.5 text-sm transition-colors ${viewMode === "list"
                        ? "bg-background text-foreground shadow-sm"
                        : "text-muted-foreground hover:text-foreground"
                        } ${isPending ? "pointer-events-none opacity-50" : ""}`}
                      aria-label="Vista de lista"
                    >
                      <List className="h-4 w-4" />
                      <span>Lista</span>
                    </Link>
                    <Link
                      href={getViewUrl("charts")}
                      prefetch={false}
                      onClick={(e) => handleLinkClick(e, getViewUrl("charts"))}
                      className={`flex items-center gap-1.5 rounded px-3 py-1.5 text-sm transition-colors ${viewMode === "charts"
                        ? "bg-background text-foreground shadow-sm"
                        : "text-muted-foreground hover:text-foreground"
                        } ${isPending ? "pointer-events-none opacity-50" : ""}`}
                      aria-label="Vista de gráficos"
                    >
                      <BarChart3 className="h-4 w-4" />
                      <span>Gráficos</span>
                    </Link>
                    <Link
                      href={getViewUrl("map")}
                      prefetch={false}
                      onClick={(e) => handleLinkClick(e, getViewUrl("map"))}
                      className={`flex items-center gap-1.5 rounded px-3 py-1.5 text-sm transition-colors ${viewMode === "map"
                        ? "bg-background text-foreground shadow-sm"
                        : "text-muted-foreground hover:text-foreground"
                        } ${isPending ? "pointer-events-none opacity-50" : ""}`}
                      aria-label="Vista de mapa"
                    >
                      <MapPin className="h-4 w-4" />
                      <span>Mapa</span>
                    </Link>
                  </div>
                  <div className="flex flex-col items-end gap-0.5">
                    <div className="text-foreground flex items-center gap-1.5 text-sm font-semibold">
                      <span className="text-muted-foreground text-xs font-normal">
                        Total
                      </span>
                      <span title="Recaudación total">
                        {formatUR(summaryData.total_ur)} UR
                      </span>
                    </div>
                    <div className="text-muted-foreground flex items-center gap-1.5 text-xs">
                      <span>Promedio</span>
                      <span
                        className="text-foreground font-medium"
                        title="Recaudación teórica promedio"
                      >
                        {formatUR(summaryData.avg_ur)} UR
                      </span>
                    </div>
                  </div>
                </div>
              </div>

              {activeFilters.length > 0 && (
                <div className="flex flex-wrap items-center gap-2">
                  {activeFilters.flatMap((filter) => {
                    const config = getDimensionConfig(filter.dimension)
                    const IconComponent = config.icon

                    return filter.values.map((valueLabel, index) => {
                      const isEmpty = valueLabel.value === ""
                      const displayText = isEmpty
                        ? config.empty
                        : valueLabel.label || valueLabel.value
                      const pillClassName = isEmpty
                        ? "inline-flex items-center gap-1 bg-muted/50 text-muted-foreground px-2 py-0.5 rounded-full text-xs italic"
                        : "inline-flex items-center gap-1 bg-primary/10 text-primary px-2 py-0.5 rounded-full text-xs"

                      const removeUrl = buildUrlWithoutFilter(
                        searchParams,
                        filter.dimension,
                        valueLabel.value
                      )

                      return (
                        <span
                          key={`${filter.dimension}-${valueLabel.value}-${index}`}
                          className={pillClassName}
                          title={`${config.label}: ${displayText}`}
                        >
                          <IconComponent className="h-3 w-3 flex-shrink-0" />
                          <span className="max-w-[200px] truncate">
                            {displayText}
                          </span>
                          <Link
                            href={removeUrl}
                            prefetch={false}
                            className="hover:bg-primary/20 ml-0.5 flex-shrink-0 rounded-full transition-colors"
                            aria-label={`Eliminar filtro ${displayText}`}
                          >
                            <X className="h-3 w-3" />
                          </Link>
                        </span>
                      )
                    })
                  })}
                  <Link
                    href={`?${(() => {
                      const p = new URLSearchParams()
                      searchParams
                        .getAll("facet")
                        .forEach((f) => p.append("facet", f))

                      // Preserve view and map parameters
                      const keysToPreserve = [
                        "view",
                        "groupBy",
                        "lat",
                        "lng",
                        "zoom",
                        "h3_index",
                      ]
                      keysToPreserve.forEach((key) => {
                        const value = searchParams.get(key)
                        if (value) p.set(key, value)
                      })

                      p.sort()
                      return p.toString()
                    })()}`}
                    prefetch={false}
                    className="text-muted-foreground hover:text-foreground text-xs underline transition-colors"
                  >
                    Limpiar todos
                  </Link>
                </div>
              )}
            </div>
          </Card>

          {viewMode === "map" && !isPending ? (
            <OffenseMap
              params={offensesParamsFromQueryParams(searchParams)}
              summaryData={summaryData}
            />
          ) : viewMode === "charts" && !isPending && chartData ? (
            <OffenseCharts
              dayOfWeekData={chartData.dayOfWeek}
              dayOfYearData={chartData.dayOfYear}
              timeOfDayData={chartData.timeOfDay}
              isLoading={chartsLoading}
              groupBy={groupBy}
              onGroupByChange={handleGroupByChange}
            />
          ) : isPending || (viewMode === "charts" && !chartData) ? (
            <div className="grid gap-2 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
              {Array.from({ length: 8 }).map((_, i) => (
                <OffenseCardSkeleton key={`skeleton-${i}`} />
              ))}
            </div>
          ) : (
            <>
              {accumulatedOffenses.length > 0 ? (
                <div className="space-y-4 print:space-y-0">
                  {accumulatedOffenses.map((offense, index) => (
                    <OffenseCard
                      key={`${offense.doc_source}#${offense.record_id}#${index}`}
                      offense={offense}
                      repos={repos}
                      articles={articles}
                      params={offensesParamsFromQueryParams(searchParams)}
                    />
                  ))}
                </div>
              ) : (
                <div className="text-muted-foreground py-12 text-center">
                  No se encontraron infracciones
                </div>
              )}

              {pagination && currentPage < pagination.total_pages && (
                <div ref={sentinelRef} className="py-8">
                  {isLoadingMore ? (
                    <div className="flex justify-center">
                      <OffenseCardSkeleton />
                    </div>
                  ) : (
                    loadMoreError && (
                      <div className="flex flex-col items-center gap-2">
                        <div className="text-destructive flex items-center gap-2">
                          <AlertCircle className="h-4 w-4" />
                          <span>Error al cargar más infracciones</span>
                        </div>
                        <Button
                          onClick={handleRetryLoadMore}
                          variant="outline"
                          size="default"
                        >
                          Reintentar
                        </Button>
                      </div>
                    )
                  )}
                </div>
              )}
            </>
          )}
        </div>
      ) : null}
    </div>
  )
}
