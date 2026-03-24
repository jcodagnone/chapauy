"use client"

import Link from "next/link"
import { BookOpen } from "lucide-react"
import { cn } from "@/lib/utils"
import { Button } from "@/components/ui/button"

interface GlobalLinksProps {
    className?: string
    includeDocsLink?: boolean
    variant?: "ghost" | "outline" | "default" | "secondary"
    size?: "default" | "sm" | "lg" | "icon"
}

function GitHubMark(props: React.SVGProps<SVGSVGElement>) {
    return (
        <svg
            viewBox="0 0 24 24"
            fill="currentColor"
            aria-hidden="true"
            focusable="false"
            {...props}
        >
            <path d="M12 .5a12 12 0 0 0-3.79 23.39c.6.1.82-.26.82-.58v-2.23c-3.34.73-4.04-1.43-4.04-1.43-.55-1.38-1.34-1.74-1.34-1.74-1.1-.75.08-.74.08-.74 1.21.09 1.86 1.22 1.86 1.22 1.08 1.82 2.84 1.3 3.53.99.11-.77.42-1.3.76-1.6-2.67-.3-5.47-1.32-5.47-5.88 0-1.3.47-2.35 1.23-3.18-.13-.3-.53-1.52.12-3.17 0 0 1-.32 3.3 1.22a11.52 11.52 0 0 1 6 0c2.3-1.54 3.3-1.22 3.3-1.22.65 1.65.25 2.87.12 3.17.77.83 1.23 1.88 1.23 3.18 0 4.57-2.8 5.58-5.48 5.88.43.37.81 1.1.81 2.22v3.29c0 .32.22.69.83.58A12 12 0 0 0 12 .5Z" />
        </svg>
    )
}

export function GlobalLinks({
    className,
    includeDocsLink = true,
    variant = "ghost",
    size = "sm",
}: GlobalLinksProps) {
    return (
        <div className={cn("flex items-center gap-2", className)}>
            <Button variant={variant} size={size} asChild>
                <Link
                    href="https://github.com/jcodagnone/chapauy"
                    target="_blank"
                    rel="noopener noreferrer"
                    title="Ver en GitHub"
                >
                    <GitHubMark className="h-4 w-4" />
                    <span className="sr-only">GitHub</span>
                </Link>
            </Button>
            {includeDocsLink && (
                <Button variant={variant} size={size} asChild>
                    <Link href="/docs" title="Documentación">
                        <BookOpen className="h-4 w-4" />
                        <span className="sr-only">Documentación</span>
                    </Link>
                </Button>
            )}
        </div>
    )
}
