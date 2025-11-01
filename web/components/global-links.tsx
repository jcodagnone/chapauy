"use client"

import Link from "next/link"
import { Github, BookOpen } from "lucide-react"
import { cn } from "@/lib/utils"
import { Button } from "@/components/ui/button"

interface GlobalLinksProps {
    className?: string
    includeDocsLink?: boolean
    variant?: "ghost" | "outline" | "default" | "secondary"
    size?: "default" | "sm" | "lg" | "icon"
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
                    <Github className="h-4 w-4" />
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
