"use client"

import * as React from "react"
import { cn } from "@/lib/utils"
import type { Heading } from "@/lib/docs"

interface TableOfContentsProps {
    headings: Heading[]
}

export function TableOfContents({ headings }: TableOfContentsProps) {
    const [activeId, setActiveId] = React.useState<string>("")

    React.useEffect(() => {
        const observer = new IntersectionObserver(
            (entries) => {
                entries.forEach((entry) => {
                    if (entry.isIntersecting) {
                        setActiveId(entry.target.id)
                    }
                })
            },
            { rootMargin: "0% 0% -80% 0%" }
        )

        headings.forEach((heading) => {
            const element = document.getElementById(heading.slug)
            if (element) {
                observer.observe(element)
            }
        })

        return () => {
            headings.forEach((heading) => {
                const element = document.getElementById(heading.slug)
                if (element) {
                    observer.unobserve(element)
                }
            })
        }
    }, [headings])

    if (!headings?.length) return null

    return (
        <div className="space-y-2">
            <p className="font-medium text-sm">En este artículo</p>
            <ul className="m-0 list-none">
                {headings.map((heading) => (
                    <li key={heading.slug} className="mt-0 pt-2">
                        <a
                            href={`#${heading.slug}`}
                            className={cn(
                                "inline-block no-underline transition-colors hover:text-foreground",
                                heading.level === 3 ? "pl-4" : "",
                                activeId === heading.slug
                                    ? "font-medium text-foreground"
                                    : "text-muted-foreground"
                            )}
                        >
                            {heading.text}
                        </a>
                    </li>
                ))}
            </ul>
        </div>
    )
}
