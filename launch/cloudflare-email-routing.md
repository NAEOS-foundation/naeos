# DNS Records — naeos.dev Email Routing

Cloudflare Email Routing DNS records. Copy-paste ke Cloudflare Dashboard → DNS → Records.

**Domain:** naeos.dev
**Email forwarding:** Cloudflare Email Routing (gratis)

---

## Email Routing Records (WAJIB)

### MX Records (menerima email)

| Type | Name | Priority | Content | Proxy |
|------|------|----------|---------|-------|
| MX | @ | 0 | `mx1.forwardemail.net` | DNS only |
| MX | @ | 0 | `mx2.forwardemail.net` | DNS only |

### TXT Records (verifikasi & SPF)

| Type | Name | Content | Kegunaan |
|------|------|---------|----------|
| TXT | @ | `forward-email=bayupriatno007@gmail.com` | Default forward (fallback) |
| TXT | @ | `v=spf1 include:_spf.google.com ~all` | SPF (kirim dari Gmail) |

### DMARC Record

| Type | Name | Content |
|------|------|---------|
| TXT | _dmarc | `v=DMARC1; p=none; rua=mailto:dmarc@naeos.dev` |

---

## Email Addresses (Cloudflare Dashboard)

Setelah MX records propagate, buat di **Email → Email Routing → Create address**:

| Email Address | Forward Ke | Kegunaan |
|---------------|-----------|----------|
| `bayu@naeos.dev` | bayupriatno007@gmail.com | Personal / founder |
| `support@naeos.dev` | bayupriatno007@gmail.com | User support |
| `hello@naeos.dev` | bayupriatno007@gmail.com | General inquiries |
| `security@naeos.dev` | bayupriatno007@gmail.com | Security reports |
| `press@naeos.dev` | bayupriatno007@gmail.com | Media/press |

---

## Gmail "Send Mail As" Setup

Agar bisa **kirim email** dari `bayu@naeos.dev` lewat Gmail:

### 1. Buat App Password

1. Buka https://myaccount.google.com/security
2. Aktifkan **2-Step Verification**
3. Buka https://myaccount.google.com/apppasswords
4. Pilih app: **Mail**, device: **Other (Custom name)**
5. Nama: `NAEOS Email`
6. Copy password yang di-generate

### 2. Tambah "Send Mail As" di Gmail

1. Gmail → ⚙️ Settings → **See all settings** → **Accounts and Import**
2. Di bagian **Send mail as**, klik **Add another email address**
3. Isi:
   - Name: `Bayu Priatno`
   - Email: `bayu@naeos.dev`
   - **Uncheck** "Treat as an alias"
4. Klik **Next Step**
5. Isi SMTP settings:
   - SMTP Server: `smtp.gmail.com`
   - Port: `587`
   - Username: `bayupriatno007@gmail.com`
   - Password: *(App Password dari step 1)*
   - Secured connection: **TLS**
6. Klik **Add Account**
7. Gmail kirim verification ke `bayu@naeos.dev` → klik link atau masukkan code

---

## DNS Records (Copy-Paste Ready)

Jika mau tambah manual di Cloudflare Dashboard:

```
MX  @    0  mx1.forwardemail.net      (DNS only)
MX  @    0  mx2.forwardemail.net      (DNS only)
TXT @       forward-email=bayupriatno007@gmail.com
TXT @       v=spf1 include:_spf.google.com ~all
TXT _dmarc  v=DMARC1; p=none; rua=mailto:dmarc@naeos.dev
```

---

## Verification Checklist

- [ ] MX records added (2 records, priority 0)
- [ ] TXT `forward-email` added
- [ ] TXT `v=spf1` added
- [ ] TXT `_dmarc` added
- [ ] Email Routing enabled di Cloudflare
- [ ] 5 email addresses created
- [ ] Destination emails verified
- [ ] Gmail App Password created
- [ ] Gmail "Send Mail As" configured
- [ ] Test: kirim email ke `bayu@naeos.dev` → terima di Gmail
- [ ] Test: kirim dari Gmail pakai `bayu@naeos.dev` → terima di recipient
