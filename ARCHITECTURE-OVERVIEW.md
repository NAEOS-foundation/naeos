# Architecture Overview

Dokumen ini memberikan gambaran besar arsitektur konseptual NAEOS.

## Tujuan arsitektur
NAEOS dirancang untuk menghubungkan empat lapisan utama:
1. Governance — menetapkan aturan organisasi dan proses.
2. Specification — mendefinisikan kebutuhan, desain, dan kontrak.
3. Constitution — memegang prinsip normatif yang tidak boleh dilanggar.
4. Policy Compiler — mengubah kebijakan menjadi aturan eksekusi.

## Alur konseptual
Requirement dan intent masuk ke layer specification.
Setelah itu, policy dan governance dipetakan ke aturan yang dapat divalidasi.
Output akhirnya adalah artefak implementasi, dokumentasi, dan aturan eksekusi yang konsisten.

## Komponen utama
- Governance layer
- Specification layer
- Constitution layer
- Policy layer
- Validation and compiler pipeline
- Reference implementation

## Prinsip desain
- human readable,
- machine readable,
- vendor neutral,
- extensible,
- deterministic.

## Kaitan dengan repositori
Repositori ini menyimpan dokumen-dokumen yang menjelaskan lapisan-lapisan tersebut secara terstruktur sehingga implementasi dan review dapat dilakukan dengan konsistensi yang lebih tinggi.
