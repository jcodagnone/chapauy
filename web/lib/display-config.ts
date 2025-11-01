/**
 * Copyright 2025 The ChapaUY Authors
 * SPDX-License-Identifier: Apache-2.0
 */

import {
  Database,
  Calendar,
  Globe,
  Truck,
  Car,
  FileText,
  MapPin,
  AlignLeft,
  File,
  Hash,
  Clock,
} from "lucide-react"
import { Dimension } from "@/lib/types"
import type { LucideIcon } from "lucide-react"

// Dimension configuration
export interface DimensionConfig {
  icon: LucideIcon
  label: string
  empty: string
}

const dimensionConfig: Record<Dimension, DimensionConfig> = {
  [Dimension.Database]: {
    icon: Database,
    label: "Base de datos",
    empty: "Sin base de datos",
  },
  [Dimension.Year]: {
    icon: Calendar,
    label: "Año",
    empty: "Sin año",
  },
  [Dimension.Country]: {
    icon: Globe,
    label: "País de la matrícula",
    empty: "Sin información",
  },
  [Dimension.VehicleType]: {
    icon: Truck,
    label: "Tipo de vehículo",
    empty: "Sin información",
  },

  [Dimension.Location]: {
    icon: MapPin,
    label: "Ubicación",
    empty: "Sin ubicación",
  },
  [Dimension.Vehicle]: {
    icon: Car,
    label: "Matrícula",
    empty: "Sin matrícula",
  },
  [Dimension.Description]: {
    icon: AlignLeft,
    label: "Descripción",
    empty: "Sin descripción",
  },
  [Dimension.ArticleCode]: {
    icon: AlignLeft,
    label: "Capítulo",
    empty: "Sin capítulo",
  },
  [Dimension.ArticleID]: {
    icon: AlignLeft,
    label: "Artículo",
    empty: "Sin artículo",
  },
  [Dimension.DocSource]: {
    icon: File,
    label: "Fuente",
    empty: "Sin fuente",
  },
  [Dimension.Features]: {
    icon: File,
    label: "Características",
    empty: "Sin características",
  },
  [Dimension.Date]: {
    icon: Clock,
    label: "Fecha",
    empty: "Sin fecha",
  },
}

export function getDimensionConfig(dimension: string): DimensionConfig {
  return (
    dimensionConfig[dimension as Dimension] ||
    dimensionConfig[Dimension.Description]
  )
}

// Country configuration
interface CountryData {
  flag: string
  name: string
}

const COUNTRY_DATA: Record<string, CountryData> = {
  UY: { flag: "🇺🇾", name: "Uruguay" },
  AR: { flag: "🇦🇷", name: "Argentina" },
  BR: { flag: "🇧🇷", name: "Brasil" },
  PY: { flag: "🇵🇾", name: "Paraguay" },
  CL: { flag: "🇨🇱", name: "Chile" },
}

export function getCountryFlag(countryCode: string): string {
  return COUNTRY_DATA[countryCode]?.flag || "🏳️"
}

export function getCountryName(countryCode: string): string {
  return COUNTRY_DATA[countryCode]?.name || countryCode
}

export function getCountryLabel(countryCode: string): string {
  const data = COUNTRY_DATA[countryCode]
  return data ? `${data.flag} ${data.name}` : countryCode
}
