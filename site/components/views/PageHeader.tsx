import Link from "next/link";
import type { Page } from "@/lib/content";
import type { Lang } from "@/lib/site";

export function Breadcrumbs({ page, lang }: { page: Page; lang: Lang }) {
  const base = lang === "en" ? "" : "/id";
  const homeLabel = lang === "id" ? "Beranda" : "Home";
  const crumbs: { name: string; url: string }[] = [
    { name: homeLabel, url: base === "" ? "/" : `${base}/` },
  ];
  const segments = page.url.split("/").filter(Boolean);
  let acc = base;
  for (const seg of segments) {
    acc += `/${seg}`;
    crumbs.push({ name: seg.replace(/-/g, " "), url: acc });
  }
  return (
    <nav className="breadcrumbs" aria-label="Breadcrumb">
      {crumbs.map((c, i) => {
        const isLast = i === crumbs.length - 1;
        const label = c.name.charAt(0).toUpperCase() + c.name.slice(1);
        return (
          <span key={`${c.url}-${i}`}>
            {i > 0 && <span className="breadcrumb-sep">/</span>}
            {isLast ? (
              <span className="breadcrumb-current" aria-current="page">{label}</span>
            ) : (
              <Link href={c.url}>{label}</Link>
            )}
          </span>
        );
      })}
    </nav>
  );
}

export function PageHeader({
  page,
  lang,
}: {
  page: Page;
  lang: Lang;
}) {
  return (
    <div className="page-header">
      <div className="container">
        <Breadcrumbs page={page} lang={lang} />
        <h1>{page.title}</h1>
        {page.description && <p className="page-description">{page.description}</p>}
      </div>
    </div>
  );
}
