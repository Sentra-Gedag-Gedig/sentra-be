-- Seed data untuk page_mappings
-- Sesuaikan dengan routeMap dari FE

-- Home / Beranda
INSERT INTO page_mappings (page_id, url, display_name, keywords, synonyms, category, description, is_active)
VALUES (
    'home',
    '/home',
    'Beranda',
    '["beranda", "home", "utama", "halaman utama", "dashboard"]',
    '["rumah", "awal", "mulai", "main page"]',
    'main',
    'Halaman utama aplikasi Sentra',
    true
) ON CONFLICT (page_id) DO UPDATE SET
    url = EXCLUDED.url,
    display_name = EXCLUDED.display_name,
    keywords = EXCLUDED.keywords,
    synonyms = EXCLUDED.synonyms,
    updated_at = now();

-- Deteksi Uang
INSERT INTO page_mappings (page_id, url, display_name, keywords, synonyms, category, description, is_active)
VALUES (
    'money-detection',
    '/deteksi/money-detection',
    'Deteksi Uang',
    '["deteksi", "deteksi uang", "money detection", "scan uang", "cek uang"]',
    '["periksa uang", "identifikasi uang", "lihat uang", "money scanner"]',
    'detection',
    'Fitur untuk mendeteksi dan mengidentifikasi nilai uang',
    true
) ON CONFLICT (page_id) DO UPDATE SET
    url = EXCLUDED.url,
    display_name = EXCLUDED.display_name,
    keywords = EXCLUDED.keywords,
    synonyms = EXCLUDED.synonyms,
    updated_at = now();

-- Literasi
INSERT INTO page_mappings (page_id, url, display_name, keywords, synonyms, category, description, is_active)
VALUES (
    'literasi',
    '/literasi',
    'Literasi',
    '["literasi", "edukasi", "belajar", "artikel", "blog"]',
    '["bacaan", "pengetahuan", "informasi", "education"]',
    'education',
    'Konten edukasi dan literasi keuangan',
    true
) ON CONFLICT (page_id) DO UPDATE SET
    url = EXCLUDED.url,
    display_name = EXCLUDED.display_name,
    keywords = EXCLUDED.keywords,
    synonyms = EXCLUDED.synonyms,
    updated_at = now();

-- Profil
INSERT INTO page_mappings (page_id, url, display_name, keywords, synonyms, category, description, is_active)
VALUES (
    'profiles',
    '/profiles',
    'Profil',
    '["profil", "profile", "akun", "account", "saya"]',
    '["pengaturan akun", "data diri", "my profile"]',
    'account',
    'Halaman profil dan pengaturan akun pengguna',
    true
) ON CONFLICT (page_id) DO UPDATE SET
    url = EXCLUDED.url,
    display_name = EXCLUDED.display_name,
    keywords = EXCLUDED.keywords,
    synonyms = EXCLUDED.synonyms,
    updated_at = now();

-- Edit Email
INSERT INTO page_mappings (page_id, url, display_name, keywords, synonyms, category, description, is_active)
VALUES (
    'edit-email',
    '/profiles/edit-email',
    'Edit Email',
    '["edit email", "ubah email", "ganti email", "change email"]',
    '["update email", "modifikasi email"]',
    'account',
    'Halaman untuk mengubah alamat email',
    true
) ON CONFLICT (page_id) DO UPDATE SET
    url = EXCLUDED.url,
    display_name = EXCLUDED.display_name,
    keywords = EXCLUDED.keywords,
    synonyms = EXCLUDED.synonyms,
    updated_at = now();

-- Edit Phone
INSERT INTO page_mappings (page_id, url, display_name, keywords, synonyms, category, description, is_active)
VALUES (
    'edit-phone',
    '/profiles/edit-phone',
    'Edit Nomor HP',
    '["edit nomor", "edit phone", "ubah nomor", "ganti nomor hp", "change phone"]',
    '["update nomor", "modifikasi nomor telepon", "edit hp"]',
    'account',
    'Halaman untuk mengubah nomor telepon',
    true
) ON CONFLICT (page_id) DO UPDATE SET
    url = EXCLUDED.url,
    display_name = EXCLUDED.display_name,
    keywords = EXCLUDED.keywords,
    synonyms = EXCLUDED.synonyms,
    updated_at = now();

-- Edit Profil
INSERT INTO page_mappings (page_id, url, display_name, keywords, synonyms, category, description, is_active)
VALUES (
    'edit-profile',
    '/profiles/edit-profile',
    'Edit Profil',
    '["edit profil", "ubah profil", "ganti profil", "edit profile"]',
    '["update profil", "modifikasi profil", "change profile"]',
    'account',
    'Halaman untuk mengubah data profil',
    true
) ON CONFLICT (page_id) DO UPDATE SET
    url = EXCLUDED.url,
    display_name = EXCLUDED.display_name,
    keywords = EXCLUDED.keywords,
    synonyms = EXCLUDED.synonyms,
    updated_at = now();

-- Verifikasi Email
INSERT INTO page_mappings (page_id, url, display_name, keywords, synonyms, category, description, is_active)
VALUES (
    'verif-email',
    '/profiles/verification-email',
    'Verifikasi Email',
    '["verifikasi email", "verif email", "verify email", "konfirmasi email"]',
    '["validasi email", "email verification"]',
    'account',
    'Halaman verifikasi alamat email',
    true
) ON CONFLICT (page_id) DO UPDATE SET
    url = EXCLUDED.url,
    display_name = EXCLUDED.display_name,
    keywords = EXCLUDED.keywords,
    synonyms = EXCLUDED.synonyms,
    updated_at = now();

-- Verifikasi HP
INSERT INTO page_mappings (page_id, url, display_name, keywords, synonyms, category, description, is_active)
VALUES (
    'verif-hp',
    '/profiles/verification-hp',
    'Verifikasi Nomor HP',
    '["verifikasi hp", "verif nomor", "verify phone", "konfirmasi nomor"]',
    '["validasi hp", "phone verification", "verifikasi telepon"]',
    'account',
    'Halaman verifikasi nomor telepon',
    true
) ON CONFLICT (page_id) DO UPDATE SET
    url = EXCLUDED.url,
    display_name = EXCLUDED.display_name,
    keywords = EXCLUDED.keywords,
    synonyms = EXCLUDED.synonyms,
    updated_at = now();

-- QR Code
INSERT INTO page_mappings (page_id, url, display_name, keywords, synonyms, category, description, is_active)
VALUES (
    'qr',
    '/qr',
    'QR Code',
    '["qr", "qris", "kode qr", "qr code"]',
    '["quick response", "barcode"]',
    'payment',
    'Menu QR Code untuk pembayaran',
    true
) ON CONFLICT (page_id) DO UPDATE SET
    url = EXCLUDED.url,
    display_name = EXCLUDED.display_name,
    keywords = EXCLUDED.keywords,
    synonyms = EXCLUDED.synonyms,
    updated_at = now();

-- Scan QR
INSERT INTO page_mappings (page_id, url, display_name, keywords, synonyms, category, description, is_active)
VALUES (
    'scan-qr',
    '/qrCode/scan-qr',
    'Scan QR',
    '["scan qr", "pindai qr", "scan kode", "baca qr"]',
    '["scan barcode", "pindai kode qr", "qr scanner"]',
    'payment',
    'Halaman untuk memindai kode QR',
    true
) ON CONFLICT (page_id) DO UPDATE SET
    url = EXCLUDED.url,
    display_name = EXCLUDED.display_name,
    keywords = EXCLUDED.keywords,
    synonyms = EXCLUDED.synonyms,
    updated_at = now();

-- Show QR
INSERT INTO page_mappings (page_id, url, display_name, keywords, synonyms, category, description, is_active)
VALUES (
    'show-qr',
    '/qrCode/show-qr',
    'Tampilkan QR',
    '["tampilkan qr", "show qr", "lihat qr", "kode qr saya"]',
    '["my qr", "qr code saya", "display qr"]',
    'payment',
    'Halaman menampilkan kode QR untuk menerima pembayaran',
    true
) ON CONFLICT (page_id) DO UPDATE SET
    url = EXCLUDED.url,
    display_name = EXCLUDED.display_name,
    keywords = EXCLUDED.keywords,
    synonyms = EXCLUDED.synonyms,
    updated_at = now();

-- PIN QR
INSERT INTO page_mappings (page_id, url, display_name, keywords, synonyms, category, description, is_active)
VALUES (
    'pin-qr',
    '/qrCode/pin-qr',
    'PIN QR',
    '["pin qr", "password qr", "keamanan qr"]',
    '["security qr", "qr password"]',
    'payment',
    'Pengaturan PIN untuk transaksi QR',
    true
) ON CONFLICT (page_id) DO UPDATE SET
    url = EXCLUDED.url,
    display_name = EXCLUDED.display_name,
    keywords = EXCLUDED.keywords,
    synonyms = EXCLUDED.synonyms,
    updated_at = now();

-- Sentra Pay
INSERT INTO page_mappings (page_id, url, display_name, keywords, synonyms, category, description, is_active)
VALUES (
    'sentra-pay',
    '/sentra-pay',
    'Sentra Pay',
    '["sentra pay", "sentrapay", "dompet", "wallet", "saldo"]',
    '["dompet digital", "e-wallet", "uang elektronik"]',
    'payment',
    'Dompet digital Sentra Pay',
    true
) ON CONFLICT (page_id) DO UPDATE SET
    url = EXCLUDED.url,
    display_name = EXCLUDED.display_name,
    keywords = EXCLUDED.keywords,
    synonyms = EXCLUDED.synonyms,
    updated_at = now();

-- Notifikasi
INSERT INTO page_mappings (page_id, url, display_name, keywords, synonyms, category, description, is_active)
VALUES (
    'notification',
    '/notification',
    'Notifikasi',
    '["notifikasi", "notification", "pemberitahuan", "pesan"]',
    '["peringatan", "alert", "inbox"]',
    'communication',
    'Halaman notifikasi dan pemberitahuan',
    true
) ON CONFLICT (page_id) DO UPDATE SET
    url = EXCLUDED.url,
    display_name = EXCLUDED.display_name,
    keywords = EXCLUDED.keywords,
    synonyms = EXCLUDED.synonyms,
    updated_at = now();

-- Update timestamps untuk semua records
UPDATE page_mappings SET updated_at = now() WHERE page_id IN (
    'home', 'money-detection', 'literasi', 'profiles', 'edit-email', 
    'edit-phone', 'edit-profile', 'verif-email', 'verif-hp', 
    'qr', 'scan-qr', 'show-qr', 'pin-qr', 'sentra-pay', 'notification'
);