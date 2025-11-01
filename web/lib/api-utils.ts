/**
 * Copyright 2025 The ChapaUY Authors
 * SPDX-License-Identifier: Apache-2.0
 */

// Ensures consistent cache keys by sorting params alphabetically
export function buildApiUrl(
    base: string,
    params: Record<string, string | string[] | undefined | null>
): string {
    // Use a dummy origin if running on server side or if window is undefined,
    // though this function is typically used on client side.
    const origin =
        typeof window !== "undefined" ? window.location.origin : "http://localhost"
    const url = new URL(base, origin)
    const sortedKeys = Object.keys(params).sort()

    for (const key of sortedKeys) {
        const val = params[key]
        if (val === undefined || val === null) continue

        if (Array.isArray(val)) {
            val.sort().forEach((v) => url.searchParams.append(key, v))
        } else {
            url.searchParams.set(key, val)
        }
    }
    return url.toString()
}
