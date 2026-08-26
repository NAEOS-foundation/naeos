import Link from "next/link";
import enDict from "@/lib/i18n/en.json";
import idDict from "@/lib/i18n/id.json";
import NewsletterForm from "./NewsletterForm";
import CookieBanner from "./CookieBanner";
import { SITE, type Lang } from "@/lib/site";

type Dict = Record<string, string>;
const DICTS: Record<Lang, Dict> = {
  en: enDict as Dict,
  id: idDict as Dict,
};

interface FooterProps {
  lang: Lang;
  tagline: string;
}

function GithubIcon() {
  return (
    <svg width="20" height="20" viewBox="0 0 24 24" fill="currentColor"><path d="M12 0C5.37 0 0 5.37 0 12c0 5.31 3.435 9.795 8.205 11.385.6.105.825-.255.825-.57 0-.285-.015-1.23-.015-2.235-3.015.555-3.795-.735-4.035-1.41-.135-.345-.72-1.41-1.23-1.695-.42-.225-1.02-.78-.015-.795.945-.015 1.62.87 1.845 1.23 1.08 1.815 2.805 1.305 3.495.99.105-.78.42-1.305.765-1.605-2.67-.3-5.46-1.335-5.46-5.925 0-1.305.465-2.385 1.23-3.225-.12-.3-.54-1.53.12-3.18 0 0 1.005-.315 3.3 1.23.96-.27 1.98-.405 3-.405s2.04.135 3 .405c2.295-1.56 3.3-1.23 3.3-1.23.66 1.65.24 2.88.12 3.18.765.84 1.23 1.905 1.23 3.225 0 4.605-2.805 5.625-5.475 5.925.435.375.81 1.095.81 2.22 0 1.605-.015 2.895-.015 3.3 0 .315.225.69.825.57A12.02 12.02 0 0 0 24 12c0-6.63-5.37-12-12-12z" /></svg>
  );
}

export default function Footer({ lang, tagline }: FooterProps) {
  const t = (key: string, fallback?: string) => DICTS[lang][key] ?? fallback ?? key;
  const base = lang === "en" ? "" : "/id";

  return (
    <>
      <footer className="site-footer" role="contentinfo">
        <div className="container">
          <div className="footer-grid">
            <div className="footer-col brand">
              <Link href={base === "" ? "/" : `${base}/`} className="logo">
                <img className="logo-icon" src="/images/Logo.png" width={28} height={28} loading="lazy" alt="" aria-hidden="true" />
                <span className="logo-text">NAEOS</span>
              </Link>
              <p className="tagline">{tagline}</p>
              <div className="social-links">
                <a href={SITE.repo} target="_blank" rel="noopener" aria-label="GitHub"><GithubIcon /></a>
                <a href={SITE.twitter} target="_blank" rel="noopener" aria-label="Twitter / X">
                  <svg width="20" height="20" viewBox="0 0 24 24" fill="currentColor"><path d="M18.244 2.25h3.308l-7.227 8.26 8.502 11.24H16.17l-5.214-6.817L4.99 21.75H1.68l7.73-8.835L1.254 2.25H8.08l4.713 6.231zm-1.161 17.52h1.833L7.084 4.126H5.117z" /></svg>
                </a>
                <a href="https://discord.gg/naeos" target="_blank" rel="noopener" aria-label="Discord">
                  <svg width="20" height="20" viewBox="0 0 24 24" fill="currentColor"><path d="M20.317 4.3698a19.7913 19.7913 0 00-4.8851-1.5152.0741.0741 0 00-.0785.0371c-.211.3753-.4447.8648-.6083 1.2495-1.8447-.2762-3.68-.2762-5.4868 0-.1636-.3933-.4058-.8742-.6177-1.2495a.077.077 0 00-.0785-.037 19.7363 19.7363 0 00-4.8852 1.515.0699.0699 0 00-.0321.0277C.5334 9.0458-.319 13.5799.0992 18.0578a.0824.0824 0 00.0312.0561c2.0528 1.5076 4.0413 2.4228 5.9929 3.0294a.0777.0777 0 00.0842-.0276c.4616-.6304.8731-1.2952 1.226-1.9942a.076.076 0 00-.0416-.1057c-.6528-.2476-1.2743-.5495-1.8722-.8923a.077.077 0 01-.0076-.1277c.1258-.0943.2517-.1923.3718-.2914a.0743.0743 0 01.0776-.0105c3.9278 1.7933 8.18 1.7933 12.0614 0a.0739.0739 0 01.0785.0095c.1202.099.246.1981.3728.2924a.077.077 0 01-.0066.1276 12.2986 12.2986 0 01-1.873.8914.0766.0766 0 00-.0407.1067c.3604.698.7719 1.3628 1.225 1.9932a.076.076 0 00.0842.0286c1.961-.6067 3.9495-1.5219 6.0023-3.0294a.077.077 0 00.0313-.0552c.5004-5.177-.8382-9.6739-3.5485-13.6604a.061.061 0 00-.0312-.0286zM8.02 15.3312c-1.1825 0-2.1569-1.0857-2.1569-2.419 0-1.3332.9555-2.4189 2.157-2.4189 1.2108 0 2.1757 1.0952 2.1568 2.419 0 1.3332-.9555 2.4189-2.1569 2.4189zm7.9748 0c-1.1825 0-2.1569-1.0857-2.1569-2.419 0-1.3332.9554-2.4189 2.1569-2.4189 1.2108 0 2.1757 1.0952 2.1568 2.419 0 1.3332-.946 2.4189-2.1568 2.4189z" /></svg>
                </a>
              </div>
              <NewsletterForm lang={lang} />
            </div>
            <div className="footer-col">
              <h4>{t("footer_product")}</h4>
              <a href={`${base}/features`}>{t("nav_features")}</a>
              <a href={`${base}/download`}>{t("nav_download")}</a>
              <a href={`${SITE.repo}/releases`} target="_blank" rel="noopener">{t("footer_releases")}</a>
              <a href={`${base}/roadmap`}>{t("footer_roadmap")}</a>
              <a href={`${base}/blog`}>{t("nav_blog")}</a>
              <a href={`${base}/plugins`}>{t("nav_plugins")}</a>
              <a href={`${base}/showcase`}>{t("nav_showcase")}</a>
              <a href={`${base}/cookbook`}>{t("nav_cookbook")}</a>
              <a href={`${base}/use-cases`}>{t("footer_use_cases")}</a>
              <a href={`${base}/comparison`}>{t("nav_comparison")}</a>
            </div>
            <div className="footer-col">
              <h4>{t("footer_docs")}</h4>
              <a href={`${base}/docs/getting-started`}>{t("footer_getting_started")}</a>
              <a href={`${base}/docs/installation`}>{t("footer_installation")}</a>
              <a href={`${base}/docs/cli-reference`}>{t("footer_cli_reference")}</a>
              <a href={`${base}/docs/architecture`}>{t("footer_architecture")}</a>
              <a href={`${base}/docs/spec-language`}>{t("nav_spec_language")}</a>
              <a href={`${base}/templates`}>Templates</a>
              <a href={`${base}/schemaregistry`}>Schema Registry</a>
              <a href={`${base}/faq`}>{t("footer_faq")}</a>
            </div>
            <div className="footer-col">
              <h4>{t("footer_community")}</h4>
              <a href={`${base}/community`}>{t("nav_community")}</a>
              <a href={`${base}/about`}>{t("footer_about")}</a>
              <a href={`${SITE.repo}/blob/main/CONTRIBUTING.md`} target="_blank" rel="noopener">{t("footer_contribute")}</a>
              <a href={`${SITE.repo}/discussions`} target="_blank" rel="noopener">{t("footer_discussions")}</a>
              <a href={`${SITE.repo}/issues`} target="_blank" rel="noopener">{t("footer_issues")}</a>
              <a href={`${base}/blog/index.xml`}>{t("footer_rss")}</a>
            </div>
          </div>
          <div className="footer-bottom">
            <p>{SITE.copyright}</p>
            <div style={{ display: "flex", gap: "1.5rem", alignItems: "center", flexWrap: "wrap" }}>
              <a href={`${base}/privacy`} style={{ color: "var(--color-text-dim)", fontSize: "var(--font-size-sm)" }}>{t("privacy_title", "Privacy")}</a>
              <a href={`${base}/terms`} style={{ color: "var(--color-text-dim)", fontSize: "var(--font-size-sm)" }}>{t("terms_title", "Terms")}</a>
              <a href={`${base}/cookie-policy`} style={{ color: "var(--color-text-dim)", fontSize: "var(--font-size-sm)" }}>{t("footer_cookies", "Cookies")}</a>
              <a href={`${base}/disclaimer`} style={{ color: "var(--color-text-dim)", fontSize: "var(--font-size-sm)" }}>{t("footer_disclaimer", "Disclaimer")}</a>
              <span className="license">Apache License 2.0</span>
            </div>
          </div>
        </div>
      </footer>
      <CookieBanner lang={lang} />
    </>
  );
}
