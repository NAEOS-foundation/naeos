---
title: Contributors
description: The amazing people who contribute to NAEOS.
---

## Our Contributors

NAEOS is built by an amazing community of contributors. Thank you to everyone who has helped make this project possible.

<div id="contributors-container">
  <div class="contributors-grid" id="contributors-grid">
    <p style="grid-column:1/-1;text-align:center;padding:2rem;color:var(--color-text-dim);">Loading contributors from GitHub...</p>
  </div>
</div>

## How to Contribute

We welcome contributions of all kinds. Here are some ways you can help:

- **Code** — Submit pull requests for bug fixes or new features
- **Documentation** — Improve docs, write tutorials, or translate content
- **Issues** — Report bugs or suggest features via GitHub Issues
- **Community** — Help answer questions in GitHub Discussions
- **Design** — Improve the website, logo, or branding

To get started, read our [Contributing Guide](https://github.com/NAEOS-foundation/naeos/blob/main/CONTRIBUTING.md).

<script>
document.addEventListener('DOMContentLoaded', function() {
  var grid = document.getElementById('contributors-grid');
  if (!grid) return;
  fetch('https://api.github.com/repos/NAEOS-foundation/naeos/contributors?per_page=48')
    .then(function(r) { return r.json(); })
    .then(function(data) {
      if (!Array.isArray(data) || data.length === 0) {
        grid.innerHTML = '<p style="grid-column:1/-1;text-align:center;padding:2rem;color:var(--color-text-dim);">No contributors found.</p>';
        return;
      }
      var html = '';
      data.forEach(function(c) {
        html += '<div class="contributor-card">';
        html += '  <img class="contributor-avatar" src="' + c.avatar_url + '&s=80" alt="' + escapeHtml(c.login) + '" loading="lazy" width="56" height="56">';
        html += '  <h4><a href="' + c.html_url + '" target="_blank" rel="noopener">@' + escapeHtml(c.login) + '</a></h4>';
        html += '  <p>' + c.contributions + ' contribution' + (c.contributions !== 1 ? 's' : '') + '</p>';
        html += '</div>';
      });
      grid.innerHTML = html;
    })
    .catch(function() {
      grid.innerHTML = '<p style="grid-column:1/-1;text-align:center;padding:2rem;color:var(--color-text-dim);">Unable to load contributors. <a href="https://github.com/NAEOS-foundation/naeos/graphs/contributors" target="_blank">View on GitHub</a></p>';
    });
});
</script>
