/**
 * Copyright 2025 The ChapaUY Authors
 * SPDX-License-Identifier: Apache-2.0
 */

import { type ClassValue, clsx } from "clsx"
import { twMerge } from "tailwind-merge"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

export function formatUR(urValue: number): string {
  return (urValue / 100.0).toLocaleString().replace(/\.?0+$/, "")
}

export function normalizeVehicleId(s: string): string {
  return s.replace(/\s+/g, "").toUpperCase()
}
