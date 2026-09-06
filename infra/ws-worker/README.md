# WebSocket Worker for `ws.naeos.dev`

Worker Cloudflare yang melayani `wss://ws.naeos.dev/ws`. Ini endpoint "Realtime"
yang direferensikan oleh:

- `site/lib/site.ts` → `SITE.websocketUrl`
- `site/data/status.json` / `.github/workflows/status.yml` → probe Realtime
- Hero terminal interaktif di homepage (`site/components/home/terminalBridge.ts`)

Sebelumnya `ws.naeos.dev` tidak punya record DNS sama sekali, sehingga kartu
"Realtime" pada halaman status menampilkan **Down** dan hero terminal tidak
punya backend.

## Struktur

```
infra/ws-worker/
├── src/index.ts          # fetch handler + WebSocket protocol
├── wrangler.jsonc        # nama worker, route custom domain ws.naeos.dev
├── package.json          # wrangler + workers-types (dev)
├── tsconfig.json
└── README.md
```

## Deploy

Worker di-deploy otomatis oleh `.github/workflows/ws-deploy.yml`
(push ke `main` untuk perubahan `infra/ws-worker/**`, atau manual via
workflow_dispatch). Deploy memakai `CLOUDFLARE_API_TOKEN` /
`CLOUDFLARE_ACCOUNT_ID` yang sama dengan workflow website, jadi tidak ada
secret baru.

Deploy manual:

```sh
cd infra/ws-worker
npx wrangler deploy
```

### DNS

Routes `ws.naeos.dev` menggunakan `custom_domain: true`, sehingga wrangler akan
membuat rute Cloudflare secara otomatis bila zonanya ada di akun yang sama
dengan token. Verifikasi setelah deploy:

```sh
curl https://ws.naeos.dev/healthz
# {"ok":true,"service":"naeos-ws","version":"3.4.0"}
```

## Protokol

### Terminal interaktif (hero homepage)

Client mengirim perintah sebagai teks baris, server membalas dengan prefix:

| Prefix   | Makna                              |
|----------|------------------------------------|
| `!ready` | Info/banner (client memanggil `onReady`) |
| `!prompt`| Tampilkan prompt `$ `              |
| `!error` | Pesan error (ditampilkan merah)    |

Contoh dialog:

```text
> (connect)
S: !ready NAEOS Playground 3.4.0. Type 'help'.
S: !prompt
> help
S: !ready Available commands: help, version, status, ping
S: !prompt
> version
S: !ready NAEOS 3.4.0 (Cloudflare Worker playground sandbox)
S: !prompt
```

**Catatan penting:** worker ini adalah *sandbox tanpa eksekusi*. Perintah selain
whitelist (`help`, `version`, `status`) akan dibalas `!error`. Worker tidak
menjalankan pipeline/compile karena itu membutuhkan backend Go. Untuk eksekusi
nyata, jalankan image `ghcr.io/naeos-foundation/naeos` (lihat
`.goreleaser.yaml` / `Dockerfile`) dengan `naeos dashboard --port 3000` di host
yang selalu aktif (Fly.io / Cloud Run / VPS) dan arahkan `ws.naeos.dev` ke host
tersebut (CNAME proxied) — atau perluas worker ini untuk memanggil API backend.

### JSON (client non-terminal)

Client boleh mengirim frame JSON:

```json
{"type":"ping"}
```

Dibalas:

```json
{"type":"pong","payload":{},"time":"..."}
```

## Uji lokal

```sh
cd infra/ws-worker
npm install
npm run typecheck
npm run dev            # wrangler dev → http://localhost:8787
```

Uji dengan `websocat` (atau `wscat`):

```sh
# JSON ping/pong
echo '{"type":"ping"}' | websocat ws://localhost:8787/ws
# terminal protocol
printf 'help\n' | websocat ws://localhost:8787/ws
```

## Mengaktifkan hero terminal di website

`site/app/[lang]/layout.tsx` saat ini menetapkan `data-ws-url="disabled"` yang
membuat hero menampilkan terminal statis. Setelah worker ini hidup, ubah menjadi
`data-ws-url={SITE.websocketUrl}` untuk mengaktifkan terminal interaktif yang
terhubung ke `wss://ws.naeos.dev/ws`.