package model

type Student struct {
	ID       string  `json:"id"`
	NIM      string  `json:"nim"`
	Name     string  `json:"name"`
	Grade    float64 `json:"grade"`
	IsActive bool    `json:"is_active"`
	OwnerID  int     `json:"owner_id"`
}

type CreateStudentRequest struct {
	NIM      string  `json:"nim" validate:"required,max=20,notblank,nimformat"`
	Name     string  `json:"name" validate:"required,min=2,max=100,notblank"`
	Grade    float64 `json:"grade" validate:"gte=0,lte=100"`
	IsActive bool    `json:"is_active"`
}

type UpdateStudentRequest struct {
	NIM      string  `json:"nim" validate:"required,max=20,notblank,nimformat"`
	Name     string  `json:"name" validate:"required,min=2,max=100,notblank"`
	Grade    float64 `json:"grade" validate:"gte=0,lte=100"`
	IsActive bool    `json:"is_active"`
}

type PatchStudentRequest struct {
	NIM      *string  `json:"nim" validate:"omitnil,max=20,notblank,nimformat"`
	Name     *string  `json:"name" validate:"omitnil,min=2,max=100,notblank"`
	Grade    *float64 `json:"grade" validate:"omitnil,gte=0,lte=100"`
	IsActive *bool    `json:"is_active" validate:"omitnil"`
}
