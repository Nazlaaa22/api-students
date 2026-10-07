package models

type MataKuliah struct {
	ID      int64  `json:"id"`
	KodeMK  string `json:"kode_mk"`
	NamaMK  string `json:"nama_mk"`
	SKS     int    `json:"sks"`
	DosenID *int64 `json:"dosen_id"`
}
