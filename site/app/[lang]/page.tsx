import type { Metadata } from "next";
import Link from "next/link";
import HeroEffects from "@/components/home/HeroEffects";
import Playground from "@/components/home/Playground";
import CopyButton from "@/components/home/CopyButton";
import { CountUpNumber, GithubStats } from "@/components/home/HomeEffects";
import { getPage, getBlogPosts } from "@/lib/content";
import { LANGUAGES, DEFAULT_LANG, SITE, type Lang } from "@/lib/site";
import { pageMetadata } from "@/lib/metadata";
import enDict from "@/lib/i18n/en.json";
import idDict from "@/lib/i18n/id.json";

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

function Html({ children }: { children: string }) {
  return <span dangerouslySetInnerHTML={{ __html: children }} />;
}

export default async function HomePage(props: {
  params: Promise<{ lang: string }>;
}) {
  const { lang: raw } = await props.params;
  const lang = (LANGUAGES as readonly string[]).includes(raw) ? (raw as Lang) : DEFAULT_LANG;
  const t = (key: string) => DICTS[lang][key] ?? key;
  const base = lang === "en" ? "" : "/id";
  const posts = getBlogPosts(lang).slice(0, 3);

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
      code: "naeos run --input spec.yaml --output-dir ./out",
    },
    {
      title: t("home_step_ai_title"),
      desc: t("home_step_ai_desc"),
      lang: "bash",
      code: "naeos ai compile --input-file spec.yaml --target opencode",
    },
  ];

  const testimonials = [
    { quote: t("home_testimonial_one"), initials: "SK", name: "Sarah Kim", role: t("home_testimonial_one_role") },
    { quote: t("home_testimonial_two"), initials: "JR", name: "James Reyes", role: t("home_testimonial_two_role") },
    { quote: t("home_testimonial_three"), initials: "AN", name: "Aditya Nugraha", role: t("home_testimonial_three_role") },
  ];

  return (
    <>
      {/* Hero */}
      <section className="hero">
        <div className="hero-bg" />
        <div className="hero-grid" />
        <HeroEffects />
        <div className="container hero-content">
          <div className="hero-badges">
            <span className="badge badge-green">v{SITE.version}</span>
            <span className="badge badge-blue">{t("hero_open_source")}</span>
            <span className="badge badge-orange">Apache 2.0</span>
            <span className="badge badge-success">{t("mk_hero_chip1")}</span>
            <span className="badge badge-success">{t("mk_hero_chip2")}</span>
            <span className="badge badge-success">{t("mk_hero_chip3")}</span>
          </div>
          <h1 className="hero-title"><Html>{t("hero_title")}</Html></h1>
          <p className="hero-subtitle">{t("hero_subtitle")}</p>
          <div className="hero-actions">
            <Link href={`${base}/docs/getting-started`} className="btn btn-primary btn-lg">
              {t("cta_get_started")}
            </Link>
            <a href={SITE.repo} className="btn btn-secondary btn-lg" target="_blank" rel="noopener">
              {t("cta_view_on_github")}
            </a>
          </div>

          <div className="hero-visual">
            <div className="hero-terminal fade-in-scale" id="hero-terminal-static">
              <div className="terminal-header">
                <span className="terminal-dot red" aria-hidden="true" />
                <span className="terminal-dot yellow" aria-hidden="true" />
                <span className="terminal-dot green" aria-hidden="true" />
                <span className="terminal-title">terminal — naeos</span>
              </div>
              <div className="terminal-body">
                <div className="terminal-line prompt">$ naeos init</div>
                <div className="terminal-line output">✓ Config initialized: naeos.yaml</div>
                <div className="terminal-line prompt">$ cat spec.yaml</div>
                <div className="terminal-line output">project: my-app</div>
                <div className="terminal-line output">modules:</div>
                <div className="terminal-line output">  - name: api</div>
                <div className="terminal-line output">    path: ./api</div>
                <div className="terminal-line output">generation:</div>
                <div className="terminal-line output">  languages: [go, typescript]</div>
                <div className="terminal-line prompt">$ naeos run --input spec.yaml</div>
                <div className="terminal-line output">✓ Parsed → Normalized → Resolved</div>
                <div className="terminal-line output">✓ NEIR built | <span className="highlight">4 modules</span>, <span className="highlight">8 services</span></div>
                <div className="terminal-line output">✓ Policy evaluation: <span className="highlight">passed</span></div>
                <div className="terminal-line output">✓ Generated: go/, typescript/</div>
                <div className="terminal-line output">✓ <span className="highlight">Done</span> in 1.2s</div>
                <div className="terminal-line terminal-cursor" />
              </div>
            </div>
            <div
              className="hero-terminal fade-in-scale"
              id="interactive-terminal"
              style={{ display: "none" }}
            >
              <div className="terminal-header">
                <span className="terminal-dot red" aria-hidden="true" />
                <span className="terminal-dot yellow" aria-hidden="true" />
                <span className="terminal-dot green" aria-hidden="true" />
                <span className="terminal-title">demo — naeos interactive</span>
              </div>
            </div>

            <div className="hero-float-card hero-float-neir fade-in-scale" aria-hidden="true">
              <span className="hero-float-icon">◆</span>
              <div>
                <div className="hero-float-label">{t("mk_hero_float_neir")}</div>
                <div className="hero-float-value">4 modules · 8 services</div>
              </div>
            </div>
            <div className="hero-float-card hero-float-validated fade-in-scale" aria-hidden="true">
              <span className="hero-float-icon">✓</span>
              <div>
                <div className="hero-float-label">{t("mk_hero_float_validated")}</div>
                <div className="hero-float-value">0 errors · 0 warnings</div>
              </div>
            </div>
            <div className="hero-float-card hero-float-generated fade-in-scale" aria-hidden="true">
              <span className="hero-float-icon">⚙</span>
              <div>
                <div className="hero-float-label">{t("mk_hero_float_generated")}</div>
                <div className="hero-float-value">go/ · typescript/ · infra/</div>
              </div>
            </div>
          </div>
        </div>
      </section>

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
                {["GitHub Copilot", "Claude Code", "Cursor", "Gemini CLI", "Codex", "OpenCode"].map((name) => (
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
              <Link href={`${base}/docs/getting-started`} className="btn btn-primary btn-sm">
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
          <p className="section-subtitle fade-in">{t("home_features_subtitle").replace("{{.CLI}}", String(SITE.stats.cli))}</p>
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
          <h2 className="section-title">{t("home_testimonials_title")}</h2>
          <p className="section-subtitle">{t("home_testimonials_desc")}</p>
          <div className="testimonials-grid stagger-fade">
            {testimonials.map((tm) => (
              <div key={tm.name} className="testimonial-card">
                <p>{tm.quote}</p>
                <div className="testimonial-author">
                  <div className="testimonial-avatar">{tm.initials}</div>
                  <div className="testimonial-info">
                    <h4>{tm.name}</h4>
                    <span>{tm.role}</span>
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
