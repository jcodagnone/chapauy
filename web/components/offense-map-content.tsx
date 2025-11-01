/**
 * Copyright 2025 The ChapaUY Authors
 * SPDX-License-Identifier: Apache-2.0
 */

"use client"

import { useEffect, useState, useRef, useCallback } from "react"
import {
  MapContainer,
  TileLayer,
  GeoJSON,
  CircleMarker,
  Popup,
  useMap,
  useMapEvents,
  Marker,
} from "react-leaflet"
import Link from "next/link"
import * as h3 from "h3-js"
import { fetchMapData, fetchFacetValues } from "@/lib/api/client"
import type {
  MapFeatureCollection,
  OffensesParams,
  MapFeature,
  ClusterProperties,
  LocationProperties,
  FacetValue,
} from "@/lib/types"
import { Dimension } from "@/lib/types"
import { useOffenseSearchParams } from "@/lib/search-params"
import L from "leaflet"
import type { Map as LeafletMap, LatLngBounds, PathOptions } from "leaflet"
import "leaflet/dist/leaflet.css"

interface MapContentProps {
  center: [number, number]
  zoom: number
  params: OffensesParams
  initialLocation?: string
  onViewChange: (center: [number, number], zoom: number) => void
}

import { ZOOM_TO_PARENT_RES } from "./offense-map"

function getMarkerColor(
  count: number,
  isDark: boolean,
  isCluster: boolean = false
): string {
  if (isDark) {
    if (count > 1000) return "bg-red-900/80 border-red-500 text-red-100"
    if (count > 100) return "bg-orange-900/80 border-orange-500 text-orange-100"
    if (count > 10) return "bg-yellow-900/80 border-yellow-500 text-yellow-100"
    // Default
    if (isCluster) return "bg-amber-900/80 border-amber-500 text-amber-100"
    return "bg-slate-800/80 border-slate-400 text-slate-200"
  } else {
    if (count > 1000) return "bg-red-500 border-red-700 text-white"
    if (count > 100) return "bg-orange-400 border-orange-600 text-white"
    if (count > 10) return "bg-yellow-400 border-yellow-600 text-black"
    // Default
    if (isCluster) return "bg-yellow-200 border-yellow-400 text-yellow-900"
    return "bg-blue-100 border-blue-300 text-blue-800"
  }
}

interface LocationMarkerProps {
  feature: MapFeature
  searchParams: URLSearchParams
  selectedLocation?: string
  onSelect: (location: string) => void
  isDarkMode: boolean
}

function LocationMarker({
  feature,
  searchParams,
  selectedLocation,
  onSelect,
  isDarkMode,
}: LocationMarkerProps) {
  const props = feature.properties as LocationProperties
  const markerRef = useRef<L.Marker>(null)
  const [articleData, setArticleData] = useState<FacetValue[]>([])
  const [isLoading, setIsLoading] = useState(false)
  const loadedLocationRef = useRef<string | null>(null)

  useEffect(() => {
    if (selectedLocation === props.location && markerRef.current) {
      if (!markerRef.current.isPopupOpen()) {
        markerRef.current.openPopup()
      }

      // Fetch article codes if not already loaded for this location
      if (loadedLocationRef.current !== props.location) {
        // Set ref immediately to prevent double requests
        loadedLocationRef.current = props.location
        setIsLoading(true)

        // Construct params for the specific location
        const params: OffensesParams = {
          predicates: [],
        }

        // Reconstruct predicates from searchParams
        Object.values(Dimension).forEach((dim) => {
          const d = dim as Dimension
          const values = searchParams.getAll(d)
          if (values.length > 0) {
            params.predicates.push({ dimension: d, values })
          }
        })

        // Add location predicate
        params.predicates.push({
          dimension: Dimension.Location,
          values: [props.location],
        })

        // Determine which dimension to fetch
        // If we are already filtering by ArticleCode, show ArticleID (articles) instead
        const isFilteringByCode = searchParams.has(Dimension.ArticleCode)
        const targetDimension = isFilteringByCode
          ? Dimension.ArticleID
          : Dimension.ArticleCode

        fetchFacetValues(params, targetDimension)
          .then((facet) => {
            if (facet && facet.values) {
              let sorted = [...facet.values]

              // Check if any value is selected
              const hasSelected = sorted.some((v) => v.selected)

              if (hasSelected) {
                // If selected, show only selected values
                sorted = sorted.filter((v) => v.selected)
              } else {
                // Otherwise, sort by count descending and take top 5
                sorted = sorted.sort((a, b) => b.count - a.count).slice(0, 5)
              }

              setArticleData(sorted)
            }
          })
          .catch((err) => {
            console.error("Error fetching location details:", err)
            // Reset ref on error to allow retry if needed, or handle otherwise
            loadedLocationRef.current = null
          })
          .finally(() => {
            setIsLoading(false)
          })
      }
    } else if (markerRef.current && markerRef.current.isPopupOpen()) {
      markerRef.current.closePopup()
    }
  }, [selectedLocation, props.location, searchParams, isDarkMode])

  const createLocationHref = () => {
    const newParams = new URLSearchParams(searchParams.toString())
    const existing = newParams.getAll("location")
    if (!existing.includes(props.location)) {
      newParams.append("location", props.location)
    }
    newParams.delete("view")
    return `?${newParams.toString()}`
  }

  let lat, lng
  if (feature.geometry.type === "Point") {
    ;[lng, lat] = feature.geometry.coordinates as number[]
  } else {
    return null
  }

  const totalOffenses = (props as any).offenses || props.count || 0

  // Calculate others
  const displayedCount = articleData.reduce((acc, item) => acc + item.count, 0)
  const othersCount = totalOffenses - displayedCount

  const compactNumber = new Intl.NumberFormat("es-UY", {
    notation: "compact",
    maximumFractionDigits: 0,
  })

  const colorClass = getMarkerColor(totalOffenses, isDarkMode, false)

  const formattedOffenses = compactNumber.format(totalOffenses)

  const icon = L.divIcon({
    className: "custom-div-icon",
    html: `<div class="${colorClass} border-2 rounded-md w-6 h-6 flex items-center justify-center text-[10px] font-bold shadow-sm" title="${totalOffenses.toLocaleString("es-UY")} infracciones">${formattedOffenses}</div>`,
    iconSize: [24, 24],
    iconAnchor: [12, 12],
  })

  return (
    <Marker
      ref={markerRef}
      position={[lat, lng]}
      icon={icon}
      eventHandlers={{
        popupopen: () => onSelect(props.location),
        popupclose: () => {
          if (selectedLocation === props.location) {
            onSelect("")
          }
        },
      }}
    >
      <Popup
        key={isDarkMode ? "dark" : "light"}
        minWidth={250}
        className={isDarkMode ? "dark-popup" : ""}
      >
        <div className={`text-sm ${isDarkMode ? "text-gray-200" : ""}`}>
          <Link
            href={createLocationHref()}
            className={`mb-2 block cursor-pointer text-base font-medium hover:underline ${isDarkMode ? "text-blue-400" : "text-blue-600"}`}
            title="Filtrar por esta ubicación"
          >
            {props.location}
          </Link>

          <div className="mt-2">
            <div
              className={`mb-1 text-xs font-semibold tracking-wider uppercase ${isDarkMode ? "text-gray-400" : "text-gray-500"}`}
            >
              Principales Infracciones
            </div>
            {isLoading ? (
              <div className="animate-pulse space-y-2">
                <div
                  className={`h-4 w-3/4 rounded ${isDarkMode ? "bg-gray-700" : "bg-gray-200"}`}
                ></div>
                <div
                  className={`h-4 w-1/2 rounded ${isDarkMode ? "bg-gray-700" : "bg-gray-200"}`}
                ></div>
                <div
                  className={`h-4 w-2/3 rounded ${isDarkMode ? "bg-gray-700" : "bg-gray-200"}`}
                ></div>
              </div>
            ) : articleData.length > 0 ? (
              <ul className="space-y-1">
                {articleData.map((item) => {
                  const percent =
                    totalOffenses > 0 ? (item.count / totalOffenses) * 100 : 0
                  const percentFormatted = `${percent.toLocaleString("es-UY", { maximumFractionDigits: 1 })}%`

                  return (
                    <li
                      key={item.value}
                      className="flex items-start justify-between text-xs"
                    >
                      <span
                        className={`mr-2 flex-1 truncate ${isDarkMode ? "text-gray-300" : "text-gray-600"}`}
                        title={item.label || item.value}
                      >
                        {item.label || item.value}
                      </span>
                      <span
                        className={`rounded px-1.5 py-0.5 font-medium ${isDarkMode ? "bg-gray-700 text-gray-200" : "bg-gray-100 text-gray-900"}`}
                        title={percentFormatted}
                      >
                        {item.count.toLocaleString("es-UY")}
                      </span>
                    </li>
                  )
                })}
                {othersCount > 0 && (
                  <li className="flex items-start justify-between text-xs text-gray-500 italic">
                    <span className="mr-2 flex-1 truncate">Otros</span>
                    <span
                      className="px-1.5 py-0.5 font-medium"
                      title={`${((othersCount / totalOffenses) * 100).toLocaleString("es-UY", { maximumFractionDigits: 1 })}%`}
                    >
                      {othersCount.toLocaleString("es-UY")}
                    </span>
                  </li>
                )}
              </ul>
            ) : (
              <div className="text-xs text-gray-400 italic">
                No hay datos disponibles
              </div>
            )}
          </div>

          <div
            className="mt-3 flex items-center justify-between border-t pt-2"
            style={{ borderColor: isDarkMode ? "#374151" : "#e5e7eb" }}
          >
            <span
              className={`font-semibold ${isDarkMode ? "text-gray-300" : "text-gray-700"}`}
            >
              Total
            </span>
            <span
              className={`text-base font-bold ${isDarkMode ? "text-white" : "text-gray-900"}`}
            >
              {totalOffenses.toLocaleString("es-UY")}
            </span>
          </div>
        </div>
      </Popup>
    </Marker>
  )
}

function MapEventHandler({
  params,
  initialLocation,
  onViewChange,
  isDarkMode,
}: {
  params: OffensesParams
  initialLocation?: string
  onViewChange: (center: [number, number], zoom: number) => void
  isDarkMode: boolean
}) {
  const map = useMap()
  const [mapData, setMapData] = useState<Record<string, MapFeatureCollection>>(
    {}
  )
  const [selectedLocation, setSelectedLocation] = useState<string | undefined>(
    initialLocation
  )
  const loadedCellsRef = useRef<Set<string>>(new Set())
  const paramsRef = useRef<string>(JSON.stringify(params))
  const currentParamsRef = useRef(params)
  const { applyUpdates, searchParams } = useOffenseSearchParams()

  useEffect(() => {
    currentParamsRef.current = params
  }, [params])

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape") {
        setSelectedLocation(undefined)
        // Restore focus to the map container for keyboard navigation
        map.getContainer().focus()
      }
    }

    window.addEventListener("keydown", handleKeyDown)
    return () => window.removeEventListener("keydown", handleKeyDown)
  }, [map])

  const updateUrlHash = useCallback(() => {
    const center = map.getCenter()
    const zoom = map.getZoom()
    let hash = `#${zoom}/${center.lat.toFixed(5)}/${center.lng.toFixed(5)}`
    if (selectedLocation) {
      hash += `/${encodeURIComponent(selectedLocation)}`
    }
    // Debounce or check if changed?
    // For now, just replaceState. The loop fix should prevent spam.
    if (window.location.hash !== hash) {
      history.replaceState(null, "", hash)
      onViewChange([center.lat, center.lng], zoom)
    }
  }, [map, onViewChange, selectedLocation])

  const fetchAndDrawCell = useCallback(async (h3Index: string) => {
    if (loadedCellsRef.current.has(h3Index)) {
      return
    }
    loadedCellsRef.current.add(h3Index)

    try {
      const data = await fetchMapData(h3Index, currentParamsRef.current)
      if (data.features && data.features.length > 0) {
        setMapData((prev) => ({ ...prev, [h3Index]: data }))
      }
    } catch (error) {
      console.error("[v0] Error fetching H3 data:", error)
      loadedCellsRef.current.delete(h3Index)
    }
  }, [])

  const updateMap = useCallback(() => {
    const zoom = map.getZoom()
    const resolution = ZOOM_TO_PARENT_RES[zoom] || null

    if (!resolution) {
      setMapData({})
      loadedCellsRef.current.clear()
      console.log("[v0] Zoom level too low, clearing map.")
      return
    }

    const bounds = map.getBounds()
    const ne = bounds.getNorthEast()
    const sw = bounds.getSouthWest()
    const nw = bounds.getNorthWest()
    const se = bounds.getSouthEast()

    try {
      const cellsInView = h3.polygonToCells(
        [
          [ne.lat, ne.lng],
          [nw.lat, nw.lng],
          [sw.lat, sw.lng],
          [se.lat, se.lng],
          [ne.lat, ne.lng],
        ],
        resolution
      )

      setMapData((prev) => {
        const newData: Record<string, MapFeatureCollection> = {}
        Object.keys(prev).forEach((h3Index) => {
          const resOfCell = h3.getResolution(h3Index)
          if (resOfCell === resolution) {
            newData[h3Index] = prev[h3Index]
          } else {
            loadedCellsRef.current.delete(h3Index)
          }
        })
        return newData
      })

      cellsInView.forEach((h3Index) => {
        fetchAndDrawCell(h3Index)
      })
    } catch (error) {
      console.error("[v0] Error updating map:", error)
    }
  }, [map, fetchAndDrawCell])

  useEffect(() => {
    // hack to avoid refreshing map when opening a facet but not selection a value
    const predicates = JSON.stringify(params.predicates)
    if (paramsRef.current !== predicates) {
      console.log(
        "[v0] Filter params changed, clearing cache and refetching cells"
      )
      // Clear all cached data
      setMapData({})
      loadedCellsRef.current.clear()
      paramsRef.current = predicates

      // Refetch visible cells with new predicates
      updateMap()
    }
  }, [params, updateMap])

  useMapEvents({
    moveend: () => {
      updateMap()
      updateUrlHash()
    },
  })

  // Initial map load
  useEffect(() => {
    updateMap()
    // Focus map on load to enable keyboard navigation immediately
    map.getContainer().focus()
  }, [updateMap, map])

  // Update hash when selection changes
  useEffect(() => {
    updateUrlHash()
  }, [selectedLocation, updateUrlHash])

  const compactNumber = new Intl.NumberFormat("es-UY", {
    notation: "compact",
    maximumFractionDigits: 0,
  })

  return (
    <>
      {Object.entries(mapData).map(([h3Index, data]) =>
        data.features.map((feature, index) => {
          if (feature.properties.type === "cluster") {
            const props = feature.properties as ClusterProperties
            const offenses = props.offenses || 0

            // Use coordinates from the feature geometry (provided by backend)
            // Backend sends [lng, lat] in Point geometry
            let lat, lng
            if (feature.geometry.type === "Point") {
              ;[lng, lat] = feature.geometry.coordinates as number[]
            } else {
              // Fallback for safety
              ;[lat, lng] = h3.cellToLatLng(props.h3_index)
            }

            const colorClass = getMarkerColor(offenses, isDarkMode, true)

            const formattedOffenses = compactNumber.format(offenses)

            const icon = L.divIcon({
              className: "custom-div-icon",
              html: `<div class="${colorClass} border-2 rounded-full w-8 h-8 flex items-center justify-center text-xs font-bold shadow-md" title="${offenses.toLocaleString("es-UY")} infracciones&#10;H3: ${props.h3_index}">${formattedOffenses}</div>`,
              iconSize: [32, 32],
              iconAnchor: [16, 16],
            })

            return (
              <Marker
                key={`${h3Index}-${index}`}
                position={[lat, lng]}
                icon={icon}
                eventHandlers={{
                  click: () => {
                    console.log("Cluster H3:", props.h3_index)
                    navigator.clipboard
                      .writeText(props.h3_index)
                      .catch(console.error)
                    map.flyTo([lat, lng], map.getZoom() + 2)
                  },
                }}
              />
            )
          } else if (feature.properties.type === "location") {
            const props = feature.properties as LocationProperties
            let lat, lng
            if (feature.geometry.type === "Point") {
              ;[lng, lat] = feature.geometry.coordinates as number[]
            } else {
              return null
            }

            const createLocationHref = () => {
              const newParams = new URLSearchParams(searchParams.toString())
              const existing = newParams.getAll("location")
              if (!existing.includes(props.location)) {
                newParams.append("location", props.location)
              }
              newParams.delete("view")
              return `?${newParams.toString()}`
            }

            return (
              <LocationMarker
                key={`${h3Index}-${index}`}
                feature={feature}
                searchParams={searchParams}
                selectedLocation={selectedLocation}
                onSelect={(loc) => {
                  setSelectedLocation(loc)
                  // Trigger URL update immediately when selection changes
                  // We need to wait for state update, but we can also force updateUrlHash
                  // However, updateUrlHash depends on selectedLocation state.
                  // So we rely on useEffect below.
                }}
                isDarkMode={isDarkMode}
              />
            )
          }
          return null
        })
      )}
    </>
  )
}

function MapUpdater({
  center,
  zoom,
}: {
  center: [number, number]
  zoom: number
}) {
  const map = useMap()

  useEffect(() => {
    map.setView(center, zoom)
  }, [center, zoom, map])

  return null
}

export default function MapContent({
  center,
  zoom,
  params,
  initialLocation,
  onViewChange,
}: MapContentProps) {
  const [isDarkMode, setIsDarkMode] = useState(true)

  const toggleDarkMode = useCallback(() => {
    setIsDarkMode((prev) => !prev)
  }, [])

  return (
    <>
      <MapContainer
        center={center}
        zoom={zoom}
        style={{ height: "100%", width: "100%", background: "#242424" }}
        minZoom={6}
        maxZoom={18}
      >
        <TileLayer
          key={isDarkMode ? "dark" : "light"}
          attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
          url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
          className={
            isDarkMode
              ? "[filter:invert(100%)_hue-rotate(180deg)_brightness(95%)_contrast(90%)]"
              : ""
          }
        />

        <MapEventHandler
          params={params}
          initialLocation={initialLocation}
          onViewChange={onViewChange}
          isDarkMode={isDarkMode}
        />
        <MapUpdater center={center} zoom={zoom} />
      </MapContainer>
      <div
        className="leaflet-bottom leaflet-left"
        style={{
          position: "absolute",
          bottom: "20px",
          left: "10px",
          pointerEvents: "auto",
          zIndex: 1000,
        }}
      >
        <div className="leaflet-control leaflet-bar">
          <button
            onClick={toggleDarkMode}
            className="flex cursor-pointer items-center gap-1 rounded border border-gray-400 bg-white px-2 py-1 text-xs font-semibold text-gray-800 shadow hover:bg-gray-100"
            title="Toggle Dark Mode"
          >
            {isDarkMode ? (
              <>
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  width="14"
                  height="14"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  strokeWidth="2"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                >
                  <circle cx="12" cy="12" r="5" />
                  <line x1="12" y1="1" x2="12" y2="3" />
                  <line x1="12" y1="21" x2="12" y2="23" />
                  <line x1="4.22" y1="4.22" x2="5.64" y2="5.64" />
                  <line x1="18.36" y1="18.36" x2="19.78" y2="19.78" />
                  <line x1="1" y1="12" x2="3" y2="12" />
                  <line x1="21" y1="12" x2="23" y2="12" />
                  <line x1="4.22" y1="19.78" x2="5.64" y2="18.36" />
                  <line x1="18.36" y1="5.64" x2="19.78" y2="4.22" />
                </svg>
                <span>Light</span>
              </>
            ) : (
              <>
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  width="14"
                  height="14"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  strokeWidth="2"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                >
                  <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z" />
                </svg>
                <span>Dark</span>
              </>
            )}
          </button>
        </div>
      </div>
    </>
  )
}
