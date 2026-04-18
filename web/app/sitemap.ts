/**
 * Copyright 2025 The ChapaUY Authors
 * SPDX-License-Identifier: Apache-2.0
 */

import { MetadataRoute } from 'next'
import { BASE_URL } from '@/lib/config'

export default function sitemap(): MetadataRoute.Sitemap {
  const baseUrl = BASE_URL

  return [
    {
      url: baseUrl,
      lastModified: new Date(),
      changeFrequency: 'daily',
      priority: 1,
    },
    {
      url: `${baseUrl}/offenses`,
      lastModified: new Date(),
      changeFrequency: 'daily',
      priority: 0.8,
    },
    {
      url: `${baseUrl}/docs`,
      lastModified: new Date(),
      changeFrequency: 'weekly',
      priority: 0.5,
    },
  ]
}
