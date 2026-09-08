interface Env {}

const VERSION = "3.4.0";

const COMMANDS: Record<string, string> = {
	help: "Available commands: help, version, status, ping",
	version: `NAEOS ${VERSION} (Cloudflare Worker playground sandbox)`,
	status: "sandbox connected to wss://ws.naeos.dev/ws; all services operational",
};

function jsonMessage(type: string, payload: unknown): string {
	return JSON.stringify({ type, payload, time: new Date().toISOString() });
}

function isJSONFrame(data: string): boolean {
	const first = data.charCodeAt(0);
	return first === 0x7b || first === 0x5b;
}

function respondToLine(line: string, server: WebSocket): void {
	if (COMMANDS[line]) {
		server.send(`!ready ${COMMANDS[line]}`);
		server.send("!prompt");
		return;
	}
	server.send(`!error unknown command: '${line}'. Type 'help'.`);
}

export default {
	async fetch(request: Request, _env: Env): Promise<Response> {
		const url = new URL(request.url);

		if (request.method === "GET" && url.pathname === "/healthz") {
			return new Response(
				JSON.stringify({ ok: true, service: "naeos-ws", version: VERSION }),
				{
					headers: { "content-type": "application/json", "cache-control": "no-store" },
				},
			);
		}

		if (url.pathname === "/ws") {
			if (request.headers.get("Upgrade")?.toLowerCase() !== "websocket") {
				return new Response("Expected WebSocket upgrade", { status: 426 });
			}
			const pair = new WebSocketPair();
			const [client, server] = Object.values(pair) as [WebSocket, WebSocket];
			server.accept();
			server.send(`!ready NAEOS Playground ${VERSION}. Type 'help'.`);
			server.send("!prompt");

			server.addEventListener("message", (event) => {
				const raw = typeof event.data === "string" ? event.data : "";
				if (!raw) return;

				if (isJSONFrame(raw)) {
					try {
						const msg = JSON.parse(raw) as { type?: string };
						if (msg.type === "ping") {
							server.send(jsonMessage("pong", {}));
						}
					} catch {
						/* ignore malformed frame */
					}
					return;
				}

				respondToLine(raw.trim(), server);
			});

			return new Response(null, { status: 101, webSocket: client });
		}

		return new Response("naeos-ws: use GET /ws for WebSocket or /healthz for health", {
			status: 404,
		});
	},
} as ExportedHandler<Env>;