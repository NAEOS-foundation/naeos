import type { NextConfig } from "next";

const securityHeaders = [
  { key: "X-Frame-Options", value: "SAMEORIGIN" },
  { key: "X-Content-Type-Options", value: "nosniff" },
  { key: "Referrer-Policy", value: "strict-origin-when-cross-origin" },
  { key: "Strict-Transport-Security", value: "max-age=63072000; includeSubDomains; preload" },
  { key: "X-XSS-Protection", value: "1; mode=block" },
  {
    key: "Permissions-Policy",
    value: "camera=(), microphone=(), geolocation=()",
  },
  {
    key: "Content-Security-Policy",
    value: [
      "default-src 'self'",
      "script-src 'self' 'unsafe-inline' https://api.github.com https://cdn.jsdelivr.net https://cloud.umami.is https://unpkg.com",
      "style-src 'self' 'unsafe-inline' https://fonts.googleapis.com https://cdn.jsdelivr.net https://unpkg.com",
      "font-src 'self' https://fonts.gstatic.com data:",
      "img-src 'self' data: blob: https:",
      "connect-src 'self' https://api.github.com wss://ws.naeos.dev",
      "frame-ancestors 'none'",
    ].join("; "),
  },
];

const nextConfig: NextConfig = {
  trailingSlash: true,
  staticPageGenerationTimeout: 300,
  experimental: {
    staticGenerationMaxConcurrency: 2,
  },
  async headers() {
    return [
      { source: "/(.*)", headers: securityHeaders },
      {
        source: "/:path*.(css|js|svg|png|ico|webmanifest|woff2|woff)",
        headers: [
          { key: "Cache-Control", value: "public, max-age=31536000, immutable" },
        ],
      },
      {
        source: "/service-worker.js",
        headers: [{ key: "Cache-Control", value: "max-age=0" }],
      },
      {
        source: "/manifest.json",
        headers: [{ key: "Cache-Control", value: "max-age=3600" }],
      },
    ];
  },
  async redirects() {
    return [
      { source: "/install.sh", destination: "/downloads/install.sh", permanent: true },
      { source: "/docs/api-reference", destination: "/docs/api/", permanent: false },
    ];
  },
};

export default nextConfig;

import { initOpenNextCloudflareForDev } from "@opennextjs/cloudflare";
initOpenNextCloudflareForDev();
