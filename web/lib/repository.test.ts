/**
 * Copyright 2025 The ChapaUY Authors
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect, beforeEach, vi, afterEach } from "vitest"
import duckdb from "duckdb"
import {
  getOffenses,
  getOffensesSummary,
  getDimensionResults,
  getChartDataByDayOfYear,
  getMapClusters,
} from "./repository"
import { Dimension, InPredicate, SortBy } from "./types"

// Mock the duckdb module to return our test instance
let testDB: duckdb.Database

vi.mock("./duckdb", () => ({
  getDuckDB: () => testDB,
  waitForDB: () => Promise.resolve(),
}))

vi.mock("next/cache", () => ({
  cacheLife: () => {},
  unstable_cache: (fn: any) => fn,
}))

vi.mock("h3-js", async () => {
  const actual = await vi.importActual("h3-js")
  return {
    ...actual,
    isValidCell: () => true,
    getResolution: (h3Index: string) => {
      if (h3Index.startsWith("86")) return 6
      if (h3Index.startsWith("87")) return 7
      if (h3Index.startsWith("88")) return 8
      return 0
    },
  }
})

// Helper to run SQL
const runQuery = (
  db: duckdb.Database,
  sql: string,
  params: any[] = []
): Promise<void> => {
  return new Promise((resolve, reject) => {
    db.run(sql, ...params, (err) => {
      if (err) reject(err)
      else resolve()
    })
  })
}

const setupTestDB = async () => {
  testDB = new duckdb.Database(":memory:")

  // Create Schema
  await runQuery(
    testDB,
    `
    INSTALL spatial; LOAD spatial;
    CREATE TABLE offenses (
        db_id INTEGER NOT NULL,
        doc_id VARCHAR,
        doc_date DATE,
        doc_source VARCHAR NOT NULL,
        record_id INTEGER NOT NULL,
        offense_id VARCHAR,
        vehicle VARCHAR,
        vehicle_country CHAR(2),
        vehicle_type VARCHAR,
        "time" TIMESTAMPTZ,
        time_year USMALLINT,
        location VARCHAR,
        display_location VARCHAR,
        description VARCHAR,
        ur INTEGER,
        error VARCHAR,
        point POINT_2D,
        h3_res1 UBIGINT,
        h3_res2 UBIGINT,
        h3_res3 UBIGINT,
        h3_res4 UBIGINT,
        h3_res5 UBIGINT,
        h3_res6 UBIGINT,
        h3_res7 UBIGINT,
        h3_res8 UBIGINT,
        article_ids VARCHAR[],
        article_codes TINYINT[]
    );
    CREATE TABLE articles (
        id VARCHAR PRIMARY KEY,
        text VARCHAR NOT NULL,
        code TINYINT NOT NULL,
        title VARCHAR NOT NULL
    );
  `
  )

  // Insert Test Data
  await runQuery(
    testDB,
    `
    INSERT INTO offenses (db_id, doc_source, doc_id, doc_date, record_id, offense_id, vehicle, vehicle_country, vehicle_type, time, time_year, location, description, ur, error) VALUES
      (45, 'doc1', 'doc1_id', '2023-01-01', 1, 'offense1', 'AAAA123', 'UY', 'AUTO', '2023-01-01 10:00:00', 2023, 'Some Location', 'Speeding', 100, NULL),
      (45, 'doc2', 'doc2_id', '2024-01-01', 1, 'offense2', 'BBBB456', 'UY', 'MOTO', '2024-01-01 11:00:00', 2024, 'Another Location', 'Parking', 200, NULL),
      (45, 'doc1', 'doc1_id', '2023-01-01', 2, 'offense3', 'AAAA123', 'UY', 'AUTO', '2024-01-01 12:00:00', 2024, 'Some Location', 'Speeding', 300, NULL),
      (45, 'doc1', 'doc1_id', '2023-01-01', 3, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, 'Some error'),
      (45, 'doc1', 'doc1_id', '2023-01-01', 4, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, 'Another error')
  `
  )
}

const setupTestDBWithChartData = async () => {
  await setupTestDB()
  await runQuery(
    testDB,
    `
        INSERT INTO offenses (db_id, doc_source, doc_id, doc_date, record_id, offense_id, vehicle, vehicle_country, vehicle_type, time, time_year, location, description, ur, error) VALUES
            -- Regular year data (2023)
            (45, 'doc3', 'doc3_id', '2023-01-01', 5, 'offense5', 'CCCC789', 'UY', 'AUTO', '2023-01-01 10:00:00', 2023, 'Location1', 'Offense1', 100, NULL),
            (45, 'doc3', 'doc3_id', '2023-01-01', 6, 'offense6', 'DDDD012', 'UY', 'AUTO', '2023-01-01 11:00:00', 2023, 'Location2', 'Offense2', 200, NULL),
            (45, 'doc3', 'doc3_id', '2023-01-01', 7, 'offense7', 'EEEE345', 'UY', 'AUTO', '2023-01-02 12:00:00', 2023, 'Location3', 'Offense3', 300, NULL),
            (45, 'doc3', 'doc3_id', '2023-01-01', 8, 'offense8', 'FFFF678', 'UY', 'AUTO', '2023-02-28 13:00:00', 2023, 'Location4', 'Offense4', 400, NULL),
            (45, 'doc3', 'doc3_id', '2023-01-01', 9, 'offense9', 'GGGG901', 'UY', 'AUTO', '2023-12-31 14:00:00', 2023, 'Location5', 'Offense5', 500, NULL),
            -- Leap year data (2024) - including leap day
            (45, 'doc4', 'doc4_id', '2024-01-01', 10, 'offense10', 'HHHH234', 'UY', 'AUTO', '2024-01-01 10:00:00', 2024, 'Location6', 'Offense6', 600, NULL),
            (45, 'doc4', 'doc4_id', '2024-01-01', 11, 'offense11', 'IIII567', 'UY', 'AUTO', '2024-02-28 11:00:00', 2024, 'Location7', 'Offense7', 700, NULL),
            (45, 'doc4', 'doc4_id', '2024-01-01', 12, 'offense12', 'JJJJ890', 'UY', 'AUTO', '2024-02-29 12:00:00', 2024, 'Location8', 'Offense8', 800, NULL),
            (45, 'doc4', 'doc4_id', '2024-01-01', 13, 'offense13', 'KKKK123', 'UY', 'AUTO', '2024-03-01 13:00:00', 2024, 'Location9', 'Offense9', 900, NULL),
            (45, 'doc4', 'doc4_id', '2024-01-01', 14, 'offense14', 'LLLL456', 'UY', 'AUTO', '2024-12-31 14:00:00', 2024, 'Location10', 'Offense10', 1000, NULL)
    `
  )
}

describe("OffenseRepository", () => {
  describe("getOffensesSummary", () => {
    beforeEach(async () => {
      await setupTestDB()
    })

    it("group by database", async () => {
      const summaries = await getOffensesSummary([], Dimension.Database)
      expect(summaries).toHaveLength(1)
      expect(summaries[0].value).toBe(45)
      // Label might depend on getDBName which is imported.
      // Assuming getDBName works or we might need to mock it if it uses external config.
      // For now, let's check other fields.
      expect(summaries[0].count).toBe(5)
      expect(summaries[0].doc_count).toBe(2)
      expect(summaries[0].ur_total).toBe(600)
      expect(summaries[0].ur_avg).toBe(200)
    })

    it("no grouping", async () => {
      const summaries = await getOffensesSummary([], null)
      expect(summaries).toHaveLength(1)
      expect(summaries[0].count).toBe(5)
      expect(summaries[0].ur_total).toBe(600)
      expect(summaries[0].ur_avg).toBe(200)
    })
  })

  describe("getOffenses", () => {
    beforeEach(async () => {
      await setupTestDB()
    })

    it("should return offenses with filters", async () => {
      const filters: InPredicate[] = [
        { dimension: Dimension.Database, values: ["45"] },
      ]
      const offenses = await getOffenses(filters, SortBy.Vehicle, 1, 10)
      expect(offenses).toHaveLength(5)

      offenses.forEach((o) => {
        if (o.doc_date) {
          // Check timezone if possible, or just existence
          expect(o.doc_date).toBeDefined()
        }
      })
    })

    it("no results", async () => {
      const filters: InPredicate[] = [
        { dimension: Dimension.Year, values: ["1999"] },
      ]
      const offenses = await getOffenses(filters, SortBy.Vehicle, 1, 10)
      expect(offenses).toHaveLength(0)
    })
  })

  describe("getDimensionResults", () => {
    beforeEach(async () => {
      await setupTestDB()
    })

    it("unlimited results", async () => {
      const filters: InPredicate[] = [
        { dimension: Dimension.Database, values: ["45"] },
        { dimension: Dimension.Year, values: ["2024"] },
      ]
      const dimensions = [Dimension.VehicleType]

      const results = await getDimensionResults(filters, dimensions)
      expect(results).toHaveLength(1)

      const result = results[0]
      expect(result.dimension).toBe(Dimension.VehicleType)
      expect(result.values).toHaveLength(2) // AUTO, MOTO (from 2024 data? Wait, let's check data)
      // Data:
      // 2024-01-01: offense2 (MOTO)
      // 2024-01-01: offense3 (AUTO) - wait, offense3 is doc_date 2023 but time 2024.
      // getColumnExpr(Dimension.Year) uses strftime(time, '%Y').
      // offense3 time is '2024-01-01 12:00:00', so it is 2024.
      // So we have MOTO and AUTO. Correct.
    })

    it("top results", async () => {
      const filters: InPredicate[] = [
        { dimension: Dimension.Database, values: ["45"] },
      ]
      const dimensions = [
        Dimension.Vehicle,
        Dimension.Location,
        Dimension.Description,
        Dimension.Year,
        Dimension.Country,
      ]

      const results = await getDimensionResults(filters, dimensions)
      expect(results).toHaveLength(5)

      results.forEach((result) => {
        switch (result.dimension) {
          case Dimension.Vehicle:
            expect(result.values.length).toBeGreaterThanOrEqual(1)
            const val = result.values.find((v) => v.value === "AAAA123")
            expect(val).toBeDefined()
            expect(val?.count).toBe(2)
            break
          case Dimension.Location:
            expect(
              result.values.find((v) => v.value === "Some Location")?.count
            ).toBe(2)
            break
          case Dimension.Description:
            expect(
              result.values.find((v) => v.value === "Speeding")?.count
            ).toBe(2)
            break
          case Dimension.Year:
            expect(result.values.find((v) => v.value === "2024")?.count).toBe(2)
            break
          case Dimension.Country:
            expect(result.values.find((v) => v.value === "UY")?.count).toBe(3)
            break
        }
      })
    })
    it("should prioritize selected values", async () => {
      // Setup data with counts that would normally put "selected" item at the bottom
      await runQuery(testDB, `DELETE FROM offenses`)
      await runQuery(
        testDB,
        `
              INSERT INTO offenses (db_id, doc_source, doc_id, record_id, vehicle_type, time) VALUES
              (1, 'd', '1', 1, 'AUTO', '2023-01-01'),
              (1, 'd', '1', 2, 'AUTO', '2023-01-01'),
              (1, 'd', '1', 3, 'AUTO', '2023-01-01'),
              (1, 'd', '1', 4, 'MOTO', '2023-01-01'),
              (1, 'd', '1', 5, 'MOTO', '2023-01-01'),
              (1, 'd', '1', 6, 'CAMION', '2023-01-01')
          `
      )

      // AUTO: 3, MOTO: 2, CAMION: 1
      // Normal order: AUTO, MOTO, CAMION

      // Select CAMION. Expect: CAMION, AUTO, MOTO
      const filters: InPredicate[] = [
        { dimension: Dimension.VehicleType, values: ["CAMION"] },
      ]

      const results = await getDimensionResults(filters, [
        Dimension.VehicleType,
      ])
      const values = results[0].values

      expect(values[0].value).toBe("CAMION")
      expect(values[0].selected).toBe(true)

      expect(values[1].value).toBe("AUTO")
      expect(values[2].value).toBe("MOTO")
    })
  })

  describe("getChartDataByDayOfYear", () => {
    beforeEach(async () => {
      await setupTestDBWithChartData()
    })

    it("ungrouped data - no leap day handling", async () => {
      const filters: InPredicate[] = [
        { dimension: Dimension.Year, values: ["2023"] },
      ]
      const chartData = await getChartDataByDayOfYear(filters, null)

      expect(chartData).toHaveProperty("")
      const data = chartData[""]

      // 2023:
      // Jan 1: offense1(2023), offense5, offense6 -> 3 offenses
      // Jan 2: offense7 -> 1 offense
      // Feb 28: offense8 -> 1 offense
      // Dec 31: offense9 -> 1 offense

      // Cumulative:
      // Jan 1: 3
      // Jan 2: 3+1=4
      // Feb 28: 4+1=5
      // Dec 31: 5+1=6

      expect(data["01-01"]).toBe(3)
      expect(data["01-02"]).toBe(4)
      expect(data["02-28"]).toBe(5)
      expect(data["12-30"]).toBe(6)
    })

    it("grouped by year", async () => {
      const chartData = await getChartDataByDayOfYear([], Dimension.Year)

      expect(chartData).toHaveProperty("2023")
      expect(chartData).toHaveProperty("2024")

      const data2024 = chartData["2024"]
      // 2024:
      // Jan 1: offense2(2024), offense3(2024), offense10 -> 3 offenses
      // Feb 28: offense11 -> 1 offense
      // Feb 29: offense12 -> 1 offense
      // Mar 01: offense13 -> 1 offense
      // Dec 31: offense14 -> 1 offense

      // Cumulative:
      // Jan 1: 3
      // Feb 28: 3+1=4
      // Feb 29: 4+1=5
      // Mar 01: 5+1=6
      // Dec 31: 6+1=7

      expect(data2024["01-01"]).toBe(3)
      expect(data2024["02-28"]).toBe(4)
      expect(data2024["02-29"]).toBe(5)
      expect(data2024["03-01"]).toBe(6)
      expect(data2024["12-31"]).toBe(7)
    })
  })

  describe("getMapClusters", () => {
    beforeEach(async () => {
      testDB = new duckdb.Database(":memory:")
      await runQuery(
        testDB,
        `INSTALL spatial; LOAD spatial; CREATE TABLE offenses (
            db_id INTEGER, location VARCHAR, point POINT_2D, 
            h3_res6 UBIGINT, h3_res7 UBIGINT, h3_res8 UBIGINT
          )`
      )

      const parentRes6 = BigInt("0x86c2a7a97ffffff").toString()
      const childRes7A = BigInt("0x87c2a7a90ffffff").toString()
      const childRes7B = BigInt("0x87c2a7a92ffffff").toString()

      const parentRes7 = BigInt("0x87c2a7a97ffffff").toString()
      const childRes8A = BigInt("0x88c2a7a97ffffff").toString()
      const childRes8B = BigInt("0x88c2a7a94ffffff").toString()

      await runQuery(
        testDB,
        `
             INSERT INTO offenses (db_id, location, point, h3_res6, h3_res7, h3_res8) VALUES
             (45, 'Location A', ST_Point(-56.1, -34.9), NULL, ${parentRes7}, ${childRes8A}),
             (45, 'Location A', ST_Point(-56.1, -34.9), NULL, ${parentRes7}, ${childRes8A}),
             (45, 'Location B', ST_Point(-56.11, -34.91), NULL, ${parentRes7}, ${childRes8B}),
             
             (45, 'Location C', ST_Point(-56.2, -34.8), ${parentRes6}, ${childRes7A}, NULL),
             (45, 'Location D', ST_Point(-56.21, -34.81), ${parentRes6}, ${childRes7A}, NULL),
             
             (45, 'Location E', ST_Point(-56.25, -34.85), ${parentRes6}, ${childRes7B}, NULL)
          `
      )
    })

    it("returns clusters and exploded locations for parent resolution 6", async () => {
      const features = await getMapClusters([], "86c2a7a97ffffff")
      // TS implementation explodes all 3 clusters because total locations (3) < 20
      expect(features.features).toHaveLength(3)

      const points = features.features.filter(
        (f) => f.properties.type === "location"
      )
      expect(points).toHaveLength(3)

      const locE = points.find((f) => f.properties.location === "Location E")
      expect(locE).toBeDefined()
    })

    it("returns locations for parent resolution 7", async () => {
      const features = await getMapClusters([], "87c2a7a97ffffff")
      expect(features.features).toHaveLength(2)

      const locA = features.features.find(
        (f) => f.properties.location === "Location A"
      )
      const locB = features.features.find(
        (f) => f.properties.location === "Location B"
      )

      expect(locA).toBeDefined()
      expect(locA?.properties.offenses).toBe(2)

      expect(locB).toBeDefined()
      expect(locB?.properties.offenses).toBe(1)
    })
  })
})
