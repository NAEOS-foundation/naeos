---
title: Pertanyaan yang Sering Diajukan
description: Pertanyaan umum tentang NAEOS, engineering control plane untuk AI coding agents, evidence, dan Golden Path.
---

<div class="faq-list">
<div class="faq-item"><button class="faq-question"><span>Apa itu NAEOS?</span><span class="faq-arrow">▾</span></button><div class="faq-answer"><p>NAEOS (Nusantara AI Engineering Operating System) adalah engineering control plane open-source untuk AI coding agents. NAEOS memberikan batas eksplisit antara intent engineering, specification, policy, execution terotorisasi, observation, evidence, dan independent verification.</p></div></div>

<div class="faq-item"><button class="faq-question"><span>Masalah apa yang diselesaikan NAEOS?</span><span class="faq-arrow">▾</span></button><div class="faq-answer"><p>AI coding agents dapat menghasilkan dan mengusulkan perubahan dengan cepat, tetapi tim engineering tetap perlu mengendalikan apa yang boleh dilakukan agent dan memiliki bukti tentang apa yang benar-benar terjadi. NAEOS menyediakan control boundary yang terstruktur untuk workflow tersebut.</p></div></div>

<div class="faq-item"><button class="faq-question"><span>Apa model control NAEOS?</span><span class="faq-arrow">▾</span></button><div class="faq-answer"><p>Model saat ini adalah: Specification → NEIR → Validation + Policy → Agent Intent → Authorized Execution → Observation → Evidence → Independent Verification. Batas utamanya adalah agent tidak memberikan authority kepada dirinya sendiri.</p></div></div>

<div class="faq-item"><button class="faq-question"><span>Apa itu Golden Path?</span><span class="faq-arrow">▾</span></button><div class="faq-answer"><p>Golden Path adalah jalur proof developer yang reproducible untuk workflow control plane. Jalur adopsi saat ini adalah Control Plane → Golden Path → Reference Demo → Evidence → Independent Verification. P1.6–P1.10 menjadi urutan proof utama.</p></div></div>

<div class="faq-item"><button class="faq-question"><span>Apa yang ditambahkan P1.11?</span><span class="faq-arrow">▾</span></button><div class="faq-answer"><p>P1.11 menambahkan independent verifier CLI untuk EvidenceBundle yang telah diserialisasi. Verifier memeriksa binding identitas decision/execution dan integritas evidence tanpa mengevaluasi policy, menjalankan action, atau membutuhkan control plane live.</p></div></div>

<div class="faq-item"><button class="faq-question"><span>Apakah NAEOS sudah production-ready?</span><span class="faq-arrow">▾</span></button><div class="faq-answer"><p>NAEOS memiliki banyak capability engineering dan governance, tetapi proof publik dan adoption sengaja diperkuat terlebih dahulu sebelum komersialisasi yang lebih luas. Evaluasi scope, Golden Path, dan evidence repository untuk kebutuhan Anda.</p></div></div>

<div class="faq-item"><button class="faq-question"><span>Apakah NAEOS open source?</span><span class="faq-arrow">▾</span></button><div class="faq-answer"><p>Ya. NAEOS menggunakan Apache License 2.0. Kontribusi dikelola melalui proses contribution berbasis DCO di repository.</p></div></div>

<div class="faq-item"><button class="faq-question"><span>Bagaimana cara mengevaluasi NAEOS?</span><span class="faq-arrow">▾</span></button><div class="faq-answer"><p>Mulai dari Control Plane publik, lalu jalankan Golden Path dan Reference Demo dari fresh checkout. Periksa evidence yang dihasilkan dan verifikasi secara independen. Ini adalah jalur evaluasi teknis yang direkomendasikan.</p></div></div>
</div>