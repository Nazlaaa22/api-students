package service

import (
	"errors"

	"api-students/app/model"
)

var (
	ErrInvalidNIM   = errors.New("NIM wajib diisi")
	ErrInvalidName  = errors.New("Nama wajib diisi")
	ErrInvalidGrade = errors.New("Grade harus berada di antara 0 sampai 100")
)

// ValidateCreateStudent melakukan validasi data saat POST.
func ValidateCreateStudent(input model.CreateStudentRequest) error {
	if input.NIM == "" {
		return ErrInvalidNIM
	}

	if input.Name == "" {
		return ErrInvalidName
	}

	if input.Grade < 0 || input.Grade > 100 {
		return ErrInvalidGrade
	}

	return nil
}

// ValidateUpdateStudent melakukan validasi data saat PUT.
func ValidateUpdateStudent(input model.UpdateStudentRequest) error {
	if input.NIM == "" {
		return ErrInvalidNIM
	}

	if input.Name == "" {
		return ErrInvalidName
	}

	if input.Grade < 0 || input.Grade > 100 {
		return ErrInvalidGrade
	}

	return nil
}

// ApplyPatchStudent menerapkan perubahan data saat PATCH.
func ApplyPatchStudent(
	current *model.Student,
	input model.PatchStudentRequest,
) error {

	if input.NIM != nil {
		if *input.NIM == "" {
			return ErrInvalidNIM
		}

		current.NIM = *input.NIM
	}

	if input.Name != nil {
		if *input.Name == "" {
			return ErrInvalidName
		}

		current.Name = *input.Name
	}

	if input.Grade != nil {
		if *input.Grade < 0 || *input.Grade > 100 {
			return ErrInvalidGrade
		}

		current.Grade = *input.Grade
	}

	if input.IsActive != nil {
		current.IsActive = *input.IsActive
	}

	return nil
}
