// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

import type { Metadata } from "next";
import Link from "next/link";
import ControlPlaneHero from "@/components/home/ControlPlaneHero";
import Playground from "@/components/home/Playground";
import CopyButton from "@/components/home/CopyButton";
import { CountUpNumber, GithubStats } from "@/components/home/HomeEffects";
import { getPage, getBlogPosts } from "@/lib/content";
import { LANGUAGES, DEFAULT_LANG, SITE, type Lang } from "@/lib/site";
import { pageMetadata } from "@/lib/metadata";
import { t as translate } from "@/lib/i18n";
import enDict from "@/lib/i18n/en.json";
import idDict from "@/lib/i18n/id.json";
import statusData from "@/data/status.json";

type Dict = Record<string, string>;
const DICTS: Record<Lang, Dict> = { en: enDict as Dict, id: idDict as Dict };

const PIPELINE_STAGES = [
  "parse",
  "normalize",
  "resolve",
  "build",
  "validate",
  "graph",
  "policy",
  "schedule",
  "generate",
  "review",
  "write",
] as const;

export function generateStaticParams() {
  return LANGUAGES.map((lang) => ({ lang }));
}

export async function generateMetadata(props: {
  params: Promise<{ lang: string }>;
}): Promise<Metadata> {
  const { lang: raw } = await props.params;
  const lang = (LANGUAGES as readonly string[]).includes(raw) ? (raw as Lang) : DEFAULT_LANG;
  return pageMetadata(getPage("/", lang), lang);
}

export default async function HomePage(props: {
  params: Promise<{ lang: string }>;
}) {
  const { lang: raw } = await props.params;
  const lang = (LANGUAGES as readonly string[]).includes(raw) ? (raw as Lang) : DEFAULT_LANG;
  const t = (key: string) => DICTS[lang][key] ?? key;
  const base = lang === "en" ? "" : "/id";
  const posts = getBlogPosts(lang).slice(0, 3);

  const githubStats =
    (statusData as { github?: { stars?: number; forks?: number; openIssues?: number; contributors?: number } })
      .github ?? {};

  const features = [
    { key: "pipeline", icon: <><circle cx="12" cy="12" r="10" /><path d="M12 6v6l4 2" /></> },
    { key: "compiler", icon: <><path d="M12 2a10 10 0 1 0 10 10" /><path d="M12 12 2 12" /><path d="M12 2 12 12" /><path d="M12 12 12 22" /></> },
    { key: "generator", icon: <><polyline points="16 18 22 12 16 6" /><polyline points="8 6 2 12 8 18" /></> },
    { key: "neir", icon: <><polygon points="12 2 22 8.5 22 15.5 12 22 2 15.5 2 8.5" /><line x1="12" y1="22" x2="12" y2="15.5" /><polyline points="22 8.5 12 15.5 2 8.5" /></> },
    { key: "governance", icon: <><rect x="3" y="11" width="18" height="11" rx="2" /><path d="M7 11V7a5 5 0 0 1 10 0v4" /></> },
    { key: "marketplace", icon: <><circle cx="12" cy="12" r="3" /><path d="M12 2v4m0 12v4m10-10h-4M6 12H2" /></> },
  ];

  const useCases = [
    {
      title: t("home_use_case_microservices_title"),
      desc: t("home_use_case_microservices_desc"),
      badge: "Go · TypeScript",
      badgeClass: "badge-green",
      icon: (
        <>
          <rect x="3" y="5" width="10" height="8" rx="2" />
          <rect x="19" y="5" width="10" height="8" rx="2" />
          <rect x="11" y="19" width="10" height="8" rx="2" />
          <path d="M8 13v3h8m8-3v3h-8m0 0v3" />
        </>
      ),
    },
    {
      title: t("home_use_case_serverless_title"),
      desc: t("home_use_case_serverless_desc"),
      badge: "Python · TypeScript",
      badgeClass: "badge-green",
      icon: (
        <>
          <path d="M18 2 7 18h8l-1 12 11-17h-8l1-11Z" />
        </>
      ),
    },
    {
      title: t("home_use_case_ai_title"),
      desc: t("home_use_case_ai_desc"),
      badge: t("home_ai_platforms_badge"),
      badgeClass: "badge-blue",
      icon: (
        <>
          <path d="M12 6a5 5 0 0 0-8 4 5 5 0 0 0 1 3 6 6 0 0 0 3 11h4V6Zm8 0a5 5 0 0 1 8 4 5 5 0 0 1-1 3 6 6 0 0 1-3 11h-4V6Z" />
          <path d="M8 13h4m-4 6h4m8-6h4m-4 6h4M16 4v24" />
        </>
      ),
    },
    {
      title: t("home_use_case_governance_title"),
      desc: t("home_use_case_governance_desc"),
      badge: t("home_governance_badge"),
      badgeClass: "badge-orange",
      icon: (
        <>
          <path d="m16 3 13 6H3l13-6Z" />
          <path d="M5 27h22M7 9v16m6-16v16m6-16v16m6-16v16" />
        </>
      ),
    },
  ];

  const quickSteps = [
    {
      title: t("home_step_install_title"),
      desc: t("home_step_install_desc"),
      lang: "bash",
      code: t("mk_install_cmd"),
    },
    {
      title: t("home_step_spec_title"),
      desc: t("home_step_spec_desc"),
      lang: "yaml",
      code: `project: my-app\nmodules:\n  - name: auth\n    path: ./auth\n  - name: api\n    path: ./api\n    dependencies: [auth]\nservices:\n  - name: gateway\n    kind: http\n    port: 8080\ngeneration:\n  languages: [go, typescript]`,
    },
    {
      title: t("home_step_run_title"),
      desc: t("home_step_run_desc"),
      lang: "bash",
      code: "naeos run --config naeos.yaml --input-file spec.yaml",
    },
    {
      title: t("home_step_ai_title"),
      desc: t("home_step_ai_desc"),
      lang: "bash",
      code: "naeos ai compile --input-file spec.yaml --target opencode",
    },
  ];

  const cliDemoSteps = [
    {
      label: t("home_cli_demo_validate"),
      code: "naeos validate --input-file spec.yaml --output json",
    },
    {
      label: t("home_cli_demo_context"),
      code: "naeos context --input-file spec.yaml --output markdown --output-file context.md",
    },
    {
      label: t("home_cli_demo_generate"),
      code: "naeos run --config naeos.yaml --input-file spec.yaml --output json",
    },
  ];

  const proofPoints = [
    { quote: t("home_proof_one"), initials: "01", name: "README", role: t("home_proof_one_source") },
    { quote: t("home_proof_two"), initials: "02", name: "Quick Start", role: t("home_proof_two_source") },
    { quote: t("home_proof_three"), initials: "03", name: "Architecture", role: t("home_proof_three_source") },
  ];

  return (
    <>
      <ControlPlaneHero base={base} lang={lang} />

      {/* Announcement */}
      <div className="announcement-bar">
        <div className="container announcement-bar-inner">
          <span className="announcement-badge">{t("home_announcement_new")}</span>
          <span className="announcement-text">
            <strong>v{SITE.version}</strong> — {t("home_announcement_desc")}{" "}
            <a href={`${SITE.repo}/releases`} target="_blank" rel="noopener">{t("home_announcement_link")}</a>
          </span>
        </div>
      </div>

      {/* Executive summary */}
      <section className="section">
        <div className="container fade-in">
          <div className="problem-grid">
            <div className="problem-card problem-card-before">
              <span className="problem-tag problem-tag-red">
                {lang === "id" ? "Ringkasan eksekutif" : "Executive summary"}
              </span>
              <h3>
                {lang === "id"
                  ? "NAEOS adalah control plane untuk engineering AI-native."
                  : "NAEOS is the control plane for AI-native engineering."}
              </h3>
              <p>
                {lang === "id"
                  ? "NAEOS mengubah spesifikasi perangkat lunak menjadi workflow yang tervalidasi, teratur, dan siap untuk AI. Platform ini menggabungkan NEIR, policy enforcement, traceability, dan evidence agar AI dapat mempercepat delivery tanpa mengorbankan kontrol."
                  : "NAEOS turns software specifications into validated, governed, AI-ready workflows. It combines NEIR, policy enforcement, traceability, and evidence so AI can accelerate delivery without sacrificing control."}
              </p>
            </div>
            <div className="problem-arrow" aria-hidden="true">→</div>
            <div className="problem-card problem-card-after">
              <span className="problem-tag problem-tag-green">
                {lang === "id" ? "Kenapa penting" : "Why it matters"}
              </span>
              <h3>
                {lang === "id"
                  ? "AI menghasilkan kode. NAEOS mengelola engineering di sekitarnya."
                  : "AI generates code. NAEOS governs the engineering around it."}
              </h3>
              <p>
                {lang === "id"
                  ? "Masalah utama sekarang bukan lagi kualitas generation, tetapi trust, auditability, dan repeatability. NAEOS hadir untuk memberi tim engineering lapisan sistem yang dibutuhkan agar adopsi AI bisa berkembang secara aman dan terukur."
                  : "The main bottleneck is no longer generation quality; it is trust, auditability, and repeatability. NAEOS gives engineering teams the system layer they need to scale AI adoption safely and predictably."}
              </p>
              <Link href={`${base}/investors`} className="btn btn-primary btn-sm" data-umami-event="home-investor-summary">
                {lang === "id" ? "Lihat investor page" : "View investor page"}
              </Link>
            </div>
          </div>
        </div>
      </section>

      {/* What makes NAEOS different */}
      <section className="section">
        <div className="container fade-in">
          <h2 className="section-title">
            {lang === "id" ? "Apa yang membuat NAEOS berbeda" : "What makes NAEOS different"}
          </h2>
          <p className="section-subtitle">
            {lang === "id"
              ? "NAEOS tidak hanya menghasilkan kode. NAEOS mengorganisasi seluruh alur engineering di sekitar spesifikasi, governance, dan bukti."
              : "NAEOS does not only generate code. It organizes the entire engineering workflow around specification, governance, and evidence."}
          </p>
          <div className="use-cases-grid stagger-fade">
            <div className="use-case-card">
              <span className="use-case-icon" aria-hidden="true">
                <svg viewBox="0 0 32 32" fill="none"><path d="M5 10.5 16 4l11 6.5v11L16 28l-11-6.5v-11Zm11 0V28M5 10.5l11 6.5 11-6.5" /></svg>
              </span>
              <h3>{lang === "id" ? "Spesifikasi sebagai sumber kebenaran" : "Specification as the source of truth"}</h3>
              <p>
                {lang === "id"
                  ? "Semua intent engineering dimodelkan sebagai spesifikasi yang dapat dipahami mesin, dipetakan, dan dijalankan secara konsisten."
                  : "Engineering intent is modeled as a machine-readable specification that can be understood, mapped, and executed consistently."}
              </p>
            </div>
            <div className="use-case-card">
              <span className="use-case-icon" aria-hidden="true">
                <svg viewBox="0 0 32 32" fill="none"><path d="M7 20.5 13.5 14l4 4L25 6.5M25 6.5H18M25 6.5v7" /></svg>
              </span>
              <h3>{lang === "id" ? "AI workflow yang terukur" : "Measurable AI workflows"}</h3>
              <p>
                {lang === "id"
                  ? "NAEOS menambahkan validasi, policy, dan konteks yang jelas di setiap tahap agar AI bisa bekerja lebih aman dan lebih repeatable."
                  : "NAEOS adds validation, policy, and clear context at every stage so AI can work more safely and more repeatably."}
              </p>
            </div>
            <div className="use-case-card">
              <span className="use-case-icon" aria-hidden="true">
                <svg viewBox="0 0 32 32" fill="none"><path d="M8 6h16v20H8zM12 10h8M12 16h8M12 22h8" /></svg>
              </span>
              <h3>{lang === "id" ? "Delivery yang bisa dilacak" : "Traceable delivery"}</h3>
              <p>
                {lang === "id"
                  ? "Artifacts, keputusan, dan bukti dapat ditelusuri kembali ke spesifikasi awal, sehingga review dan audit menjadi lebih mudah."
                  : "Artifacts, decisions, and evidence can be traced back to the original specification, making review and audit far easier."}
              </p>
            </div>
          </div>
          <div style={{ textAlign: "center", marginTop: "2rem" }}>
            <Link href={`${base}/docs/getting-started`} className="btn btn-primary btn-lg" data-umami-event="home-structure-get-started">
              {lang === "id" ? "Mulai dengan NAEOS" : "Get started with NAEOS"}
            </Link>
          </div>
        </div>
      </section>

      {/* Social proof */}
      <section className="section social-proof">
        <div className="container fade-in">
          <h2 className="section-title">{t("mk_social_proof_title")}</h2>
          <p className="section-subtitle">{t("mk_social_proof_desc")}</p>
          <div className="logo-strip stagger-fade">
            <div className="logo-strip-group">
              <h3 className="logo-strip-label">{t("home_strip_languages")}</h3>
              <div className="logo-strip-items">
                {[["#00add8", "Go"], ["#3178c6", "TypeScript"], ["#3776ab", "Python"], ["#ed8b00", "Java"], ["#dea584", "Rust"]].map(([color, name]) => (
                  <span key={name} className="stack-logo">
                    <span className="lang-dot" style={{ background: color }} />
                    {name}
                  </span>
                ))}
              </div>
            </div>
            <div className="logo-strip-divider" />
            <div className="logo-strip-group">
              <h3 className="logo-strip-label">{t("home_strip_ai_platforms")}</h3>
              <div className="logo-strip-items">
                {["GitHub Copilot", "Claude Code", "Cursor", "Gemini CLI", "Codex", "OpenCode", "Windsurf"].map((name) => (
                  <span key={name} className="stack-logo">{name}</span>
                ))}
              </div>
            </div>
            <div className="logo-strip-divider" />
            <div className="logo-strip-group">
              <h3 className="logo-strip-label">{t("home_strip_infrastructure")}</h3>
              <div className="logo-strip-items">
                {["Docker", "Kubernetes", "GitHub Actions", "Terraform"].map((name) => (
                  <span key={name} className="stack-logo">{name}</span>
                ))}
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* Problem / Solution */}
      <section className="section section-problem">
        <div className="container">
          <div className="problem-grid fade-in">
            <div className="problem-card problem-card-before">
              <span className="problem-tag problem-tag-red">{t("mk_problem_label")}</span>
              <h3>{t("mk_problem_title")}</h3>
              <p>{t("mk_problem_desc")}</p>
              <ul className="problem-list">
                <li>{t("mk_problem_item1")}</li>
                <li>{t("mk_problem_item2")}</li>
                <li>{t("mk_problem_item3")}</li>
              </ul>
            </div>
            <div className="problem-arrow" aria-hidden="true">→</div>
            <div className="problem-card problem-card-after">
              <span className="problem-tag problem-tag-green">{t("mk_solution_label")}</span>
              <h3>{t("mk_solution_title")}</h3>
              <p>{t("mk_solution_desc")}</p>
              <ul className="problem-list">
                <li>{t("mk_solution_item1")}</li>
                <li>{t("mk_solution_item2")}</li>
                <li>{t("mk_solution_item3")}</li>
              </ul>
              <Link href={`${base}/docs/getting-started`} className="btn btn-primary btn-sm" data-umami-event="solution-get-started">
                {t("cta_get_started")}
              </Link>
            </div>
          </div>
        </div>
      </section>

      {/* How it works */}
      <section className="section">
        <div className="container">
          <h2 className="section-title fade-in">{t("mk_how_title")}</h2>
          <p className="section-subtitle fade-in">{t("mk_how_desc")}</p>
          <div className="how-grid stagger-fade">
            {[
              {
                step: "1",
                title: t("mk_how_step1_title"),
                desc: t("mk_how_step1_desc"),
                header: "spec.yaml",
                code: "project: my-app\narchitecture:\n  pattern: microservices\nservices:\n  - name: gateway\n    kind: http\n    port: 8080",
              },
              {
                step: "2",
                title: t("mk_how_step2_title"),
                desc: t("mk_how_step2_desc"),
                header: "NEIR",
                code: "✓ modules: 4\n✓ services: 8\n✓ dependencies: 12\n✓ architecture: microservices\n✓ policies: passed",
              },
              {
                step: "3",
                title: t("mk_how_step3_title"),
                desc: t("mk_how_step3_desc"),
                header: "output/",
                code: "├── go/            # Go services\n├── typescript/    # TS services\n├── infra/         # K8s + Terraform\n├── ci/            # GitHub Actions\n└── ai/            # Context bundles",
              },
            ].map((card) => (
              <div key={card.step} className="how-card">
                <div className="how-step">{card.step}</div>
                <h3>{card.title}</h3>
                <p>{card.desc}</p>
                <div className="how-code">
                  <div className="code-block-header"><span>{card.header}</span></div>
                  <pre><code>{card.code}</code></pre>
                </div>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Pipeline */}
      <section className="section section-pipeline">
        <div className="container fade-in">
          <h2 className="section-title">{t("mk_pipeline_title")}</h2>
          <p className="section-subtitle">{t("mk_pipeline_desc")}</p>
          <div className="pipeline-strip stagger-fade">
            {PIPELINE_STAGES.map((stage, i) => (
              <div key={`wrap-${stage}`} style={{ display: "contents" }}>
                <div className="pipeline-stage">
                  <div className="pipeline-stage-index">{i + 1}</div>
                  <div className="pipeline-stage-name">{t(`mk_stage_${stage}`)}</div>
                </div>
                {i < PIPELINE_STAGES.length - 1 && (
                  <div className="pipeline-arrow" aria-hidden="true">→</div>
                )}
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Features */}
      <section className="section">
        <div className="container">
          <h2 className="section-title fade-in">{t("section_features")}</h2>
          <p className="section-subtitle fade-in">{translate(lang, "home_features_subtitle", { CLI: SITE.stats.cli })}</p>
          <div className="features-grid stagger-fade">
            {features.map((f) => (
              <div key={f.key} className="feature-card">
                <svg className="feature-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" aria-hidden="true">
                  {f.icon}
                </svg>
                <h3>{t(`feature_${f.key}`)}</h3>
                <p>{t(`feature_${f.key}_desc`)}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Stats */}
      <div className="container">
        <div className="stats-grid stagger-fade">
          <div className="stat-card" role="figure">
            <CountUpNumber target={SITE.stats.cli} />
            <div className="stat-label">{t("stats_cli_commands")}+</div>
          </div>
          <div className="stat-card" role="figure">
            <CountUpNumber target={SITE.stats.languages} />
            <div className="stat-label">{t("stats_languages")}</div>
          </div>
          <div className="stat-card" role="figure">
            <CountUpNumber target={SITE.stats.ai_platforms} />
            <div className="stat-label">{t("stats_ai_platforms")}</div>
          </div>
          <div className="stat-card" role="figure">
            <CountUpNumber target={SITE.stats.specs} />
            <div className="stat-label">{t("stats_specs")}</div>
          </div>
        </div>
      </div>

      {/* Use cases */}
      <section className="section">
        <div className="container fade-in">
          <h2 className="section-title">{t("use_cases_title")}</h2>
          <p className="section-subtitle">{t("home_use_cases_desc")}</p>
          <div className="use-cases-grid stagger-fade">
            {useCases.map((uc) => (
              <div key={uc.title} className="use-case-card">
                <span className="use-case-icon" aria-hidden="true">
                  <svg viewBox="0 0 32 32" fill="none">{uc.icon}</svg>
                </span>
                <h3>{uc.title}</h3>
                <p>{uc.desc}</p>
                <span className={`badge ${uc.badgeClass}`}>{uc.badge}</span>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Playground */}
      <section className="section">
        <div className="container fade-in">
          <h2 className="section-title">{t("home_playground_title")}</h2>
          <p className="section-subtitle">{t("home_playground_desc")}</p>
          <Playground lang={lang} />
        </div>
      </section>

      {/* Runnable CLI demo */}
      <section className="section">
        <div className="container fade-in">
          <h2 className="section-title">{t("home_cli_demo_title")}</h2>
          <p className="section-subtitle">{t("home_cli_demo_desc")}</p>
          <div className="quick-start-steps stagger-fade">
            {cliDemoSteps.map((step, i) => (
              <div key={step.label} className="quick-start-step">
                <div className="step-number">{i + 1}</div>
                <div className="step-content">
                  <h4>{step.label}</h4>
                  <div className="code-block">
                    <div className="code-block-header">
                      <span>bash</span>
                      <CopyButton text={step.code} label={t("copy_code")} />
                    </div>
                    <pre><code>{step.code}</code></pre>
                  </div>
                </div>
              </div>
            ))}
          </div>
          <div style={{ textAlign: "center", marginTop: "2rem" }}>
            <a
              href={`${SITE.repo}/tree/main/examples/demo-cli`}
              className="btn btn-primary"
              target="_blank"
              rel="noopener"
              data-umami-event="cli-demo-source"
            >
              {t("home_cli_demo_source")}
            </a>
          </div>
        </div>
      </section>

      {/* Quick start */}
      <section className="section">
        <div className="container">
          <h2 className="section-title fade-in">{t("section_quick_start")}</h2>
          <p className="section-subtitle fade-in">{t("home_quick_start_subtitle")}</p>
          <div className="quick-start-steps stagger-fade">
            {quickSteps.map((step, i) => (
              <div key={step.title} className="quick-start-step">
                <div className="step-number">{i + 1}</div>
                <div className="step-content">
                  <h4>{step.title}</h4>
                  <p>{step.desc}</p>
                  <div className="code-block">
                    <div className="code-block-header">
                      <span>{step.lang}</span>
                      <CopyButton text={step.code} label={t("copy_code")} />
                    </div>
                    <pre><code>{step.code}</code></pre>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Testimonials */}
      <section className="section">
        <div className="container fade-in">
          <h2 className="section-title">{t("home_proof_title")}</h2>
          <p className="section-subtitle">{t("home_proof_desc")}</p>
          <div className="testimonials-grid stagger-fade">
            {proofPoints.map((point) => (
              <div key={point.name} className="testimonial-card">
                <p>{point.quote}</p>
                <div className="testimonial-author">
                  <div className="testimonial-avatar">{point.initials}</div>
                  <div className="testimonial-info">
                    <h4>{point.name}</h4>
                    <span>{point.role}</span>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Latest blog */}
      <section className="section">
        <div className="container fade-in">
          <h2 className="section-title">{t("section_latest_blog")}</h2>
          <p className="section-subtitle">{t("section_latest_blog_desc")}</p>
          <div className="blog-grid" id="home-blog-preview">
            {posts.map((post) => (
              <article key={post.url} className="blog-card">
                <div className="blog-card-categories">
                  {(post.categories ?? []).map((c) => (
                    <span key={c} className="category-badge">{c}</span>
                  ))}
                </div>
                <div className="blog-date">
                  {post.date
                    ? new Date(post.date).toLocaleDateString(lang === "id" ? "id-ID" : "en-US", {
                        year: "numeric",
                        month: "short",
                        day: "numeric",
                      })
                    : ""}
                </div>
                <h3><Link href={post.url}>{post.title}</Link></h3>
                <p>{post.summary}</p>
                <div className="blog-meta" style={{ marginTop: "auto", justifyContent: "flex-start" }}>
                  <span>{t("blog_read")} {post.readingTime} min</span>
                </div>
              </article>
            ))}
          </div>
          <div style={{ textAlign: "center", margin: "2rem 0" }}>
            <Link href={`${base}/blog`} className="btn btn-secondary">
              {t("nav_blog")} →
            </Link>
          </div>
        </div>
      </section>

      {/* GitHub stats */}
      <section className="section section-github-stats">
        <div className="container fade-in">
          <h2 className="section-title github-stats-title fade-in">{t("home_community_title")}</h2>
          <div className="github-stats stagger-fade">
            <GithubStats
              initial={{
                stars: githubStats.stars,
                forks: githubStats.forks,
                issues: githubStats.openIssues,
                contributors: githubStats.contributors,
              }}
              labels={[
                t("home_github_stars"),
                t("home_github_forks"),
                t("home_github_issues"),
                t("home_github_contributors"),
              ]}
            />
          </div>
        </div>
      </section>

      {/* Supported languages */}
      <section className="section section-supported">
        <div className="container fade-in">
          <h2 className="section-title">{t("section_languages")}</h2>
          <p className="section-subtitle">{t("home_supported_subtitle")}</p>
          <div className="supported-grid stagger-fade">
            <div className="supported-group">
              <h3 className="supported-group-title">{t("home_supported_languages_heading")}</h3>
              <div className="supported-items">
                {[["#00add8", "Go"], ["#3178c6", "TypeScript"], ["#3776ab", "Python"], ["#ed8b00", "Java"], ["#dea584", "Rust"]].map(([color, name]) => (
                  <span key={name} className="lang-badge">
                    <span className="lang-dot" style={{ background: color }} />
                    {name}
                  </span>
                ))}
              </div>
            </div>
            <div className="supported-divider" />
            <div className="supported-group">
              <h3 className="supported-group-title">{t("home_supported_ai_heading")}</h3>
              <div className="supported-items ai-items">
                {[
                  ["⟐", "GitHub Copilot"],
                  ["◉", "Claude Code"],
                  ["⟡", "Cursor"],
                  ["◇", "Gemini CLI"],
                  ["⊡", "Codex"],
                  ["◈", "OpenCode"],
                ].map(([icon, name]) => (
                  <div key={name} className="ai-card">
                    <span className="ai-icon">{icon}</span>
                    {name}
                  </div>
                ))}
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* CTA band */}
      <section className="section section-cta-band">
        <div className="container">
          <div className="cta-band fade-in">
            <h2>{t("mk_cta_band_title")}</h2>
            <p>{t("mk_cta_band_desc")}</p>
            <div className="cta-band-install">
              <div className="code-block">
                <div className="code-block-header">
                  <span>bash</span>
                  <CopyButton text={t("mk_install_cmd")} label={t("copy_code")} />
                </div>
                <pre><code>{t("mk_install_cmd")}</code></pre>
              </div>
            </div>
            <div className="cta-band-actions">
              <Link href={`${base}/docs/getting-started`} className="btn btn-primary btn-lg">
                {t("cta_get_started")}
              </Link>
              <Link href={`${base}/download`} className="btn btn-secondary btn-lg">
                {t("cta_install_now")}
              </Link>
            </div>
          </div>
        </div>
      </section>
    </>
  );
}
