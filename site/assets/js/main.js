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
  var els = document.querySelectorAll('.fade-in');
  if (!els.length) return;
  var observer = new IntersectionObserver(function (entries) {
    entries.forEach(function (entry) {
      if (entry.isIntersecting) {
        entry.target.classList.add('visible');
        observer.unobserve(entry.target);
      }
    });
  }, { threshold: 0.1 });
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
      navigator.clipboard.writeText(text).then(function () {
        btn.textContent = 'Copied!';
        btn.classList.add('copied');
        setTimeout(function () {
          btn.textContent = 'Copy';
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
      stars.textContent = '—';
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
  var btn = document.querySelector('.cookie-banner .btn');
  if (!banner) return;
  if (localStorage.getItem('cookies-accepted')) return;
  setTimeout(function () { banner.classList.add('show'); }, 1000);
  if (btn) {
    btn.addEventListener('click', function () {
      localStorage.setItem('cookies-accepted', 'true');
      banner.classList.remove('show');
    });
  }
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
  root.setAttribute('data-theme', next);
  localStorage.setItem('theme', next);
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

function performSearch(query) {
  if (!searchResults) return;
  var hint = document.querySelector('.search-hint');
  if (!query.trim()) {
    if (hint) hint.style.display = 'block';
    searchResults.innerHTML = '';
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
    searchResults.innerHTML = '<div class="search-hint">' + SEARCH_NO_RESULTS + '</div>';
    return;
  }
  var html = '';
  var maxResults = 20;
  for (var i = 0; i < Math.min(results.length, maxResults); i++) {
    var r = results[i];
    var item = r.item || r;
    var title = item.title || '';
    var section = item.section || '';
    var url = item.permalink || item.url || '#';
    var excerpt = item.content ? item.content.substring(0, 120) : '';
    html += '<a href="' + url + '" class="search-result-item" data-index="' + i + '">';
    html += '  <div class="result-title">' + escapeHtml(title) + '</div>';
    if (section) html += '  <div class="result-section">' + escapeHtml(section) + '</div>';
    if (excerpt) html += '  <div class="result-excerpt">' + escapeHtml(excerpt) + '</div>';
    html += '</a>';
  }
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