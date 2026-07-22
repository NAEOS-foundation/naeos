---
title: Download
description: Install NAEOS and start engineering your next project.
---

## Installation Methods

<div class="download-grid">
<div class="download-card">
<h3>Go Install</h3>
<p>Install directly using Go. Requires Go 1.25+.</p>
  <div class="code-block">
      <div class="code-block-header"><span>bash</span><button class="copy-btn" aria-label="Copy code"><svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true"><rect x="9" y="9" width="13" height="13" rx="2"/><path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1"/></svg>Copy</button></div>
      <pre><code>go install github.com/NAEOS-foundation/naeos/cmd/naeos@latest</code></pre>
  </div>
  </div>

  <div class="download-card">
  <h3>Docker</h3>
  <p>Run using Docker container.</p>
  <div class="code-block">
      <div class="code-block-header"><span>bash</span><button class="copy-btn" aria-label="Copy code"><svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true"><rect x="9" y="9" width="13" height="13" rx="2"/><path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1"/></svg>Copy</button></div>
      <pre><code>docker pull ghcr.io/naeos-foundation/naeos:latest
docker run --rm ghcr.io/naeos-foundation/naeos:latest naeos version</code></pre>
  </div>
  </div>

  <div class="download-card">
  <h3>Binary Release</h3>
  <p>Download the latest binary from GitHub Releases.</p>
  <a href="https://github.com/NAEOS-foundation/naeos/releases" class="btn btn-primary" target="_blank" rel="noopener">View Releases</a>
  </div>

  <div class="download-card">
  <h3>Build from Source</h3>
  <p>Clone the repository and build manually.</p>
  <div class="code-block">
      <div class="code-block-header"><span>bash</span><button class="copy-btn" aria-label="Copy code"><svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true"><rect x="9" y="9" width="13" height="13" rx="2"/><path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1"/></svg>Copy</button></div>
      <pre><code>git clone https://github.com/NAEOS-foundation/naeos.git
cd naeos
go build ./cmd/naeos/</code></pre>
</div>
</div>
</div>

## Platform Support

| Platform | Support |
|----------|---------|
| Linux (amd64) | ✅ |
| Linux (arm64) | ✅ |
| macOS (amd64) | ✅ |
| macOS (arm64) | ✅ |
| Windows (amd64) | ✅ |

## Verify Installation

```bash
naeos version
```

## Quick Start

After installation, initialize your first project:

```bash
naeos init
naeos run --help
```