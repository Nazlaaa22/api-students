# REST API Students

REST API sederhana untuk mengelola data mahasiswa menggunakan bahasa pemrograman Go.

API ini menyediakan operasi CRUD (Create, Read, Update, Delete), validasi input, pagination, pencarian, sorting, dan filtering data mahasiswa.

---

## Teknologi

- Go
- REST API
- JSON
- MySQL
- Thunder Client

---

## Struktur Project

```text
api-students/
├── go.mod
├── go.sum
├── main.go
├── handler.go
├── helper.go
├── model.go
└── README.md
````

---

## Base URL

```text
http://127.0.0.1:3000/api/v1
```

---

## Kontrak API

| Method | Endpoint        | Keterangan                          |
| ------ | --------------- | ----------------------------------- |
| GET    | `/students`     | Mengambil daftar mahasiswa          |
| GET    | `/students/:id` | Mengambil mahasiswa berdasarkan ID  |
| POST   | `/students`     | Menambahkan mahasiswa               |
| PUT    | `/students/:id` | Memperbarui seluruh data mahasiswa  |
| PATCH  | `/students/:id` | Memperbarui sebagian data mahasiswa |
| DELETE | `/students/:id` | Menghapus mahasiswa                 |

---

# 1. GET /students

Digunakan untuk mengambil daftar mahasiswa.

## Query Parameter

| Parameter   | Keterangan                                                    |
| ----------- | ------------------------------------------------------------- |
| `page`      | Nomor halaman, default 1                                      |
| `limit`     | Jumlah data per halaman, default 10                           |
| `search`    | Pencarian berdasarkan nama tanpa membedakan huruf besar/kecil |
| `sort`      | Pengurutan berdasarkan `name`, `grade`, atau `nim`            |
| `active`    | Filter mahasiswa aktif (`true` / `false`)                     |
| `min_grade` | Nilai minimum                                                 |
| `max_grade` | Nilai maksimum                                                |

## Contoh Request

### Pagination

```text
GET /api/v1/students?page=1&limit=2
```

### Search

```text
GET /api/v1/students?search=nazla
```

### Sort

```text
GET /api/v1/students?sort=-grade
```

Tanda `-` digunakan untuk mengurutkan secara descending.

### Filter Mahasiswa Aktif

```text
GET /api/v1/students?active=true
```

### Filter Nilai Minimum

```text
GET /api/v1/students?min_grade=85
```

### Filter Nilai Maksimum

```text
GET /api/v1/students?max_grade=90
```

## Contoh Response

```json
{
  "data": {
    "items": [
      {
        "id": "1",
        "nim": "001",
        "name": "Nazla",
        "grade": 90,
        "is_active": true
      },
      {
        "id": "2",
        "nim": "002",
        "name": "Nafisa",
        "grade": 85,
        "is_active": true
      }
    ],
    "meta": {
      "limit": 2,
      "page": 1,
      "total": 4,
      "total_pages": 2
    }
  }
}
```

---

# 2. GET /students/:id

Digunakan untuk mengambil data mahasiswa berdasarkan ID.

## Contoh Request

```text
GET /api/v1/students/4
```

## Response Berhasil

```json
{
  "data": {
    "id": "4",
    "nim": "005",
    "name": "Nazla",
    "grade": 88,
    "is_active": true
  },
  "message": "Data mahasiswa ditemukan",
  "success": true
}
```

## Response Jika Data Tidak Ditemukan

```json
{
  "message": "Data mahasiswa tidak ditemukan",
  "success": false
}
```

Status:

```text
404 Not Found
```

## Validasi ID

Jika ID bukan berupa angka, API akan memberikan response:

```json
{
  "message": "ID harus berupa angka",
  "success": false
}
```

Status:

```text
400 Bad Request
```

Contoh:

```text
GET /api/v1/students/abc
```

---

# 3. POST /students

Digunakan untuk menambahkan data mahasiswa baru.

## Contoh Request

```text
POST /api/v1/students
```

## Request Body

```json
{
  "nim": "006",
  "name": "Testing Delete",
  "grade": 80,
  "is_active": true
}
```

## Response Berhasil

```json
{
  "data": {
    "id": "10",
    "nim": "006",
    "name": "Testing Delete",
    "grade": 80,
    "is_active": true
  },
  "message": "Mahasiswa berhasil ditambahkan",
  "success": true
}
```

Status:

```text
201 Created
```

## Validasi NIM Duplikat

Jika NIM sudah digunakan, API memberikan response:

```json
{
  "message": "NIM sudah digunakan",
  "success": false
}
```

Status:

```text
409 Conflict
```

## Validasi Body

Jika body bukan JSON yang valid:

```json
{
  "message": "Body bukan JSON yang valid",
  "success": false
}
```

Status:

```text
400 Bad Request
```

## Validasi Content-Type

Request harus menggunakan:

```text
Content-Type: application/json
```

Jika tidak sesuai:

```json
{
  "message": "Content-Type harus application/json",
  "success": false
}
```

Status:

```text
415 Unsupported Media Type
```

## Validasi Field

Field `name` wajib diisi.

Jika field `name` kosong:

```json
{
  "message": "Field name wajib diisi",
  "success": false
}
```

Status:

```text
422 Unprocessable Entity
```

---

# 4. PUT /students/:id

Digunakan untuk memperbarui seluruh data mahasiswa berdasarkan ID.

## Contoh Request

```text
PUT /api/v1/students/4
```

## Request Body

```json
{
  "nim": "005",
  "name": "Nazla Putri",
  "grade": 95,
  "is_active": true
}
```

## Response Berhasil

```json
{
  "data": {
    "id": "4",
    "nim": "005",
    "name": "Nazla Putri",
    "grade": 95,
    "is_active": true
  },
  "message": "Data mahasiswa berhasil diperbarui",
  "success": true
}
```

Status:

```text
200 OK
```

---

# 5. PATCH /students/:id

Digunakan untuk memperbarui sebagian data mahasiswa.

## Contoh Request

```text
PATCH /api/v1/students/4
```

## Request Body

Contoh hanya mengubah nilai:

```json
{
  "grade": 98
}
```

## Response Berhasil

```json
{
  "data": {
    "id": "4",
    "nim": "005",
    "name": "Nazla Putri",
    "grade": 98,
    "is_active": true
  },
  "message": "Sebagian data mahasiswa berhasil diperbarui",
  "success": true
}
```

Status:

```text
200 OK
```

---

# 6. DELETE /students/:id

Digunakan untuk menghapus data mahasiswa berdasarkan ID.

## Contoh Request

```text
DELETE /api/v1/students/10
```

## Response Berhasil

Jika data berhasil dihapus, API memberikan:

```text
204 No Content
```

Tidak terdapat response body.

## Response Jika Data Tidak Ditemukan

```json
{
  "message": "Data mahasiswa tidak ditemukan",
  "success": false
}
```

Status:

```text
404 Not Found
```

---

# 7. Testing DELETE

Setelah data mahasiswa dengan ID `10` berhasil dihapus, dilakukan pengecekan kembali menggunakan GET.

## Request

```text
GET /api/v1/students/10
```

## Response

```json
{
  "message": "Data mahasiswa tidak ditemukan",
  "success": false
}
```

Status:

```text
404 Not Found
```

Hasil tersebut menunjukkan bahwa data dengan ID `10` telah berhasil dihapus.

---

# 8. Daftar Validasi

API memiliki beberapa validasi sebagai berikut:

| Kondisi                        | Status                     |
| ------------------------------ | -------------------------- |
| ID bukan angka                 | 400 Bad Request            |
| Body bukan JSON                | 400 Bad Request            |
| NIM sudah digunakan            | 409 Conflict               |
| Content-Type tidak sesuai      | 415 Unsupported Media Type |
| Field `name` kosong            | 422 Unprocessable Entity   |
| Data mahasiswa tidak ditemukan | 404 Not Found              |
| Data berhasil dibuat           | 201 Created                |
| Data berhasil diambil          | 200 OK                     |
| Data berhasil diperbarui       | 200 OK                     |
| Data berhasil dihapus          | 204 No Content             |

---

# 9. Fitur yang Telah Diimplementasikan

* [x] GET seluruh data mahasiswa
* [x] GET mahasiswa berdasarkan ID
* [x] POST mahasiswa baru
* [x] PUT mahasiswa
* [x] PATCH mahasiswa
* [x] DELETE mahasiswa
* [x] Pagination
* [x] Search berdasarkan nama
* [x] Sorting berdasarkan data
* [x] Filter mahasiswa aktif
* [x] Filter berdasarkan nilai minimum
* [x] Filter berdasarkan nilai maksimum
* [x] Validasi ID
* [x] Validasi JSON
* [x] Validasi Content-Type
* [x] Validasi NIM duplikat
* [x] Validasi field wajib

---

# 10. Pengujian dengan Thunder Client

Pengujian API dilakukan menggunakan Thunder Client pada Visual Studio Code.

Endpoint yang telah diuji:

```text
GET     /api/v1/students
GET     /api/v1/students/:id
POST    /api/v1/students
PUT     /api/v1/students/:id
PATCH   /api/v1/students/:id
DELETE  /api/v1/students/:id
```

Pengujian juga mencakup kondisi berhasil dan kondisi error/validasi.

---

# 11. Cara Menjalankan Project

Pastikan Go sudah terinstall.

Masuk ke folder project:

```text
cd api-students
```

Kemudian jalankan:

```text
go run .
```

Server berjalan pada:

```text
http://127.0.0.1:3000
```

API dapat diakses melalui:

```text
http://127.0.0.1:3000/api/v1/students
```

---

# 12. Kesimpulan

REST API Students berhasil dibuat menggunakan Go dengan menerapkan konsep REST API dan operasi CRUD.

API telah dilengkapi dengan fitur pagination, search, sorting, filtering, serta validasi input. Seluruh endpoint telah diuji menggunakan Thunder Client, termasuk pengujian terhadap kondisi error seperti ID tidak valid, NIM duplikat, body JSON tidak valid, field wajib, dan data yang tidak ditemukan.



WKWK iya 😭 kepanjangan.

Kalau buat **README Tugas 3**, cukup yang penting-penting aja. **Copy ini full, ganti README lama:**

````markdown
# REST API Students - Tugas 3

REST API untuk mengelola data mahasiswa menggunakan Go, Fiber, dan PostgreSQL.

## Teknologi

- Go
- Fiber v2
- PostgreSQL
- pgx/v5
- Thunder Client

## Fitur

- GET mahasiswa
- GET mahasiswa berdasarkan ID
- POST mahasiswa
- PUT mahasiswa
- PATCH mahasiswa
- DELETE mahasiswa
- Pagination
- Search nama
- Sorting
- Filter mahasiswa aktif
- Filter nilai minimum dan maksimum
- Validasi input
- Error handling

## Repository Pattern

Akses database dipisahkan menggunakan Repository Pattern.

Method repository:

```text
FindAll
FindByID
Create
Update
Delete
````

Repository juga menggunakan sentinel error:

```go
ErrNotFound
ErrDuplicate
```

## Database

Database menggunakan PostgreSQL dengan tabel:

```text
students
```

Kolom utama:

```text
id
nim
name
grade
is_active
created_at
updated_at
```

`id` menggunakan UUID dan `nim` memiliki constraint UNIQUE.

## Query Database

Fitur berikut diproses pada PostgreSQL:

* Search menggunakan `ILIKE`
* Filtering menggunakan `WHERE`
* Sorting menggunakan `ORDER BY`
* Pagination menggunakan `LIMIT` dan `OFFSET`
* Total data menggunakan `COUNT(*)`

Query menggunakan parameterized query.

## Endpoint

| Method | Endpoint        | Fungsi               |
| ------ | --------------- | -------------------- |
| GET    | `/students`     | Daftar mahasiswa     |
| GET    | `/students/:id` | Detail mahasiswa     |
| POST   | `/students`     | Tambah mahasiswa     |
| PUT    | `/students/:id` | Update seluruh data  |
| PATCH  | `/students/:id` | Update sebagian data |
| DELETE | `/students/:id` | Hapus mahasiswa      |

## Status HTTP yang Diuji

* `200 OK` - Request berhasil
* `201 Created` - Data berhasil dibuat
* `204 No Content` - Data berhasil dihapus
* `404 Not Found` - Data tidak ditemukan
* `409 Conflict` - NIM duplikat
* `500 Internal Server Error` - Database tidak dapat diakses

## Cara Menjalankan

Clone repository:

```bash
git clone https://github.com/Nazlaaa22/api-students.git
cd api-students
```

Siapkan file `.env` sesuai konfigurasi PostgreSQL.

Kemudian jalankan:

```bash
go run .
```

Server:

```text
http://127.0.0.1:3000
```

API:

```text
http://127.0.0.1:3000/api/v1/students
```

## Testing

Pengujian dilakukan menggunakan Thunder Client dengan menguji endpoint CRUD, pagination, search, sorting, filtering, serta error `404`, `409`, dan `500`.

## GitHub

[https://github.com/Nazlaaa22/api-students](https://github.com/Nazlaaa22/api-students)

