---
title: Build Terdistribusi
slug: distributed-builds
description: Paralelkan eksekusi pipeline di banyak worker untuk build yang lebih cepat.
weight: 13
---

NAEOS mendukung eksekusi pipeline terdistribusi: 9 tahap pipeline dijadwalkan
sebagai task independen dan diproses secara paralel di seluruh kumpulan
worker yang dapat dikonfigurasi.

## Kapan Menggunakannya

Mode terdistribusi membantu ketika:

- Spesifikasi memiliki banyak modul atau layanan yang dapat diproses secara
  independen
- Anda ingin mengurangi waktu build end-to-end untuk spesifikasi besar
- Anda sedang melakukan benchmark throughput tahap pipeline

Untuk spesifikasi kecil, pipeline in-process lebih cepat — biaya distribusi
hanya terbayar setelah ada cukup pekerjaan untuk diparalelkan.

## Penggunaan

```bash
# Jalankan pipeline dengan 4 worker (default)
naeos distributed --input spec.yaml

# Gunakan lebih banyak worker untuk spesifikasi besar
naeos distributed --config naeos.yaml --workers 8
```

| Flag | Deskripsi | Default |
|------|-----------|---------|
| `--input` | Path ke file spesifikasi | — |
| `--config` | Path ke file konfigurasi NAEOS | — |
| `-w, --workers` | Jumlah worker paralel | `4` |

## Cara Kerja

1. Konfigurasi pipeline dimuat (resolusi konfigurasi sama seperti
   `naeos run`: `naeos.yaml`, `naeos.yml`, `naeos.json`, atau
   `.naeos/config.yaml`).
2. Setiap tahap pipeline (`parse`, `normalize`, `resolve`, `build-neir`,
   `validate`, `schedule`, `generate`, `review`) menjadi sebuah task.
3. Task dikirim ke koordinator yang menyeimbangkan beban ke seluruh
   kumpulan worker menggunakan distribusi round-robin.
4. Hasil diagregasi; task yang gagal dilaporkan per worker.

## Output

Perintah mencetak ringkasan dan setiap task yang gagal:

```text
Distributed pipeline: 4 workers, 8 tasks
Results: completed: 8, failed: 0
Pipeline: my-project
```

## Terkait

- [Pipeline Engine](/id/docs/pipeline-engine/) — pipeline DAG 9-tahap
- [Dashboard](/id/docs/dashboard/) — pantau aktivitas pipeline secara real time
