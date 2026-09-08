# NAEOS CLI Demo

## Jalankan demo dalam 3 langkah

### 1. Clone repository

```bash
git clone https://github.com/NAEOS-foundation/naeos.git
cd naeos
```

### 2. Build CLI

```bash
go build -o naeos ./cmd/naeos
```

> Membutuhkan Go 1.25 atau yang lebih baru.

### 3. Jalankan demo

```bash
./examples/demo-cli/run-demo.sh
```

## Alur demo

```text
spec.yaml
    |
    +--> 1. validate
    |
    +--> 2. context.md
    |
    +--> 3. generated/
```

Demo memvalidasi spesifikasi, membuat context bundle untuk AI, lalu
menghasilkan project Go dan TypeScript.

## Output

Hasil tersimpan di:

```text
examples/demo-cli/.run/
├── context.md
├── summary.md
└── generated/
    ├── README.md
    ├── go.mod
    └── package.json
```

Script juga memverifikasi file penting secara otomatis. Jumlah artifact dapat
berubah ketika generator NAEOS berkembang.

## Simpan output di lokasi lain

```bash
NAEOS_DEMO_OUTPUT_DIR=/tmp/naeos-demo \
  ./examples/demo-cli/run-demo.sh
```

## Gunakan binary lain

```bash
NAEOS_BIN=/path/to/naeos \
  ./examples/demo-cli/run-demo.sh
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

**Mulai ulang dari output bersih**

```bash
rm -rf examples/demo-cli/.run
./examples/demo-cli/run-demo.sh
```

## Pelajari lebih lanjut

- Demo source: https://github.com/NAEOS-foundation/naeos/tree/main/examples/demo-cli
- Studi kasus: https://naeos.dev/blog/first-cli-demo-case-study/
- Website: https://naeos.dev/
