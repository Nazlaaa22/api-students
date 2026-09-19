-- 1. Tambahkan owner_id ke tabel students
--    Awalnya boleh NULL agar data lama tidak langsung gagal.
ALTER TABLE students
ADD COLUMN IF NOT EXISTS owner_id INTEGER;

-- 2. Isi owner_id untuk data students lama.
--    Data lama diberikan kepada user pertama yang tersedia.
UPDATE students
SET owner_id = (
    SELECT id
    FROM users
    ORDER BY id
    LIMIT 1
)
WHERE owner_id IS NULL;

-- 3. Tambahkan foreign key ke users.
ALTER TABLE students
ADD CONSTRAINT students_owner_id_fkey
FOREIGN KEY (owner_id)
REFERENCES users(id);

-- 4. Setelah seluruh data lama memiliki owner,
--    owner_id wajib diisi.
ALTER TABLE students
ALTER COLUMN owner_id SET NOT NULL;


-- ============================================
-- 5. Tabel permissions
-- ============================================

CREATE TABLE IF NOT EXISTS permissions (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE
);


-- ============================================
-- 6. Relasi role dengan permission
-- ============================================

CREATE TABLE IF NOT EXISTS role_permissions (
    role VARCHAR(20) NOT NULL,
    permission_id INTEGER NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role, permission_id)
);


-- ============================================
-- 7. Masukkan daftar permission
-- ============================================

INSERT INTO permissions (name)
VALUES
    ('student:list'),
    ('student:read:any'),
    ('student:create'),
    ('student:update:any'),
    ('student:delete')
ON CONFLICT (name) DO NOTHING;


-- ============================================
-- 8. Permission untuk ADMIN
-- ============================================

INSERT INTO role_permissions (role, permission_id)
SELECT 'admin', id
FROM permissions
WHERE name IN (
    'student:list',
    'student:read:any',
    'student:create',
    'student:update:any',
    'student:delete'
)
ON CONFLICT DO NOTHING;


-- ============================================
-- 9. Permission untuk STAFF
-- ============================================

INSERT INTO role_permissions (role, permission_id)
SELECT 'staff', id
FROM permissions
WHERE name IN (
    'student:list',
    'student:read:any',
    'student:create'
)
ON CONFLICT DO NOTHING;


-- ============================================
-- 10. User biasa tidak mendapatkan permission
--     :any / list secara otomatis.
--     User hanya boleh mengakses data miliknya
--     melalui pemeriksaan owner_id di service.
-- ============================================