/**
 * Copyright 2025 The ChapaUY Authors
 * SPDX-License-Identifier: Apache-2.0
 */

import fs from "fs"
import path from "path"
import matter from "gray-matter"
import GithubSlugger from "github-slugger"

let DOCS_DIR = path.join(process.cwd(), "docs")

// For testing purposes
export function setDocsDirForTesting(dir: string) {
    DOCS_DIR = dir
}

export interface DocPage {
    slug: string[]
    title: string
    description?: string
    author?: string
    content: string
    relativePath: string
    headings: Heading[]
}

export interface Heading {
    level: number
    text: string
    slug: string
}

export interface DocNode {
    title: string
    slug?: string
    children?: DocNode[]
}

// Helper to remove 3-digit prefix from name
export function cleanName(name: string): string {
    return name.replace(/^\d{3}-/, "")
}

// Helper to get all files recursively
function getFiles(dir: string): string[] {
    if (!fs.existsSync(dir)) return []
    const dirents = fs.readdirSync(dir, { withFileTypes: true })
    const files: string[] = []
    for (const dirent of dirents) {
        const res = path.resolve(dir, dirent.name)
        if (dirent.isDirectory()) {
            files.push(...getFiles(res))
        } else {
            files.push(res)
        }
    }
    return files
}

function buildTreeStructure(dir: string, parentSlug: string = ""): DocNode[] {
    if (!fs.existsSync(dir)) return []
    const entries = fs.readdirSync(dir, { withFileTypes: true })

    // Sort: directories first, then files
    entries.sort((a, b) => {
        if (a.isDirectory() && !b.isDirectory()) return -1
        if (!a.isDirectory() && b.isDirectory()) return 1
        return a.name.localeCompare(b.name)
    })

    const nodes: DocNode[] = []

    for (const entry of entries) {
        const cleanEntryName = cleanName(entry.name)

        if (entry.isDirectory()) {
            const children = buildTreeStructure(
                path.join(dir, entry.name),
                `${parentSlug}/${cleanEntryName}`
            )
            if (children.length) {
                nodes.push({
                    title: capitalize(cleanEntryName),
                    children,
                })
            }
        } else if (entry.name.endsWith(".mdx") || entry.name.endsWith(".md")) {
            const content = fs.readFileSync(path.join(dir, entry.name), "utf-8")
            const { data } = matter(content)
            const basename = entry.name.replace(/\.mdx?$/, "")
            const cleanBasename = cleanName(basename)

            // Handle index files
            if (cleanBasename === "index") {
                if (!parentSlug) {
                    // Root index
                    nodes.unshift({
                        // Put home first
                        title: (data.title as string) || "Home",
                        slug: "/docs",
                    })
                }
                continue
            }

            nodes.push({
                title: (data.title as string) || capitalize(cleanBasename),
                slug: `/docs${parentSlug}/${cleanBasename}`,
            })
        }
    }
    return nodes
}

export async function getDocPages(): Promise<DocNode[]> {
    if (!fs.existsSync(DOCS_DIR)) return []
    return buildTreeStructure(DOCS_DIR)
}

// Resolves a clean slug parts to a real file path on disk
function resolveSlugPath(dir: string, parts: string[]): string | null {
    if (parts.length === 0) {
        // Look for index in current dir
        if (!fs.existsSync(dir)) return null
        const entries = fs.readdirSync(dir)
        for (const entry of entries) {
            if (
                (entry.endsWith(".mdx") || entry.endsWith(".md")) &&
                cleanName(entry.replace(/\.mdx?$/, "")) === "index"
            ) {
                return path.join(dir, entry)
            }
        }
        return null
    }

    const [head, ...tail] = parts
    if (!fs.existsSync(dir)) return null

    const entries = fs.readdirSync(dir, { withFileTypes: true })

    for (const entry of entries) {
        const name = entry.name
        const isDir = entry.isDirectory()

        if (isDir) {
            if (cleanName(name) === head) {
                // Recursive step
                if (tail.length === 0) {
                    // We matched the directory, now check for index inside it
                    return resolveSlugPath(path.join(dir, name), [])
                }
                return resolveSlugPath(path.join(dir, name), tail)
            }
        } else {
            // File
            if (tail.length === 0 && (name.endsWith(".mdx") || name.endsWith(".md"))) {
                const basename = name.replace(/\.mdx?$/, "")
                if (cleanName(basename) === head) {
                    return path.join(dir, name)
                }
            }
        }
    }

    return null
}


export async function getDocBySlug(
    slug: string[] = []
): Promise<DocPage | null> {
    const resolvedPath = resolveSlugPath(DOCS_DIR, slug)

    if (!resolvedPath) return null

    // SECURITY: Ensure resolved path is within DOCS_DIR
    if (!path.resolve(resolvedPath).startsWith(path.resolve(DOCS_DIR))) {
        return null
    }

    if (fs.existsSync(resolvedPath)) {
        const source = fs.readFileSync(resolvedPath, "utf8")
        const { content, data } = matter(source)
        const headings = extractHeadings(content)
        return {
            slug,
            title: (data.title as string) || "Untitled",
            description: data.description as string,
            author: data.author as string,
            content,
            relativePath: path.relative(DOCS_DIR, resolvedPath),
            headings,
        }
    }

    return null
}

function capitalize(s: string) {
    return s.charAt(0).toUpperCase() + s.slice(1)
}

export function getAllDocSlugs(): string[][] {
    if (!fs.existsSync(DOCS_DIR)) return []
    const files = getFiles(DOCS_DIR)
    const slugs: string[][] = []
    const seenSlugs = new Set<string>()

    for (const file of files) {
        if (!file.endsWith(".mdx") && !file.endsWith(".md")) continue

        const relativePath = path.relative(DOCS_DIR, file)
        const parts = relativePath.split(path.sep)

        const cleanParts = parts.map(p => {
            if (p.endsWith(".mdx") || p.endsWith(".md")) {
                return cleanName(p.replace(/\.mdx?$/, ""))
            }
            return cleanName(p)
        })

        // Handle index files
        if (cleanParts[cleanParts.length - 1] === "index") {
            cleanParts.pop()
        }

        const slugStr = cleanParts.join("/")
        if (seenSlugs.has(slugStr)) {
            throw new Error(`Duplicate slug detected: ${slugStr}. Source: ${relativePath}`)
        }
        seenSlugs.add(slugStr)

        slugs.push(cleanParts)
    }
    return slugs
}

function extractHeadings(content: string): Heading[] {
    const headingRegex = /^(#{2,3})\s+(.+)$/gm
    const headings: Heading[] = []
    const slugger = new GithubSlugger()

    let match
    while ((match = headingRegex.exec(content)) !== null) {
        const level = match[1].length
        const text = match[2].trim()
        const slug = slugger.slug(text)
        headings.push({ level, text, slug })
    }
    return headings
}
