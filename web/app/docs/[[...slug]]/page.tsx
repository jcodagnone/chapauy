import { notFound } from "next/navigation"
import { MdxContent } from "@/components/mdx-content"
import { Metadata } from "next"
import { getDocBySlug, getAllDocSlugs, getDocPages } from "@/lib/docs"
import { Edit } from "lucide-react"
import { TableOfContents } from "@/components/toc"

export async function generateStaticParams() {
  const slugs = getAllDocSlugs()
  return slugs.map((slug) => ({
    slug: slug,
  }))
}

interface PageProps {
  params: Promise<{ slug?: string[] }>
}

export default async function DocsPage(props: PageProps) {
  const params = await props.params
  const slug = params.slug || []
  const doc = await getDocBySlug(slug)

  if (!doc) {
    notFound()
  }

  return (
    <article className="relative container py-6 lg:py-10">
      <div className="grid grid-cols-1 gap-10 xl:grid-cols-[1fr_200px]">
        <div className="min-w-0">
          <div className="space-y-2">
            <h1 className="scroll-m-20 text-4xl font-extrabold tracking-tight lg:text-5xl">
              {doc.title}
            </h1>
            {doc.description && (
              <p className="text-muted-foreground text-xl">{doc.description}</p>
            )}
            {doc.author && (
              <p className="text-muted-foreground text-sm">Por {doc.author}</p>
            )}
            <div className="mb-4 flex justify-end">
              <a
                href={`https://github.com/jcodagnone/chapauy/blob/master/web/docs/${doc.relativePath}`}
                target="_blank"
                rel="noopener noreferrer"
                className="text-muted-foreground hover:text-foreground flex items-center text-sm"
              >
                <Edit className="mr-2 h-4 w-4" />
                Editar en GitHub
              </a>
            </div>
          </div>
          <div className="mt-8 border-t pt-8">
            <MdxContent source={doc.content} />
          </div>
        </div>
        <div className="hidden text-sm xl:block">
          <div className="sticky top-16 -mt-10 max-h-[calc(100vh-3.5rem)] overflow-y-auto pt-10">
            <TableOfContents headings={doc.headings} />
          </div>
        </div>
      </div>
    </article>
  )
}

export async function generateMetadata(props: PageProps): Promise<Metadata> {
  const params = await props.params
  const slug = params.slug || []
  const doc = await getDocBySlug(slug)

  if (!doc) return {}

  return {
    title: `${doc.title} | Documentación | ChapaUY`,
    description: doc.description,
    authors: doc.author ? [{ name: doc.author }] : undefined,
  }
}
