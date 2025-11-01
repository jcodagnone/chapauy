/**
 * Copyright 2025 The ChapaUY Authors
 * SPDX-License-Identifier: Apache-2.0
 */

import { MDXRemote } from "next-mdx-remote/rsc"
import Link from "next/link"
import { cn } from "@/lib/utils"
import rehypeSlug from "rehype-slug"
import remarkGfm from "remark-gfm"
import { FileCode } from "lucide-react"

const components = {
  section: ({ className, ...props }: React.HTMLAttributes<HTMLElement>) => {
    if (
      className === "footnotes" ||
      (props as any)["data-footnotes"] !== undefined
    ) {
      return (
        <section className={cn("mt-10 border-t pt-8", className)} {...props}>
          <h2 className="mb-4 text-2xl font-semibold tracking-tight">Notas</h2>
          {props.children}
        </section>
      )
    }
    return <section className={className} {...props} />
  },
  h1: ({ className, ...props }: React.HTMLAttributes<HTMLHeadingElement>) => (
    <h1
      className={cn(
        "mt-2 scroll-m-20 text-4xl font-bold tracking-tight",
        className
      )}
      {...props}
    />
  ),
  h2: ({ className, ...props }: React.HTMLAttributes<HTMLHeadingElement>) => (
    <h2
      className={cn(
        "mt-10 scroll-m-20 border-b pb-1 text-3xl font-semibold tracking-tight first:mt-0",
        className
      )}
      {...props}
    />
  ),
  h3: ({ className, ...props }: React.HTMLAttributes<HTMLHeadingElement>) => (
    <h3
      className={cn(
        "mt-8 scroll-m-20 text-2xl font-semibold tracking-tight",
        className
      )}
      {...props}
    />
  ),
  h4: ({ className, ...props }: React.HTMLAttributes<HTMLHeadingElement>) => (
    <h4
      className={cn(
        "mt-8 scroll-m-20 text-xl font-semibold tracking-tight",
        className
      )}
      {...props}
    />
  ),
  p: ({ className, ...props }: React.HTMLAttributes<HTMLParagraphElement>) => (
    <p
      className={cn("leading-7 [&:not(:first-child)]:mt-6 text-justify", className)}
      {...props}
    />
  ),
  ul: ({ className, ...props }: React.HTMLAttributes<HTMLUListElement>) => (
    <ul
      className={cn("my-6 ml-6 list-disc [&>li]:mt-2", className)}
      {...props}
    />
  ),
  ol: ({ className, ...props }: React.HTMLAttributes<HTMLOListElement>) => (
    <ol
      className={cn("my-6 ml-6 list-decimal [&>li]:mt-2", className)}
      {...props}
    />
  ),
  li: ({ className, ...props }: React.HTMLAttributes<HTMLLIElement>) => (
    <li className={cn("mt-2 text-justify [&_ul]:my-2", className)} {...props} />
  ),
  blockquote: ({
    className,
    ...props
  }: React.HTMLAttributes<HTMLQuoteElement>) => (
    <blockquote
      className={cn("mt-6 border-l-2 pl-6 italic", className)}
      {...props}
    />
  ),
  img: ({
    className,
    alt,
    ...props
  }: React.ImgHTMLAttributes<HTMLImageElement>) => (
    <img className={cn("rounded-md border", className)} alt={alt} {...props} />
  ),
  hr: ({ ...props }) => <hr className="my-4 md:my-8" {...props} />,
  table: ({ className, ...props }: React.HTMLAttributes<HTMLTableElement>) => (
    <div className="my-6 w-full overflow-y-auto">
      <table className={cn("w-full", className)} {...props} />
    </div>
  ),
  tr: ({ className, ...props }: React.HTMLAttributes<HTMLTableRowElement>) => (
    <tr
      className={cn("even:bg-muted m-0 border-t p-0", className)}
      {...props}
    />
  ),
  th: ({ className, ...props }: React.HTMLAttributes<HTMLTableCellElement>) => (
    <th
      className={cn(
        "border px-4 py-2 text-left font-bold [&[align=center]]:text-center [&[align=right]]:text-right",
        className
      )}
      {...props}
    />
  ),
  td: ({ className, ...props }: React.HTMLAttributes<HTMLTableCellElement>) => (
    <td
      className={cn(
        "border px-4 py-2 text-left [&[align=center]]:text-center [&[align=right]]:text-right",
        className
      )}
      {...props}
    />
  ),
  pre: ({ className, ...props }: React.HTMLAttributes<HTMLPreElement>) => (
    <pre
      className={cn(
        "mt-6 mb-4 overflow-x-auto rounded-lg border bg-black py-4 px-4 [&_code]:bg-transparent [&_code]:p-0",
        className
      )}
      {...props}
    />
  ),
  code: ({ className, ...props }: React.HTMLAttributes<HTMLElement>) => (
    <code
      className={cn(
        "bg-muted relative rounded px-[0.3rem] py-[0.2rem] font-mono text-sm",
        className
      )}
      {...props}
    />
  ),
  Link: ({ className, ...props }: React.ComponentProps<typeof Link>) => (
    <Link
      className={cn(
        "text-blue-600 dark:text-blue-400 font-medium no-underline hover:underline",
        className
      )}
      {...props}
    />
  ),
  a: ({
    className,
    href,
    children,
    ...props
  }: React.AnchorHTMLAttributes<HTMLAnchorElement>) => {
    const isExternal = href?.startsWith("http")

    if (isExternal && href) {
      let domain = ""
      let xDisplay = ""
      try {
        const url = new URL(href)
        domain = url.hostname.replace(/^www\./, "")
        if (domain === "x.com" || domain === "twitter.com") {
          const username = url.pathname.split("/").filter(Boolean)[0]
          if (username) {
            xDisplay = `${domain}/${username}`
          }
        }
      } catch (e) {
        // ignore invalid urls
      }

      const isGithub = domain === "github.com"

      return (
        <a
          className={cn(
            "text-blue-600 dark:text-blue-400 font-medium no-underline hover:underline",
            className
          )}
          href={href}
          target="_blank"
          rel="noopener noreferrer"
          {...props}
        >
          {children}
          {isGithub ? (
            <FileCode className="ml-1 inline h-4 w-4 align-text-bottom text-muted-foreground" />
          ) : (
            (xDisplay || domain) && (
              <span className="text-muted-foreground ml-1 no-underline text-xs">
                [{xDisplay || domain}]
              </span>
            )
          )}
        </a>
      )
    }

    return (
      <Link
        className={cn(
          "text-blue-600 dark:text-blue-400 font-medium no-underline hover:underline",
          className
        )}
        href={href as string}
        {...props}
      >
        {children}
      </Link>
    )
  },
}

export function MdxContent(props: React.ComponentProps<typeof MDXRemote>) {
  return (
    <MDXRemote
      {...props}
      components={{ ...components, ...(props.components || {}) }}
      options={{
        mdxOptions: {
          remarkPlugins: [remarkGfm],
          rehypePlugins: [
            rehypeSlug,
          ],
        },
      }}
    />
  )
}
