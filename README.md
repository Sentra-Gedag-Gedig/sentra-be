# Sentra 💰💳
![alt text](https://github.com/user-attachments/assets/8ba32535-0d1f-4a96-a0ae-27a0ce0dd57a)

Sentra adalah platform manajemen keuangan komprehensif yang dibangun dengan Go yang menyediakan otentikasi pengguna, manajemen anggaran, verifikasi identitas, dan kemampuan dompet digital.

## Tim Pengembang ✨

### Nama Tim: Gedag Gedig

#### Anggota:
- Richard (virgobulan05@student.ub.ac.id)
- Jason Surya Wijaya (jasonsurya17@student.ub.ac.id)
- Kadek Nandana Tyo Nayotama (nandanatyon@student.ub.ac.id)

## Fitur ✨

### Otentikasi Pengguna 🔐
- Pendaftaran dengan verifikasi email/telepon
- Otentikasi multi-faktor
- Login sosial (Google OAuth)
- Otentikasi biometrik (Touch ID)
- Pengelolaan profil pengguna

### Manajemen Anggaran 📊
- Pelacakan pemasukan dan pengeluaran
- Kategorisasi transaksi
- Laporan keuangan berbasis periode
- Catatan audio untuk transaksi
- Kategori yang dapat disesuaikan

### Verifikasi Identitas 🔐
- Deteksi KTP dan ekstraksi data
- Pengenalan wajah untuk otentikasi
- Pemindaian kode QRIS
- Pengenalan mata uang

### Dompet Digital 👛
- Pengelolaan saldo
- Pembuatan akun virtual melalui DOKU
- Pemrosesan pembayaran dengan callback
- Riwayat transaksi
- Transfer dana yang aman

## Arsitektur 🏗️

SentraPay mengikuti arsitektur yang bersih dan modular:

- **Lapisan API**: API RESTful yang dibangun dengan framework Fiber
- **Lapisan Layanan**: Implementasi logika bisnis inti
- **Lapisan Repositori**: Pola akses data untuk persistensi
- **Lapisan Infrastruktur**: Integrasi layanan eksternal

## Tech Stack 🛠️

- **Backend**: Go (Golang)
- **Framework Web**: Fiber
- **Database**: PostgreSQL
- **Caching**: Redis
- **Penyimpanan**: AWS S3
- **Perpesanan**: WhatsApp API
- **Layanan AI**: Google Gemini untuk analisis gambar
- **Payment Gateway**: DOKU API

## Prasyarat ✅

- Go 1.18+
- PostgreSQL 13+
- Redis 6+
- Kredensial AWS S3
- Akun payment gateway DOKU
- Proyek Google Cloud untuk Gemini AI
- Integrasi WhatsApp untuk notifikasi


## Instalasi 📥

1. Klon repositori:
   ```bash
   git clone [https://github.com/yourusername/sentrapay.git](https://github.com/Sentra-Gedag-Gedig/sentra-backend.git)
   cd sentrapay
   ```

2. Instal dependensi:
   ```bash
   go mod download
   ```

3. Siapkan database:
   ```bash
   # Jalankan migrasi PostgreSQL
   migrate -database "postgres://postgres:yourpassword@localhost:5432/sentrapay?sslmode=disable" -path database/migrations up
   ```

4. Build aplikasi:
   ```bash
   go build -o sentrapay ./cmd/app
   ```

5. Jalankan aplikasi:
   ```bash
   ./sentrapay
   ```

## Deployment Docker 🐳

Anda juga dapat menggunakan Docker Compose untuk menjalankan seluruh stack aplikasi:

```bash
docker-compose up -d
```

Ini akan memulai:
- Aplikasi Go utama
- Database PostgreSQL
- Cache Redis
- Layanan deteksi wajah
- Layanan deteksi KTP
- Layanan deteksi QRIS
- NGINX sebagai reverse proxy

## Dokumentasi API 📘

### Koleksi Postman

Akses dokumentasi API lengkap kami dan uji endpoint menggunakan koleksi Postman kami:

[![Run in Postman](https://run.pstmn.io/button.svg)](https://braciate-backend.postman.co/workspace/My-Workspace~3c0895d0-8f47-45ff-8232-9471b36c8289/collection/32354585-ae5b5ec5-ccbf-46a0-b4a5-1375abc5d2e4?action=share&creator=32354585&active-environment=32354585-f992d894-dc2a-4b75-8494-aefe3fa343d9)

## Struktur Proyek 📂

```
ProjectGolang/
├── cmd/app/                  # Titik masuk aplikasi
├── database/                 # Migrasi dan konfigurasi database
│   ├── migrations/           # File migrasi SQL
│   └── postgres/             # Koneksi PostgreSQL
├── internal/                 # Kode aplikasi internal
│   ├── api/                  # Handler dan layanan API
│   │   ├── auth/             # Modul otentikasi
│   │   ├── budget_manager/   # Modul manajemen anggaran
│   │   ├── detection/        # Layanan deteksi
│   │   └── sentra_pay/       # Dompet dan pembayaran
│   ├── config/               # Konfigurasi aplikasi
│   ├── entity/               # Entitas domain
│   └── middleware/           # Middleware HTTP
├── nginx/                    # Konfigurasi NGINX
├── pkg/                      # Paket bersama
│   ├── bcrypt/               # Hashing kata sandi
│   ├── context/              # Utilitas konteks
│   ├── doku/                 # Payment gateway DOKU
│   ├── gemini/               # Google Gemini AI
│   ├── google/               # Google OAuth
│   ├── handlerUtil/          # Utilitas handler
│   ├── jwt/                  # Otentikasi JWT
│   ├── log/                  # Logging
│   ├── redis/                # Klien Redis
│   ├── response/             # Utilitas respons HTTP
│   ├── s3/                   # Klien AWS S3
│   ├── smtp/                 # Pengiriman email
│   ├── utils/                # Utilitas umum
│   ├── websocket/            # Utilitas WebSocket
│   └── whatsapp/             # Perpesanan WhatsApp
└── .env                      # Variabel lingkungan
```

## Kontribusi 🤝

1. Fork repositori
2. Buat branch fitur Anda: `git checkout -b feature/fitur-saya`
3. Commit perubahan Anda: `git commit -am 'Tambahkan fitur saya'`
4. Push ke branch: `git push origin feature/fitur-saya`
5. Kirim pull request

## Lisensi 📝

Proyek ini dilisensikan di bawah Lisensi MIT - lihat file LICENSE untuk detailnya.

## Penghargaan 🙏

- Tim Go Fiber untuk framework web mereka yang luar biasa
- Semua kontributor pada pustaka open-source yang digunakan dalam proyek ini
