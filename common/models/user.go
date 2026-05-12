package models

const (
	RoleSuperAdmin = "superadmin"
	RoleAdmin      = "admin"
	RoleBasicUser  = "basic"
)

type User struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Password string `json:"-"`
	Role     string `json:"role"`
}
