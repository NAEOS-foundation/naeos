(function () {
  'use strict';

  var wsUrl = window.NAEOS_WS_URL || '';
  var container = document.getElementById('interactive-terminal');
  if (!container || !wsUrl) return;

  var term, fitAddon, ws, reconnectTimer;
  var inputBuffer = '';
  var history = [];
  var historyIdx = -1;

  function initTerminal() {
    if (typeof Terminal === 'undefined') {
      loadXterm();
      return;
    }

    document.getElementById('hero-terminal-static').style.display = 'none';

    term = new Terminal({
      cursorBlink: true,
      cursorStyle: 'block',
      fontSize: 14,
      fontFamily: "'JetBrains Mono', 'Cascadia Code', 'Fira Code', monospace",
      theme: {
        background: 'transparent',
        foreground: '#e0e0e0',
        cursor: '#00ff88',
        cursorAccent: '#0a0a0a',
        selectionBackground: 'rgba(0,255,136,0.3)',
        black: '#1a1a1a',
        red: '#ff4444',
        green: '#00ff88',
        yellow: '#ffaa00',
        blue: '#569cd6',
        magenta: '#c586c0',
        cyan: '#4ec9b0',
        white: '#e0e0e0',
        brightBlack: '#666666',
        brightRed: '#ff6666',
        brightGreen: '#66ffaa',
        brightYellow: '#ffbb33',
        brightBlue: '#77b3e6',
        brightMagenta: '#d499d4',
        brightCyan: '#77d4c0',
        brightWhite: '#ffffff',
      },
      allowTransparency: true,
      convertEol: true,
    });

    fitAddon = new FitAddon.FitAddon();
    term.loadAddon(fitAddon);

    term.open(container);
    fitAddon.fit();

    window.addEventListener('resize', function () {
      if (fitAddon) fitAddon.fit();
    });

    term.writeln('\x1b[32m▸ NAEOS Interactive Demo\x1b[0m');
    term.writeln('\x1b[90m  Type \x1b[33mnaeos init\x1b[0m\x1b[90m to get started\x1b[0m');
    term.writeln('');

    setTimeout(connectWS, 500);
  }

  function loadXterm() {
    var link = document.createElement('link');
    link.rel = 'stylesheet';
    link.href = 'https://cdn.jsdelivr.net/npm/xterm@5/css/xterm.min.css';
    document.head.appendChild(link);

    var script = document.createElement('script');
    script.src = 'https://cdn.jsdelivr.net/npm/xterm@5/lib/xterm.min.js';
    script.onload = function () {
      var fitScript = document.createElement('script');
      fitScript.src = 'https://cdn.jsdelivr.net/npm/xterm-addon-fit@0.8/lib/xterm-addon-fit.min.js';
      fitScript.onload = initTerminal;
      document.body.appendChild(fitScript);
    };
    document.body.appendChild(script);
  }

  function connectWS() {
    if (ws && ws.readyState === WebSocket.OPEN) return;

    term.writeln('\x1b[90mConnecting to demo server...\x1b[0m');

    try {
      ws = new WebSocket(wsUrl);
    } catch (e) {
      term.writeln('\x1b[31mFailed to connect: ' + e.message + '\x1b[0m');
      return;
    }

    ws.onopen = function () {
      term.writeln('\x1b[32mConnected ✓\x1b[0m');
      if (reconnectTimer) {
        clearTimeout(reconnectTimer);
        reconnectTimer = null;
      }
      term.focus();
    };

    ws.onmessage = function (e) {
      var data = e.data;
      if (data.startsWith('!prompt')) {
        writePrompt();
      } else if (data.startsWith('!error ')) {
        term.writeln('\x1b[31m' + data.substring(7) + '\x1b[0m');
      } else if (data.startsWith('!ready ')) {
        term.writeln('\x1b[90m' + data.substring(7) + '\x1b[0m');
      } else {
        term.write(data);
      }
    };

    ws.onclose = function () {
      if (term) {
        term.writeln('');
        term.writeln('\x1b[31mDisconnected from demo server\x1b[0m');
        term.writeln('\x1b[90mReconnecting in 5s...\x1b[0m');
      }
      if (!reconnectTimer) {
        reconnectTimer = setTimeout(function () {
          reconnectTimer = null;
          connectWS();
        }, 5000);
      }
    };

    ws.onerror = function () {
      if (term) term.writeln('\x1b[31mConnection error\x1b[0m');
    };
  }

  function writePrompt() {
    term.write('\r\n\x1b[32m$\x1b[0m ');
    inputBuffer = '';
    historyIdx = -1;
  }

  term.onKey(function (e) {
    var ev = e.domEvent;

    if (ev.ctrlKey && ev.key === 'c') {
      ws.send('\x03');
      term.write('^C');
      writePrompt();
      return;
    }

    if (ev.key === 'Enter') {
      if (inputBuffer.trim()) {
        history.push(inputBuffer.trim());
        if (history.length > 50) history.shift();
      }
      term.write('\r\n');
      ws.send(inputBuffer);
      inputBuffer = '';
      return;
    }

    if (ev.key === 'Backspace') {
      if (inputBuffer.length > 0) {
        inputBuffer = inputBuffer.slice(0, -1);
        term.write('\b \b');
      }
      return;
    }

    if (ev.key === 'ArrowUp') {
      if (history.length > 0) {
        var newIdx = historyIdx < 0 ? history.length - 1 : Math.max(0, historyIdx - 1);
        if (newIdx !== historyIdx) {
          while (inputBuffer.length > 0) {
            term.write('\b \b');
            inputBuffer = inputBuffer.slice(0, -1);
          }
          inputBuffer = history[newIdx];
          term.write(inputBuffer);
          historyIdx = newIdx;
        }
      }
      return;
    }

    if (ev.key === 'ArrowDown') {
      if (historyIdx >= 0) {
        while (inputBuffer.length > 0) {
          term.write('\b \b');
          inputBuffer = inputBuffer.slice(0, -1);
        }
        if (historyIdx < history.length - 1) {
          historyIdx++;
          inputBuffer = history[historyIdx];
          term.write(inputBuffer);
        } else {
          historyIdx = -1;
        }
      }
      return;
    }

    if (ev.key === 'Tab') {
      ev.preventDefault();
      return;
    }

    if (e.key.length === 1 && !ev.ctrlKey && !ev.altKey && !ev.metaKey) {
      inputBuffer += e.key;
      term.write(e.key);
    }
  });

  if (typeof Terminal !== 'undefined') {
    initTerminal();
  }
})();
