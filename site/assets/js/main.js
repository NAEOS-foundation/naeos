document.addEventListener('DOMContentLoaded', function () {
  initMobileMenu();
  initScrollAnimations();
  initCountUp();
  initCopyButtons();
  initTerminalAnimation();
  initGitHubStats();
  initPlayground();
  initFAQ();
  initCookieBanner();
  initNewsletter();
  initTheme();
  initSearch();
  initKeyboardShortcuts();
  initHeaderScroll();
  initBackToTop();
  initParticles();
  initSmoothScroll();
  initScrollProgress();
  initAnchorHeadings();
  initCopyOnHover();
  initSidebarFilter();
  initImageLightbox();
  initPageTransitions();
  initDocsDrawer();
  initTocScrollspy();
});

function toggleMobileMenu(force) {
  var menu = document.getElementById('mobile-menu');
  var btn = document.querySelector('.mobile-menu-btn');
  if (!menu || !btn) return;
  if (force === true) { menu.classList.add('open'); btn.classList.add('open'); }
  else if (force === false) { menu.classList.remove('open'); btn.classList.remove('open'); }
  else { menu.classList.toggle('open'); btn.classList.toggle('open'); }
  document.body.style.overflow = menu.classList.contains('open') ? 'hidden' : '';
  if (menu.classList.contains('open')) {
    var first = menu.querySelector('a, button');
    if (first) first.focus();
  }
}

function initMobileMenu() {
  var btn = document.querySelector('.mobile-menu-btn');
  var menu = document.getElementById('mobile-menu');
  if (!btn || !menu) return;
  btn.addEventListener('click', function () { toggleMobileMenu(); });
  menu.querySelectorAll('a, button').forEach(function (el) {
    el.addEventListener('click', function () { toggleMobileMenu(false); });
  });
  menu.addEventListener('keydown', function (e) {
    if (e.key === 'Escape') { toggleMobileMenu(false); btn.focus(); return; }
    if (e.key !== 'Tab') return;
    var focusable = menu.querySelectorAll('a, button');
    if (!focusable.length) return;
    var first = focusable[0];
    var last = focusable[focusable.length - 1];
    if (e.shiftKey && document.activeElement === first) { e.preventDefault(); last.focus(); }
    else if (!e.shiftKey && document.activeElement === last) { e.preventDefault(); first.focus(); }
  });
}

function initScrollAnimations() {
  var els = document.querySelectorAll('.fade-in, .fade-in-left, .fade-in-right, .fade-in-scale, .stagger-fade');
  if (!els.length) return;
  var mql = window.matchMedia('(prefers-reduced-motion: reduce)');
  if (mql.matches) {
    els.forEach(function (el) { el.classList.add('visible'); });
    return;
  }
  var observer = new IntersectionObserver(function (entries) {
    entries.forEach(function (entry) {
      if (entry.isIntersecting) {
        entry.target.classList.add('visible');
        observer.unobserve(entry.target);
      }
    });
  }, { threshold: 0.1, rootMargin: '0px 0px -40px 0px' });
  els.forEach(function (el) { observer.observe(el); });
}

function initCountUp() {
  var counters = document.querySelectorAll('.stat-number');
  if (!counters.length) return;
  var observer = new IntersectionObserver(function (entries) {
    entries.forEach(function (entry) {
      if (entry.isIntersecting) {
        var el = entry.target;
        var target = parseInt(el.getAttribute('data-count'), 10);
        if (isNaN(target)) return;
        animateCounter(el, target);
        observer.unobserve(el);
      }
    });
  }, { threshold: 0.5 });
  counters.forEach(function (el) { observer.observe(el); });
}

function animateCounter(el, target) {
  var duration = 1500;
  var start = 0;
  var startTime = null;
  function step(timestamp) {
    if (!startTime) startTime = timestamp;
    var progress = Math.min((timestamp - startTime) / duration, 1);
    var eased = 1 - Math.pow(1 - progress, 3);
    el.textContent = Math.floor(eased * target);
    if (progress < 1) {
      requestAnimationFrame(step);
    } else {
      el.textContent = target;
    }
  }
  requestAnimationFrame(step);
}

function initCopyButtons() {
  var btns = document.querySelectorAll('.copy-btn');
  btns.forEach(function (btn) {
    btn.addEventListener('click', function () {
      var code = this.closest('.code-block').querySelector('code');
      if (!code) return;
      var text = code.textContent;
      var original = btn.innerHTML;
      navigator.clipboard.writeText(text).then(function () {
        btn.innerHTML = '<svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" style="vertical-align:middle;margin-right:4px;"><polyline points="20 6 9 17 4 12"/></svg>Copied!';
        btn.classList.add('copied');
        setTimeout(function () {
          btn.innerHTML = original;
          btn.classList.remove('copied');
        }, 2000);
      });
    });
  });
}

function initTerminalAnimation() {
  var lines = document.querySelectorAll('.terminal-line');
  if (!lines.length) return;
  lines.forEach(function (line, i) {
    line.style.animationDelay = (i * 0.4 + 0.5) + 's';
  });
}

function initGitHubStats() {
  var stars = document.getElementById('gh-stars');
  var forks = document.getElementById('gh-forks');
  var issues = document.getElementById('gh-issues');
  var contributors = document.getElementById('gh-contributors');
  if (!stars) return;
  fetch('https://api.github.com/repos/NAEOS-foundation/naeos')
    .then(function (r) { return r.json(); })
    .then(function (data) {
      if (data.stargazers_count !== undefined) {
        animateCounter(stars, data.stargazers_count);
      }
      if (data.forks_count !== undefined) {
        animateCounter(forks, data.forks_count);
      }
      if (data.open_issues_count !== undefined) {
        animateCounter(issues, data.open_issues_count);
      }
    })
    .catch(function () {
      stars.textContent = '\u2014';
    });
  fetch('https://api.github.com/repos/NAEOS-foundation/naeos/contributors?per_page=1&anon=true')
    .then(function (r) {
      var link = r.headers.get('Link');
      if (link) {
        var m = link.match(/page=(\d+)>; rel="last"/);
        if (m) { animateCounter(contributors, parseInt(m[1], 10)); }
      }
    })
    .catch(function () {});
}

var playgroundSamples = {
  yaml: 'project: my-service\nversion: "1.0"\nmodules:\n  - name: api-gateway\n    path: ./api-gateway\n    dependencies: [user-service, order-service]\n  - name: user-service\n    path: ./services/users\n    dependencies: [database]\n  - name: order-service\n    path: ./services/orders\n    dependencies: [user-service, payment-service]\n  - name: payment-service\n    path: ./services/payments\n  - name: database\n    path: ./infra/db\nservices:\n  - name: api-gateway\n    kind: reverse-proxy\n    port: 8080\n  - name: user-api\n    kind: rest\n    port: 9001\n  - name: order-api\n    kind: rest\n    port: 9002\narchitecture:\n  pattern: microservices\ngeneration:\n  languages: [go, typescript]\n  output_dir: ./generated',
  serverless: 'project: serverless-app\nversion: "1.0"\nmodules:\n  - name: auth\n    path: ./functions/auth\n  - name: api\n    path: ./functions/api\n    dependencies: [auth]\n  - name: processor\n    path: ./functions/processor\n    dependencies: [api]\nservices:\n  - name: auth-function\n    kind: lambda\n  - name: api-function\n    kind: lambda\n  - name: processor-function\n    kind: lambda\narchitecture:\n  pattern: serverless\ndeployment:\n  strategy: serverless-framework\ngeneration:\n  languages: [python, typescript]',
  monolith: 'project: monolith-app\nversion: "1.0"\nmodules:\n  - name: core\n    path: ./core\n  - name: web\n    path: ./web\n    dependencies: [core]\n  - name: database\n    path: ./infra/db\n    dependencies: [core]\nservices:\n  - name: web-server\n    kind: http\n    port: 8080\narchitecture:\n  pattern: monolithic\ndeployment:\n  strategy: docker-compose\ngeneration:\n  languages: [go]\n  output_dir: ./cmd',
  'ai-context': 'project: my-genai-service\nversion: "1.0"\nmodules:\n  - name: agent-orchestrator\n    path: ./orchestrator\n    dependencies: [llm-provider, memory-store]\n  - name: llm-provider\n    path: ./providers/llm\n    dependencies: [vector-db]\n  - name: memory-store\n    path: ./stores/memory\n  - name: vector-db\n    path: ./infra/vector\n    kind: database\n    engine: qdrant\nservices:\n  - name: api-gateway\n    kind: reverse-proxy\n    port: 8080\n  - name: chat-api\n    kind: rest\n    port: 9001\n  - name: streaming-ws\n    kind: websocket\n    port: 9002\narchitecture:\n  pattern: microservices\nai:\n  providers:\n    - name: openai\n      models: [gpt-4o, gpt-4o-mini]\n    - name: anthropic\n      models: [claude-opus-4, claude-sonnet-4]\n  context:\n    format: neir\n    compression: semantic\n    max_tokens: 128000\ngeneration:\n  languages: [go, typescript, python]\n  ai_instructions: true\n  output_dir: ./generated'
};

function initPlayground() {
  var input = document.getElementById('playground-input');
  var output = document.getElementById('playground-output');
  if (!input || !output) return;
  input.value = playgroundSamples.yaml;
  updatePlaygroundPreview();
  input.addEventListener('input', updatePlaygroundPreview);
}

function switchPlayground(btn, name) {
  var tabs = document.querySelectorAll('.playground-tab');
  tabs.forEach(function (t) { t.classList.remove('active'); });
  btn.classList.add('active');
  var input = document.getElementById('playground-input');
  if (input && playgroundSamples[name]) {
    input.value = playgroundSamples[name];
    updatePlaygroundPreview();
  }
}

function updatePlaygroundPreview() {
  var input = document.getElementById('playground-input');
  var output = document.getElementById('playground-output');
  if (!input || !output) return;
  var text = input.value;
  var lines = text.split('\n').filter(function (l) { return l.trim(); });
  var html = '<h4>NEIR Model Preview</h4>';
  html += '<div style="font-family:var(--font-mono);font-size:0.8125rem;line-height:1.8;">';
  var indent = 0;
  lines.forEach(function (line) {
    var trimmed = line.trim();
    if (trimmed.endsWith(':')) {
      html += '<div class="tree-node" style="margin-left:' + (indent * 16) + 'px"><span class="tree-key">' + escapeHtml(trimmed) + '</span></div>';
      indent++;
    } else if (trimmed.startsWith('- ')) {
      var parts = trimmed.split(': ');
      if (parts.length === 2) {
        html += '<div class="tree-node" style="margin-left:' + (indent * 16) + 'px"><span class="tree-key">' + escapeHtml(parts[0]) + ':</span> <span class="tree-str">' + escapeHtml(parts[1]) + '</span></div>';
      } else {
        html += '<div class="tree-node" style="margin-left:' + (indent * 16) + 'px"><span class="tree-val">' + escapeHtml(trimmed) + '</span></div>';
      }
    } else {
      var parts = trimmed.split(': ');
      if (parts.length === 2) {
        html += '<div class="tree-node" style="margin-left:' + ((indent - 1) * 16) + 'px"><span class="tree-key">' + escapeHtml(parts[0]) + ':</span> <span class="tree-str">' + escapeHtml(parts[1]) + '</span></div>';
      }
    }
  });
  html += '</div>';
  output.innerHTML = html;
}

function escapeHtml(text) {
  var d = document.createElement('div');
  d.textContent = text;
  return d.innerHTML;
}

function initFAQ() {
  var items = document.querySelectorAll('.faq-question');
  items.forEach(function (q) {
    q.addEventListener('click', function () {
      var item = this.parentElement;
      item.classList.toggle('open');
    });
  });
}

function initCookieBanner() {
  var banner = document.querySelector('.cookie-banner');
  if (!banner) return;
  if (localStorage.getItem('cookies-accepted') || localStorage.getItem('cookies-declined')) return;
  setTimeout(function () { banner.classList.add('show'); }, 1000);
  var acceptBtn = banner.querySelector('.btn-primary');
  var declineBtn = banner.querySelector('.btn-secondary');
  if (acceptBtn) {
    acceptBtn.addEventListener('click', function () {
      localStorage.setItem('cookies-accepted', 'true');
      banner.classList.remove('show');
    });
  }
  if (declineBtn) {
    declineBtn.addEventListener('click', function () {
      localStorage.setItem('cookies-declined', 'true');
      banner.classList.remove('show');
    });
  }
}
function declineCookies() {
  localStorage.setItem('cookies-declined', 'true');
  var banner = document.querySelector('.cookie-banner');
  if (banner) banner.classList.remove('show');
}

function initNewsletter() {
  var form = document.querySelector('.newsletter-form');
  var msg = document.querySelector('.newsletter-message');
  if (!form || !msg) return;
  form.addEventListener('submit', function (e) {
    e.preventDefault();
    var email = form.querySelector('input').value.trim();
    if (!email) { msg.textContent = 'Please enter your email.'; return; }
    msg.textContent = 'Thank you! You\'ve been subscribed.';
    msg.style.color = 'var(--color-accent)';
    form.querySelector('input').value = '';
    setTimeout(function () { msg.textContent = ''; }, 3000);
  });
}

function initTheme() {
  var saved = localStorage.getItem('theme');
  var prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
  var theme = saved || (prefersDark ? 'dark' : 'dark');
  document.documentElement.setAttribute('data-theme', theme);
  localStorage.setItem('theme', theme);
}

function toggleTheme() {
  var root = document.documentElement;
  var current = root.getAttribute('data-theme');
  var next = current === 'dark' ? 'light' : 'dark';
  root.classList.add('theme-transitioning');
  root.setAttribute('data-theme', next);
  localStorage.setItem('theme', next);
  setTimeout(function () { root.classList.remove('theme-transitioning'); }, 400);
}

function switchTab(btn, tabId) {
  var container = btn.closest('.tab-container');
  if (!container) return;
  var tabs = container.querySelectorAll('.tab-item');
  tabs.forEach(function (t) { t.classList.remove('active'); });
  btn.classList.add('active');
  var panels = container.querySelectorAll('.tab-content');
  panels.forEach(function (p) {
    var pid = p.getAttribute('id');
    var target = tabId + '-panel';
    if (pid === target) { p.classList.add('active'); }
    else { p.classList.remove('active'); }
  });
}

var searchData = null;
var fuseInstance = null;
var searchOverlay, searchModal, searchInput, searchResults;
var selectedIndex = -1;

function openSearch() {
  if (!searchOverlay || !searchModal) return;
  searchOverlay.classList.add('open');
  searchModal.classList.add('open');
  searchModal.style.display = 'flex';
  searchOverlay.style.display = 'block';
  setTimeout(function () { if (searchInput) searchInput.focus(); }, 100);
  if (typeof Fuse !== 'undefined' && fuseInstance === null && searchData) {
    fuseInstance = new Fuse(searchData, {
      keys: ['title', 'sections', 'content'],
      threshold: 0.4,
      includeScore: true,
      includeMatches: true
    });
  }
}

function closeSearch() {
  if (!searchOverlay || !searchModal) return;
  searchOverlay.classList.remove('open');
  searchModal.classList.remove('open');
  searchOverlay.style.display = 'none';
  searchModal.style.display = 'none';
  if (searchInput) searchInput.value = '';
  if (searchResults) searchResults.innerHTML = '';
  selectedIndex = -1;
}

function getRecentSearches() {
  try { return JSON.parse(localStorage.getItem('recent-searches') || '[]'); } catch (e) { return []; }
}
function saveRecentSearch(query) {
  var searches = getRecentSearches().filter(function (s) { return s !== query; });
  searches.unshift(query);
  if (searches.length > 5) searches = searches.slice(0, 5);
  localStorage.setItem('recent-searches', JSON.stringify(searches));
}

function initSearch() {
  searchOverlay = document.getElementById('search-overlay');
  searchModal = document.getElementById('search-modal');
  searchInput = document.getElementById('search-input');
  searchResults = document.getElementById('search-results');
  if (!searchOverlay || !searchModal || !searchInput || !searchResults) return;

  if (typeof Fuse === 'undefined' && !document.querySelector('script[src*="fuse.js"]')) {
    var script = document.createElement('script');
    script.src = 'https://cdn.jsdelivr.net/npm/fuse.js@7.0.0/dist/fuse.min.js';
    script.onload = function () { loadSearchIndex(); };
    document.head.appendChild(script);
  } else if (typeof Fuse !== 'undefined') {
    loadSearchIndex();
  }

  searchOverlay.addEventListener('click', function (e) {
    if (e.target === searchOverlay) closeSearch();
  });

  searchInput.addEventListener('keydown', function (e) {
    if (e.key === 'Escape') { closeSearch(); return; }
    if (e.key === 'ArrowDown') { e.preventDefault(); navigateResults(1); return; }
    if (e.key === 'ArrowUp') { e.preventDefault(); navigateResults(-1); return; }
    if (e.key === 'Enter') { e.preventDefault(); selectResult(); return; }
  });

  searchInput.addEventListener('input', function () { performSearch(searchInput.value); });

  searchInput.addEventListener('focus', function () {
    if (!this.value.trim()) showRecentSearches();
  });
}

function showRecentSearches() {
  if (!searchResults) return;
  var recent = getRecentSearches();
  if (!recent.length) {
    searchResults.innerHTML = '<div class="search-hint">' + (window.SEARCH_PLACEHOLDER || 'Type to search') + '</div>';
    return;
  }
  var html = '<div class="search-recent-header">Recent searches</div>';
  recent.forEach(function (q) {
    html += '<button class="search-recent-item" onclick="document.getElementById(\'search-input\').value=\'' + escapeHtml(q) + '\';performSearch(\'' + escapeHtml(q) + '\');this.focus();">';
    html += '<svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>';
    html += escapeHtml(q);
    html += '</button>';
  });
  searchResults.innerHTML = html;
}

function loadSearchIndex() {
  var indexURL = '/index.json';
  if (document.documentElement.lang === 'id' || window.location.pathname.startsWith('/id/')) {
    indexURL = '/id/index.json';
  }
  fetch(indexURL)
    .then(function (r) { return r.json(); })
    .then(function (data) {
      searchData = data;
      if (typeof Fuse !== 'undefined') {
        fuseInstance = new Fuse(searchData, {
          keys: ['title', 'sections', 'content'],
          threshold: 0.4,
          includeScore: true,
          includeMatches: true
        });
      }
    })
    .catch(function () {});
}

function highlightMatches(text, query) {
  if (!query || !text) return escapeHtml(text);
  var escaped = escapeHtml(text);
  var words = query.trim().split(/\s+/).filter(function (w) { return w.length > 0; });
  words.forEach(function (word) {
    var re = new RegExp('(' + word.replace(/[.*+?^${}()|[\]\\]/g, '\\$&') + ')', 'gi');
    escaped = escaped.replace(re, '<mark>$1</mark>');
  });
  return escaped;
}

function performSearch(query) {
  if (!searchResults) return;
  var hint = document.querySelector('.search-hint');
  if (!query.trim()) {
    if (hint) hint.style.display = 'block';
    showRecentSearches();
    selectedIndex = -1;
    return;
  }
  if (hint) hint.style.display = 'none';
  var results = [];
  if (fuseInstance) {
    results = fuseInstance.search(query);
  } else if (searchData) {
    var q = query.toLowerCase();
    results = searchData.filter(function (item) {
      return (item.title && item.title.toLowerCase().indexOf(q) !== -1) ||
             (item.content && item.content.toLowerCase().indexOf(q) !== -1);
    }).map(function (item) { return { item: item }; });
  }
  selectedIndex = -1;
  if (results.length === 0) {
    searchResults.innerHTML = '<div class="search-empty">' +
      '<svg width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="var(--color-text-dim)" stroke-width="1.5" aria-hidden="true"><circle cx="11" cy="11" r="8"/><path d="m21 21-4.35-4.35"/><line x1="8" y1="11" x2="14" y2="11"/></svg>' +
      '<div class="search-empty-title">' + SEARCH_NO_RESULTS + '</div>' +
      '<div class="search-empty-hint">Try different keywords or browse the documentation.</div>' +
      '</div>';
    return;
  }
  saveRecentSearch(query);
  var maxResults = 20;
  var categories = {};
  for (var i = 0; i < Math.min(results.length, maxResults); i++) {
    var r = results[i];
    var item = r.item || r;
    var section = item.section || 'General';
    if (!categories[section]) categories[section] = [];
    categories[section].push({ item: item, r: r, i: i });
  }
  var html = '';
  var globalIndex = 0;
  var catKeys = Object.keys(categories);
  catKeys.sort();
  catKeys.forEach(function (cat) {
    html += '<div class="search-category-header">' + escapeHtml(cat) + '</div>';
    categories[cat].forEach(function (entry) {
      var item = entry.item;
      var title = item.title || '';
      var section = item.section || '';
      var url = item.permalink || item.url || '#';
      var excerpt = item.content ? item.content.substring(0, 120) : '';
      html += '<a href="' + url + '" class="search-result-item" data-index="' + globalIndex + '">';
      html += '  <div class="result-title">' + highlightMatches(title, query) + '</div>';
      if (section) html += '  <div class="result-section">' + escapeHtml(section) + '</div>';
      if (excerpt) html += '  <div class="result-excerpt">' + highlightMatches(excerpt, query) + '</div>';
      html += '</a>';
      globalIndex++;
    });
  });
  searchResults.innerHTML = html;
  var items = searchResults.querySelectorAll('.search-result-item');
  items.forEach(function (item) {
    item.addEventListener('click', function (e) { closeSearch(); });
    item.addEventListener('mouseenter', function () {
      items.forEach(function (i) { i.classList.remove('selected'); });
      this.classList.add('selected');
    });
  });
}

function navigateResults(dir) {
  var items = searchResults.querySelectorAll('.search-result-item');
  if (!items.length) return;
  items.forEach(function (i) { i.classList.remove('selected'); });
  selectedIndex += dir;
  if (selectedIndex < 0) selectedIndex = 0;
  if (selectedIndex >= items.length) selectedIndex = items.length - 1;
  items[selectedIndex].classList.add('selected');
  items[selectedIndex].scrollIntoView({ block: 'nearest' });
}

function selectResult() {
  var selected = searchResults.querySelector('.search-result-item.selected');
  if (selected) { window.location.href = selected.getAttribute('href'); closeSearch(); }
}

function initKeyboardShortcuts() {
  document.addEventListener('keydown', function (e) {
    if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
      e.preventDefault();
      openSearch();
    }
  });
}

function initHeaderScroll() {
  var header = document.querySelector('.site-header');
  if (!header) return;
  var ticking = false;
  window.addEventListener('scroll', function () {
    if (!ticking) {
      requestAnimationFrame(function () {
        if (window.scrollY > 20) {
          header.classList.add('scrolled');
        } else {
          header.classList.remove('scrolled');
        }
        ticking = false;
      });
      ticking = true;
    }
  }, { passive: true });
}

function initBackToTop() {
  var btn = document.getElementById('back-to-top');
  if (!btn) return;
  var ticking = false;
  window.addEventListener('scroll', function () {
    if (!ticking) {
      requestAnimationFrame(function () {
        if (window.scrollY > 400) {
          btn.classList.add('visible');
        } else {
          btn.classList.remove('visible');
        }
        ticking = false;
      });
      ticking = true;
    }
  }, { passive: true });
  btn.addEventListener('click', function () {
    window.scrollTo({ top: 0, behavior: 'smooth' });
  });
}

function initParticles() {
  var container = document.querySelector('.hero-particles');
  if (!container) return;
  var prefersReduced = window.matchMedia('(prefers-reduced-motion: reduce)').matches;
  if (prefersReduced) return;
  for (var i = 0; i < 20; i++) {
    var particle = document.createElement('div');
    particle.className = 'hero-particle';
    particle.style.left = (Math.random() * 100) + '%';
    particle.style.width = (2 + Math.random() * 3) + 'px';
    particle.style.height = particle.style.width;
    particle.style.animationDelay = (Math.random() * 8) + 's';
    particle.style.animationDuration = (6 + Math.random() * 6) + 's';
    container.appendChild(particle);
  }
}

function initSmoothScroll() {
  document.querySelectorAll('a[href^="#"]').forEach(function (anchor) {
    anchor.addEventListener('click', function (e) {
      var href = this.getAttribute('href');
      if (href === '#') return;
      var target = document.querySelector(href);
      if (target) {
        e.preventDefault();
        var headerHeight = parseInt(getComputedStyle(document.documentElement).getPropertyValue('--header-height'), 10) || 72;
        var targetPos = target.getBoundingClientRect().top + window.scrollY - headerHeight - 16;
        window.scrollTo({ top: targetPos, behavior: 'smooth' });
      }
    });
  });
}

function initAnchorHeadings() {
  var containers = document.querySelectorAll('.content-section, .single-content');
  containers.forEach(function (container) {
    var headings = container.querySelectorAll('h2, h3, h4');
    headings.forEach(function (h) {
      if (!h.id) return;
      if (h.querySelector('.anchor-link')) return;
      var link = document.createElement('a');
      link.className = 'anchor-link';
      link.href = '#' + h.id;
      link.setAttribute('aria-label', 'Link to this section');
      link.innerHTML = '<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true"><path d="M10 13a5 5 0 007.54.54l3-3a5 5 0 00-7.07-7.07l-1.72 1.71"/><path d="M14 11a5 5 0 00-7.54-.54l-3 3a5 5 0 007.07 7.07l1.71-1.71"/></svg>';
      h.appendChild(link);
    });
  });
}

function initSidebarFilter() {
  var input = document.getElementById('sidebar-search');
  var nav = document.getElementById('sidebar-nav');
  if (!input || !nav) return;
  input.addEventListener('input', function () {
    var q = this.value.toLowerCase().trim();
    var links = nav.querySelectorAll('a[data-title]');
    links.forEach(function (link) {
      var title = link.getAttribute('data-title').toLowerCase();
      if (!q || title.indexOf(q) !== -1) {
        link.classList.remove('hidden');
      } else {
        link.classList.add('hidden');
      }
    });
  });
}

function initImageLightbox() {
  var images = document.querySelectorAll('.content-section img, .single-content img, .doc-layout img');
  images.forEach(function (img) {
    if (img.closest('.not-found-art') || img.closest('.hero-terminal')) return;
    if (img.parentElement.tagName === 'A') return;
    img.style.cursor = 'zoom-in';
    img.addEventListener('click', function () {
      var overlay = document.createElement('div');
      overlay.className = 'lightbox-overlay';
      overlay.setAttribute('role', 'dialog');
      overlay.setAttribute('aria-label', 'Image lightbox');
      overlay.innerHTML = '<div class="lightbox-content"><img src="' + img.src + '" alt="' + (img.alt || '') + '"><button class="lightbox-close" aria-label="Close">&times;</button></div>';
      document.body.appendChild(overlay);
      document.body.style.overflow = 'hidden';
      requestAnimationFrame(function () { overlay.classList.add('open'); });
      overlay.addEventListener('click', function (e) {
        if (e.target === overlay || e.target.classList.contains('lightbox-close')) {
          overlay.classList.remove('open');
          setTimeout(function () { document.body.removeChild(overlay); document.body.style.overflow = ''; }, 300);
        }
      });
      document.addEventListener('keydown', function handler(e) {
        if (e.key === 'Escape') {
          overlay.classList.remove('open');
          setTimeout(function () { document.body.removeChild(overlay); document.body.style.overflow = ''; }, 300);
          document.removeEventListener('keydown', handler);
        }
      });
    });
  });
}

function initCopyOnHover() {
  var blocks = document.querySelectorAll('.content-section pre, .single-content pre');
  blocks.forEach(function (pre) {
    if (pre.closest('.code-block')) return;
    if (pre.querySelector('.copy-hover-btn')) return;
    var btn = document.createElement('button');
    btn.className = 'copy-hover-btn';
    btn.setAttribute('aria-label', 'Copy code');
    btn.innerHTML = '<svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true"><rect x="9" y="9" width="13" height="13" rx="2"/><path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1"/></svg>';
    pre.style.position = 'relative';
    pre.appendChild(btn);
    btn.addEventListener('click', function () {
      var code = pre.querySelector('code');
      var text = code ? code.textContent : pre.textContent;
      navigator.clipboard.writeText(text).then(function () {
        btn.innerHTML = '<svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" aria-hidden="true"><polyline points="20 6 9 17 4 12"/></svg>';
        btn.classList.add('copied');
        setTimeout(function () {
          btn.innerHTML = '<svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true"><rect x="9" y="9" width="13" height="13" rx="2"/><path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1"/></svg>';
          btn.classList.remove('copied');
        }, 2000);
      });
    });
  });
}

function initScrollProgress() {
  var bar = document.getElementById('scroll-progress');
  if (!bar) return;
  var ticking = false;
  window.addEventListener('scroll', function () {
    if (!ticking) {
      requestAnimationFrame(function () {
        var scrollTop = window.scrollY;
        var docHeight = document.documentElement.scrollHeight - window.innerHeight;
        var progress = docHeight > 0 ? (scrollTop / docHeight) * 100 : 0;
        bar.style.width = progress + '%';
        bar.style.opacity = progress > 2 ? '1' : '0';
        ticking = false;
      });
      ticking = true;
    }
  }, { passive: true });
}

function initPageTransitions() {
  var mql = window.matchMedia('(prefers-reduced-motion: reduce)');
  if (mql.matches) return;

  var overlay = document.getElementById('transition-overlay');
  var styleEl = document.getElementById('overlay-init');
  if (!overlay) return;

  var navTimeout;

  function removeOverlayStyle() {
    if (styleEl && styleEl.parentNode) styleEl.remove();
  }

  removeOverlayStyle();

  window.addEventListener('pageshow', function (e) {
    overlay.classList.remove('active');
    removeOverlayStyle();
  });

  document.addEventListener('mousedown', function (e) {
    var link = e.target.closest('a[href]');
    if (!link || e.button !== 0) return;
    if (link.hasAttribute('download') || link.hasAttribute('target')) return;
    if (link.getAttribute('rel') === 'external') return;
    if (link.closest('.search-overlay') || link.closest('#search-modal')) return;

    var url = new URL(link.href, window.location.origin);
    if (url.origin !== window.location.origin) return;
    if (url.pathname + url.search === window.location.pathname + window.location.search) return;

    e.preventDefault();
    overlay.classList.add('active');
    if (navTimeout) clearTimeout(navTimeout);
    navTimeout = setTimeout(function () { window.location.href = url.href; }, 280);
  }, { passive: false });

  var prefetched = {};
  document.addEventListener('mouseover', function (e) {
    var link = e.target.closest('a[href]');
    if (!link) return;
    if (link.hostname !== window.location.hostname) return;
    if (link.hasAttribute('download') || link.hasAttribute('target')) return;
    var href = link.href;
    if (prefetched[href]) return;
    prefetched[href] = true;
    var preload = document.createElement('link');
    preload.rel = 'prefetch';
    preload.href = href;
    preload.as = 'document';
    document.head.appendChild(preload);
  }, { passive: true });
}

function initDocsDrawer() {
  var toggle = document.getElementById('doc-drawer-toggle');
  var sidebar = document.querySelector('.doc-sidebar');
  var overlay = document.getElementById('doc-drawer-overlay');
  if (!toggle || !sidebar || !overlay) return;

  function open() { sidebar.classList.add('open'); overlay.classList.add('open'); document.body.style.overflow = 'hidden'; }
  function close() { sidebar.classList.remove('open'); overlay.classList.remove('open'); document.body.style.overflow = ''; }

  toggle.addEventListener('click', function (e) {
    e.stopPropagation();
    if (sidebar.classList.contains('open')) close(); else open();
  });

  overlay.addEventListener('click', close);

  document.addEventListener('keydown', function (e) {
    if (e.key === 'Escape' && sidebar.classList.contains('open')) close();
  });

  sidebar.querySelectorAll('a, button').forEach(function (el) {
    el.addEventListener('click', function () {
      if (window.innerWidth <= 968) close();
    });
  });
}

function initTocScrollspy() {
  var toc = document.querySelector('.toc');
  if (!toc) return;
  var headings = document.querySelectorAll('.single-content h2, .single-content h3');
  if (!headings.length) return;
  var links = toc.querySelectorAll('nav a');
  if (!links.length) return;

  var linkMap = {};
  links.forEach(function (link) {
    var href = link.getAttribute('href');
    if (href) linkMap[href.substring(1)] = link;
  });

  var mql = window.matchMedia('(prefers-reduced-motion: reduce)');

  var observer = new IntersectionObserver(function (entries) {
    var visible = [];
    entries.forEach(function (entry) {
      if (entry.isIntersecting) visible.push(entry.target.id);
    });
    if (!visible.length) return;
    links.forEach(function (l) { l.classList.remove('active'); });
    var last = visible[visible.length - 1];
    if (linkMap[last]) linkMap[last].classList.add('active');
  }, { threshold: 0, rootMargin: '-80px 0px -65% 0px' });

  headings.forEach(function (h) {
    if (h.id) observer.observe(h);
  });
}
