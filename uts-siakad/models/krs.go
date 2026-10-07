package models

type KRS struct {
	ID           int64  `json:"id"`
	MahasiswaID  int64  `json:"mahasiswa_id"`
	MataKuliahID int64  `json:"mata_kuliah_id"`
	Semester     string `json:"semester"`
}
