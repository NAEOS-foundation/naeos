import http from "node:http";

export function createServer(port: number) {
  const server = http.createServer((req, res) => {
    res.writeHead(200, { "Content-Type": "application/json" });
    res.end(JSON.stringify({ service: "gateway", status: "ok" }));
  });

  server.listen(port, () => {
    console.log("gateway listening on port", port);
  });

  return server;
}
