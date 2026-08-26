"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { jsYaml } from "@/components/home/jsyamlBridge";
import { useTranslation } from "@/lib/useTranslation";
import type { Lang } from "@/lib/site";

interface PlaygroundSpec {
  project?: string;
  version?: string;
  modules?: { name?: string; path?: string; dependencies?: string[] }[];
  services?: { name?: string; kind?: string; port?: number }[];
  architecture?: { pattern?: string };
  generation?: { languages?: string[]; output_dir?: string };
}

const SAMPLES: Record<string, string> = {
  microservices:
    'project: my-service\nversion: "1.0"\nmodules:\n  - name: api-gateway\n    path: ./api-gateway\n    dependencies: [user-service, order-service]\n  - name: user-service\n    path: ./services/users\n    dependencies: [database]\n  - name: order-service\n    path: ./services/orders\n    dependencies: [user-service, payment-service]\n  - name: payment-service\n    path: ./services/payments\n  - name: database\n    path: ./infra/db\nservices:\n  - name: api-gateway\n    kind: reverse-proxy\n    port: 8080\n  - name: user-api\n    kind: rest\n    port: 9001\n  - name: order-api\n    kind: rest\n    port: 9002\narchitecture:\n  pattern: microservices\ngeneration:\n  languages: [go, typescript]\n  output_dir: ./generated',
  serverless:
    'project: serverless-app\nversion: "1.0"\nmodules:\n  - name: auth\n    path: ./functions/auth\n  - name: api\n    path: ./functions/api\n    dependencies: [auth]\n  - name: processor\n    path: ./functions/processor\n    dependencies: [api]\nservices:\n  - name: auth-function\n    kind: lambda\n  - name: api-function\n    kind: lambda\n  - name: processor-function\n    kind: lambda\narchitecture:\n  pattern: serverless\ndeployment:\n  strategy: serverless-framework\ngeneration:\n  languages: [python, typescript]',
  monolith:
    'project: monolith-app\nversion: "1.0"\nmodules:\n  - name: core\n    path: ./core\n  - name: web\n    path: ./web\n    dependencies: [core]\n  - name: database\n    path: ./infra/db\n    dependencies: [core]\nservices:\n  - name: web-server\n    kind: http\n    port: 8080\narchitecture:\n  pattern: monolithic\ndeployment:\n  strategy: docker-compose\ngeneration:\n  languages: [go]\n  output_dir: ./cmd',
  hexagonal:
    'project: clean-arch-app\nversion: "1.0"\nmodules:\n  - name: domain\n    path: ./internal/domain\n  - name: application\n    path: ./internal/application\n    dependencies: [domain]\n  - name: adapters-inbound\n    path: ./internal/adapters/inbound\n    dependencies: [application]\n  - name: adapters-outbound\n    path: ./internal/adapters/outbound\n    dependencies: [application]\n  - name: infrastructure\n    path: ./internal/infrastructure\n    dependencies: [adapters-outbound]\nservices:\n  - name: rest-api\n    kind: rest\n    port: 8080\n  - name: grpc-api\n    kind: grpc\n    port: 9090\narchitecture:\n  pattern: hexagonal\ngeneration:\n  languages: [go, java]\n  output_dir: ./src',
  "event-driven":
    'project: events-platform\nversion: "1.0"\nmodules:\n  - name: event-ingestor\n    path: ./ingest\n  - name: stream-processor\n    path: ./process\n    dependencies: [event-ingestor]\n  - name: analytics\n    path: ./analytics\n    dependencies: [stream-processor]\n  - name: notification\n    path: ./notify\n    dependencies: [stream-processor]\nservices:\n  - name: ingestion-api\n    kind: rest\n    port: 8080\n  - name: stream-worker\n    kind: worker\n    port: 9001\n  - name: notification-ws\n    kind: websocket\n    port: 9002\narchitecture:\n  pattern: event-driven\ndeployment:\n  strategy: kubernetes\ngeneration:\n  languages: [go, typescript, python]\n  output_dir: ./generated',
  "ai-context":
    'project: my-genai-service\nversion: "1.0"\nmodules:\n  - name: agent-orchestrator\n    path: ./orchestrator\n    dependencies: [llm-provider, memory-store]\n  - name: llm-provider\n    path: ./providers/llm\n    dependencies: [vector-db]\n  - name: memory-store\n    path: ./stores/memory\n  - name: vector-db\n    path: ./infra/vector\n    kind: database\n    engine: qdrant\nservices:\n  - name: api-gateway\n    kind: reverse-proxy\n    port: 8080\n  - name: chat-api\n    kind: rest\n    port: 9001\n  - name: streaming-ws\n    kind: websocket\n    port: 9002\narchitecture:\n  pattern: microservices\nai:\n  providers:\n    - name: openai\n      models: [gpt-4o, gpt-4o-mini]\n    - name: anthropic\n      models: [claude-opus-4, claude-sonnet-4]\n  context:\n    format: neir\n    compression: semantic\n    max_tokens: 128000\ngeneration:\n  languages: [go, typescript, python]\n  ai_instructions: true\n  output_dir: ./generated',
};

const KIND_COLORS: Record<string, string> = {
  rest: "#60a5fa",
  grpc: "#a78bfa",
  websocket: "#34d399",
  lambda: "#fbbf24",
  "reverse-proxy": "#f87171",
  http: "#60a5fa",
  worker: "#fb923c",
};

function countDeps(modules: { dependencies?: string[] }[]): number {
  return modules.reduce((acc, m) => acc + (m.dependencies?.length ?? 0), 0);
}

export default function Playground({ lang }: { lang: Lang }) {
  const { t } = useTranslation(lang);
  const [tab, setTab] = useState("microservices");
  const [view, setView] = useState<"neir" | "ai">("neir");
  const [spec, setSpec] = useState<PlaygroundSpec | undefined>(undefined);
  const [error, setError] = useState("");
  const [copied, setCopied] = useState(false);
  const textareaRef = useRef<HTMLTextAreaElement>(null);
  const debounceRef = useRef<number | undefined>(undefined);

  const parse = useCallback((text: string) => {
    try {
      const parsed = jsYaml.load(text);
      if (!parsed || typeof parsed !== "object") throw new Error("Empty spec");
      setSpec(parsed as PlaygroundSpec);
      setError("");
    } catch (e) {
      setSpec(undefined);
      setError(e instanceof Error ? e.message : String(e));
    }
  }, []);

  const onInput = useCallback(
    (text: string) => {
      window.clearTimeout(debounceRef.current);
      debounceRef.current = window.setTimeout(() => parse(text), 200);
    },
    [parse],
  );

  useEffect(() => {
    const text = SAMPLES.microservices;
    if (textareaRef.current) textareaRef.current.value = text;
    parse(text);
    return () => window.clearTimeout(debounceRef.current);
  }, [parse]);

  function switchTab(name: string) {
    setTab(name);
    const text = SAMPLES[name];
    if (text !== undefined && textareaRef.current) {
      textareaRef.current.value = text;
      parse(text);
    }
  }

  function copySpec() {
    const text = textareaRef.current?.value ?? "";
    void navigator.clipboard.writeText(text).then(() => {
      setCopied(true);
      window.setTimeout(() => setCopied(false), 2000);
    });
  }

  function tryInTerminal() {
    const interactive = document.getElementById("interactive-terminal");
    const staticEl = document.getElementById("hero-terminal-static");
    if (interactive && staticEl) {
      staticEl.style.display = "none";
      interactive.style.display = "";
      interactive.scrollIntoView({ behavior: "smooth", block: "center" });
    }
    window.scrollTo({ top: 0, behavior: "smooth" });
  }

  const labels = {
    preview: t("home_playground_preview"),
    modules: t("home_playground_modules"),
    services: t("home_playground_services"),
    dependencies: t("home_playground_dependencies"),
    languages: t("home_playground_languages"),
  };

  function renderNeir(s: PlaygroundSpec) {
    const gen = s.generation ?? {};
    const deps = countDeps(s.modules ?? []);
    return (
      <>
        <h4>{labels.preview}</h4>
        <div className="playground-stats">
          <div className="playground-stat"><span className="playground-stat-num">{s.modules?.length ?? 0}</span><span className="playground-stat-label">{labels.modules}</span></div>
          <div className="playground-stat"><span className="playground-stat-num">{s.services?.length ?? 0}</span><span className="playground-stat-label">{labels.services}</span></div>
          <div className="playground-stat"><span className="playground-stat-num">{deps}</span><span className="playground-stat-label">{labels.dependencies}</span></div>
          <div className="playground-stat"><span className="playground-stat-num">{gen.languages?.length ?? 0}</span><span className="playground-stat-label">{labels.languages}</span></div>
        </div>
        {s.project && (
          <div className="playground-tree-section">
            <div className="playground-tree-header">{t("home_playground_project")}</div>
            <div className="tree-node"><span className="tree-key">name:</span> <span className="tree-str">{String(s.project)}</span></div>
            {s.version && <div className="tree-node"><span className="tree-key">version:</span> <span className="tree-str">{String(s.version)}</span></div>}
            {s.architecture?.pattern && <div className="tree-node"><span className="tree-key">pattern:</span> <span className="tree-val">{s.architecture.pattern}</span></div>}
          </div>
        )}
        {(s.modules?.length ?? 0) > 0 && (
          <div className="playground-tree-section">
            <div className="playground-tree-header">{labels.modules}</div>
            {s.modules!.map((m, i) => (
              <div key={`${m.name}-${i}`}>
                <div className="tree-node">
                  <span className="tree-key">{m.name ?? t("home_playground_unnamed")}</span>
                  {m.path && <span className="tree-dim"> {m.path}</span>}
                </div>
                {(m.dependencies?.length ?? 0) > 0 && (
                  <div className="tree-node tree-dep">
                    <span className="tree-dim">  └─ deps:</span>{" "}
                    <span className="tree-str">{m.dependencies!.join(", ")}</span>
                  </div>
                )}
              </div>
            ))}
          </div>
        )}
        {(s.services?.length ?? 0) > 0 && (
          <div className="playground-tree-section">
            <div className="playground-tree-header">{labels.services}</div>
            {s.services!.map((sv, i) => {
              const kind = sv.kind ?? "unknown";
              const color = KIND_COLORS[kind] ?? "#999";
              return (
                <div key={`${sv.name}-${i}`} className="tree-node">
                  <span className="tree-key">{sv.name ?? t("home_playground_unnamed")}</span>{" "}
                  <span className="tree-badge" style={{ color, borderColor: color }}>{kind}</span>
                  {sv.port ? <span className="tree-dim">:{sv.port}</span> : null}
                </div>
              );
            })}
          </div>
        )}
        {(gen.languages?.length ?? 0) > 0 && (
          <div className="playground-tree-section">
            <div className="playground-tree-header">{t("home_playground_generation")}</div>
            <div className="tree-node"><span className="tree-key">languages:</span> <span className="tree-str">{gen.languages!.join(", ")}</span></div>
            {gen.output_dir && <div className="tree-node"><span className="tree-key">output:</span> <span className="tree-str">{gen.output_dir}</span></div>}
          </div>
        )}
        {s.architecture?.pattern === "hexagonal" && (
          <div className="tree-node tree-dep" style={{ marginTop: "0.5rem" }}>
            <span className="tree-dim">  └─ architecture:</span>{" "}
            <span className="tree-str">{t("home_playground_arch_desc")}</span>
          </div>
        )}
      </>
    );
  }

  function renderAiContext(s: PlaygroundSpec) {
    const gen = s.generation ?? {};
    const languages = Array.isArray(gen.languages) ? gen.languages.join(", ") : t("home_playground_not_specified");
    return (
      <>
        <h4>{t("home_playground_ai_bundle")}</h4>
        <div className="playground-stats">
          <div className="playground-stat"><span className="playground-stat-num">{s.modules?.length ?? 0}</span><span className="playground-stat-label">{labels.modules}</span></div>
          <div className="playground-stat"><span className="playground-stat-num">{s.services?.length ?? 0}</span><span className="playground-stat-label">{labels.services}</span></div>
          <div className="playground-stat"><span className="playground-stat-num">{languages.split(",").length}</span><span className="playground-stat-label">{labels.languages}</span></div>
        </div>
        <div className="ai-context-section">
          <div className="ai-context-header">{t("home_playground_system_overview")}</div>
          <div className="ai-context-block">
            Project <strong>{s.project ?? t("home_playground_unnamed")}</strong> is a {s.architecture?.pattern ?? "microservices"} system with{" "}
            {s.modules?.length ?? 0} modules and {s.services?.length ?? 0} services.
            <br /><br />
            Code will be generated in: <strong>{languages}</strong>
            {gen.output_dir && (
              <>
                <br />
                {t("home_playground_output_dir")} <strong>{gen.output_dir}</strong>
              </>
            )}
          </div>
        </div>
        {(s.modules?.length ?? 0) > 0 && (
          <div className="ai-context-section">
            <div className="ai-context-header">{t("home_playground_dep_graph")}</div>
            <div className="ai-context-block">
              {s.modules!.map((m, i) => (
                <div key={i}>
                  • <strong>{m.name ?? t("home_playground_unnamed")}</strong>
                  {m.dependencies && m.dependencies.length > 0
                    ? ` → ${t("home_playground_depends_on")} ${m.dependencies.join(", ")}`
                    : ` ${t("home_playground_no_deps")}`}
                </div>
              ))}
            </div>
          </div>
        )}
        {(s.services?.length ?? 0) > 0 && (
          <div className="ai-context-section">
            <div className="ai-context-header">{t("home_playground_service_contracts")}</div>
            <div className="ai-context-block">
              {s.services!.map((sv, i) => (
                <div key={i}>
                  • <strong>{sv.name ?? t("home_playground_unnamed")}</strong> — {sv.kind ?? "http"}
                  {sv.port ? `:${sv.port}` : ""}
                </div>
              ))}
            </div>
          </div>
        )}
        <div className="ai-context-section">
          <div className="ai-context-header">{t("home_playground_ai_instructions")}</div>
          <div className="ai-context-block ai-context-prompt">
            You are working on project <strong>{s.project ?? t("home_playground_unnamed")}</strong>. The system uses{" "}
            {s.architecture?.pattern ?? "microservices"} architecture. When writing code, follow the existing module
            structure. All new services should be added to the appropriate module path. Generate code in{" "}
            <strong>{languages}</strong>. Maintain the dependency direction: downstream modules should not import
            upstream modules.
          </div>
        </div>
      </>
    );
  }

  return (
    <div className="spec-playground">
      <div className="playground-tabs">
        {[
          ["microservices", t("home_tab_microservice")],
          ["serverless", t("home_playground_tab_serverless")],
          ["monolith", t("home_playground_tab_monolith")],
          ["hexagonal", t("home_playground_tab_hexagonal")],
          ["event-driven", t("home_playground_tab_event")],
          ["ai-context", t("home_tab_ai_context")],
        ].map(([name, label]) => (
          <button
            key={name}
            className={`playground-tab${tab === name ? " active" : ""}`}
            onClick={() => switchTab(name)}
          >
            {label}
          </button>
        ))}
      </div>
      <div className="playground-toolbar">
        <span className="playground-toolbar-label">spec.yaml</span>
        <div className="playground-toolbar-actions">
          <div className="playground-view-toggle">
            <button
              className={`btn btn-sm view-toggle-btn${view === "neir" ? " active" : ""}`}
              data-view="neir"
              onClick={() => setView("neir")}
            >
              NEIR
            </button>
            <button
              className={`btn btn-sm view-toggle-btn${view === "ai" ? " active" : ""}`}
              data-view="ai"
              onClick={() => setView("ai")}
            >
              {t("home_playground_ai_ctx")}
            </button>
          </div>
          <button className="btn btn-sm" onClick={copySpec} aria-label={t("home_playground_copy_spec")}>
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" aria-hidden="true"><rect x="9" y="9" width="13" height="13" rx="2" /><path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1" /></svg>
            {copied ? t("home_playground_copied") : t("home_playground_copy")}
          </button>
          <button className="btn btn-sm" onClick={tryInTerminal} aria-label={t("home_playground_try_terminal")}>
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" aria-hidden="true"><polyline points="4 17 10 11 4 5" /><line x1="12" y1="19" x2="20" y2="19" /></svg>
            {t("home_playground_terminal")}
          </button>
        </div>
      </div>
      <div className="playground-body">
        <div className="playground-editor">
          <textarea
            ref={textareaRef}
            placeholder={t("home_playground_placeholder")}
            spellCheck={false}
            onChange={(e) => onInput(e.target.value)}
            defaultValue={SAMPLES.microservices}
          />
        </div>
        <div className="playground-preview" id="playground-output">
          {error && <div className="playground-error">{t("home_playground_invalid_yaml")}: {error}</div>}
          {!error && spec && (view === "neir" ? renderNeir(spec) : renderAiContext(spec))}
        </div>
      </div>
    </div>
  );
}
