document.addEventListener("DOMContentLoaded",function(){initMobileMenu(),initScrollAnimations(),initCountUp(),initCopyButtons(),initTerminalAnimation(),initGitHubStats(),initPlayground(),initFAQ(),initCookieBanner(),initNewsletter(),initTheme(),initSearch(),initKeyboardShortcuts()});function toggleMobileMenu(e){var s,t=document.getElementById("mobile-menu"),n=document.querySelector(".mobile-menu-btn");if(!t||!n)return;e===!0?(t.classList.add("open"),n.classList.add("open")):e===!1?(t.classList.remove("open"),n.classList.remove("open")):(t.classList.toggle("open"),n.classList.toggle("open")),document.body.style.overflow=t.classList.contains("open")?"hidden":"",t.classList.contains("open")&&(s=t.querySelector("a, button"),s&&s.focus())}function initMobileMenu(){var t=document.querySelector(".mobile-menu-btn"),e=document.getElementById("mobile-menu");if(!t||!e)return;t.addEventListener("click",function(){toggleMobileMenu()}),e.querySelectorAll("a, button").forEach(function(e){e.addEventListener("click",function(){toggleMobileMenu(!1)})}),e.addEventListener("keydown",function(n){if(n.key==="Escape"){toggleMobileMenu(!1),t.focus();return}if(n.key!=="Tab")return;var o,i,s=e.querySelectorAll("a, button");if(!s.length)return;o=s[0],i=s[s.length-1],n.shiftKey&&document.activeElement===o?(n.preventDefault(),i.focus()):!n.shiftKey&&document.activeElement===i&&(n.preventDefault(),o.focus())})}function initScrollAnimations(){var e,t=document.querySelectorAll(".fade-in");if(!t.length)return;e=new IntersectionObserver(function(t){t.forEach(function(t){t.isIntersecting&&(t.target.classList.add("visible"),e.unobserve(t.target))})},{threshold:.1}),t.forEach(function(t){e.observe(t)})}function initCountUp(){var e,t=document.querySelectorAll(".stat-number");if(!t.length)return;e=new IntersectionObserver(function(t){t.forEach(function(t){if(t.isIntersecting){var n=t.target,s=parseInt(n.getAttribute("data-count"),10);if(isNaN(s))return;animateCounter(n,s),e.unobserve(n)}})},{threshold:.5}),t.forEach(function(t){e.observe(t)})}function animateCounter(e,t){var o=1500,i=0,n=null;function s(i){n||(n=i);var a=Math.min((i-n)/o,1),r=1-Math.pow(1-a,3);e.textContent=Math.floor(r*t),a<1?requestAnimationFrame(s):e.textContent=t}requestAnimationFrame(s)}function initCopyButtons(){var e=document.querySelectorAll(".copy-btn");e.forEach(function(e){e.addEventListener("click",function(){var n,t=this.closest(".code-block").querySelector("code");if(!t)return;n=t.textContent,navigator.clipboard.writeText(n).then(function(){e.textContent="Copied!",e.classList.add("copied"),setTimeout(function(){e.textContent="Copy",e.classList.remove("copied")},2e3)})})})}function initTerminalAnimation(){var e=document.querySelectorAll(".terminal-line");if(!e.length)return;e.forEach(function(e,t){e.style.animationDelay=t*.4+.5+"s"})}function initGitHubStats(){var e=document.getElementById("gh-stars"),t=document.getElementById("gh-forks"),n=document.getElementById("gh-issues"),s=document.getElementById("gh-contributors");if(!e)return;fetch("https://api.github.com/repos/NAEOS-foundation/naeos").then(function(e){return e.json()}).then(function(s){s.stargazers_count!==void 0&&animateCounter(e,s.stargazers_count),s.forks_count!==void 0&&animateCounter(t,s.forks_count),s.open_issues_count!==void 0&&animateCounter(n,s.open_issues_count)}).catch(function(){e.textContent="—"}),fetch("https://api.github.com/repos/NAEOS-foundation/naeos/contributors?per_page=1&anon=true").then(function(e){var t,n=e.headers.get("Link");n&&(t=n.match(/page=(\d+)>; rel="last"/),t&&animateCounter(s,parseInt(t[1],10)))}).catch(function(){})}var searchData,fuseInstance,searchOverlay,searchModal,searchInput,searchResults,selectedIndex,playgroundSamples={yaml:`project: my-service
version: "1.0"
modules:
  - name: api-gateway
    path: ./api-gateway
    dependencies: [user-service, order-service]
  - name: user-service
    path: ./services/users
    dependencies: [database]
  - name: order-service
    path: ./services/orders
    dependencies: [user-service, payment-service]
  - name: payment-service
    path: ./services/payments
  - name: database
    path: ./infra/db
services:
  - name: api-gateway
    kind: reverse-proxy
    port: 8080
  - name: user-api
    kind: rest
    port: 9001
  - name: order-api
    kind: rest
    port: 9002
architecture:
  pattern: microservices
generation:
  languages: [go, typescript]
  output_dir: ./generated`,serverless:`project: serverless-app
version: "1.0"
modules:
  - name: auth
    path: ./functions/auth
  - name: api
    path: ./functions/api
    dependencies: [auth]
  - name: processor
    path: ./functions/processor
    dependencies: [api]
services:
  - name: auth-function
    kind: lambda
  - name: api-function
    kind: lambda
  - name: processor-function
    kind: lambda
architecture:
  pattern: serverless
deployment:
  strategy: serverless-framework
generation:
  languages: [python, typescript]`,monolith:`project: monolith-app
version: "1.0"
modules:
  - name: core
    path: ./core
  - name: web
    path: ./web
    dependencies: [core]
  - name: database
    path: ./infra/db
    dependencies: [core]
services:
  - name: web-server
    kind: http
    port: 8080
architecture:
  pattern: monolithic
deployment:
  strategy: docker-compose
generation:
  languages: [go]
  output_dir: ./cmd`,"ai-context":`project: my-genai-service
version: "1.0"
modules:
  - name: agent-orchestrator
    path: ./orchestrator
    dependencies: [llm-provider, memory-store]
  - name: llm-provider
    path: ./providers/llm
    dependencies: [vector-db]
  - name: memory-store
    path: ./stores/memory
  - name: vector-db
    path: ./infra/vector
    kind: database
    engine: qdrant
services:
  - name: api-gateway
    kind: reverse-proxy
    port: 8080
  - name: chat-api
    kind: rest
    port: 9001
  - name: streaming-ws
    kind: websocket
    port: 9002
architecture:
  pattern: microservices
ai:
  providers:
    - name: openai
      models: [gpt-4o, gpt-4o-mini]
    - name: anthropic
      models: [claude-opus-4, claude-sonnet-4]
  context:
    format: neir
    compression: semantic
    max_tokens: 128000
generation:
  languages: [go, typescript, python]
  ai_instructions: true
  output_dir: ./generated`};function initPlayground(){var e=document.getElementById("playground-input"),t=document.getElementById("playground-output");if(!e||!t)return;e.value=playgroundSamples.yaml,updatePlaygroundPreview(),e.addEventListener("input",updatePlaygroundPreview)}function switchPlayground(e,t){var n,s=document.querySelectorAll(".playground-tab");s.forEach(function(e){e.classList.remove("active")}),e.classList.add("active"),n=document.getElementById("playground-input"),n&&playgroundSamples[t]&&(n.value=playgroundSamples[t],updatePlaygroundPreview())}function updatePlaygroundPreview(){var e,t,o,i,n=document.getElementById("playground-input"),s=document.getElementById("playground-output");if(!n||!s)return;o=n.value,i=o.split(`
`).filter(function(e){return e.trim()}),e="<h4>NEIR Model Preview</h4>",e+='<div style="font-family:var(--font-mono);font-size:0.8125rem;line-height:1.8;">',t=0,i.forEach(function(n){var s,o=n.trim();o.endsWith(":")?(e+='<div class="tree-node" style="margin-left:'+t*16+'px"><span class="tree-key">'+escapeHtml(o)+"</span></div>",t++):o.startsWith("- ")?(s=o.split(": "),s.length===2?e+='<div class="tree-node" style="margin-left:'+t*16+'px"><span class="tree-key">'+escapeHtml(s[0])+':</span> <span class="tree-str">'+escapeHtml(s[1])+"</span></div>":e+='<div class="tree-node" style="margin-left:'+t*16+'px"><span class="tree-val">'+escapeHtml(o)+"</span></div>"):(s=o.split(": "),s.length===2&&(e+='<div class="tree-node" style="margin-left:'+(t-1)*16+'px"><span class="tree-key">'+escapeHtml(s[0])+':</span> <span class="tree-str">'+escapeHtml(s[1])+"</span></div>"))}),e+="</div>",s.innerHTML=e}function escapeHtml(e){var t=document.createElement("div");return t.textContent=e,t.innerHTML}function initFAQ(){var e=document.querySelectorAll(".faq-question");e.forEach(function(e){e.addEventListener("click",function(){var e=this.parentElement;e.classList.toggle("open")})})}function initCookieBanner(){var e=document.querySelector(".cookie-banner"),t=document.querySelector(".cookie-banner .btn");if(!e)return;if(localStorage.getItem("cookies-accepted"))return;setTimeout(function(){e.classList.add("show")},1e3),t&&t.addEventListener("click",function(){localStorage.setItem("cookies-accepted","true"),e.classList.remove("show")})}function initNewsletter(){var t=document.querySelector(".newsletter-form"),e=document.querySelector(".newsletter-message");if(!t||!e)return;t.addEventListener("submit",function(n){n.preventDefault();var s=t.querySelector("input").value.trim();if(!s){e.textContent="Please enter your email.";return}e.textContent="Thank you! You've been subscribed.",e.style.color="var(--color-accent)",t.querySelector("input").value="",setTimeout(function(){e.textContent=""},3e3)})}function initTheme(){var t=localStorage.getItem("theme"),n=window.matchMedia("(prefers-color-scheme: dark)").matches,e=t||(n?"dark":"dark");document.documentElement.setAttribute("data-theme",e),localStorage.setItem("theme",e)}function toggleTheme(){var e=document.documentElement,n=e.getAttribute("data-theme"),t=n==="dark"?"light":"dark";e.setAttribute("data-theme",t),localStorage.setItem("theme",t)}searchData=null,fuseInstance=null,selectedIndex=-1;function openSearch(){if(!searchOverlay||!searchModal)return;searchOverlay.classList.add("open"),searchModal.classList.add("open"),searchModal.style.display="flex",searchOverlay.style.display="block",setTimeout(function(){searchInput&&searchInput.focus()},100),typeof Fuse!="undefined"&&fuseInstance===null&&searchData&&(fuseInstance=new Fuse(searchData,{keys:["title","sections","content"],threshold:.4,includeScore:!0,includeMatches:!0}))}function closeSearch(){if(!searchOverlay||!searchModal)return;searchOverlay.classList.remove("open"),searchModal.classList.remove("open"),searchOverlay.style.display="none",searchModal.style.display="none",searchInput&&(searchInput.value=""),searchResults&&(searchResults.innerHTML=""),selectedIndex=-1}function initSearch(){if(searchOverlay=document.getElementById("search-overlay"),searchModal=document.getElementById("search-modal"),searchInput=document.getElementById("search-input"),searchResults=document.getElementById("search-results"),!searchOverlay||!searchModal||!searchInput||!searchResults)return;if(typeof Fuse=="undefined"&&!document.querySelector('script[src*="fuse.js"]')){var e=document.createElement("script");e.src="https://cdn.jsdelivr.net/npm/fuse.js@7.0.0/dist/fuse.min.js",e.onload=function(){loadSearchIndex()},document.head.appendChild(e)}else typeof Fuse!="undefined"&&loadSearchIndex();searchOverlay.addEventListener("click",function(e){e.target===searchOverlay&&closeSearch()}),searchInput.addEventListener("keydown",function(e){if(e.key==="Escape"){closeSearch();return}if(e.key==="ArrowDown"){e.preventDefault(),navigateResults(1);return}if(e.key==="ArrowUp"){e.preventDefault(),navigateResults(-1);return}if(e.key==="Enter"){e.preventDefault(),selectResult();return}}),searchInput.addEventListener("input",function(){performSearch(searchInput.value)})}function loadSearchIndex(){var e="/index.json";(document.documentElement.lang==="id"||window.location.pathname.startsWith("/id/"))&&(e="/id/index.json"),fetch(e).then(function(e){return e.json()}).then(function(e){searchData=e,typeof Fuse!="undefined"&&(fuseInstance=new Fuse(searchData,{keys:["title","sections","content"],threshold:.4,includeScore:!0,includeMatches:!0}))}).catch(function(){})}function performSearch(e){if(!searchResults)return;var t,n,s,o,a,r,c,l,d,u,h,m,i=document.querySelector(".search-hint");if(!e.trim()){i&&(i.style.display="block"),searchResults.innerHTML="",selectedIndex=-1;return}if(i&&(i.style.display="none"),s=[],fuseInstance?s=fuseInstance.search(e):searchData&&(r=e.toLowerCase(),s=searchData.filter(function(e){return e.title&&e.title.toLowerCase().indexOf(r)!==-1||e.content&&e.content.toLowerCase().indexOf(r)!==-1}).map(function(e){return{item:e}})),selectedIndex=-1,s.length===0){searchResults.innerHTML='<div class="search-hint">'+SEARCH_NO_RESULTS+"</div>";return}for(n="",u=20,o=0;o<Math.min(s.length,u);o++)a=s[o],t=a.item||a,h=t.title||"",c=t.section||"",m=t.permalink||t.url||"#",l=t.content?t.content.substring(0,120):"",n+='<a href="'+m+'" class="search-result-item" data-index="'+o+'">',n+='  <div class="result-title">'+escapeHtml(h)+"</div>",c&&(n+='  <div class="result-section">'+escapeHtml(c)+"</div>"),l&&(n+='  <div class="result-excerpt">'+escapeHtml(l)+"</div>"),n+="</a>";searchResults.innerHTML=n,d=searchResults.querySelectorAll(".search-result-item"),d.forEach(function(e){e.addEventListener("click",function(){closeSearch()}),e.addEventListener("mouseenter",function(){d.forEach(function(e){e.classList.remove("selected")}),this.classList.add("selected")})})}function navigateResults(e){var t=searchResults.querySelectorAll(".search-result-item");if(!t.length)return;t.forEach(function(e){e.classList.remove("selected")}),selectedIndex+=e,selectedIndex<0&&(selectedIndex=0),selectedIndex>=t.length&&(selectedIndex=t.length-1),t[selectedIndex].classList.add("selected"),t[selectedIndex].scrollIntoView({block:"nearest"})}function selectResult(){var e=searchResults.querySelector(".search-result-item.selected");e&&(window.location.href=e.getAttribute("href"),closeSearch())}function initKeyboardShortcuts(){document.addEventListener("keydown",function(e){(e.metaKey||e.ctrlKey)&&e.key==="k"&&(e.preventDefault(),openSearch())})}