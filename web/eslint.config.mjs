/**
 * Copyright 2025 The ChapaUY Authors
 * SPDX-License-Identifier: Apache-2.0
 */

import { defineConfig, globalIgnores } from "eslint/config"
import nextCoreWebVitals from "eslint-config-next/core-web-vitals"
import nextTypeScript from "eslint-config-next/typescript"
import prettier from "eslint-config-prettier/flat"

export default defineConfig([
  ...nextCoreWebVitals,
  ...nextTypeScript,
  prettier,
  globalIgnores([".next/**", "out/**", "build/**", "next-env.d.ts"]),
])
