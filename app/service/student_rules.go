package service

import "api-students/app/model"

// ApplyPatchStudent menerapkan perubahan data saat PATCH.
// Validasi dilakukan secara deklaratif melalui ValidateStruct.
func ApplyPatchStudent(
	current *model.Student,
	input model.PatchStudentRequest,
) {
	if input.NIM != nil {
		current.NIM = *input.NIM
	}

	if input.Name != nil {
		current.Name = *input.Name
	}

	if input.Grade != nil {
		current.Grade = *input.Grade
	}

	if input.IsActive != nil {
		current.IsActive = *input.IsActive
	}
}
