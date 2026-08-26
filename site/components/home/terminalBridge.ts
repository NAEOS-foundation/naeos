"use client";

interface XtermLike {
  writeln(data: string): void;
  write(data: string): void;
  onKey(cb: (e: { key: string }) => void): void;
  focus(): void;
  dispose(): void;
  loadAddon(a: FitAddonLike): void;
  open(container: HTMLElement): void;
}

type FitAddonLike = { fit(): void; proposeDimensions(): { cols: number } | undefined };

function loadScript(src: string): Promise<void> {
  return new Promise((resolve, reject) => {
    if (document.querySelector(`script[src="${src}"]`)) {
      resolve();
      return;
    }
    const script = document.createElement("script");
    script.src = src;
    script.onload = () => resolve();
    script.onerror = () => reject(new Error(`failed to load ${src}`));
    document.body.appendChild(script);
  });
}

export function startInteractiveTerminal(
  container: HTMLElement,
  wsUrl: string,
  onReady?: () => void,
): () => void {
  let disposed = false;
  let term: XtermLike | null = null;
  let ws: WebSocket | null = null;
  let reconnectTimer: number | undefined;
  let inputBuffer = "";
  const history: string[] = [];
  let historyIndex = -1;

  const staticEl = document.getElementById("hero-terminal-static");

  async function boot() {
    try {
      await Promise.all([
        loadScript("https://cdn.jsdelivr.net/npm/xterm@5.3.0/lib/xterm.js"),
        new Promise<void>((resolve, reject) => {
          if (document.querySelector('link[href*="xterm@5"]')) {
            resolve();
            return;
          }
          const link = document.createElement("link");
          link.rel = "stylesheet";
          link.href = "https://cdn.jsdelivr.net/npm/xterm@5.3.0/css/xterm.min.css";
          link.onload = () => resolve();
          link.onerror = () => reject(new Error("xterm css failed"));
          document.head.appendChild(link);
        }),
      ]);
      if (disposed) return;
      const w = window as unknown as {
        Terminal: new (opts: Record<string, unknown>) => XtermLike & {
          loadAddon(a: FitAddonLike): void;
        };
        FitAddon: {
          FitAddon: new () => FitAddonLike;
        };
      };
      staticEl?.style.setProperty("display", "none");
      container.style.display = "";
      term = new w.Terminal({
        cursorBlink: true,
        fontSize: 13,
        fontFamily: "'JetBrains Mono', monospace",
        theme: {
          background: "rgba(0,0,0,0)",
          foreground: "#e8e8e8",
          cursor: "#00ff88",
        },
      });
      const fit = new w.FitAddon.FitAddon();
      term.loadAddon(fit);
      term.open(container);
      fit.fit();
      term.onKey(({ key }) => {
        if (!ws) return;
        if (key === "\r") {
          if (inputBuffer.trim()) {
            history.unshift(inputBuffer);
            if (history.length > 50) history.pop();
          }
          historyIndex = -1;
          ws.send(inputBuffer);
          inputBuffer = "";
        } else if (key === "\x7f") {
          if (inputBuffer.length > 0) {
            inputBuffer = inputBuffer.slice(0, -1);
            term?.write("\b \b");
          }
        } else if (key === "\x03") {
          ws.send("\x03");
          inputBuffer = "";
        } else if (key === "\x1b[A" || key === "\x1b[B") {
          if (history.length === 0) return;
          historyIndex =
            key === "\x1b[A"
              ? Math.min(historyIndex + 1, history.length - 1)
              : Math.max(historyIndex - 1, -1);
          while (inputBuffer.length > 0) {
            inputBuffer = inputBuffer.slice(0, -1);
            term?.write("\b \b");
          }
          inputBuffer = historyIndex >= 0 ? history[historyIndex] : "";
          term?.write(inputBuffer);
        } else if (key >= " " && key <= "~") {
          inputBuffer += key;
          term?.write(key);
        }
      });

      let reconnectDelay = 1000;

      function connect() {
        if (disposed) return;
        ws = new WebSocket(wsUrl);
        ws.onmessage = (event) => {
          reconnectDelay = 1000;
          const data = String(event.data);
          if (data.startsWith("!prompt")) term?.write("$ ");
          else if (data.startsWith("!error")) term?.writeln(`\x1b[31m${data.slice(6)}\x1b[0m`);
          else if (data.startsWith("!ready")) {
            term?.writeln(data.slice(6));
            onReady?.();
          } else term?.write(`${data}\r\n`);
        };
        ws.onclose = () => {
          if (!disposed) {
            reconnectTimer = window.setTimeout(connect, reconnectDelay);
            reconnectDelay = Math.min(reconnectDelay * 2, 30000);
          }
        };
      }
      connect();

      cleanupFns.push(() => {
        ws?.close();
        term?.dispose();
      });
    } catch {
      if (staticEl) staticEl.style.display = "";
    }
  }

  const cleanupFns: (() => void)[] = [];
  void boot();

  return () => {
    disposed = true;
    if (reconnectTimer) window.clearTimeout(reconnectTimer);
    for (const fn of cleanupFns) fn();
    if (staticEl) staticEl.style.display = "";
  };
}
