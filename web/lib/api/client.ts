/**
 * Copyright 2025 The ChapaUY Authors
 * SPDX-License-Identifier: Apache-2.0
 */

import type {
  OffensesListResponse,
  OffensesSummaryResponse,
  OffensesParams,
  Dimension,
  Facet,
  MapFeatureCollection,
} from "@/lib/types"

function buildQueryString(params: OffensesParams): string {
  const searchParams = new URLSearchParams()

  // Add predicates as query parameters
  params.predicates.forEach((predicate) => {
    predicate.values.forEach((value) => {
      searchParams.append(predicate.dimension, value)
    })
  })

  // Add page
  if (params.page) {
    searchParams.set("page", params.page.toString())
  }

  // Add facets
  if (params.facets) {
    params.facets.forEach((facet) => {
      searchParams.append("facet", facet)
    })
  }

  return searchParams.toString()
}

export async function fetchFacetValues(
  params: OffensesParams,
  dimension: Dimension
): Promise<Facet> {
  const queryString = buildQueryString(params)
  const response = await fetch(
    `/api/v1/suggest?${queryString}&dimension=${dimension}`
  )

  if (!response.ok) {
    throw new Error(`Failed to fetch facet values: ${response.statusText}`)
  }

  return response.json()
}

export async function fetchMapData(
  h3Index: string,
  params: OffensesParams
): Promise<MapFeatureCollection> {
  const searchParams = new URLSearchParams()

  // Add predicates as query parameters
  params.predicates.forEach((predicate) => {
    predicate.values.forEach((value) => {
      searchParams.append(predicate.dimension, value)
    })
  })

  const queryString = searchParams.toString()
  const url = `/api/v1/map/${h3Index}${queryString ? `?${queryString}` : ""}`

  const response = await fetch(url)

  if (!response.ok) {
    throw new Error(`Failed to fetch map data: ${response.statusText}`)
  }

  return response.json()
}
