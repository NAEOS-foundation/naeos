"use client";

import { useEffect } from "react";

type Lang = "en" | "id";

declare global {
  interface Window {
    SwaggerUIBundle?: {
      (opts: Record<string, unknown>): unknown;
      presets: { apis: unknown };
      SwaggerUIStandalonePreset: unknown;
    };
  }
}

export default function SwaggerApi({ specUrl, lang }: { specUrl: string; lang: Lang }) {
  useEffect(() => {
    let cancelled = false;

    async function load() {
      const container = document.getElementById("swagger-ui");
      if (!container) return;
      try {
        await new Promise<void>((resolve, reject) => {
          if (window.SwaggerUIBundle) {
            resolve();
            return;
          }
          const cssLink = document.createElement("link");
          cssLink.rel = "stylesheet";
          cssLink.href = "https://unpkg.com/swagger-ui-dist@5/swagger-ui.css";
          document.head.appendChild(cssLink);

          const script = document.createElement("script");
          script.src = "https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js";
          script.onload = () => resolve();
          script.onerror = () => reject(new Error("failed to load swagger-ui"));
          document.body.appendChild(script);
        });
        if (cancelled || !window.SwaggerUIBundle) return;
        window.SwaggerUIBundle({
          url: specUrl,
          dom_id: "#swagger-ui",
          deepLinking: true,
          presets: [
            window.SwaggerUIBundle.presets.apis,
            window.SwaggerUIBundle.SwaggerUIStandalonePreset,
          ],
          layout: "BaseLayout",
          defaultModelsExpandDepth: 1,
          filter: true,
          tryItOutEnabled: false,
        });
      } catch {
        container.textContent =
          lang === "id"
            ? "Gagal memuat dokumentasi API."
            : "Failed to load the API documentation.";
      }
    }

    void load();
    return () => {
      cancelled = true;
    };
  }, [specUrl, lang]);

  return <div id="swagger-ui" />;
}
