package models

type User struct {
	ID          int64  `json:"id"`
	Nama        string `json:"nama"`
	Email       string `json:"email"`
	Role        string `json:"role"`
	MahasiswaID *int64 `json:"mahasiswa_id,omitempty"`
	DosenID     *int64 `json:"dosen_id,omitempty"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
