/**
 * Copyright 2025 The ChapaUY Authors
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect } from "vitest"
import { Dimension } from "./types"

describe("Dimension", () => {
  it("should have valid values", () => {
    expect(Dimension.Database).toBe("database")
    expect(Dimension.Year).toBe("year")
    expect(Dimension.Country).toBe("country")
    expect(Dimension.VehicleType).toBe("vehicle_type")
    expect(Dimension.Vehicle).toBe("vehicle")
    expect(Dimension.DocSource).toBe("doc_source")
    expect(Dimension.Location).toBe("location")
    expect(Dimension.Description).toBe("description")
    expect(Dimension.ArticleID).toBe("article_id")
    expect(Dimension.ArticleCode).toBe("article_code")
  })

  it("should validate input strings against dimensions", () => {
    // In TS, we don't have a direct "NewDimension" constructor that validates strings like in Go.
    // But we can check if a string is a valid value of the enum.
    const isValidDimension = (input: string): boolean => {
      return Object.values(Dimension).includes(input as Dimension)
    }

    expect(isValidDimension("database")).toBe(true)
    expect(isValidDimension("year")).toBe(true)
    expect(isValidDimension("country")).toBe(true)
    expect(isValidDimension("vehicle_type")).toBe(true)

    expect(isValidDimension("")).toBe(false)
    expect(isValidDimension("foo")).toBe(false)
    expect(isValidDimension("DATABASE")).toBe(false) // Case sensitive
    expect(isValidDimension(" ")).toBe(false)
  })
})
