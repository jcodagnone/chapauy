/**
 * Copyright 2025 The ChapaUY Authors
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * URL Utility Functions
 *
 * This file contains pure utility functions for parsing and manipulating URL parameters.
 *
 * It was split from `lib/search-params.ts` to resolve build errors where Server Components
 * (like `app/offenses/page.tsx`) were importing a file containing client-side hooks (`useSearchParams`, `usePathname`).
 * By extracting these pure functions here, we can share logic between Server and Client components
 * without violating the "use client" boundary.
 */
import { Dimension, type OffensesParams, type Facet } from "@/lib/types"
import { ReadonlyURLSearchParams } from "next/navigation"

export function offensesParamsFromQueryParams(
  paramsInput:
    | URLSearchParams
    | ReadonlyURLSearchParams
    | Record<string, string | string[] | undefined>
): OffensesParams {
  let searchParams: URLSearchParams

  if (
    paramsInput instanceof URLSearchParams ||
    (typeof ReadonlyURLSearchParams !== "undefined" &&
      paramsInput instanceof ReadonlyURLSearchParams)
  ) {
    searchParams = paramsInput as URLSearchParams
  } else {
    searchParams = new URLSearchParams()
    Object.entries(paramsInput).forEach(([key, value]) => {
      if (Array.isArray(value)) {
        value.forEach((v) => searchParams.append(key, v))
      } else if (value !== undefined) {
        searchParams.append(key, value)
      }
    })
  }
  const facets = searchParams
    .getAll("facet")
    .filter((f) => f !== "") as Dimension[]

  return {
    predicates: Object.values(Dimension)
      .map((dimension) => ({
        dimension: dimension as Dimension,
        values: searchParams.getAll(dimension).sort(),
      }))
      .filter((filter) => filter.values.length > 0),
    page: Number.parseInt(searchParams.get("page") || "1"),
    facets: facets.length > 0 ? facets : undefined,
  }
}

/**
 * Constructs a new URL query string by adding a specific filter value to the current search parameters.
 * Automatically resets the page number to 1.
 *
 * @param currentState The current URL search parameters.
 * @param dimension The dimension (filter key) to add.
 * @param value The value of the filter to add.
 * @returns A string representing the new query string (e.g., "?key=value").
 */
export function buildUrlWithFilter(
  currentState: URLSearchParams | ReadonlyURLSearchParams,
  dimension: string,
  value: string
): string {
  const params = new URLSearchParams(currentState.toString())
  const currentValues = params.getAll(dimension)

  if (!currentValues.includes(value)) {
    params.append(dimension, value)
    // Reset page to 1 on filter change
    params.delete("page")
  }

  params.sort()
  const search = params.toString()
  return search ? `?${search}` : ""
}

/**
 * Constructs a new URL query string by removing a specific filter value from the current search parameters.
 * Automatically resets the page number to 1.
 *
 * @param currentState The current URL search parameters.
 * @param dimension The dimension (filter key) to remove.
 * @param value The value of the filter to remove.
 * @returns A string representing the new query string.
 */
export function buildUrlWithoutFilter(
  currentState: URLSearchParams | ReadonlyURLSearchParams,
  dimension: string,
  value: string
): string {
  const params = new URLSearchParams(currentState.toString())
  const currentValues = params.getAll(dimension)
  const newValues = currentValues.filter((v) => v !== value)

  params.delete(dimension)
  newValues.forEach((v) => params.append(dimension, v))

  // Reset page to 1 on filter change
  params.delete("page")

  params.sort()
  const search = params.toString()
  return search ? `?${search}` : "?"
}

/**
 * Constructs a new URL query string by toggling a filter value:
 * - If the value exists for the dimension, it is removed.
 * - If the value does not exist, it is added.
 * Use this for toggle behavior in UI elements.
 *
 * @param currentState The current URL search parameters.
 * @param dimension The dimension (filter key) to toggle.
 * @param value The value of the filter.
 * @returns A string representing the new query string.
 */
export function buildUrlToggleFilter(
  currentState: URLSearchParams | ReadonlyURLSearchParams,
  dimension: string,
  value: string
): string {
  const params = new URLSearchParams(currentState.toString())
  const currentValues = params.getAll(dimension)

  if (currentValues.includes(value)) {
    return buildUrlWithoutFilter(currentState, dimension, value)
  } else {
    return buildUrlWithFilter(currentState, dimension, value)
  }
}

/**
 * Checks if a specific filter value is currently active within the provided parameters.
 *
 * @param params The parsed OffensesParams object.
 * @param dimension The dimension to check.
 * @param value The value to check for.
 * @returns True if the filter is active, false otherwise.
 */
export function isFilterActiveInParams(
  params: OffensesParams,
  dimension: string,
  value: string
): boolean {
  return params.predicates.some(
    (p) => p.dimension === dimension && p.values.includes(value)
  )
}

/**
 * Converts a structured `OffensesParams` object back into a standard `URLSearchParams` instance.
 * Useful for regenerating the URL or ensuring consistent parameter ordering.
 *
 * @param params The OffensesParams object to convert.
 * @returns A URLSearchParams instance containing the equivalent query parameters.
 */
export function toSearchParams(params: OffensesParams): URLSearchParams {
  const searchParams = new URLSearchParams()

  params.predicates.forEach((predicate) => {
    predicate.values.forEach((value) => {
      searchParams.append(predicate.dimension, value)
    })
  })

  if (params.page && params.page > 1) {
    searchParams.set("page", params.page.toString())
  }

  if (params.facets) {
    params.facets.forEach((facet) => {
      searchParams.append("facet", facet)
    })
  }

  return searchParams
}
