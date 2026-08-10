---
title: Download
description: Install NAEOS and start engineering your next project.
---

<div class="download-page">

## Installation Methods

<div class="os-detected">
  <span class="os-detected-label">Detected platform:</span>
  <span class="os-detected-value" id="detected-os">detecting...</span>
</div>

<div class="os-tabs">
  <button class="os-tab" data-os="linux">Linux</button>
  <button class="os-tab" data-os="macos">macOS</button>
  <button class="os-tab" data-os="windows">Windows</button>
</div>

<div class="os-pane" data-os="linux">
  <div class="download-grid">
    <div class="download-card">
      <h3>Binary (amd64)</h3>
      <p>Linux x86_64 binary — download and run.</p>
      <a href="https://github.com/NAEOS-foundation/naeos/releases/latest/download/naeos-linux-amd64.tar.gz" class="btn btn-primary">Download for Linux</a>
    </div>
    <div class="download-card">
      <h3>Binary (arm64)</h3>
      <p>Linux ARM64 binary — for Raspberry Pi, AWS Graviton, etc.</p>
      <a href="https://github.com/NAEOS-foundation/naeos/releases/latest/download/naeos-linux-arm64.tar.gz" class="btn btn-primary">Download ARM64</a>
    </div>
  </div>
</div>

<div class="os-pane" data-os="macos">
  <div class="download-grid">
    <div class="download-card">
      <h3>Binary (amd64)</h3>
      <p>macOS x86_64 binary — for Intel Macs.</p>
      <a href="https://github.com/NAEOS-foundation/naeos/releases/latest/download/naeos-darwin-amd64.tar.gz" class="btn btn-primary">Download for macOS Intel</a>
    </div>
    <div class="download-card">
      <h3>Binary (arm64)</h3>
      <p>macOS ARM64 binary — for Apple Silicon (M1/M2/M3/M4).</p>
      <a href="https://github.com/NAEOS-foundation/naeos/releases/latest/download/naeos-darwin-arm64.tar.gz" class="btn btn-primary">Download for macOS Apple Silicon</a>
    </div>
  </div>
</div>

<div class="os-pane" data-os="windows">
  <div class="download-grid">
    <div class="download-card">
      <h3>Binary (amd64)</h3>
      <p>Windows x86_64 binary.</p>
      <a href="https://github.com/NAEOS-foundation/naeos/releases/latest/download/naeos-windows-amd64.zip" class="btn btn-primary">Download for Windows</a>
    </div>
  </div>
</div>

### Install via Go

Requires Go 1.25+.

<div class="code-block">
    <div class="code-block-header"><span>bash</span><button class="copy-btn" aria-label="Copy code"><svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true"><rect x="9" y="9" width="13" height="13" rx="2"/><path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1"/></svg>Copy</button></div>
    <pre><code>go install github.com/NAEOS-foundation/naeos/cmd/naeos@latest</code></pre>
</div>

### Run with Docker

<div class="code-block">
    <div class="code-block-header"><span>bash</span><button class="copy-btn" aria-label="Copy code"><svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true"><rect x="9" y="9" width="13" height="13" rx="2"/><path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1"/></svg>Copy</button></div>
    <pre><code>docker pull ghcr.io/naeos-foundation/naeos:latest
docker run --rm ghcr.io/naeos-foundation/naeos:latest naeos version</code></pre>
</div>

### Build from Source

<div class="code-block">
    <div class="code-block-header"><span>bash</span><button class="copy-btn" aria-label="Copy code"><svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true"><rect x="9" y="9" width="13" height="13" rx="2"/><path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1"/></svg>Copy</button></div>
    <pre><code>git clone https://github.com/NAEOS-foundation/naeos.git
cd naeos
go build ./cmd/naeos/</code></pre>
</div>

## Platform Support

| Platform | amd64 | arm64 |
|----------|-------|-------|
| Linux | ✅ | ✅ |
| macOS | ✅ | ✅ |
| Windows | ✅ | — |

## Documentation PDFs

| Document | Download |
|----------|----------|
| **NAEOS Whitepaper** (English) | [naeos-whitepaper.pdf](/downloads/naeos-whitepaper.pdf) |
| **Whitepaper NAEOS** (Bahasa Indonesia) | [naeos-whitepaper-id.pdf](/downloads/naeos-whitepaper-id.pdf) |
| CLI Reference | [naeos-cli-reference.pdf](/downloads/naeos-cli-reference.pdf) |
| Getting Started | [naeos-getting-started.pdf](/downloads/naeos-getting-started.pdf) |

## Verify Installation

```bash
naeos version
```

## Quick Start

```bash
naeos init
naeos run --help
```

</div>
