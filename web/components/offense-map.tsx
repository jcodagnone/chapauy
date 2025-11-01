/**
 * Copyright 2025 The ChapaUY Authors
 * SPDX-License-Identifier: Apache-2.0
 */

"use client"

import { useEffect, useState, useCallback, useRef } from "react"
import dynamic from "next/dynamic"
import { fetchMapData } from "@/lib/api/client"
import type { MapFeatureCollection, OffensesParams } from "@/lib/types"
import type { Map as LeafletMap } from "leaflet"
import * as h3 from "h3-js"

interface OffenseMapProps {
  params: OffensesParams
  summaryData: any
}

export const ZOOM_TO_PARENT_RES: Record<number, number> = {
  6: 1,
  7: 2,
  8: 2,
  9: 3,
  10: 3,
  11: 4,
  12: 5,
  13: 6,
  14: 7,
  15: 8,
  16: 8,
  17: 8,
  18: 8,
}

// Inverse mapping for initial zoom
export const RES_TO_ZOOM: Record<number, number> = {
  2: 6,
  3: 8,
  4: 11,
  5: 12,
  6: 13,
  7: 14,
  8: 15,
  9: 16,
}

const MapContent = dynamic(() => import("./offense-map-content"), {
  ssr: false,
  loading: () => (
    <div className="bg-muted flex h-[calc(100vh-12rem)] w-full items-center justify-center rounded-lg border">
      <p className="text-muted-foreground">Cargando mapa...</p>
    </div>
  ),
})

function parseFragment(
  hash: string
): { zoom: number; center: [number, number]; location?: string } | null {
  if (!hash || !hash.startsWith("#")) return null

  const parts = hash.slice(1).split("/")
  if (parts.length < 3) return null

  const zoom = parseInt(parts[0])
  const lat = parseFloat(parts[1])
  const lng = parseFloat(parts[2])
  const location = parts[3] ? decodeURIComponent(parts[3]) : undefined

  if (isNaN(zoom) || isNaN(lat) || isNaN(lng)) return null

  return { zoom, center: [lat, lng], location }
}

export function OffenseMap({ params, summaryData }: OffenseMapProps) {
  const initialState =
    typeof window !== "undefined" ? parseFragment(window.location.hash) : null

  // Calculate default center/zoom from summaryData if available
  let defaultCenter: [number, number] = [-34.9, -56.16]
  let defaultZoom = 12

  if (summaryData?.viewport_h3_index) {
    try {
      const [lat, lng] = h3.cellToLatLng(summaryData.viewport_h3_index)
      defaultCenter = [lat, lng]
      const res = h3.getResolution(summaryData.viewport_h3_index)
      defaultZoom = RES_TO_ZOOM[res] || 12
    } catch (e) {
      console.error("Error parsing viewport h3:", e)
    }
  }

  const [center, setCenter] = useState<[number, number]>(
    initialState?.center || defaultCenter
  )
  const [zoom, setZoom] = useState<number>(initialState?.zoom || defaultZoom)

  // Update view if summaryData changes (e.g. filter applied) AND no hash is present (or we want to force update?)
  // Actually, if the user filters, the URL params change, causing a re-render of SearchInterface,
  // which re-renders OffenseMap with new summaryData.
  // We want to update the map view when the filter changes, effectively "resetting" the view to the new data.
  // BUT we don't want to reset if the user just moved the map (which updates the hash).
  // The hash update logic is in MapEventHandler.

  // If params change, we might want to flyTo the new center.
  // However, OffenseMap state (center, zoom) is passed to MapContent only as initial props?
  // No, MapContent uses them as props. But MapContainer (inside MapContent) only respects center/zoom on initial render
  // unless we use a component to flyTo.
  // MapEventHandler handles updates?

  // Let's look at MapContent.
  // MapContainer props `center` and `zoom` are mutable? No, they are immutable after mount usually.
  // We need a component inside MapContainer to update view.
  // MapEventHandler calls `onViewChange` but doesn't seem to listen to props to update view?
  // Wait, MapEventHandler has `updateUrlHash` which reads map state.

  // We need to pass the "target" view to MapContent, and MapContent should update the map.
  // Currently MapContent just passes center/zoom to MapContainer.

  // If we want dynamic updates, we need to handle it.
  // For now, let's just set the initial state correctly.
  // If the user changes filters, the component might be re-mounted or updated.
  // If it's updated, we need to ensure the map moves.

  // We can use a useEffect here to update state when summaryData changes?
  // If we do that, we overwrite user's manual movement?
  // Only if the filter changed.

  const handleViewChange = useCallback(
    (newCenter: [number, number], newZoom: number) => {
      setCenter(newCenter)
      setZoom(newZoom)
    },
    []
  )

  return (
    <div className="relative h-[calc(100vh-12rem)] w-full overflow-hidden rounded-lg border bg-zinc-900">
      <MapContent
        center={center}
        zoom={zoom}
        params={params}
        initialLocation={initialState?.location}
        onViewChange={handleViewChange}
      />
    </div>
  )
}
