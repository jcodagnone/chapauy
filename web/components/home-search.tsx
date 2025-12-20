/**
 * Copyright 2025 The ChapaUY Authors
 * SPDX-License-Identifier: Apache-2.0
 */

"use client"

import * as React from "react"
import { useRouter } from "next/navigation"
import { Search, Loader2, ChevronDown, X } from "lucide-react"
import { Dimension } from "@/lib/types"
import type { FacetValue, Facet } from "@/lib/types"
import { getDimensionConfig } from "@/lib/display-config"
import { cn, normalizeVehicleId } from "@/lib/utils"
import { Button } from "@/components/ui/button"
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { useDebounce } from "@/hooks/use-debounce"

const AVAILABLE_DIMENSIONS = [
    Dimension.Vehicle,
    Dimension.Location,
    Dimension.Description,
    Dimension.VehicleType,
    Dimension.Year,
    Dimension.Country,
    Dimension.ArticleCode,
    Dimension.ArticleID,
]

export function HomeSearch() {
    const router = useRouter()
    const [dimension, setDimension] = React.useState<Dimension>(Dimension.Vehicle)
    const [query, setQuery] = React.useState("")
    const [suggestions, setSuggestions] = React.useState<FacetValue[]>([])
    const [loading, setLoading] = React.useState(false)
    const [isOpen, setIsOpen] = React.useState(false)
    const wrapperRef = React.useRef<HTMLDivElement>(null)

    const [selectedIndex, setSelectedIndex] = React.useState(-1)

    const debouncedQuery = useDebounce(query, 300)

    React.useEffect(() => {
        async function fetchSuggestions() {
            setLoading(true)
            try {
                const params = new URLSearchParams()
                params.set("dimension", dimension)
                if (debouncedQuery) {
                    params.set("q", debouncedQuery)
                }

                const res = await fetch(`/api/v1/suggest?${params.toString()}`)
                if (!res.ok) throw new Error("Failed to fetch")
                const data: Facet = await res.json()
                setSuggestions(data.values || [])

                // Reset selected index when suggestions update
                setSelectedIndex(-1)

                if (debouncedQuery) {
                    setIsOpen(true)
                }
            } catch (error) {
                console.error("Error fetching suggestions:", error)
                setSuggestions([])
            } finally {
                setLoading(false)
            }
        }

        fetchSuggestions()
    }, [debouncedQuery, dimension])

    // ... (handleClickOutside)

    const handleSearch = (value?: string) => {
        const searchTerm = value !== undefined ? value : query
        // ... (existing logic)
        // Check if value is undefined and we have a selected index
        if (value === undefined && selectedIndex >= 0 && suggestions[selectedIndex]) {
            const selectedItem = suggestions[selectedIndex]
            const finalTerm = selectedItem.value
            // Update input to match selected item
            setQuery(finalTerm)

            const params = new URLSearchParams()
            // Standard search will handle normalization on the server
            params.set(dimension, finalTerm) // Use finalTerm not searchTerm (which is just query)

            router.push(`/offenses?${params.toString()}`)
            setIsOpen(false)
            return
        }

        // ... (rest of existing logic using searchTerm)
        if (value === undefined && !query) return

        const params = new URLSearchParams()
        params.set(dimension, searchTerm)
        router.push(`/offenses?${params.toString()}`)
    }

    const handleKeyDown = (e: React.KeyboardEvent) => {
        if (e.key === "ArrowDown") {
            e.preventDefault()
            setSelectedIndex((prev) => (prev + 1) % suggestions.length)
        } else if (e.key === "ArrowUp") {
            e.preventDefault()
            setSelectedIndex((prev) => (prev - 1 + suggestions.length) % suggestions.length)
        } else if (e.key === "Enter") {
            e.preventDefault() // Prevent form submission if inside a form (though it's not) and ensure custom handler runs
            handleSearch()
            setIsOpen(false)
        } else if (e.key === "Escape") {
            setIsOpen(false)
            setSelectedIndex(-1)
        }
    }

    const clearSearch = () => {
        setQuery("")
        setSuggestions([])
        setIsOpen(false)
        setSelectedIndex(-1) // Reset index
    }

    const currentConfig = getDimensionConfig(dimension)

    return (
        <div className="relative w-full max-w-2xl px-4" ref={wrapperRef}>
            <div className="bg-background ring-border focus-within:ring-primary flex h-14 w-full items-center gap-2 rounded-full border shadow-sm transition-all focus-within:ring-2 focus-within:shadow-md">

                {/* Search Icon */}
                <div className="pl-4">
                    <Search className="text-muted-foreground h-5 w-5" />
                </div>

                {/* Dimension Selector */}
                <div className="flex-shrink-0">
                    <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                            <Button
                                variant="ghost"
                                className="text-muted-foreground hover:text-foreground h-10 gap-2 rounded-full px-3 font-normal"
                            >
                                <currentConfig.icon className="h-4 w-4" />
                                <span className="hidden sm:inline-block">Por {currentConfig.label}</span>
                                <ChevronDown className="h-3 w-3 opacity-50" />
                            </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="start" className="w-[200px]">
                            {AVAILABLE_DIMENSIONS.map((dim) => {
                                const config = getDimensionConfig(dim)
                                return (
                                    <DropdownMenuItem
                                        key={dim}
                                        onClick={() => {
                                            setDimension(dim)
                                            setQuery("") // Reset query on dimension change
                                            // Focus input after dimension change
                                            setTimeout(() => {
                                                const input = wrapperRef.current?.querySelector('input')
                                                if (input) {
                                                    input.focus()
                                                    // Trigger suggestions fetch by setting query to empty string (already done)
                                                    // We might need to manually trigger isOpen for top results if query was already empty
                                                    setIsOpen(true)
                                                }
                                            }, 0)
                                        }}
                                        className={cn(
                                            "flex items-center gap-2",
                                            dimension === dim && "bg-accent"
                                        )}
                                    >
                                        <config.icon className="text-muted-foreground h-4 w-4" />
                                        <span>Por {config.label}</span>
                                    </DropdownMenuItem>
                                )
                            })}
                        </DropdownMenuContent>
                    </DropdownMenu>
                </div>

                {/* Input */}
                <div className="relative flex-1">
                    <input
                        type="text"
                        value={query}
                        onChange={(e) => {
                            setQuery(e.target.value)
                            if (!isOpen && e.target.value.length >= 2) setIsOpen(true)
                        }}
                        onKeyDown={handleKeyDown}
                        onFocus={() => {
                            if (suggestions.length > 0) setIsOpen(true)
                        }}
                        placeholder={`Buscar ${currentConfig.label.toLowerCase()}...`}
                        className="placeholder:text-muted-foreground/50 h-full w-full bg-transparent text-lg outline-none"
                        autoFocus
                    />
                </div>

                {/* Actions */}
                <div className="pr-4 flex items-center gap-2">
                    {loading ? (
                        <Loader2 className="text-muted-foreground h-5 w-5 animate-spin" />
                    ) : query ? (
                        <Button
                            variant="ghost"
                            size="icon"
                            className="h-8 w-8 hover:bg-transparent"
                            onClick={clearSearch}
                        >
                            <X className="text-muted-foreground hover:text-foreground h-5 w-5" />
                        </Button>
                    ) : null}
                </div>
            </div>

            {/* Suggestions Dropdown */}
            {isOpen && suggestions.length > 0 && (
                <div className="bg-popover text-popover-foreground absolute top-16 right-4 left-4 z-50 overflow-hidden rounded-xl border shadow-lg animate-in fade-in-0 zoom-in-95">
                    <div className="max-h-[300px] overflow-y-auto py-2">
                        {suggestions.map((item, i) => (
                            <button
                                key={`${item.value}-${i}`}
                                className={cn(
                                    "flex w-full items-center gap-3 px-4 py-3 text-left text-sm transition-colors",
                                    "hover:bg-accent hover:text-accent-foreground",
                                    i === selectedIndex && "bg-accent text-accent-foreground"
                                )}
                                onClick={() => {
                                    setQuery(item.value) // Update input for visual feedback
                                    handleSearch(item.value)
                                    setIsOpen(false)
                                }}
                            >
                                <Search className="text-muted-foreground h-4 w-4" />
                                <div className="flex flex-1 flex-col items-start">
                                    {!item.value ? (
                                        <span className="text-muted-foreground bg-muted rounded px-2 py-0.5 text-xs italic">
                                            {currentConfig.empty}
                                        </span>
                                    ) : (
                                        <span className="font-medium">{item.label || item.value}</span>
                                    )}
                                    {item.count !== undefined && (
                                        <span className="text-muted-foreground text-xs">
                                            {item.count.toLocaleString()} resultados
                                        </span>
                                    )}
                                </div>
                            </button>
                        ))}
                    </div>
                </div>
            )}
        </div>
    )
}
