/**
 * Copyright 2025 The ChapaUY Authors
 * SPDX-License-Identifier: Apache-2.0
 */

export interface DbReference {
  id: number
  name: string
}

export const databases: DbReference[] = [
  { id: 65, name: "Caminera" },
  { id: 40, name: "Canelones" },
  { id: 48, name: "Colonia" },
  { id: 26, name: "Lavalleja" },
  { id: 45, name: "Maldonado" },
  { id: 6, name: "Montevideo" },
  { id: 43, name: "Paysandu" },
  { id: 55, name: "Rio Negro" },
  { id: 49, name: "Soriano" },
  { id: 56, name: "Tacuarembó" },
  { id: 52, name: "Treinta y Tres" },
  { id: 68, name: "Vialidad" },
]

export function getDBName(id: number): string {
  const db = databases.find((d) => d.id === id)
  if (!db) {
    return `Unknown DB (${id})`
  }
  return db.name
}

export const countryDisplay: Record<string, string> = {
  UY: "🇺🇾 Uruguay",
  AR: "🇦🇷 Argentina",
  BR: "🇧🇷 Brasil",
  PY: "🇵🇾 Paraguay",
  CL: "🇨🇱 Chile",
}
