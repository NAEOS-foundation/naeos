import type { Metadata } from "next";
import { SITE, type Lang } from "./site";
import type { Page } from "./content";

export const MAIN_ENTITY = {
  "@context": "https://schema.org",
  "@graph": [
    {
      "@type": "SoftwareApplication",
      name: "NAEOS",
      applicationCategory: "DeveloperApplication",
      operatingSystem: "Linux, macOS, Windows",
      offers: { "@type": "Offer", price: "0", priceCurrency: "USD" },
      softwareVersion: SITE.version,
      url: SITE.baseUrl,
      downloadUrl: `${SITE.baseUrl}/download/`,
    },
    {
      "@type": "WebSite",
      name: SITE.title.en,
      url: SITE.baseUrl,
    },
  ],
};

interface Crumb {
  name: string;
  url: string;
}

export function breadcrumbJsonLd(crumbs: Crumb[]) {
  return {
    "@context": "https://schema.org",
    "@type": "BreadcrumbList",
    itemListElement: crumbs.map((c, i) => ({
      "@type": "ListItem",
      position: i + 1,
      name: c.name,
      item: `${SITE.baseUrl}${c.url}/`,
    })),
  };
}

export function blogPostingJsonLd(
  page: Page,
  author: string,
  publisherName = "NAEOS Foundation",
) {
  return {
    "@context": "https://schema.org",
    "@type": "BlogPosting",
    headline: page.title,
    description: page.description,
    datePublished: page.date,
    dateModified: page.date,
    author: { "@type": "Organization", name: author },
    publisher: { "@type": "Organization", name: publisherName },
    mainEntityOfPage: `${SITE.baseUrl}${page.url}/`,
  };
}

export interface FaqQA {
  question: string;
  answer: string;
}

/** Extract Q&A pairs from raw-HTML faq-item markup inside markdown bodies. */
export function extractFaqItems(html: string): FaqQA[] {
  const items: FaqQA[] = [];
  const itemRe =
    /<div class="faq-item">([\s\S]*?)<div class="faq-answer">([\s\S]*?)<\/div>\s*<\/div>/g;
  let m: RegExpExecArray | null;
  while ((m = itemRe.exec(html)) !== null) {
    const qMatch = /<span>([^<]+)<\/span>/.exec(m[1]);
    if (!qMatch) continue;
    const answer = m[2]
      .replace(/<[^>]+>/g, " ")
      .replace(/\s+/g, " ")
      .trim();
    items.push({ question: qMatch[1].trim(), answer });
  }
  return items;
}

export function faqJsonLd(items: FaqQA[]) {
  return {
    "@context": "https://schema.org",
    "@type": "FAQPage",
    mainEntity: items.map((i) => ({
      "@type": "Question",
      name: i.question,
      acceptedAnswer: { "@type": "Answer", text: i.answer },
    })),
  };
}

export function techArticleJsonLd(page: Page) {
  return {
    "@context": "https://schema.org",
    "@type": "TechArticle",
    headline: page.title,
    description: page.description,
    timeRequired: `PT${page.readingTime}M`,
    url: `${SITE.baseUrl}${page.url}/`,
  };
}

export function pageMetadata(
  page: Page | null | undefined,
  lang: Lang,
  overrides?: { ogType?: "website" | "article" },
): Metadata {
  const siteTitle = SITE.title[lang];
  const title = page ? (page.isIndex && page.url === "/" ? siteTitle : `${page.title} — ${siteTitle}`) : siteTitle;
  const description = page?.description || SITE.description[lang];
  const canonicalPath = page?.url ?? "/";
  const languages: Record<string, string> = {};
  for (const l of ["en", "id"] as Lang[]) {
    languages[l] = page
      ? (page.translations[l] ?? (l === "en" ? "/" : "/id/"))
      : l === "en"
        ? "/"
        : "/id/";
  }
  return {
    title,
    description,
    authors: [{ name: "NAEOS Foundation" }],
    robots: { index: true, follow: true },
    referrer: "strict-origin-when-cross-origin",
    alternates: {
      canonical: `${SITE.baseUrl}${canonicalPath === "/" ? "" : `${canonicalPath}/`}`,
      types: { "application/rss+xml": `${SITE.baseUrl}/blog/index.xml` },
      languages,
    },
    openGraph: {
      title,
      description,
      type: overrides?.ogType ?? "website",
      url: canonicalPath === "/" ? SITE.baseUrl : `${SITE.baseUrl}${canonicalPath}/`,
      siteName: siteTitle,
      locale: lang === "id" ? "id_ID" : "en_US",
      images: [{ url: `${SITE.baseUrl}/images/og-default.png`, width: 1200, height: 630 }],
    },
    twitter: {
      card: "summary_large_image",
      site: SITE.twitterHandle,
      title,
      description,
      images: [`${SITE.baseUrl}/images/og-default.png`],
    },
  };
}
