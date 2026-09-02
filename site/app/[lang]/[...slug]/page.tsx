import type { Metadata } from "next";
import { notFound } from "next/navigation";
import {
  getPage,
  getAllPages,
  localizedDocsOrder,
  docsNeighbors,
} from "@/lib/content";
import { LANGUAGES, DEFAULT_LANG, type Lang } from "@/lib/site";
import {
  GenericPageView,
  DocPageView,
  ApiPageView,
  BlogPostView,
  BlogListView,
} from "@/components/views/PageViews";
import { PageHeader } from "@/components/views/PageHeader";
import PluginsGrid, { PluginPublishSteps } from "@/components/views/PluginsGrid";
import TemplatesGrid from "@/components/views/TemplatesGrid";
import StatusView from "@/components/views/StatusView";
import {
  breadcrumbJsonLd,
  faqJsonLd,
  extractFaqItems,
  techArticleJsonLd,
  blogPostingJsonLd,
  pageMetadata,
} from "@/lib/metadata";
import pluginData from "@/data/plugins.json";
import templateData from "@/data/templates/registry.json";

/** Rebuild the content-cache URL ("/id/blog" for id, unprefixed for en) from route params. */
function cacheUrl(lang: string, slug?: string[]): string {
  const rest = (slug ?? []).join("/");
  return lang === DEFAULT_LANG ? `/${rest}` : `/${lang}/${rest}`;
}

export function generateStaticParams() {
  return getAllPages()
    .filter((p) => p.url !== "/")
    .map((p) => ({
      lang: p.lang,
      slug:
        p.lang === DEFAULT_LANG
          ? p.url.split("/").filter(Boolean)
          : p.url.split("/").filter(Boolean).slice(1),
    }))
    .filter((p) => p.slug.length > 0);
}

export async function generateMetadata(props: {
  params: Promise<{ lang: string; slug: string[] }>;
}): Promise<Metadata> {
  const { lang: rawLang, slug } = await props.params;
  const lang = (LANGUAGES as readonly string[]).includes(rawLang) ? (rawLang as Lang) : DEFAULT_LANG;
  const page = getPage(cacheUrl(lang, slug), lang) ?? null;
  return pageMetadata(page, lang);
}

interface Crumb {
  name: string;
  url: string;
}

function buildCrumbs(page: ReturnType<typeof getPage> & {}): Crumb[] {
  const base = page.lang === "en" ? "" : "/id";
  const crumbs: Crumb[] = [{ name: "Home", url: base === "" ? "/" : `${base}/` }];
  const segments = page.url.split("/").filter(Boolean);
  let acc = base;
  segments.forEach((seg, i) => {
    acc += `/${seg}`;
    const last = i === segments.length - 1;
    crumbs.push({
      name:
        last && page.title.length < 40
          ? page.title
          : seg.replace(/-/g, " "),
      url: acc,
    });
  });
  return crumbs;
}

function JsonLd({ data }: { data: object }) {
  return (
    <script
      type="application/ld+json"
      dangerouslySetInnerHTML={{ __html: JSON.stringify(data) }}
    />
  );
}

export default async function ContentPage(props: {
  params: Promise<{ lang: string; slug?: string[] }>;
}) {
  const { lang: rawLang, slug } = await props.params;
  const lang = (LANGUAGES as readonly string[]).includes(rawLang) ? (rawLang as Lang) : DEFAULT_LANG;
  const page = getPage(cacheUrl(lang, slug), lang);
  if (!page) notFound();

  const crumbs = buildCrumbs(page);
  const tocTitle = lang === "id" ? "Di halaman ini" : "On this page";

  if (page.section === "docs" && page.layout !== "api" && !page.isIndex) {
    const words = page.plain.split(" ").length;
    return (
      <>
        <DocPageView
          page={page}
          docsEntries={localizedDocsOrder(lang)}
          neighbors={docsNeighbors(page.url, lang)}
          tocTitle={tocTitle}
        />
        <JsonLd data={breadcrumbJsonLd(crumbs)} />
        {words >= 1000 && <JsonLd data={techArticleJsonLd(page)} />}
      </>
    );
  }

  if (page.section === "docs" && page.layout === "api") {
    return (
      <>
        <ApiPageView
          page={page}
          docsEntries={localizedDocsOrder(lang)}
          docsTitle={lang === "id" ? "Dokumen" : "Docs"}
          tocTitle={tocTitle}
        />
        <JsonLd data={breadcrumbJsonLd(crumbs)} />
      </>
    );
  }

  if (page.section === "blog" && !page.isIndex) {
    return (
      <>
        <BlogPostView page={page} />
        <JsonLd data={breadcrumbJsonLd(crumbs)} />
        <JsonLd data={blogPostingJsonLd(page, page.author ?? "NAEOS Foundation")} />
      </>
    );
  }

  if (page.section === "blog" && page.isIndex) {
    return <BlogListView page={page} />;
  }

  if (page.url === "/faq") {
    return (
      <>
        <GenericPageView page={page} />
        <JsonLd data={faqJsonLd(extractFaqItems(page.body))} />
      </>
    );
  }

  if (page.url === "/status") {
    return (
      <>
        <PageHeader page={page} lang={lang} />
        <StatusView lang={lang} />
      </>
    );
  }

  if (page.section === "plugins" && page.isIndex) {
    return (
      <>
        <section className="section section-first">
          <div className="container">
            <h1 className="page-title">{page.title}</h1>
            <p className="page-subtitle">{page.description}</p>
            <PluginPublishSteps lang={lang} />
            <PluginsGrid plugins={pluginData.plugins} lang={lang} />
          </div>
        </section>
        <JsonLd data={breadcrumbJsonLd(crumbs)} />
      </>
    );
  }

  if (page.section === "templates" && page.isIndex) {
    return (
      <>
        <section className="section section-first">
          <div className="container">
            <h1 className="page-title">{page.title}</h1>
            <p className="page-subtitle">{page.description}</p>
            <TemplatesGrid templates={templateData.templates} lang={lang} />
          </div>
        </section>
        <JsonLd data={breadcrumbJsonLd(crumbs)} />
      </>
    );
  }

  return (
    <>
      <GenericPageView page={page} />
      {!page.isIndex && <JsonLd data={breadcrumbJsonLd(crumbs)} />}
    </>
  );
}

export const dynamicParams = false;
