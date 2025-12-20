/**
 * Copyright 2025 The ChapaUY Authors
 * SPDX-License-Identifier: Apache-2.0
 */

"use client"

import { useState } from "react"
import { Menu } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Sheet, SheetContent, SheetTrigger, SheetTitle } from "@/components/ui/sheet"
import { OffensesSidebarClient } from "@/components/offenses-sidebar-client"

export function MobileNav() {
    const [open, setOpen] = useState(false)

    return (
        <div className="flex items-center justify-between border-b p-4 md:hidden">
            <div className="flex items-center gap-2">
                <h1 className="text-lg font-semibold">ChapaUY</h1>
            </div>
            <Sheet open={open} onOpenChange={setOpen}>
                <SheetTrigger asChild>
                    <Button variant="ghost" size="icon" aria-label="Abrir menú">
                        <Menu className="h-6 w-6" />
                    </Button>
                </SheetTrigger>
                <SheetContent side="left" className="w-80 p-0">
                    <SheetTitle className="sr-only">Menú de navegación</SheetTitle>
                    <OffensesSidebarClient
                        onClose={() => setOpen(false)}
                        mode="mobile"
                    />
                </SheetContent>
            </Sheet>
        </div>
    )
}
