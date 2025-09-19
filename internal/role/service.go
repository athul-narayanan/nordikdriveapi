package role

import (
	"nordik-drive-api/internal/auth"

	"gorm.io/gorm"
)

type RoleService struct {
	DB *gorm.DB
}

func (rs *RoleService) GetAllRoles(uniqueRoles []string) ([]Role, error) {
	var roles []Role
	if len(uniqueRoles) == 0 {
		return roles, nil
	}

	result := rs.DB.Where("role NOT IN ?", uniqueRoles).Find(&roles)
	if result.Error != nil {
		return nil, result.Error
	}

	return roles, nil
}

func (rs *RoleService) GetRoleByUser(userid int) ([]auth.UserRole, error) {
	var roles []auth.UserRole
	result := rs.DB.Where("user_id = ?", userid).Order("id ASC").First(&roles)
	if result.Error != nil {
		return nil, result.Error
	}
	return roles, nil
}

func (rs *RoleService) GetRolesByUserId(userid int) ([]auth.UserRole, error) {
	var roles []auth.UserRole
	result := rs.DB.Where("user_id = ?", userid).Find(&roles)
	if result.Error != nil {
		return nil, result.Error
	}
	return roles, nil
}
