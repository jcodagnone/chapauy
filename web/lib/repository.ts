/**
 * Copyright 2025 The ChapaUY Authors
 * SPDX-License-Identifier: Apache-2.0
 */

import { getDuckDB, waitForDB } from "./duckdb"
import { getDBName, databases, countryDisplay } from "./db-refs"
import {
  OffensesParams,
  OffenseDocument,
  OffensesListResponse,
  InPredicate,
  SortBy,
  Repo,
  Dimension,
  Facet,
  FacetValue,
} from "@/lib/types"
import * as h3 from "h3-js"
import { unstable_cache, cacheLife } from "next/cache"
import { Database } from "duckdb"

function dbAll(db: Database, query: string, args: any[]): Promise<any[]> {
  return new Promise((resolve, reject) => {
    db.all(query, ...args, (err, rows) => {
      if (err) reject(err)
      else resolve(rows)
    })
  })
}

// We sighly change sorting to match source document whe filtering by a document.
// this let user compare side by side the extraction of records from the source document.
export function determineSortBy(predicates: InPredicate[]): SortBy {
  return predicates.length === 1 &&
    predicates[0].dimension === Dimension.DocSource
    ? SortBy.Document
    : SortBy.Vehicle
}

// Helper to get column expression for a dimension
function getColumnExpr(dim: Dimension): string {
  switch (dim) {
    case Dimension.Year:
      return "time_year"
    case Dimension.VehicleType:
      return "vehicle_type"
    case Dimension.Country:
      return "vehicle_country"
    case Dimension.Database:
      return "db_id"
    case Dimension.DocSource:
      return "doc_source"
    case Dimension.Vehicle:
      return "vehicle"
    case Dimension.Location:
      return "location"
    case Dimension.Description:
      return "description"
    case Dimension.ArticleCode:
      return "article_codes"
    case Dimension.ArticleID:
      return "article_ids"
    case Dimension.Features:
      return "features"
    case Dimension.Date:
      return "CAST(time AS DATE)"
    default:
      return dim
  }
}

// Helper to build WHERE clause
function buildWhereClause(
  predicates: InPredicate[],
  exclude?: Dimension
): { where: string; args: any[] } {
  const clauses: string[] = []
  const args: any[] = []

  for (const p of predicates) {
    if (exclude && p.dimension === exclude) {
      continue
    }

    if (!p.values || p.values.length === 0) {
      continue
    }

    // Status dimension is special (virtual column)
    if (p.dimension === Dimension.Features) {
      const statusClauses: string[] = []
      for (const v of p.values) {
        if (v === "with_error") statusClauses.push("error IS NOT NULL")
        else if (v === "no_error") statusClauses.push("error IS NULL")
        else if (v === "with_ur")
          statusClauses.push("(ur IS NOT NULL AND ur != 0)")
        else if (v === "no_ur") statusClauses.push("(ur IS NULL OR ur = 0)")
      }
      if (statusClauses.length > 0) {
        clauses.push(`(${statusClauses.join(" OR ")})`)
      }
      continue
    }

    const column = getColumnExpr(p.dimension)
    const placeholders = p.values.map(() => "?").join(",")

    switch (p.dimension) {
      case Dimension.Database:
      case Dimension.Year:
      case Dimension.Country:
      case Dimension.Vehicle:
      case Dimension.DocSource:
      case Dimension.Location:
      case Dimension.Date:
        clauses.push(`${column} IN (${placeholders})`)
        args.push(...p.values)
        break
      case Dimension.Description:
        const likeClauses = p.values.map(() => `${column} LIKE ?`)
        clauses.push(`(${likeClauses.join(" OR ")})`)
        args.push(...p.values.map((v) => `%${v}%`))
        break
      case Dimension.VehicleType:
        const typeClauses = p.values.map((v) => {
          if (v === "") return `${column} IS NULL`
          return `${column} = ?`
        })
        clauses.push(`(${typeClauses.join(" OR ")})`)
        p.values.forEach((v) => {
          if (v !== "") args.push(v)
        })
        break
      case Dimension.ArticleCode:
      case Dimension.ArticleID:
        const listClauses = p.values.map(() => `list_contains(${column}, ?)`)
        clauses.push(`(${listClauses.join(" OR ")})`)
        args.push(...p.values)
        break
      default:
        throw new Error(`Unhandled dimension in where clause: ${p.dimension}`)
    }
  }

  return {
    where: clauses.length > 0 ? clauses.join(" AND ") : "",
    args,
  }
}

export async function getOffensesSummary(
  predicates: InPredicate[],
  dimension: Dimension | null
): Promise<any[]> {
  "use cache"
  cacheLife("days")

  await waitForDB()
  const db = getDuckDB()
  const { where, args } = buildWhereClause(predicates || [])

  let query = "SELECT "
  if (dimension) {
    query += `${getColumnExpr(dimension)} as value, `
  }

  query += `
    COUNT(*) as count,
    SUM(ur) as ur_total,
    AVG(ur) as ur_avg
  `

  if (dimension === Dimension.Database) {
    query += `,
      COUNT(DISTINCT doc_source) as doc_count,
      MAX(doc_date) as last_doc_date
    `
  }

  // for viewport we want to choose the resolution that has one cell only
  // so we know if max = min then we have only one cell
  query += `,
  MIN(h3_res1) as min_h3_1, MAX(h3_res1) as max_h3_1,
      MIN(h3_res2) as min_h3_2, MAX(h3_res2) as max_h3_2,
      MIN(h3_res3) as min_h3_3, MAX(h3_res3) as max_h3_3,
      MIN(h3_res4) as min_h3_4, MAX(h3_res4) as max_h3_4,
      MIN(h3_res5) as min_h3_5, MAX(h3_res5) as max_h3_5,
      MIN(h3_res6) as min_h3_6, MAX(h3_res6) as max_h3_6,
      MIN(h3_res7) as min_h3_7, MAX(h3_res7) as max_h3_7,
      MIN(h3_res8) as min_h3_8, MAX(h3_res8) as max_h3_8
    `
  query += " FROM offenses"

  if (where) {
    query += ` WHERE ${where}`
  }

  if (dimension) {
    query += ` GROUP BY ${getColumnExpr(dimension)} ORDER BY ${getColumnExpr(
      dimension
    )}`
  }

  return dbAll(db, query, args).then((rows) => {
    // Post-processing for labels and BigInt conversion
    return rows.map((row: any) => {
      const newRow = { ...row }
      if (dimension === Dimension.Database) {
        newRow.label = getDBName(parseInt(row.value))
      }
      // Convert BigInts to Number
      newRow.count = Number(row.count)
      newRow.ur_total = Number(row.ur_total)
      newRow.ur_avg = Number(row.ur_avg)
      if (newRow.doc_count) newRow.doc_count = Number(row.doc_count)
      // Check resolutions from finest available (8) to coarsest (2)
      const resolutionsToCheck = [8, 7, 6, 5, 4, 3, 2]
      const last = resolutionsToCheck[resolutionsToCheck.length - 1]
      for (const res of resolutionsToCheck) {
        const min = row[`min_h3_${res}`]
        const max = row[`max_h3_${res}`]
        if (min && max && (min === max || res === last)) {
          // Found common parent!
          // Convert BigInt to hex string
          newRow.viewport_h3_index = BigInt(min).toString(16)
          break
        }
      }

      return newRow
    })
  })
}

export const getArticles = unstable_cache(
  async (): Promise<{
    byId: Record<string, string>
    byCode: Record<string, string>
  }> => {
    await waitForDB()
    const db = getDuckDB()

    return dbAll(db, "SELECT * FROM articles", []).then((rows) => {
      const byId: Record<string, string> = {}
      const byCode: Record<string, string> = {}
      rows.forEach((r: any) => {
        byId[r.id] = `${r.id} - ${r.title}`
        byCode[r.code] = `${r.code} - ${r.title}`
      })
      return { byId, byCode }
    })
  },
  ["articles-cache-v2"],
  { revalidate: 3600 }
)

export async function getDimensionResults(
  predicates: InPredicate[],
  dimensions: Dimension[],
  queries?: Record<string, string> // Map of dimension -> search query
): Promise<Facet[]> {
  "use cache"
  cacheLife("days")

  await waitForDB()
  const db = getDuckDB()

  // We will build two big UNION ALL queries:
  // 1. To get the top N values for each dimension (Values Query)
  // 2. To get the total count of distinct values for each dimension (Totals Query)

  const valueQueries: string[] = []
  const valueArgs: any[] = []

  const totalQueries: string[] = []
  const totalArgs: any[] = []

  // Keep track of which dimensions we are processing to map results back
  // Features dimension is special, it essentially contributes multiple "virtual" dimensions or values
  // but here it behaves as one Dimension.Features with static values.

  for (const dim of dimensions) {
    try {
      // Special handling for Status dimension: Unpivot into unified structure
      if (dim === Dimension.Features) {
        const { where, args } = buildWhereClause(predicates || [], dim)
        const whereClause = where ? `WHERE ${where}` : ""

        const featureParts = [
          { label: "with_error", expr: "error IS NOT NULL" },
          { label: "no_error", expr: "error IS NULL" },
          { label: "with_ur", expr: "ur IS NOT NULL AND ur != 0" },
          { label: "no_ur", expr: "ur IS NULL OR ur = 0" },
        ]

        featureParts.forEach((part) => {
          const clause = whereClause
            ? `${whereClause} AND ${part.expr}`
            : `WHERE ${part.expr}`
          valueQueries.push(`
            SELECT 
              '${Dimension.Features}' as dimension,
              '${part.label}' as value,
              COUNT(*) as count
            FROM offenses
            ${clause}
          `)
          valueArgs.push(...args)
        })
        continue
      }

      const { where, args } = buildWhereClause(predicates || [], dim)
      const column = getColumnExpr(dim)
      const searchQuery = queries ? queries[dim] : undefined

      let selectVal = ""
      let fromClause = ""

      // Base query construction
      if (dim === Dimension.ArticleID || dim === Dimension.ArticleCode) {
        selectVal = "value"
        fromClause = `(SELECT UNNEST(${column}) as value FROM offenses) sub`
      } else {
        selectVal = `${column}::VARCHAR` // Force varchar for union compatibility
        fromClause = "offenses"
      }

      // Apply filters
      let finalWhere = where
      const finalArgs = [...(args || [])]

      // Add search query filter if present
      if (searchQuery) {
        const qLower = searchQuery.toLowerCase()
        if (dim === Dimension.Database) {
          const matchingIds = databases
            .filter((db) => db.name.toLowerCase().includes(qLower))
            .map((db) => db.id.toString())
          if (matchingIds.length > 0) {
            const placeholders = matchingIds.map(() => "?").join(",")
            const clause = `${column} IN (${placeholders})`
            finalWhere = finalWhere ? `${finalWhere} AND ${clause}` : clause
            finalArgs.push(...matchingIds)
          } else {
            finalWhere = finalWhere ? `${finalWhere} AND 1=0` : "1=0"
          }
        } else if (dim === Dimension.Country) {
          const matchingCodes = Object.entries(countryDisplay)
            .filter(
              ([k, v]) =>
                v.toLowerCase().includes(qLower) ||
                k.toLowerCase().includes(qLower)
            )
            .map(([k]) => k)
          if (matchingCodes.length > 0) {
            const placeholders = matchingCodes.map(() => "?").join(",")
            const clause = `${column} IN (${placeholders})`
            finalWhere = finalWhere ? `${finalWhere} AND ${clause}` : clause
            finalArgs.push(...matchingCodes)
          } else {
            finalWhere = finalWhere ? `${finalWhere} AND 1=0` : "1=0"
          }
        } else {
          const searchCol =
            dim === Dimension.ArticleID || dim === Dimension.ArticleCode
              ? "value"
              : column
          const searchClause = `${searchCol} ILIKE ?`
          finalWhere = finalWhere
            ? `${finalWhere} AND ${searchClause}`
            : searchClause
          finalArgs.push(`%${searchQuery}%`)
        }
      }

      // Construct Value Query Part
      const whereSql = finalWhere ? `WHERE ${finalWhere}` : ""
      let queryPart = ""
      let totalPart = ""

      // Standard Dimensions
      let limit: number
      switch (dim) {
        case Dimension.Vehicle:
          limit = 30
          break
        case Dimension.ArticleCode:
          limit = 30
          break
        default:
          limit = 10
      }

      if (dim === Dimension.ArticleID || dim === Dimension.ArticleCode) {
        // Article Logic
        const predWhere = where ? `WHERE ${where}` : ""
        const searchClause = searchQuery ? `WHERE value::VARCHAR ILIKE ?` : ""
        let innerSql = `SELECT UNNEST(${column}) as value FROM offenses ${predWhere}`


        queryPart = `
            SELECT 
              '${dim}' as dimension,
              value::VARCHAR as value,
              COUNT(*) as count
            FROM (${innerSql}) sub
            ${searchClause}
            GROUP BY value
            ORDER BY count DESC, value ASC
            LIMIT ${limit}
         `
        totalPart = `
            SELECT
              '${dim}' as dimension,
              COUNT(DISTINCT value) as total
            FROM (${innerSql}) sub
            ${searchClause}
         `
        valueArgs.push(...(args || []))
        if (searchQuery) valueArgs.push(`%${searchQuery}%`)
        totalArgs.push(...(args || []))
        if (searchQuery) totalArgs.push(`%${searchQuery}%`)
      } else {

        queryPart = `
            SELECT
              '${dim}' as dimension,
              ${selectVal} as value,
              COUNT(*) as count
            FROM offenses
            ${whereSql}
            GROUP BY value
            ORDER BY count DESC, value ASC
            LIMIT ${limit}
         `
        totalPart = `
            SELECT 
              '${dim}' as dimension,
              COUNT(DISTINCT COALESCE(${selectVal}, '')) as total
            FROM offenses
            ${whereSql}
         `
        valueArgs.push(...finalArgs)
        totalArgs.push(...finalArgs)
      }

      valueQueries.push(`(${queryPart})`)
      totalQueries.push(`(${totalPart})`)
    } catch (e) {
      console.error("Error processing dim:", dim, e)
    }
  }

  // Execute Batch 1: Values
  if (valueQueries.length === 0) {
    return []
  }
  const valuesSql = valueQueries.join(" UNION ALL ")
  const valuesPromise = dbAll(db, valuesSql, valueArgs)

  // Execute Batch 2: Totals
  let totalsPromise: Promise<any[]> = Promise.resolve([])
  if (totalQueries.length > 0) {
    const totalsSql = totalQueries.join(" UNION ALL ")
    totalsPromise = dbAll(db, totalsSql, totalArgs)
  }

  const [valueRows, totalRows] = await Promise.all([
    valuesPromise,
    totalsPromise,
  ])

  // Post-process and group by dimension
  // We need to map the flat rows back to the Facet[] structure
  const resultsMap: Record<string, Facet> = {}

  // Initialize results
  dimensions.forEach((dim) => {
    resultsMap[dim] = {
      dimension: dim,
      values: [],
      total_values: 0,
    }
  })

  // Map Totals
  totalRows.forEach((row: any) => {
    if (resultsMap[row.dimension]) {
      resultsMap[row.dimension].total_values = Number(row.total || 0)
    }
  })

  // Map Values
  valueRows.forEach((row: any) => {
    const dim = row.dimension as Dimension
    if (!resultsMap[dim]) return

    let label = row.value
    if (dim === Dimension.Database) {
      label = getDBName(parseInt(row.value))
    } else if (dim === Dimension.Country) {
      label = countryDisplay[row.value] || row.value
    } else if (dim === Dimension.Features) {
      // Map back internal values to labels
      const labels: Record<string, string> = {
        with_error: "Con Error",
        no_error: "Sin Error",
        with_ur: "Con UR",
        no_ur: "Sin UR",
      }
      label = labels[row.value] || row.value
    }

    resultsMap[dim].values.push({
      value: row.value === null ? "" : String(row.value),
      label: label === null ? "" : String(label),
      count: Number(row.count),
      selected: false,
    })
  })

  // Post-processing: Predicates (Mark Selected) & Hydrate Articles
  const articleDims = dimensions.filter(
    (d) => d === Dimension.ArticleID || d === Dimension.ArticleCode
  )
  let articles: any = null
  if (articleDims.length > 0) {
    articles = await getArticles()
  }

  // Final pass for selection and hydration
  for (const dim of dimensions) {
    const facet = resultsMap[dim]
    const predicate = predicates.find((p) => p.dimension === dim)

    // Hydrate articles
    if (dim === Dimension.ArticleID && articles) {
      facet.values.forEach((v) => (v.label = articles.byId[v.value] || v.value))
    } else if (dim === Dimension.ArticleCode && articles) {
      facet.values.forEach(
        (v) => (v.label = articles.byCode[v.value] || v.value)
      )
    }

    // Mark selected
    if (predicate) {
      const existing = new Set(facet.values.map((v) => v.value))
      facet.values.forEach((v) => {
        if (predicate.values.includes(v.value)) v.selected = true
      })

      // Add missing selected
      predicate.values.forEach((val) => {
        if (!existing.has(val)) {
          let label = val
          if (dim === Dimension.Database) label = getDBName(parseInt(val))
          else if (dim === Dimension.Country) label = countryDisplay[val] || val
          else if (dim === Dimension.ArticleID && articles)
            label = articles.byId[val] || val
          else if (dim === Dimension.ArticleCode && articles)
            label = articles.byCode[val] || val
          else if (dim === Dimension.Features) {
            const labels: Record<string, string> = {
              with_error: "Con Error",
              no_error: "Sin Error",
              with_ur: "Con UR",
              no_ur: "Sin UR",
            }
            label = labels[val] || val
          }

          facet.values.push({
            value: val,
            label: label,
            count: 0,
            selected: true,
          })
        }
      })
    }

    // Sort
    facet.values.sort((a, b) => {
      if (a.selected !== b.selected) return a.selected ? -1 : 1
      if (a.count !== b.count) return b.count - a.count
      return (a.label || "").localeCompare(b.label || "")
    })
  }

  // Return in order of requested dimensions
  return dimensions.map((d) => resultsMap[d])
}

export async function getDocumentFacets(
  predicates: InPredicate[],
  dimensions: Dimension[]
): Promise<Facet[]> {
  "use cache"
  cacheLife("days")

  await waitForDB()
  const db = getDuckDB()

  const valueQueries: string[] = []
  const valueArgs: any[] = []

  const totalQueries: string[] = []
  const totalArgs: any[] = []

  // Document Uniqueness is defined by (db_id, doc_id)

  for (const dim of dimensions) {
    if (dim === Dimension.Features) {
      const { where, args } = buildWhereClause(predicates || [], dim)
      const whereClause = where ? `WHERE ${where}` : ""
      const distinctDoc = "CAST(db_id AS VARCHAR) || '-' || doc_id"

      const featureParts = [
        { label: "with_error", expr: "error IS NOT NULL" },
        { label: "no_error", expr: "error IS NULL" },
        { label: "with_ur", expr: "ur IS NOT NULL AND ur != 0" },
        { label: "no_ur", expr: "ur IS NULL OR ur = 0" },
      ]

      featureParts.forEach((part) => {
        const clause = whereClause
          ? `${whereClause} AND ${part.expr}`
          : `WHERE ${part.expr}`
        valueQueries.push(`
          SELECT 
            '${Dimension.Features}' as dimension,
            '${part.label}' as value,
            COUNT(DISTINCT CASE WHEN ${part.expr} THEN ${distinctDoc} END) as count
          FROM offenses
          ${clause}
        `)
        valueArgs.push(...args)
      })
      continue
    }

    const { where, args } = buildWhereClause(predicates || [], dim)
    const column = getColumnExpr(dim)
    const distinctDocExpr = "CAST(db_id AS VARCHAR) || '-' || doc_id"

    // Select Facet Values
    const queryPart = `
      SELECT 
        '${dim}' as dimension,
        ${column}::VARCHAR as value, 
        COUNT(DISTINCT ${distinctDocExpr}) as count 
      FROM offenses 
      ${where ? `WHERE ${where}` : ""}
      GROUP BY value 
      ORDER BY count DESC, value ASC
      LIMIT 10
    `
    // Total Unique Values Count
    const totalPart = `
      SELECT 
        '${dim}' as dimension,
        COUNT(DISTINCT COALESCE(${column}::VARCHAR, '')) as total
      FROM offenses
      ${where ? `WHERE ${where}` : ""}
    `

    valueQueries.push(`(${queryPart})`)
    totalQueries.push(`(${totalPart})`)
    valueArgs.push(...args)
    totalArgs.push(...args)
  }

  // Execute Batch
  if (valueQueries.length === 0) {
    return []
  }
  const valuesSql = valueQueries.join(" UNION ALL ")
  const valuesPromise = dbAll(db, valuesSql, valueArgs)

  let totalsPromise: Promise<any[]> = Promise.resolve([])
  if (totalQueries.length > 0) {
    const totalsSql = totalQueries.join(" UNION ALL ")
    totalsPromise = dbAll(db, totalsSql, totalArgs)
  }

  const [valueRows, totalRows] = await Promise.all([
    valuesPromise,
    totalsPromise,
  ])

  // Post-process
  const resultsMap: Record<string, Facet> = {}
  dimensions.forEach((dim) => {
    resultsMap[dim] = {
      dimension: dim,
      values: [],
      total_values: 0,
    }
  })

  totalRows.forEach((row: any) => {
    if (resultsMap[row.dimension]) {
      resultsMap[row.dimension].total_values = Number(row.total || 0)
    }
  })

  valueRows.forEach((row: any) => {
    const dim = row.dimension as Dimension
    if (!resultsMap[dim]) return

    let label = row.value
    if (dim === Dimension.Database) {
      label = getDBName(parseInt(row.value))
    } else if (dim === Dimension.Features) {
      const labels: Record<string, string> = {
        with_error: "Con Errores",
        no_error: "Sin Errores",
        with_ur: "Con UR",
        no_ur: "Sin UR",
      }
      label = labels[row.value] || row.value
    }

    resultsMap[dim].values.push({
      value: row.value === null ? "" : String(row.value),
      label: label === null ? "" : String(label),
      count: Number(row.count),
      selected: false,
    })
  })

  // Mark selected
  for (const dim of dimensions) {
    const facet = resultsMap[dim]
    const predicate = predicates.find((p) => p.dimension === dim)
    if (predicate) {
      const existing = new Set(facet.values.map((v) => v.value))
      facet.values.forEach((v) => {
        if (predicate.values.includes(v.value)) v.selected = true
      })

      // Add missing
      predicate.values.forEach((val) => {
        if (!existing.has(val)) {
          let label = val
          if (dim === Dimension.Database) label = getDBName(parseInt(val))
          else if (dim === Dimension.Features) {
            const labels: Record<string, string> = {
              with_error: "Con Errores",
              no_error: "Sin Errores",
              with_ur: "Con UR",
              no_ur: "Sin UR",
            }
            label = labels[val] || val
          }
          facet.values.push({
            value: val,
            label: label,
            count: 0,
            selected: true,
          })
        }
      })
    }

    facet.values.sort((a, b) => {
      if (a.selected !== b.selected) return a.selected ? -1 : 1
      if (a.count !== b.count) return b.count - a.count
      return (a.label || "").localeCompare(b.label || "")
    })
  }

  return dimensions.map((d) => resultsMap[d])
}

export async function getOffenses(
  predicates: InPredicate[],
  sortBy: SortBy,
  page: number,
  limit: number
): Promise<any[]> {
  "use cache"
  cacheLife("days")

  await waitForDB()
  const db = getDuckDB()
  const { where, args } = buildWhereClause(predicates || [])

  let query = `
    SELECT
      db_id,
      doc_source,
      doc_id,
      doc_date,
      record_id,
      offense_id,
      time,
      location,
      display_location,
      description,
      vehicle,
      vehicle_type,
      vehicle_country,
      ur,
      error,
      point,
      article_ids
    FROM offenses
  `

  if (where) {
    query += ` WHERE ${where}`
  }

  switch (sortBy) {
    case SortBy.Document:
      query += " ORDER BY doc_id ASC, record_id ASC"
      break
    case SortBy.Vehicle:
    default:
      query += " ORDER BY time DESC, doc_id ASC, record_id ASC"
      break
  }

  const offset = (page - 1) * limit
  query += ` LIMIT ${limit} OFFSET ${offset}`

  return dbAll(db, query, args).then((rows) => {
    return rows.map((row: any) => ({
      repo_id: Number(row.db_id),
      doc_source: row.doc_source,
      doc_id: row.doc_id,
      doc_date: row.doc_date ? new Date(row.doc_date).toISOString() : null,
      record_id: Number(row.record_id),
      id: row.offense_id,
      time: row.time ? new Date(row.time).toISOString() : null,
      location: row.location,
      display_location: row.display_location,
      description: row.description,
      vehicle: row.vehicle,
      vehicle_type: row.vehicle_type,
      country: row.vehicle_country,
      ur: Number(row.ur),
      error: row.error,
      point: row.point,
      article_id: row.article_ids,
      // Defaults for missing fields
      adm_division: "",
      mercosur_format: false,
    }))
  })
}

export async function getDocuments(
  predicates: InPredicate[],
  page: number,
  limit: number
): Promise<{ documents: OffenseDocument[]; total: number }> {
  await waitForDB()
  const db = getDuckDB()

  // We strictly filter by dimensions supported by documents view if needed,
  // but for now we trust the predicates passed from the UI (Year, Database).
  const { where, args } = buildWhereClause(predicates || [])

  let query = `
    SELECT
      db_id,
      doc_id,
      doc_date,
      doc_source,
      count(*) AS records,
      sum(ur) AS ur,
      sum(CASE WHEN error IS NOT NULL THEN 1 ELSE 0 END) AS errors
    FROM offenses
  `

  if (where) {
    query += ` WHERE ${where}`
  }

  query += `
    GROUP BY db_id, doc_id, doc_date, doc_source
    ORDER BY doc_date DESC, doc_id DESC,db_id ASC
  `

  const countQuery = `
    SELECT COUNT(*) as total FROM (
      SELECT 1
      FROM offenses
      ${where ? `WHERE ${where}` : ""}
      GROUP BY db_id, doc_id, doc_date, doc_source
    )
  `

  const offset = (page - 1) * limit
  query += ` LIMIT ${limit} OFFSET ${offset}`

  const [documents, total] = await Promise.all([
    new Promise<OffenseDocument[]>((resolve, reject) => {
      db.all(query, ...args, (err, rows) => {
        if (err) reject(err)
        else {
          const mapped = rows.map((row: any) => ({
            db_id: Number(row.db_id),
            doc_id: row.doc_id,
            doc_date: row.doc_date
              ? new Date(row.doc_date).toISOString().split("T")[0]
              : "",
            doc_source: row.doc_source,
            records: Number(row.records),
            ur: Number(row.ur),
            errors: Number(row.errors),
          }))
          resolve(mapped)
        }
      })
    }),
    new Promise<number>((resolve, reject) => {
      db.all(countQuery, ...args, (err, rows) => {
        if (err) reject(err)
        else resolve(Number(rows[0]?.total || 0))
      })
    }),
  ])

  return { documents, total }
}

// Charts

function chartBuildWhereClause(predicates: InPredicate[]): {
  where: string
  args: any[]
} {
  let { where, args } = buildWhereClause(predicates)
  if (where) {
    where += " AND error IS NULL"
  } else {
    where = "error IS NULL"
  }
  return { where, args }
}

export async function getChartDataByDayOfYear(
  predicates: InPredicate[],
  groupBy: Dimension | null
): Promise<Record<string, Record<string, number>>> {
  await waitForDB()
  const db = getDuckDB()
  const { where, args } = chartBuildWhereClause(predicates)

  let selectGroup = "''"
  let groupByClause = "strftime(time, '%j')"

  if (groupBy === Dimension.Year) {
    selectGroup = "strftime(time, '%Y')"
    groupByClause = "strftime(time, '%Y'), strftime(time, '%j')"
  }

  const query = `
        SELECT
            ${selectGroup} as grp,
            CAST(strftime(time, '%j') AS INTEGER) as day_of_year,
            count(*) as count
        FROM offenses
        WHERE ${where}
        GROUP BY ${groupByClause}
        ORDER BY grp, day_of_year
    `

  return new Promise((resolve, reject) => {
    db.all(query, ...args, (err, rows) => {
      if (err) reject(err)
      else {
        const tmp: Record<string, number[]> = {}

        rows.forEach((row: any) => {
          const year = row.grp || ""
          if (!tmp[year]) tmp[year] = new Array(367).fill(0)
          tmp[year][row.day_of_year] = Number(row.count)
        })

        const ret: Record<string, Record<string, number>> = {}

        Object.entries(tmp).forEach(([yearStr, dayCounts]) => {
          ret[yearStr] = {}

          // Merge leap day
          if (yearStr === "" && dayCounts[366] > 0) {
            dayCounts[59] += dayCounts[366]
            dayCounts[366] = 0
          }

          // Find range
          let minDay = 366
          let maxDay = 0
          for (let i = 1; i <= 366; i++) {
            if (dayCounts[i] > 0) {
              if (i < minDay) minDay = i
              if (i > maxDay) maxDay = i
            }
          }

          const formatYear = 2024 // Always use a leap year to handle day 366 and ensure consistent keys across years
          let cumulativeSum = 0

          // Check if we need to reduce points
          const shouldSample =
            groupBy === Dimension.Year && Object.keys(tmp).length > 2

          for (let day = minDay; day <= maxDay; day++) {
            cumulativeSum += dayCounts[day]

            // If sampling is enabled, only include every 7th day (starting from day 1: 1, 8, 15...)
            // We always include the last day if it's not included by the step?
            // The requirement says "one per week, but keeping the day as label: 01-01, 01-08, …"
            // So strictly (day - 1) % 7 === 0
            if (shouldSample && (day - 1) % 7 !== 0) {
              continue
            }

            // Date calculation
            const date = new Date(Date.UTC(formatYear, 0, 1)) // Jan 1st
            date.setUTCDate(date.getUTCDate() + (day - 1))
            const month = String(date.getUTCMonth() + 1).padStart(2, "0")
            const d = String(date.getUTCDate()).padStart(2, "0")
            const monthDay = `${month}-${d}`

            ret[yearStr][monthDay] = cumulativeSum
          }
        })
        resolve(ret)
      }
    })
  })
}

async function getChartDataByDimension(
  predicates: InPredicate[],
  dimension: "dayOfWeek" | "timeOfDay",
  groupBy: Dimension | null
): Promise<Record<string, Record<string, number>>> {
  await waitForDB()
  const db = getDuckDB()
  let { where, args } = chartBuildWhereClause(predicates)

  let dimExpr = ""
  if (dimension === "dayOfWeek") {
    dimExpr = "%w"
  } else {
    dimExpr = "%H"
    // Exclude midnight
    where += " AND strftime(time, '%H:%M:%S') != '00:00:00'"
  }

  let selectGroup = "''"
  let groupByClause = `GROUP BY strftime(time, '${dimExpr}')`

  if (groupBy === Dimension.Year) {
    selectGroup = "strftime(time, '%Y')"
    groupByClause = `GROUP BY strftime(time, '%Y'), strftime(time, '${dimExpr}')`
  }

  const query = `
        SELECT
            ${selectGroup} as grp,
            strftime(time, '${dimExpr}') as dim,
            count(*) as count
        FROM offenses
        WHERE ${where}
        ${groupByClause}
        ORDER BY grp, dim
    `

  const dayOfWeekSpanish: Record<string, string> = {
    "0": "Domingo",
    "1": "Lunes",
    "2": "Martes",
    "3": "Miércoles",
    "4": "Jueves",
    "5": "Viernes",
    "6": "Sábado",
  }

  return new Promise((resolve, reject) => {
    db.all(query, ...args, (err, rows) => {
      if (err) reject(err)
      else {
        const ret: Record<string, Record<string, number>> = {}

        rows.forEach((row: any) => {
          const year = row.grp || ""
          if (!ret[year]) ret[year] = {}

          let dimValue = row.dim
          if (dimension === "dayOfWeek") {
            dimValue = dayOfWeekSpanish[dimValue] || dimValue
          }

          ret[year][dimValue] = Number(row.count)
        })
        resolve(ret)
      }
    })
  })
}

export async function getChartDataByDayOfWeek(
  predicates: InPredicate[],
  groupBy: Dimension | null
): Promise<Record<string, Record<string, number>>> {
  await waitForDB()
  return getChartDataByDimension(predicates, "dayOfWeek", groupBy)
}

export async function getChartDataByTimeOfDay(
  predicates: InPredicate[],
  groupBy: Dimension | null
): Promise<Record<string, Record<string, number>>> {
  await waitForDB()
  return getChartDataByDimension(predicates, "timeOfDay", groupBy)
}

// Map

interface Feature {
  type: "Feature"
  geometry: {
    type: "Point" | "Polygon"
    coordinates: any
  }
  properties: {
    type: "cluster" | "location"
    h3_index?: string
    offenses: number
    locations?: number
    centroid?: number[]
    location?: string
  }
}

interface FeatureCollection {
  type: "FeatureCollection"
  features: Feature[]
}

async function getMapLocations(
  predicates: InPredicate[],
  parentCells: string[]
): Promise<FeatureCollection> {
  await waitForDB()
  if (parentCells.length === 0) {
    return { type: "FeatureCollection", features: [] }
  }

  const resolution = h3.getResolution(parentCells[0])
  // Validate resolution if needed, but we assume valid input from getMapClusters

  const db = getDuckDB()
  const { where, args } = buildWhereClause(predicates)

  const inPlaceholders = parentCells.map(() => "CAST(? AS UBIGINT)").join(",")
  const allArgs = [
    ...parentCells.map((c) => BigInt("0x" + c).toString()),
    ...args,
  ]

  // Note: We select h3_res{resolution} as parent_h3 to detect data inconsistencies.
  // Sometimes the child cell's stored parent doesn't match the parent we queried by,
  // due to H3 hierarchy mismatches (child center vs parent coverage).
  let query = `
        SELECT
            h3_res${resolution} as parent_h3,
            ST_X(point) as lng,
            ST_Y(point) as lat,
            location,
            COUNT(*) as offenses
        FROM offenses
        WHERE
            h3_res${resolution} IN (${inPlaceholders}) AND point IS NOT NULL
    `

  if (where) {
    query += " AND " + where
  }

  query += ` GROUP BY h3_res${resolution}, lng, lat, location`

  return new Promise((resolve, reject) => {
    db.all(query, ...allArgs, (err, rows) => {
      if (err) reject(err)
      else {
        const features: Feature[] = rows.map((row: any) => ({
          type: "Feature",
          geometry: {
            type: "Point",
            coordinates: [
              Number(row.lng.toFixed(6)),
              Number(row.lat.toFixed(6)),
            ],
          },
          properties: {
            type: "location",
            location: row.location,
            offenses: Number(row.offenses),
            // Pass back the parent H3 we found this under
            h3_index: row.parent_h3.toString(16),
          },
        }))

        resolve({
          type: "FeatureCollection",
          features,
        })
      }
    })
  })
}

export async function getMapClusters(
  predicates: InPredicate[],
  h3Index: string
): Promise<FeatureCollection> {
  await waitForDB()
  if (!h3.isValidCell(h3Index)) {
    throw new Error(`Invalid h3 index: ${h3Index}`)
  }

  const resolution = h3.getResolution(h3Index)

  // Resolution 8: return locations directly
  if (resolution >= 8) {
    return getMapLocations(predicates, [h3Index])
  }

  const { where, args } = buildWhereClause(predicates)
  const parentResCol = `h3_res${resolution}`
  const clusterResCol = `h3_res${resolution + 1}`
  let query = `
        SELECT
            ${clusterResCol} as h3,
            COUNT(*) as offenses,
            COUNT(DISTINCT location) as locations,
            AVG(ST_X(point)) as lng,
            AVG(ST_Y(point)) as lat
        FROM offenses
        WHERE
            ${parentResCol} = CAST(? AS UBIGINT)
    `

  const queryArgs = [BigInt("0x" + h3Index).toString(), ...args]

  if (where) {
    query += " AND " + where
  }

  query += `
     GROUP BY ${clusterResCol}
     ORDER BY locations ASC
    `

  return new Promise((resolve, reject) => {
    const db = getDuckDB()
    db.all(query, ...queryArgs, async (err, rows) => {
      if (err) return reject(err)

      const clusters = rows.map((row: any) => ({
        h3Cell: row.h3.toString(16),
        offenses: Number(row.offenses),
        locations: Number(row.locations),
        lat: row.lat,
        lng: row.lng,
      }))

      const features: Feature[] = []
      const cellsToExplode: string[] = []

      // Budget for exploded locations
      // We want to show at most MAX_VISIBLE_LOCATIONS individual points
      const MAX_VISIBLE_LOCATIONS = 15
      let count = clusters.length

      for (const cluster of clusters) {
        // Check if exploding this cluster fits in our budget
        if (count - 1 + cluster.locations <= MAX_VISIBLE_LOCATIONS) {
          cellsToExplode.push(cluster.h3Cell)
          count += cluster.locations - 1
        } else {
          // Keep as cluster
          // Use data centroid (average of offenses) instead of cell centroid
          // This ensures the marker is on land/road where offenses actually happened
          const lat = Number(cluster.lat.toFixed(6))
          const lng = Number(cluster.lng.toFixed(6))

          features.push({
            type: "Feature",
            geometry: {
              type: "Point",
              coordinates: [lng, lat],
            },
            properties: {
              type: "cluster",
              h3_index: cluster.h3Cell,
              offenses: cluster.offenses,
              locations: cluster.locations,
              centroid: [lng, lat],
            },
          })
        }
      }

      if (cellsToExplode.length > 0) {
        try {
          const locationData = await getMapLocations(predicates, cellsToExplode)

          // Robust Data Inconsistency Handling:
          // H3 cells do not fit perfectly into their parents (hierarchy mismatch).
          // This means a cell might be "inside" a parent geometrically, but its
          // stored h3_resX parent ID might differ, or vice-versa.
          // This leads to cases where 'getMapClusters' (group by parent) sees X locations,
          // but 'getMapLocations' (where parent = ?) sees Y locations (usually Y >> X).
          // If we blindly push these points, we might explode a "small" cluster into 50+ points,
          // ruining the map experience.

          // Group fetched locations by their parent H3 cell
          const locationsByCell: Record<string, Feature[]> = {}
          locationData.features.forEach((f) => {
            const parentH3 = f.properties.h3_index
            if (parentH3) {
              if (!locationsByCell[parentH3]) locationsByCell[parentH3] = []
              locationsByCell[parentH3].push(f)
            }
          })

          // Check each exploded cell against the budget/expectation
          cellsToExplode.forEach((cellH3) => {
            const points = locationsByCell[cellH3] || []

            // If the actual number of points is significantly larger than what we expected (from the cluster query)
            // OR if it's just too large in general (safety net), we re-cluster it.
            // We use a loose threshold here (e.g. > 20) to catch the "sea of blue markers" case (56 points).
            if (points.length > 20) {
              // Re-cluster!
              // Calculate centroid of the actual points
              let sumLat = 0,
                sumLng = 0,
                totalOffenses = 0
              points.forEach((p) => {
                const [lng, lat] = p.geometry.coordinates as number[]
                sumLat += lat
                sumLng += lng
                totalOffenses += p.properties.offenses || 0
              })

              const avgLat = sumLat / points.length
              const avgLng = sumLng / points.length

              features.push({
                type: "Feature",
                geometry: {
                  type: "Point",
                  coordinates: [avgLng, avgLat],
                },
                properties: {
                  type: "cluster",
                  h3_index: cellH3,
                  offenses: totalOffenses,
                  locations: points.length,
                  centroid: [avgLng, avgLat],
                },
              })
            } else {
              // Safe to add points
              features.push(...points)
            }
          })
        } catch (e) {
          return reject(e)
        }
      }

      resolve({
        type: "FeatureCollection",
        features: features,
      })
    })
  })
}

// Version Caching
let versionCache: string | null = null

export async function getDatabaseVersion(): Promise<string> {
  if (versionCache) return versionCache

  await waitForDB()
  const db = getDuckDB()

  try {
    const query = `
      SELECT 
        (SELECT COUNT(*) FROM offenses) as cnt,
        (SELECT MAX(ts) FROM (
           SELECT updated_at as ts FROM locations
           UNION ALL 
           SELECT updated_at as ts FROM descriptions
        ) t) as max_curation_ts
    `
    const rows = await dbAll(db, query, [])
    if (rows && rows.length > 0) {
      const { cnt, max_curation_ts } = rows[0]
      // Format: v:{count}:{timestamp}
      // Use efficient formatting. If ts is null, use 0.
      const ts = max_curation_ts ? new Date(max_curation_ts).getTime() : 0
      versionCache = `v:${Number(cnt).toString(16)}:${ts.toString(16)}`
    } else {
      versionCache = "v:0:0"
    }
  } catch (err) {
    console.error("Error computing database version:", err)
    // Fallback in case of error (e.g. tables don't exist yet)
    versionCache = "v:error"
  }

  return versionCache!
}
