import Link from "next/link";
import Markdown from "@/components/markdown/Markdown";
import ContentEffects from "@/components/markdown/ContentEffects";
import { PageHeader } from "./PageHeader";
import { DocsSidebar, Toc } from "./DocsNav";
import SwaggerApi from "./SwaggerApi";
import BlogListGrid from "./BlogListGrid";
import { getBlogPosts, type Page } from "@/lib/content";
import { SITE, type Lang } from "@/lib/site";
import enDict from "@/lib/i18n/en.json";
import idDict from "@/lib/i18n/id.json";

function EditLink({ page }: { page: Page }) {
  const prefix = page.lang === "id" ? "content/id/" : "content/";
  const href = `${SITE.repo}/edit/main/site/${prefix}${page.relPath}.md`;
  return (
    <div className="edit-link">
      <a href={href} target="_blank" rel="noopener" className="edit-link-btn">
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" aria-hidden="true"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7" /><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z" /></svg>
        {page.lang === "id" ? "Edit halaman ini" : "Edit this page"}
      </a>
    </div>
  );
}

export function DocPagination({
  prev,
  next,
  lang = "en",
}: {
  prev?: { title: string; url: string };
  next?: { title: string; url: string };
  lang?: Lang;
}) {
  const dicts: Record<Lang, Record<string, string>> = { en: enDict as Record<string, string>, id: idDict as Record<string, string> };
  const t = (key: string) => dicts[lang][key] ?? key;
  if (!prev && !next) return null;
  return (
    <nav className="doc-pagination">
      {prev ? (
        <Link href={prev.url} className="doc-pagination-link prev">
          <span className="doc-pagination-label">{t("previous")}</span>
          <span className="doc-pagination-title">{prev.title}</span>
        </Link>
      ) : (
        <span />
      )}
      {next ? (
        <Link href={next.url} className="doc-pagination-link next">
          <span className="doc-pagination-label">{t("next")}</span>
          <span className="doc-pagination-title">{next.title}</span>
        </Link>
      ) : (
        <span />
      )}
    </nav>
  );
}

export function GenericPageView({ page }: { page: Page }) {
  return (
    <>
      <PageHeader page={page} lang={page.lang} />
      <section className="content-section">
        <div className="container container-narrow">
          <div className="single-content">
            <Markdown content={page.body} />
          </div>
        </div>
      </section>
      <ContentEffects />
    </>
  );
}

export function BlogListView({ page }: { page: Page }) {
  const posts = getBlogPosts(page.lang);
  return (
    <>
      <PageHeader page={page} lang={page.lang} />
      <section className="content-section">
        <div className="container">
          {page.body.trim() !== "" && (
            <div className="single-content blog-index-intro">
              <Markdown content={page.body} />
            </div>
          )}
          <BlogListGrid
            posts={posts.map((p) => ({
              url: p.url,
              title: p.title,
              date: p.date,
              categories: p.categories,
              summary: p.summary,
              readingTime: p.readingTime,
            }))}
            lang={page.lang}
          />
        </div>
      </section>
    </>
  );
}

export function ApiPageView({
  page,
  docsEntries,
  docsTitle,
  tocTitle,
}: {
  page: Page;
  docsEntries: { title: string; url: string }[];
  docsTitle: string;
  tocTitle: string;
}) {
  const specUrl = "/openapi.yaml";
  return (
    <>
      <PageHeader page={page} lang={page.lang} />
      <section className="content-section api-section">
        <div className="container">
          <div className="doc-layout">
            <DocsSidebar entries={docsEntries} currentUrl={page.url} title={docsTitle} />
            <div className="single-content">
              <SwaggerApi specUrl={specUrl} lang={page.lang as Lang} />
              <Toc headings={page.headings} title={tocTitle} />
            </div>
          </div>
        </div>
      </section>
    </>
  );
}

interface DocPageExtras {
  docsEntries: { title: string; url: string }[];
  neighbors: {
    prev?: { title: string; url: string };
    next?: { title: string; url: string };
  };
  tocTitle: string;
}

export function DocPageView({
  page,
  docsEntries,
  neighbors,
  tocTitle,
}: DocPageExtras & { page: Page }) {
  return (
    <>
      <PageHeader page={page} lang={page.lang} />
      <section className="content-section">
        <div className="container">
          <div className="doc-layout">
            <DocsSidebar entries={docsEntries} currentUrl={page.url} title={page.lang === "id" ? "Dokumen" : "Docs"} lang={page.lang as Lang} />
            <article className="single-content">
              <Toc headings={page.headings} title={tocTitle} />
              <Markdown content={page.body} />
              <EditLink page={page} />
              <DocPagination prev={neighbors.prev} next={neighbors.next} lang={page.lang as Lang} />
            </article>
          </div>
        </div>
      </section>
      <ContentEffects />
    </>
  );
}

function SocialShare({ page }: { page: Page }) {
  const url = encodeURIComponent(`${SITE.baseUrl}${page.url}/`);
  const text = encodeURIComponent(page.title);
  return (
    <div className="social-share">
      <a
        href={`https://twitter.com/intent/tweet?text=${text}&url=${url}`}
        target="_blank"
        rel="noopener"
        className="btn btn-secondary btn-sm"
      >
        Share on X
      </a>
      <a
        href={`https://www.linkedin.com/sharing/share-offsite/?url=${url}`}
        target="_blank"
        rel="noopener"
        className="btn btn-secondary btn-sm"
      >
        Share on LinkedIn
      </a>
    </div>
  );
}

export function BlogPostView({ page }: { page: Page }) {
  return (
    <>
      <PageHeader page={page} lang={page.lang} />
      <section className="content-section">
        <div className="container container-narrow">
          <div className="blog-meta">
            {page.author && (
              <span className="blog-author">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" aria-hidden="true"><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2" /><circle cx="12" cy="7" r="4" /></svg>
                {page.author}
              </span>
            )}
            {page.date && (
              <time dateTime={new Date(page.date).toISOString()}>
                {new Date(page.date).toLocaleDateString(
                  page.lang === "id" ? "id-ID" : "en-US",
                  { year: "numeric", month: "long", day: "numeric" },
                )}
              </time>
            )}
            <span>{page.readingTime} {page.lang === "id" ? "menit baca" : "min read"}</span>
          </div>
          <div className="single-content blog-content">
            <Markdown content={page.body} />
          </div>
          <SocialShare page={page} />
        </div>
      </section>
      <ContentEffects />
    </>
  );
}
