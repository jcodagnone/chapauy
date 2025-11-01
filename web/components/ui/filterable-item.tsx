/**
 * Copyright 2025 The ChapaUY Authors
 * SPDX-License-Identifier: Apache-2.0
 */

import Link from "next/link"

interface FilterableItemProps {
  children: React.ReactNode
  href: string
  title?: string
}

export function FilterableItem({ children, href, title }: FilterableItemProps) {
  return (
    <Link
      href={href}
      title={title}
      prefetch={false}
      className="hover:text-foreground inline-flex cursor-pointer items-center gap-1 transition-colors hover:underline"
    >
      {children}
    </Link>
  )
}
