
import { describe, it, expect, vi, beforeAll, afterAll } from "vitest"
import fs from "fs"
import path from "path"
import os from "os"

// Setup specific test directory
const tempDir = fs.mkdtempSync(path.join(os.tmpdir(), "docs-test-"))
const listDir = path.join(tempDir, "docs")
fs.mkdirSync(listDir)

import { getAllDocSlugs, getDocBySlug, cleanName, setDocsDirForTesting } from "./docs"

// Configure it
setDocsDirForTesting(listDir)

describe("clean URLs logic", () => {
    beforeAll(() => {
        // Setup fixtures
        fs.writeFileSync(path.join(listDir, "001-intro.md"), "---\ntitle: Intro\n---\n# Title\n## Section 1\n### Subsection\ntext")

        fs.mkdirSync(path.join(listDir, "002-section"))
        fs.writeFileSync(path.join(listDir, "002-section/001-deep.mdx"), "---\ntitle: Deep\n---\nDeep content")

        fs.writeFileSync(path.join(listDir, "ambiguous.md"), "---\ntitle: Ambiguous\n---\n")
        // fs.writeFileSync(path.join(listDir, "003-ambiguous.md"), "---\ntitle: Ambiguous 2\n---\n") -> Will uncomment to test collision
    })

    afterAll(() => {
        // Cleanup
        fs.rmSync(tempDir, { recursive: true, force: true })
    })

    it("cleanName removes 3 digit prefix", () => {
        expect(cleanName("001-hello")).toBe("hello")
        expect(cleanName("hello")).toBe("hello")
        expect(cleanName("12-hello")).toBe("12-hello") // Only 3 digits
    })

    it("getAllDocSlugs returns clean slugs", () => {
        const slugs = getAllDocSlugs()
        // Expect [['intro'], ['section', 'deep'], ['ambiguous']]
        const slugStrings = slugs.map(s => s.join("/"))
        expect(slugStrings).toContain("intro")
        expect(slugStrings).toContain("section/deep")
        expect(slugStrings).toContain("ambiguous")
        expect(slugStrings).not.toContain("001-intro")
    })

    it("getDocBySlug resolves clean slug to file on disk", async () => {
        const doc = await getDocBySlug(["intro"])
        expect(doc).not.toBeNull()
        expect(doc?.title).toBe("Intro")
        // DOCS_DIR is tempDir/docs
        // file is tempDir/docs/001-intro.md
        // so relative path should be 001-intro.md
        expect(doc?.relativePath).toBe("001-intro.md")
        expect(doc?.headings).toEqual([
            { level: 2, text: "Section 1", slug: "section-1" },
            { level: 3, text: "Subsection", slug: "subsection" }
        ])
        // Verify headings
        // Content was "Hello", no headings. Let's add one. 
        // Wait, I can't modify the file easily here without rewriting it.
        // I'll update the setup in beforeAll to include headings.

        const doc2 = await getDocBySlug(["section", "deep"])
        expect(doc2).not.toBeNull()
        expect(doc2?.title).toBe("Deep")
        expect(doc2?.relativePath).toBe("002-section/001-deep.mdx")
    })

    it("detects collisions", () => {
        fs.writeFileSync(path.join(listDir, "003-ambiguous.md"), "---\ntitle: Ambiguous 2\n---\n")
        expect(() => getAllDocSlugs()).toThrow(/Duplicate slug detected/)
        fs.rmSync(path.join(listDir, "003-ambiguous.md"))
    })
})
