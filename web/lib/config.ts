/**
 * Copyright 2025 The ChapaUY Authors
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * The base URL of the application.
 * Defaults to "https://chapa.uy" if NEXT_PUBLIC_BASE_URL is not set.
 * Normalizes the URL by removing any trailing slash.
 */
const rawBaseUrl = process.env.NEXT_PUBLIC_BASE_URL || "https://chapa.uy"
export const BASE_URL = rawBaseUrl.endsWith("/") ? rawBaseUrl.slice(0, -1) : rawBaseUrl
