package service

import (
	"api-students/app/model"
	"api-students/helper"
)

// CanAccessStudent menentukan apakah user boleh mengakses
// data student berdasarkan owner dan permission.
func CanAccessStudent(
	current model.AuthUser,
	ownerID int,
	perms *helper.PermissionSet,
	anyPermission string,
) bool {
	// Pemilik data selalu boleh mengakses.
	if current.UserID == ownerID {
		return true
	}

	// Jika bukan pemilik, harus memiliki permission :any.
	if perms == nil {
		return false
	}

	return perms.Has(anyPermission)
}
