# Table Aggregation - Generate Parent Table

## Overview

Fitur ini memungkinkan pembuatan tabel parent secara otomatis dari tabel child dengan melakukan agregasi data berdasarkan parent hierarchy dari dimension values.

## Marking System

Table yang dihasilkan dari agregasi akan ditandai dengan kolom khusus di database:

### 1. **is_aggregated** (boolean)
- Default: `false`
- Set ke `true` untuk table hasil agregasi
- Indexed untuk query performance
- Bisa digunakan untuk filter: `WHERE is_aggregated = true`

### 2. **source_table_id** (uuid, nullable)
- Menyimpan ID dari child table (source)
- NULL untuk table non-agregasi
- Indexed untuk traceability
- Bisa digunakan untuk trace back ke source

### 3. **notes** (text)
- Menyimpan informasi tambahan
- Format: `Generated from child table ID: {uuid}`
- Jika child table sudah punya notes, akan digabungkan

**Contoh Data:**
```json
{
  "id": "parent-table-uuid",
  "name": "Jumlah Penduduk Menurut Kelurahan (Agregasi Kelurahan - Parent Level)",
  "is_aggregated": true,
  "source_table_id": "child-table-uuid",
  "labels": ["statistik", "penduduk"],
  "notes": "Generated from child table ID: child-table-uuid"
}
```

**Query Examples:**
```sql
-- Ambil semua table agregasi
SELECT * FROM tables WHERE is_aggregated = true;

-- Ambil parent tables dari specific child table
SELECT * FROM tables WHERE source_table_id = 'child-table-uuid';

-- Ambil non-aggregated tables only
SELECT * FROM tables WHERE is_aggregated = false;
```

**API Query:**
```bash
# Filter aggregated tables
GET /api/v1/tables?is_aggregated=true

# Get source table
GET /api/v1/tables/{parent-table-id}
# Response includes: "source_table_id": "child-table-uuid"
```

**Kegunaan:**
- ✅ **Performance** - Indexed columns untuk query cepat
- ✅ **Filtering** - Filter table agregasi vs non-agregasi
- ✅ **Traceability** - Trace kembali ke source table
- ✅ **Data Integrity** - Type-safe di database level
- ✅ **Audit** - Track history agregasi

## Use Case

Ketika Anda memiliki tabel dengan data granular (misalnya per kelurahan) dan ingin membuat tabel agregat pada level yang lebih tinggi (misalnya per kecamatan), fitur ini akan:

1. Mengidentifikasi parent hierarchy dari dimension values
2. Membuat tabel baru dengan struktur yang sama tapi menggunakan parent values
3. Mengagregasi semua facts dari child ke parent dengan penjumlahan

## Contoh Penggunaan

### Scenario

Tabel: **"Jumlah Penduduk Menurut Kelurahan dan Agama"**

Dimension **Kelurahan** memiliki struktur:
- Kelurahan 1 (parent: Kecamatan 1)
- Kelurahan 2 (parent: Kecamatan 1)
- Kelurahan 3 (parent: Kecamatan 2)
- Kelurahan 4 (parent: Kecamatan 2)
- Kelurahan 5 (parent: Kecamatan 3)

Dimension **Agama** memiliki nilai:
- Islam
- Kristen
- Hindu
- Buddha

### Data Child Table (Kelurahan Level)

| Kelurahan | Agama | Tahun | Value |
|-----------|-------|-------|-------|
| Kelurahan 1 | Islam | 2024 | 1000 |
| Kelurahan 1 | Kristen | 2024 | 500 |
| Kelurahan 2 | Islam | 2024 | 1500 |
| Kelurahan 2 | Kristen | 2024 | 700 |
| Kelurahan 3 | Islam | 2024 | 2000 |
| Kelurahan 3 | Kristen | 2024 | 800 |

### Data Parent Table yang Dihasilkan (Kecamatan Level)

| Kecamatan | Agama | Tahun | Value (Aggregated) |
|-----------|-------|-------|--------------------|
| Kecamatan 1 | Islam | 2024 | 2500 (1000+1500) |
| Kecamatan 1 | Kristen | 2024 | 1200 (500+700) |
| Kecamatan 2 | Islam | 2024 | 2000 |
| Kecamatan 2 | Kristen | 2024 | 800 |

## API Endpoint

### Generate Parent Table

**Endpoint:** `POST /api/v1/tables/generate`

**Authentication:** Required (JWT)

**Request Body:**
```json
{
  "child_table_id": "uuid-of-child-table",
  "dimension_ids": ["uuid-dimension-1", "uuid-dimension-2"]
}
```

**Catatan:** 
- User menentukan dimension mana yang akan diagregasi
- Bisa mengirim 1 atau lebih dimension IDs
- Hanya dimensions yang memiliki parent hierarchy yang akan diproses
- Dimensions lain yang tidak ada parent akan tetap dipertahankan

**Contoh Use Cases:**

1. **Single Dimension:**
```json
{
  "child_table_id": "uuid",
  "dimension_ids": ["kelurahan-dimension-id"]
}
```
Hasil: Kelurahan → Kecamatan

2. **Multiple Dimensions:**
```json
{
  "child_table_id": "uuid",
  "dimension_ids": ["kelurahan-dimension-id", "bulan-dimension-id"]
}
```
Hasil: Kelurahan + Bulan → Kecamatan + Triwulan

**Response (201 Created - New Table):**
```json
{
  "data": {
    "parent_table_id": "uuid-of-generated-parent-table",
    "is_new_table": true,
    "message": "Parent table 'Jumlah Penduduk Menurut Kelurahan dan Agama (Agregasi Kelurahan - Parent Level)' berhasil dibuat dengan data agregasi",
    "child_table_id": "uuid-of-child-table",
    "aggregated_dimensions": [
      {
        "dimension_id": "uuid-dimension-1",
        "dimension_name": "Kelurahan",
        "parent_values_used": 3
      }
    ]
  },
  "message": "Parent table 'Jumlah Penduduk Menurut Kelurahan dan Agama (Agregasi Kelurahan - Parent Level)' berhasil dibuat dengan data agregasi"
}
```

**Response untuk Multiple Dimensions:**
```json
{
  "data": {
    "parent_table_id": "uuid-of-generated-parent-table",
    "is_new_table": true,
    "message": "Parent table 'Jumlah Penduduk Menurut Kelurahan dan Bulan (Agregasi Kelurahan, Bulan - Parent Level)' berhasil dibuat dengan data agregasi",
    "child_table_id": "uuid-of-child-table",
    "aggregated_dimensions": [
      {
        "dimension_id": "uuid-kelurahan",
        "dimension_name": "Kelurahan",
        "parent_values_used": 3
      },
      {
        "dimension_id": "uuid-bulan",
        "dimension_name": "Bulan",
        "parent_values_used": 4
      }
    ]
  },
  "message": "Parent table 'Jumlah Penduduk Menurut Kelurahan dan Bulan (Agregasi Kelurahan, Bulan - Parent Level)' berhasil dibuat dengan data agregasi"
}
```

**Error Responses:**

- **400 Bad Request:** Invalid request body atau validation error
- **400 Bad Request:** dimension_ids array kosong
- **404 Not Found:** Child table tidak ditemukan
- **400 Bad Request:** Table tidak memiliki dimensions
- **400 Bad Request:** Dimension tidak ditemukan di table
- **400 Bad Request:** Tidak ada dimension dengan parent hierarchy dari yang dikirim
- **500 Internal Server Error:** Error saat processing

## Cara Kerja Internal

### 1. Validasi

- Memastikan child table exists
- Memastikan table memiliki minimal satu dimension
- Validasi semua dimension IDs yang dikirim ada di table

### 2. Load dan Validasi Dimensions

Untuk setiap dimension ID yang dikirim user:

```go
for _, dimensionID := range dimensionIDs {
    // Check apakah dimension ada di table
    if !tableDimensionMap[dimensionID] {
        return error
    }
    
    dimension := loadDimension(dimensionID)
    
    // Check apakah ada dimension values dengan ParentID
    hasParent := false
    for _, value := range dimension.Values {
        if value.ParentID != nil {
            hasParent = true
            break
        }
    }
    
    if hasParent {
        // Collect info untuk agregasi
        aggregationInfo = append(aggregationInfo, ...)
    } else {
        log.Warning("dimension tidak punya parent, skip")
    }
}
```

### 3. Build Parent Aggregation Maps

Untuk setiap dimension yang punya parent, membuat mapping:

```go
// Untuk dimension Kelurahan
map[string][]string{
  "kecamatan_1_id": ["kelurahan_1_id", "kelurahan_2_id"],
  "kecamatan_2_id": ["kelurahan_3_id", "kelurahan_4_id"],
  "kecamatan_3_id": ["kelurahan_5_id"],
}

// Untuk dimension Bulan (jika ada)
map[string][]string{
  "triwulan_1_id": ["januari_id", "februari_id", "maret_id"],
  "triwulan_2_id": ["april_id", "mei_id", "juni_id"],
}
```

### 4. Create Parent Table

- Generate nama parent table dengan suffix `(Agregasi {DimensionNames} - Parent Level)`
- Copy struktur table (indicator, organization, labels, dimensions)
- Semua dimensions tetap sama karena parent values masih bagian dari dimension yang sama
- Buat table baru di database

### 5. Aggregate Facts (Multiple Dimensions)

Untuk setiap fact di child table:

1. Untuk setiap dimension yang diagregasi, identifikasi parent value-nya
2. Group facts berdasarkan:
   - Year
   - Kombinasi semua parent value IDs
   - Kombinasi dimension values lainnya yang tidak diagregasi
3. Sum nilai dari semua facts dalam grup yang sama
4. Simpan aggregated fact ke parent table

### 6. Response

Return parent table ID, status (new/updated), dimension info (ID, name, jumlah parent values), dan message

## Struktur Database

### Models Yang Digunakan

#### DimensionValue
```go
type DimensionValue struct {
    ID          string
    DimensionID string
    Name        string
    ParentID    *string              // NULL jika tidak ada parent
    Parent      *DimensionValue      // Relasi ke parent
    Children    []DimensionValue     // Relasi ke children
}
```

#### Fact
```go
type Fact struct {
    ID       string
    TableID  string
    Value    *float64
    OldValue *float64
    Year     int
}
```

#### FactDimensionValue
```go
type FactDimensionValue struct {
    ID               string
    FactID           string
    DimensionValueID string
}
```

## Limitasi & Catatan

1. **Agregasi Sum Only:** Saat ini hanya mendukung penjumlahan (SUM). Tidak mendukung AVG, MIN, MAX, COUNT, dll.

2. **Single Level Hierarchy:** Hanya mengagregasi satu level ke atas (child -> parent). Tidak mendukung multi-level hierarchy.

3. **Naming Convention:** Parent table diberi nama otomatis dengan suffix. Belum ada mekanisme untuk custom naming.

4. **No Duplicate Check:** Jika parent table sudah ada, saat ini akan membuat table baru. Perlu implementasi logic untuk update existing table.

5. **Transaction Safety:** Semua operasi dibungkus dalam transaction untuk memastikan data consistency.

## Development Notes

### Files Created/Modified

**Created:**
- `/internal/dto/table_aggregation_dto.go` - DTO untuk request/response
- `/internal/services/table_aggregation_service.go` - Business logic untuk agregasi
- `/internal/handlers/table_aggregation_handler.go` - HTTP handler

**Modified:**
- `/internal/repositories/dimension_repository.go` - Tambah interface methods
- `/internal/repositories/dimension_repository_impl.go` - Implementasi methods
- `/internal/routes/table_route.go` - Tambah route untuk generate parent
- `/internal/di/service_di.go` - Register service di DI
- `/internal/di/handler_di.go` - Register handler di DI
- `/app/app.go` - Daftarkan routes dan inject handler

### Next Steps / Future Improvements

1. **Check Existing Parent Table:** Implementasi logic untuk detect dan update existing parent table daripada selalu create baru

2. **Multiple Aggregation Functions:** Support untuk AVG, MIN, MAX, COUNT selain SUM

3. **Multi-Level Hierarchy:** Support agregasi multi-level (grandchild -> child -> parent)

4. **Custom Naming:** Allow user untuk specify custom name untuk parent table

5. **Incremental Update:** Jika parent table sudah ada, hanya update facts yang berubah

6. **Validation Rules:** Tambah validasi untuk ensure data integrity

7. **Bulk Operations:** Support untuk generate multiple parent tables sekaligus

8. **Audit Trail:** Track history dari agregasi untuk traceability

## Testing

### Manual Testing dengan cURL

```bash
# Generate parent table (single dimension)
curl -X POST http://localhost:8080/api/v1/tables/generate \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "child_table_id": "uuid-of-child-table",
    "dimension_ids": ["uuid-of-dimension-with-parent"]
  }'

# Generate parent table (multiple dimensions)
curl -X POST http://localhost:8080/api/v1/tables/generate \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "child_table_id": "uuid-of-child-table",
    "dimension_ids": ["kelurahan-dimension-id", "bulan-dimension-id"]
  }'
```

### Test Cases

1. **Happy Path - Single Dimension:** Child table dengan 1 dimension yang memiliki parent hierarchy → berhasil create parent table
2. **Happy Path - Multiple Dimensions:** Child table dengan 2+ dimensions yang memiliki parent → berhasil create parent table dengan multiple aggregations
3. **Mixed Dimensions:** User kirim 2 dimension IDs, tapi cuma 1 yang punya parent → proses yang 1, skip yang lain
4. **No Parent Hierarchy:** Semua dimensions yang dikirim tidak ada parent → return error
5. **Dimension Not in Table:** User kirim dimension ID yang tidak ada di table → return error
6. **Empty dimension_ids:** dimension_ids array kosong → validation error
7. **Empty Facts:** Child table tanpa data → create parent table kosong
8. **Multiple Years:** Facts dengan multiple years → aggregate per year
9. **Preserve Other Dimensions:** Table dengan dimensions (Kelurahan, Agama) → aggregate Kelurahan, Agama tetap preserved

## Questions?

Jika ada pertanyaan atau butuh modifikasi, silakan hubungi tim development.
