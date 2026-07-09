# NAEOS Foundation — NAEOS (Specification & Reference)

NAEOS adalah kumpulan spesifikasi, konstitusi, kebijakan, dan arsitektur acuan untuk implementasi sistem engineering berbasis Knowledge Graph, Policy Compiler, dan Kernel modular. Repositori ini berisi dokumen normatif yang memandu implementasi dan konformance terhadap standar NAEOS.

## Tujuan
- Menyediakan arsitektur acuan dan kontrak kebijakan untuk implementasi NAEOS.
- Menyimpan konstitusi engineering, governance, dan spesifikasi kernel serta policy.

## Struktur utama
- Reference Architecture/ — dokumen arsitektur acuan (NAEOS-NRA-001.md)
- constitution/ — dokumen konstitusi (NAEOS-CON-00x.md)
- governance/ — kebijakan tata kelola (NAEOS-GOV-00x.md)
- kernel/ — spesifikasi kernel (NAEOS-KER-001.md)
- policy/ — kebijakan dan policy compiler (NAEOS-POL-001.md)
- profile/ — definisi profil (NAEOS-PRO-001.md)
- specification/ — spesifikasi tambahan dan indeks

## Cara menjelajah dan berkontribusi
Baca terlebih dahulu CONTRIBUTING.md untuk panduan kontribusi, gaya dokumen, dan proses review.

## Membangun dokumentasi situs (opsional)
Direkomendasikan menggunakan MkDocs (Material) untuk mempublikasikan dokumentasi ke GitHub Pages.

Contoh cepat (lokal):

```bash
pip install mkdocs mkdocs-material
mkdocs new .
# pindahkan atau link file markdown ke direktori docs/
mkdocs serve
