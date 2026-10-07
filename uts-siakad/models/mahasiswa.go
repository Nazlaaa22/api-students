package models

type Mahasiswa struct {
	ID           int64   `json:"id"`
	NIM          string  `json:"nim"`
	Nama         string  `json:"nama"`
	Email        *string `json:"email"`
	ProgramStudi string  `json:"program_studi"`
}
