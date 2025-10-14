package role

type Role int

const (
	Guest Role = iota // 0
	User              // 1
	Admin             // 2
)

func RoleToString(role Role) string {
	switch role {
	case Admin:
		return "admin"
	case User:
		return "user"
	default:
		return "guest"
	}
}
