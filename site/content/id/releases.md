---
title: Rilis
description: Riwayat versi dan catatan rilis untuk NAEOS.
---

## Riwayat Rilis

Rilis terbaru NAEOS dibuat dari cabang `main`. Semua rilis diberi tag dan dipublikasikan di [GitHub Releases](https://github.com/NAEOS-foundation/naeos/releases).

<div id="releases-container">
  <table class="releases-table">
    <thead>
      <tr><th>Versi</th><th>Deskripsi</th><th>Tanggal</th></tr>
    </thead>
    <tbody id="releases-body">
      <tr><td colspan="3" style="text-align:center;padding:2rem;">Memuat rilis dari GitHub...</td></tr>
    </tbody>
  </table>
</div>

## Catatan Perubahan Lengkap

<div id="changelog-container">
  <div class="changelog-filters" id="changelog-filters"></div>
  <div id="changelog-entries">Memuat catatan perubahan dari GitHub...</div>
</div>

<script>
document.addEventListener('DOMContentLoaded', function() {
  var tbody = document.getElementById('releases-body');
  if (tbody) {
    fetch('https://api.github.com/repos/NAEOS-foundation/naeos/releases?per_page=20')
      .then(function(r) { return r.json(); })
      .then(function(data) {
        if (!Array.isArray(data) || data.length === 0) {
          tbody.innerHTML = '<tr><td colspan="3" style="text-align:center;padding:2rem;color:var(--color-text-dim);">Tidak ada rilis.</td></tr>';
          return;
        }
        var html = '';
        data.forEach(function(release) {
          var date = new Date(release.published_at).toLocaleDateString('id-ID', { year: 'numeric', month: 'long', day: 'numeric' });
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
        tbody.innerHTML = '<tr><td colspan="3" style="text-align:center;padding:2rem;color:var(--color-text-dim);">Gagal memuat rilis. <a href="https://github.com/NAEOS-foundation/naeos/releases" target="_blank">Lihat di GitHub</a></td></tr>';
      });
  }

  var clContainer = document.getElementById('changelog-entries');
  if (!clContainer) return;
  fetch('https://raw.githubusercontent.com/NAEOS-foundation/naeos/main/CHANGELOG.md')
    .then(function(r) { return r.text(); })
    .then(function(md) {
      var entries = parseChangelog(md);
      if (entries.length === 0) {
        clContainer.innerHTML = '<p style="color:var(--color-text-dim);">Tidak ada entri changelog.</p>';
        return;
      }
      renderChangelog(entries);
    })
    .catch(function() {
      clContainer.innerHTML = '<p style="color:var(--color-text-dim);">Gagal memuat changelog. <a href="https://github.com/NAEOS-foundation/naeos/blob/main/CHANGELOG.md" target="_blank">Lihat di GitHub</a></p>';
    });
});

function parseChangelog(md) {
  var lines = md.split('\n');
  var entries = [];
  var current = null;
  var sectionType = '';

  lines.forEach(function(line) {
    var verMatch = line.match(/^## \[([^\]]+)\]\s*-\s*(\d{4}-\d{2}-\d{2})/);
    if (verMatch) {
      if (current) entries.push(current);
      current = { version: verMatch[1], date: verMatch[2], sections: {} };
      sectionType = '';
      return;
    }
    var sectionMatch = line.match(/^### (.+)/);
    if (sectionMatch && current) {
      sectionType = sectionMatch[1];
      current.sections[sectionType] = current.sections[sectionType] || [];
      return;
    }
    if (current && sectionType && line.trim().startsWith('-')) {
      var item = line.trim().replace(/^- /, '');
      current.sections[sectionType].push(item);
    }
  });
  if (current) entries.push(current);
  return entries;
}

var _allEntries = [];

function renderChangelog(entries) {
  _allEntries = entries;
  var versions = {};
  entries.forEach(function(e) {
    var major = e.version.split('.')[0];
    if (!versions[major]) versions[major] = [];
    versions[major].push(e);
  });

  var filterContainer = document.getElementById('changelog-filters');
  var majorKeys = Object.keys(versions).sort(function(a,b) { return parseInt(b) - parseInt(a); });
  var filterHtml = '<button class="changelog-filter-btn active" data-filter="all">Semua</button>';
  majorKeys.forEach(function(v) {
    filterHtml += '<button class="changelog-filter-btn" data-filter="v' + v + '">v' + v + '.x</button>';
  });
  filterContainer.innerHTML = filterHtml;
  filterContainer.querySelectorAll('.changelog-filter-btn').forEach(function(btn) {
    btn.addEventListener('click', function() {
      filterContainer.querySelectorAll('.changelog-filter-btn').forEach(function(b) { b.classList.remove('active'); });
      btn.classList.add('active');
      var filter = btn.dataset.filter;
      var filtered = filter === 'all' ? entries : entries.filter(function(e) { return e.version.startsWith(filter.replace('v', '')); });
      renderEntries(filtered);
    });
  });

  renderEntries(entries);
}

function renderEntries(entries) {
  var container = document.getElementById('changelog-entries');
  var html = '';
  entries.forEach(function(e) {
    html += '<div class="changelog-entry">';
    html += '<h3>[' + escapeHtml(e.version) + ']</h3>';
    html += '<div class="changelog-date">' + e.date + '</div>';
    var order = ['Added', 'Changed', 'Fixed', 'Removed', 'Deprecated', 'Security'];
    order.forEach(function(section) {
      if (e.sections[section] && e.sections[section].length) {
        html += '<h4>' + section + '</h4>';
        html += '<ul>';
        e.sections[section].forEach(function(item) {
          html += '<li>' + escapeHtml(item) + '</li>';
        });
        html += '</ul>';
      }
    });
    html += '</div>';
  });
  if (!entries.length) {
    html = '<p style="color:var(--color-text-dim);">Tidak ada entri yang cocok dengan filter.</p>';
  }
  container.innerHTML = html;
}
</script>

### Versioning

NAEOS mengikuti [Semantic Versioning](https://semver.org/). Nomor versi **MAJOR.MINOR.PATCH** berarti:

- **MAJOR** — Perubahan API yang tidak kompatibel
- **MINOR** — Fungsionalitas baru yang kompatibel dengan versi sebelumnya
- **PATCH** — Perbaikan bug yang kompatibel dengan versi sebelumnya
