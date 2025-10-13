


INSERT INTO page_mappings (page_id, url, display_name, keywords, synonyms, category, description, is_active)
VALUES (
    'contact-person',
    '/contact-person',
    'Contact Person',
    '["contact person", "kontak darurat", "emergency contact", "kontak", "cp"]',
    '["kontak emergency", "telepon darurat", "contact emergency", "kontak penting", "nomor darurat"]',
    'account',
    'Pengaturan kontak darurat atau contact person',
    true
) ON CONFLICT (page_id) DO UPDATE SET
    url = EXCLUDED.url,
    display_name = EXCLUDED.display_name,
    keywords = EXCLUDED.keywords,
    synonyms = EXCLUDED.synonyms,
    updated_at = now();


INSERT INTO page_mappings (page_id, url, display_name, keywords, synonyms, category, description, is_active)
VALUES (
    'logout',
    '/logout',
    'Logout',
    '["logout", "keluar", "sign out", "log out", "cabut"]',
    '["exit", "keluar akun", "keluar aplikasi", "sign off", "log off"]',
    'account',
    'Keluar dari aplikasi (perlu konfirmasi)',
    true
) ON CONFLICT (page_id) DO UPDATE SET
    url = EXCLUDED.url,
    display_name = EXCLUDED.display_name,
    keywords = EXCLUDED.keywords,
    synonyms = EXCLUDED.synonyms,
    updated_at = now();


UPDATE page_mappings SET updated_at = now() 
WHERE page_id IN ('contact-person', 'logout');