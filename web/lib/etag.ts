/**
 * Copyright 2025 The ChapaUY Authors
 * SPDX-License-Identifier: Apache-2.0
 */

import { NextRequest, NextResponse } from "next/server"
import { getAppVersion } from "./version"

export const CACHE_HEADERS = {
    "Cache-Control": "public, max-age=300, stale-while-revalidate=604800",
}

export interface ETagResult {
    response?: NextResponse
    options?: {
        headers: HeadersInit
    }
}

/**
 * Checks the request headers against the current database version ETag.
 * If matches, returns a 304 response (wrapped).
 * If not, returns the headers to be used in the 200 response.
 */
export async function checkETag(request: NextRequest): Promise<ETagResult> {
    const appVersion = await getAppVersion()
    const etag = `"${appVersion}"`
    const ifNoneMatch = request.headers.get("if-none-match")

    if (ifNoneMatch === etag) {
        return {
            response: new NextResponse(null, {
                status: 304,
                headers: {
                    ...CACHE_HEADERS,
                    ETag: etag,
                },
            }),
        }
    }

    return {
        options: {
            headers: {
                ...CACHE_HEADERS,
                ETag: etag,
            }
        }
    }
}
