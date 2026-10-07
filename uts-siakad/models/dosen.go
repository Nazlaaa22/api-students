package models

type Dosen struct {
	ID    int64   `json:"id"`
	NIDN  string  `json:"nidn"`
	Nama  string  `json:"nama"`
	Email *string `json:"email"`
}
