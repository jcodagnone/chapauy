/**
 * Copyright 2025 The ChapaUY Authors
 * SPDX-License-Identifier: Apache-2.0
 */
import { OffensesSidebarClient } from "@/components/offenses-sidebar-client"
import { MobileNav } from "@/components/mobile-nav"
import { OffensesFeedClient } from "./offenses-feed-client"
import { Metadata } from "next"

type Props = {
  searchParams: Promise<{ [key: string]: string | string[] | undefined }>
}

export async function generateMetadata(
  { searchParams }: Props
): Promise<Metadata> {
  const params = await searchParams
  const vehicle = params.vehicle as string | undefined
  const location = params.location as string | undefined
  const vehicleType = params.vehicle_type as string | undefined
  const year = params.year as string | undefined

  let title = "Explorar Infracciones | ChapaUY"
  let description = "Buscá y filtrá infracciones de tránsito por departamento, año, país y tipo de vehículo."

  if (vehicle) {
    title = `Infracciones para la matrícula ${vehicle} | ChapaUY`
    description = `Mirá las infracciones de tránsito detectadas para el vehículo con matrícula ${vehicle}.`
  } else if (location) {
    title = `Infracciones en ${location} | ChapaUY`
    description = `Consultá las infracciones de tránsito registradas en ${location}.`
  } else if (vehicleType) {
    title = `Infracciones para vehículos tipo ${vehicleType} | ChapaUY`
    description = `Analizá las infracciones de tránsito registradas para vehículos de tipo ${vehicleType}.`
  } else if (year) {
    title = `Infracciones del año ${year} | ChapaUY`
    description = `Explorá las infracciones de tránsito ocurridas durante el año ${year}.`
  }

  return {
    title,
    description,
    alternates: {
      canonical: vehicle ? `/offenses?vehicle=${vehicle}` : "/offenses",
    },
    openGraph: {
      title,
      description,
      url: vehicle ? `https://chapa.uy/offenses?vehicle=${vehicle}` : "https://chapa.uy/offenses",
    },
  }
}

// Force static generation - skeletons render at build time
// Force static generation - skeletons render at build time
// export const dynamic = "force-static"

export default function OffensesPage() {
  return (
    <div className="bg-background flex flex-col min-h-screen md:flex-row">
      <MobileNav />
      <OffensesSidebarClient className="hidden md:flex" />

      <main className="flex-1 w-full print:w-full min-w-0">
        <OffensesFeedClient />
      </main>
    </div>
  )
}
