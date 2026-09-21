# NAEOS CLI Demo

## Jalankan demo kanonik

### 1. Clone repository

```bash
git clone https://github.com/NAEOS-foundation/naeos.git
cd naeos
```

### 2. Build CLI

```bash
go build -o naeos ./cmd/naeos
```

### 3. Jalankan demo

```bash
./examples/demo-cli/run-demo.sh
```

## Alur demo

```text
Specification
  ↓
Parse / Normalize / Resolve
  ↓
NEIR
  ↓
Validation
  ↓
Policy Evaluation
  ↓
AI Context
  ↓
AI Compilation (opsional jika API key tersedia)
  ↓
Execution / Generation
  ↓
Artifacts / Evidence
```

Demo ini memvalidasi `spec.yaml`, memberi tahu NEIR yang dihasilkan, menunjukkan policy rejection deterministik, membuat AI context bundle, menjalankan generator, lalu memverifikasi metadata traceability (`run_id`, `specification_hash`, `neir_hash`).

## Output

Hasilnya disimpan di:

```text
examples/demo-cli/.run/
├── inspect.json
├── validate.json
├── context.md
├── context.json
├── run.json
├── summary.md
├── generated/
│   ├── README.md
│   ├── go.mod
│   └── package.json
└── ai-compile.txt   # hanya jika NAEOS_LLM_API_KEY diatur
```

Script otomatis memverifikasi file penting. Jika metadata traceability tidak lengkap, demo gagal.

## Simpan output di lokasi lain

```bash
NAEOS_DEMO_OUTPUT_DIR=/tmp/naeos-demo ./examples/demo-cli/run-demo.sh
```

## Gunakan binary lain

```bash
NAEOS_BIN=/path/to/naeos ./examples/demo-cli/run-demo.sh
```

## Troubleshooting

**`NAEOS CLI not found`**

```bash
go build -o naeos ./cmd/naeos
```

**`Permission denied`**

```bash
chmod +x examples/demo-cli/run-demo.sh
```

**Mulai reset dari output bersih**

```bash
rm -rf examples/demo-cli/.run
./examples/demo-cli/run-demo.sh
```

**AI compilation dilewati**

```bash
export NAEOS_LLM_API_KEY=your-key
export NAEOS_LLM_PROVIDER=openai
./examples/demo-cli/run-demo.sh
```

## Pelajari lebih lanjut

- Demo source: https://github.com/NAEOS-foundation/naeos/tree/main/examples/demo-cli
- Repository: https://github.com/NAEOS-foundation/naeos
- Website: https://naeos.dev/
