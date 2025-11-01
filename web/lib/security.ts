import { NextRequest } from "next/server"

// Hotlinking "protection"
export function isAllowedOrigin(request: NextRequest): boolean {
    const origin = request.headers.get("origin")
    return !origin

    // const referer = request.headers.get("referer") || ""
    // const origin = request.headers.get("origin") || ""
    // const allowedHosts = [
    //     "localhost:3000",
    // ]

    // const checkHost = (url: string) => {
    //     if (!url) return false
    //     try {
    //         const host = new URL(url).host
    //         return allowedHosts.some((h) => host.includes(h))
    //     } catch {
    //         return false
    //     }
    // }

    // if (!referer && !origin) {
    //     return true
    // }

    // return checkHost(referer) || checkHost(origin)
}
