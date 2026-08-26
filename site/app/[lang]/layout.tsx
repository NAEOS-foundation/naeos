import type { Metadata } from "next";
import Script from "next/script";
import { notFound } from "next/navigation";
import Header from "@/components/chrome/Header";
import Footer from "@/components/chrome/Footer";
import SearchModal from "@/components/chrome/SearchModal";
import {
  BackToTop,
  ScrollProgress,
  HeaderScrollEffect,
  ServiceWorkerRegister,
} from "@/components/chrome/Effects";
import { LANGUAGES, DEFAULT_LANG, SITE, type Lang } from "@/lib/site";
import { MAIN_ENTITY } from "@/lib/metadata";
import "@/styles/globals.css";

export function generateStaticParams() {
  return LANGUAGES.map((lang) => ({ lang }));
}

export async function generateMetadata(props: {
  params: Promise<{ lang: string }>;
}): Promise<Metadata> {
  const { lang: raw } = await props.params;
  const lang = (LANGUAGES as readonly string[]).includes(raw) ? (raw as Lang) : DEFAULT_LANG;
  return {
    title: {
      default: SITE.title[lang],
      template: `%s — ${SITE.title[lang].split(" — ")[0]}`,
    },
    description: SITE.description[lang],
    icons: {
      icon: [
        { url: "/favicon.svg", type: "image/svg+xml" },
        { url: "/favicon.ico", sizes: "any" },
      ],
      apple: "/images/icon-192.svg",
    },
    manifest: "/manifest.json",
    metadataBase: new URL(SITE.baseUrl),
    other: {
      "theme-color": SITE.accentColor,
      "apple-mobile-web-app-capable": "yes",
      "apple-mobile-web-app-status-bar-style": "black-translucent",
    },
  };
}

const themeInit = `try{var t=localStorage.getItem('theme')||'dark';document.documentElement.setAttribute('data-theme',t);}catch(e){}`;

export default async function RootLayout({
  children,
  params,
}: {
  children: React.ReactNode;
  params: Promise<{ lang: string }>;
}) {
  const { lang: raw } = await params;
  if (!(LANGUAGES as readonly string[]).includes(raw)) notFound();
  const lang = raw as Lang;
  const tagline =
    lang === "id"
      ? "Tentukan Sekali. Bangun di Mana Saja."
      : "Specify Once. Build Anywhere.";

  return (
    <html lang={lang} data-theme="dark" suppressHydrationWarning>
      <head>
        <script dangerouslySetInnerHTML={{ __html: themeInit }} />
        <link rel="preconnect" href="https://fonts.googleapis.com" />
        <link rel="preconnect" href="https://fonts.gstatic.com" crossOrigin="anonymous" />
        <link
          rel="stylesheet"
          href="https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700&family=JetBrains+Mono:wght@400;500&display=swap"
          media="print"
          // @ts-expect-error onLoad is valid on link elements
          onLoad="this.media='all'"
        />
        <noscript>
          <link
            rel="stylesheet"
            href="https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700&family=JetBrains+Mono:wght@400;500&display=swap"
          />
        </noscript>
      </head>
      <body>
        <a href="#main-content" className="skip-link">{lang === "id" ? "Lompat ke konten" : "Skip to content"}</a>
        <ScrollProgress />
        <Header lang={lang} />
        <main id="main-content" className="page-transition">
          {children}
        </main>
        <Footer lang={lang} tagline={tagline} />
        <SearchModal lang={lang} />
        <BackToTop />
        <HeaderScrollEffect />
        <ServiceWorkerRegister />
        {SITE.umamiWebsiteId && (
          <Script
            defer
            src="https://cloud.umami.is/script.js"
            data-website-id={SITE.umamiWebsiteId}
            strategy="afterInteractive"
          />
        )}
        <script
          type="application/ld+json"
          dangerouslySetInnerHTML={{ __html: JSON.stringify(MAIN_ENTITY) }}
        />
      </body>
    </html>
  );
}
