/**
 * Copyright 2025 The ChapaUY Authors
 * SPDX-License-Identifier: Apache-2.0
 */

import { MetadataRoute } from 'next'
import { BASE_URL } from '@/lib/config'

export default function robots(): MetadataRoute.Robots {
  return {
    rules: {
      userAgent: '*',
      allow: '/',
      disallow: '/api/',
    },
    sitemap: `${BASE_URL}/sitemap.xml`,
  }
}
