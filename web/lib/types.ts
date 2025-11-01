/**
 * Copyright 2025 The ChapaUY Authors
 * SPDX-License-Identifier: Apache-2.0
 */

// API Types for ChapaUY Traffic Offenses

export interface Repo {
  name: string
}

export interface Offense {
  doc_source: string
  doc_id: string
  doc_date: string
  country: string
  adm_division: string
  vehicle_type: string
  mercosur_format: boolean
  repo_id: number
  record_id: number
  vehicle: string
  time: string
  location: string
  display_location?: string // Added optional display_location for UI display
  id: string
  description: string
  article_id?: string[]
  article_code?: string[]
  ur: number
  point?: {
    lat: number
    lng: number
  }
  error?: string
}

export interface OffenseDocument {
  db_id: number
  doc_id: string
  doc_date: string
  doc_source: string
  records: number // count(*)
  ur: number // sum(ur)
  errors: number // sum("error" IS NOT NULL)
}

export interface ValueLabel {
  value: string
  label?: string // Optional display label
}

export interface FacetValue extends ValueLabel {
  count: number
  selected: boolean
}

export enum Dimension {
  Database = "database",
  Year = "year",
  Country = "country",
  VehicleType = "vehicle_type",
  Vehicle = "vehicle",
  DocSource = "doc_source",
  Location = "location",
  Description = "description",
  ArticleID = "article_id",
  ArticleCode = "article_code",
  Features = "features",
  Date = "date",
}

export enum SidebarMode {
  Offenses = "offenses",
  Documents = "documents",
}

export enum SortBy {
  Document = "document",
  Vehicle = "vehicle",
}

export interface Facet {
  dimension: Dimension
  values: FacetValue[]
  total_values: number // Total count of distinct values available for this dimension
}

export interface InPredicate {
  dimension: Dimension
  values: string[]
}

export interface ActiveFilter {
  dimension: Dimension
  values: ValueLabel[]
}

export interface OffensesListResponse {
  offenses: Offense[]
  pagination: {
    current_page: number
    total_pages: number
  }
  repos: Record<string, Repo>
  articles?: Record<string, string>
  chartData?: {
    dayOfWeek: ChartData | null
    dayOfYear: ChartData | null
    timeOfDay: ChartData | null
  }
}

export interface OffensesSummaryResponse {
  avg_ur: number
  facets: Facet[]
  record_count: number
  total_ur: number
  viewport_h3_index?: string
}

// Combined response for convenience (used internally by components)
export interface OffensesResponse {
  offenses: Offense[]
  pagination: {
    current_page: number
    total_pages: number
  }
  repos: Record<string, Repo>
  summary: {
    avg_ur: number
    facets: Facet[]
    record_count: number
    total_ur: number
    viewport_h3_index?: string
  }
  articles?: Record<string, string>
  chartData?: {
    dayOfWeek: ChartData | null
    dayOfYear: ChartData | null
    timeOfDay: ChartData | null
  }
}

export interface OffensesParams {
  predicates: InPredicate[]
  page?: number
  per_page?: number
  facets?: Dimension[] // Added facets parameter to specify which dimensions to compute
}

export type ChartData = Record<string, Record<string, number>>

export interface MapFeature {
  type: "Feature"
  geometry: {
    type: "Polygon" | "Point"
    coordinates: number[][] | number[][][][] | number[]
  }
  properties: ClusterProperties | LocationProperties
}

export interface ClusterProperties {
  type: "cluster"
  h3_index: string
  offenses: number
  locations: number
  centroid: [number, number]
}

export interface LocationProperties {
  type: "location"
  location: string
  count: number
}

export interface MapFeatureCollection {
  type: "FeatureCollection"
  features: MapFeature[]
}
