# Dashboard API Documentation

## Overview
API endpoints untuk dashboard monitoring pengumpulan data statistik BPS 2026.

## Authentication
Semua endpoint memerlukan authentication token. Token dikirim melalui header `Authorization` dengan format `Bearer <token>`.

## Endpoints

### 1. GET `/api/v1/dashboard/statistics`

Mengembalikan statistik dasar dashboard untuk pengumpulan data.

**Access:** Admin dan Operator (filtered by organization)

**Response:**
```json
{
  "data": {
    "total_tables": 120,
    "total_table_draft": 10,
    "total_table_submitted": 20,
    "total_table_finalized": 90
  },
  "message": "Dashboard statistics fetched successfully"
}
```

**Response Fields:**
- `total_tables`: Total jumlah tabel yang ditugaskan
- `total_table_draft`: Jumlah tabel dengan status draft
- `total_table_submitted`: Jumlah tabel yang sudah disubmit untuk review
- `total_table_finalized`: Jumlah tabel yang sudah final

---

### 2. GET `/api/v1/dashboard/organization-completion`

Mengembalikan data keterisian per organisasi.

**Access:** Admin only

**Response:**
```json
{
  "data": [
    {
      "name": "Dinas Kesehatan",
      "completion": 95.5,
      "tables": 12
    },
    {
      "name": "Dinas Pendidikan",
      "completion": 88.0,
      "tables": 10
    }
  ],
  "message": "Organization completion data fetched successfully"
}
```

**Response Fields:**
- `name`: Nama organisasi
- `completion`: Persentase keterisian (0-100)
- `tables`: Jumlah tabel yang ditugaskan

**Notes:**
- Data diurutkan berdasarkan completion descending (tertinggi dulu)
- Completion dihitung dari (finalized + submitted) / total tables

---

### 3. GET `/api/v1/dashboard/top-performers`

Mengembalikan daftar top 3 organisasi tercepat dalam pengisian data.

**Access:** Admin only

**Response:**
```json
{
  "data": [
    {
      "name": "Dinas Kesehatan",
      "avg_time": "2.3 days",
      "completion": 95.5,
      "rank": 1
    },
    {
      "name": "Dinas Pendidikan",
      "avg_time": "3.1 days",
      "completion": 88.0,
      "rank": 2
    },
    {
      "name": "Dinas Pekerjaan Umum",
      "avg_time": "4.2 days",
      "completion": 76.0,
      "rank": 3
    }
  ],
  "message": "Top performers data fetched successfully"
}
```

**Response Fields:**
- `name`: Nama organisasi
- `avg_time`: Rata-rata waktu pengisian per tabel (format: "X.X days")
- `completion`: Persentase keterisian (0-100)
- `rank`: Peringkat (1-3)

**Notes:**
- Maksimal 3 organisasi terbaik
- Waktu dihitung dari **collection start date (12 Jan 2026)** sampai updated_at untuk tabel yang finalized/submitted
- Diurutkan berdasarkan completion percentage

---

### 4. GET `/api/v1/dashboard/organizations-need-attention`

Mengembalikan daftar organisasi yang membutuhkan perhatian khusus.

**Access:** Admin only

**Response:**
```json
{
  "data": [
    {
      "name": "Dinas Pertanian",
      "completion": 15.0,
      "tables": 6,
      "status": "Kritis",
      "days_idle": 10
    },
    {
      "name": "Dinas Pariwisata",
      "completion": 22.0,
      "tables": 3,
      "status": "Kritis",
      "days_idle": 7
    },
    {
      "name": "Dinas Sosial",
      "completion": 45.0,
      "tables": 5,
      "status": "Rendah",
      "days_idle": 3
    }
  ],
  "message": "Organizations need attention data fetched successfully"
}
```

**Response Fields:**
- `name`: Nama organisasi
- `completion`: Persentase keterisian (0-100)
- `tables`: Jumlah tabel yang ditugaskan
- `status`: Status prioritas - "Kritis" atau "Rendah"
- `days_idle`: Jumlah hari sejak update terakhir

**Notes:**
- Hanya menampilkan organisasi dengan completion < 50%
- Status "Kritis" jika completion < 30% ATAU idle > 5 hari
- Status "Rendah" untuk kondisi lainnya
- Diurutkan berdasarkan completion ascending (terendah dulu)
- Days idle dihitung dari updated_at tabel terakhir ke current time (atau dari collection start date jika tidak ada aktivitas sejak collection dimulai)

---

## Error Responses

### 403 Forbidden
```json
{
  "data": null,
  "message": "Unauthorized access"
}
```

### 500 Internal Server Error
```json
{
  "data": null,
  "message": "Error message details"
}
```

## Implementation Notes

1. **Collection Period Constants**:
   - Collection Start: **January 12, 2026**
   - Collection End: **February 13, 2026**
   - These dates are used as reference points for performance calculations

2. **Organization Filtering**: 
   - Admin users melihat semua data
   - Operator users hanya melihat data organisasi mereka sendiri (untuk `/statistics`)

3. **Completion Calculation**:
   - Completed tables = tables with status "finalized" OR "submitted"
   - Completion % = (completed / total) * 100

4. **Performance Calculation**:
   - Average time hanya dihitung untuk tabel yang sudah completed
   - Time = UpdatedAt - **CollectionStartDate** (bukan CreatedAt)
   - Tables updated before collection start date will use collection start as reference
   - Hanya organisasi dengan minimal 1 completed table yang muncul

5. **Idle Time Calculation**:
   - Dihitung dari updated_at tabel terakhir (paling recent) organisasi tersebut
   - Jika tidak ada aktivitas sejak collection start, maka dihitung dari collection start date
   - Menggunakan current server time sebagai reference
