/**
 * Copyright 2025 The ChapaUY Authors
 * SPDX-License-Identifier: Apache-2.0
 */

import { execSync } from 'child_process';
import fs from 'fs';
import path from 'path';

let gitCommitSha = process.env.GIT_COMMIT_SHA || 'dev';

if (gitCommitSha === 'dev') {
  try {
    const gitShaPath = path.join(process.cwd(), 'GIT_SHA');
    if (fs.existsSync(gitShaPath)) {
      gitCommitSha = fs.readFileSync(gitShaPath, 'utf-8').trim();
    }
  } catch (e) { /* ignore */ }
}

if (gitCommitSha === 'dev') {
  try {
    gitCommitSha = execSync('git rev-parse --short HEAD').toString().trim();
  } catch (e) {
    // console.warn('Could not get git commit sha', e);
  }
}

// https://nextjs.org/docs/app/api-reference/config/next-config-js
/** @type {import('next').NextConfig} */
const nextConfig = {
  env: {
    GIT_COMMIT_SHA: gitCommitSha,
  },
  // https://nextjs.org/docs/pages/api-reference/config/next-config-js/output#automatically-copying-traced-files
  output: 'standalone',
  poweredByHeader: false,
  typescript: {
    ignoreBuildErrors: true,
  },
  images: {
    unoptimized: true,
  },
  serverExternalPackages: ["duckdb"],
  cacheComponents: true,
  cacheLife: {
    "days": {
      stale: 3600, // 1 hour
      revalidate: 86400, // 1 day
      expire: 604800, // 1 week
    },
  },
  experimental: {
  },
};

export default nextConfig;
