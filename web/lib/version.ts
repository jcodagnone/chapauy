/**
 * Copyright 2025 The ChapaUY Authors
 * SPDX-License-Identifier: Apache-2.0
 */

import { getDatabaseVersion } from "./repository"

export async function getAppVersion(): Promise<string> {
    const dbVersion = await getDatabaseVersion()
    // In a real build, we might inject GIT_COMMIT_SHA env var
    // For now, let's assume it might be there or fallback to 'dev'
    const gitSha = process.env.GIT_COMMIT_SHA
        ? process.env.GIT_COMMIT_SHA.substring(0, 7)
        : "dev"

    return `${gitSha}-${dbVersion}`
}
