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
<div class="code-block-header"><span>bash</span><button class="copy-btn">Copy</button></div>
<pre><code>go install github.com/NAEOS-foundation/naeos/cmd/naeos@latest</code></pre>
</div>
</div>

<div class="download-card">
<h3>Docker</h3>
<p>Run using Docker container.</p>
<div class="code-block">
<div class="code-block-header"><span>bash</span><button class="copy-btn">Copy</button></div>
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
<div class="code-block-header"><span>bash</span><button class="copy-btn">Copy</button></div>
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