---
title: Releases
description: Version history and release notes for NAEOS.
---

## Release History

The latest release of NAEOS is built from the `main` branch. All releases are tagged and published on [GitHub Releases](https://github.com/NAEOS-foundation/naeos/releases).

<div id="releases-container">
  <table class="releases-table">
    <thead>
      <tr><th>Version</th><th>Description</th><th>Date</th></tr>
    </thead>
    <tbody id="releases-body">
      <tr><td colspan="3" style="text-align:center;padding:2rem;">Loading releases from GitHub...</td></tr>
    </tbody>
  </table>
</div>

## Changelog

For a detailed changelog, visit the [GitHub Releases page](https://github.com/NAEOS-foundation/naeos/releases) or read the [CHANGELOG](https://github.com/NAEOS-foundation/naeos/blob/main/CHANGELOG.md) in the repository.

### Versioning

NAEOS follows [Semantic Versioning](https://semver.org/). Given a version number **MAJOR.MINOR.PATCH**:

- **MAJOR** — Incompatible API changes
- **MINOR** — New functionality in a backward-compatible manner
- **PATCH** — Backward-compatible bug fixes

<script>
document.addEventListener('DOMContentLoaded', function() {
  var tbody = document.getElementById('releases-body');
  if (!tbody) return;
  fetch('https://api.github.com/repos/NAEOS-foundation/naeos/releases?per_page=20')
    .then(function(r) { return r.json(); })
    .then(function(data) {
      if (!Array.isArray(data) || data.length === 0) {
        tbody.innerHTML = '<tr><td colspan="3" style="text-align:center;padding:2rem;color:var(--color-text-dim);">No releases found.</td></tr>';
        return;
      }
      var html = '';
      data.forEach(function(release) {
        var date = new Date(release.published_at).toLocaleDateString('en-US', { year: 'numeric', month: 'long', day: 'numeric' });
        var desc = (release.body || '').split('\n')[0] || '';
        if (desc.length > 120) desc = desc.substring(0, 120) + '...';
        html += '<tr>';
        html += '<td class="release-version"><a href="' + release.html_url + '" target="_blank" rel="noopener">' + release.tag_name + '</a></td>';
        html += '<td>' + escapeHtml(desc) + '</td>';
        html += '<td>' + date + '</td>';
        html += '</tr>';
      });
      tbody.innerHTML = html;
    })
    .catch(function() {
      tbody.innerHTML = '<tr><td colspan="3" style="text-align:center;padding:2rem;color:var(--color-text-dim);">Unable to load releases. <a href="https://github.com/NAEOS-foundation/naeos/releases" target="_blank">View on GitHub</a></td></tr>';
    });
});
</script>
