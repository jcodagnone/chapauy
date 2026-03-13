/**
 * Copyright 2025 The ChapaUY Authors
 * SPDX-License-Identifier: Apache-2.0
 */

import { defineConfig } from "vitest/config"
import react from "@vitejs/plugin-react"

export default defineConfig({
  plugins: [react()],
  devtools: true,
  test: {
    environment: "jsdom",
    globals: true,
    setupFiles: [],
    include: ["**/*.test.{ts,tsx}"],
    exclude: ["node_modules", ".next", "backend"],
  },
  resolve: {
    tsconfigPaths: true,
  },
})
