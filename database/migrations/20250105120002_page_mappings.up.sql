CREATE TABLE IF NOT EXISTS page_mappings (
    page_id VARCHAR(100) PRIMARY KEY,
    url VARCHAR(500) NOT NULL,
    display_name VARCHAR(255) NOT NULL,
    keywords JSONB NOT NULL,
    synonyms JSONB,
    category VARCHAR(100),
    description TEXT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    updated_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE INDEX idx_page_mappings_category ON page_mappings(category);
CREATE INDEX idx_page_mappings_is_active ON page_mappings(is_active);

-- Insert default page mappings
INSERT INTO page_mappings (page_id, url, display_name, keywords, synonyms, category, description, is_active, created_at, updated_at) VALUES
('home', '/', 'Beranda', '["beranda", "home", "utama", "awal", "depan"]', '["halaman utama", "halaman awal", "halaman depan", "menu utama"]', 'navigation', 'Halaman utama aplikasi', true, now(), now()),
('profile', '/profile', 'Profil', '["profil", "profile", "akun", "account"]', '["profil saya", "akun saya", "data diri", "informasi pribadi"]', 'user', 'Halaman profil pengguna', true, now(), now()),
('transaction_history', '/transactions', 'Riwayat Transaksi', '["riwayat", "transaksi", "history", "transaction"]', '["riwayat transaksi", "historis transaksi", "sejarah transaksi", "history transaksi", "catatan transaksi"]', 'finance', 'Riwayat semua transaksi keuangan', true, now(), now()),
('wallet', '/wallet', 'Dompet', '["dompet", "wallet", "saldo", "balance"]', '["dompet digital", "e-wallet", "saldo dompet", "balance dompet"]', 'finance', 'Dompet digital dan saldo', true, now(), now()),
('settings', '/settings', 'Pengaturan', '["pengaturan", "settings", "konfigurasi", "config"]', '["pengaturan aplikasi", "setelan", "konfigurasi aplikasi"]', 'system', 'Pengaturan aplikasi', true, now(), now()),
('notifications', '/notifications', 'Notifikasi', '["notifikasi", "notifications", "pemberitahuan", "alert"]', '["pemberitahuan", "pesan masuk", "inbox", "peringatan"]', 'communication', 'Daftar notifikasi dan pemberitahuan', true, now(), now()),
('help', '/help', 'Bantuan', '["bantuan", "help", "panduan", "guide"]', '["pusat bantuan", "help center", "tutorial", "cara menggunakan"]', 'support', 'Pusat bantuan dan panduan', true, now(), now());