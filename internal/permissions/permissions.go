package permissions

const (
	AlertsRead   = "alerts:read"
	AlertsCreate = "alerts:create"
	AlertsUpdate = "alerts:update"
	AlertsDelete = "alerts:delete"

	ReportsRead    = "reports:read"
	ReportsApprove = "reports:approve"

	SOSRead         = "sos:read"
	SOSUpdateStatus = "sos:update_status"

	AssistanceRead         = "assistance:read"
	AssistanceUpdateStatus = "assistance:update_status"

	ChatsRead = "chats:read"
	ChatsSend = "chats:send"

	NotificationsRead = "notifications:read"
	NotificationsSend = "notifications:send"

	PrivilegeCodesCreate = "privilege_codes:create"
	PrivilegeCodesRead   = "privilege_codes:read"
	PrivilegeCodesRevoke = "privilege_codes:revoke"

	AuditLogsRead = "audit_logs:read"
)

var AllPermissions = []string{
	AlertsRead,
	AlertsCreate,
	AlertsUpdate,
	AlertsDelete,

	ReportsRead,
	ReportsApprove,

	SOSRead,
	SOSUpdateStatus,

	AssistanceRead,
	AssistanceUpdateStatus,

	ChatsRead,
	ChatsSend,

	NotificationsRead,
	NotificationsSend,

	PrivilegeCodesCreate,
	PrivilegeCodesRead,
	PrivilegeCodesRevoke,

	AuditLogsRead,
}

func IsValidPermission(permission string) bool {
	for _, allowed := range AllPermissions {
		if permission == allowed {
			return true
		}
	}

	return false
}

func InvalidPermissions(input []string) []string {
	invalid := []string{}

	for _, permission := range input {
		if !IsValidPermission(permission) {
			invalid = append(invalid, permission)
		}
	}

	return invalid
}

func ContainsPermission(permissionList []string, required string) bool {
	for _, permission := range permissionList {
		if permission == required {
			return true
		}
	}

	return false
}

func RemoveDuplicatePermissions(input []string) []string {
	seen := map[string]bool{}
	result := []string{}

	for _, permission := range input {
		if permission == "" {
			continue
		}

		if seen[permission] {
			continue
		}

		seen[permission] = true
		result = append(result, permission)
	}

	return result
}