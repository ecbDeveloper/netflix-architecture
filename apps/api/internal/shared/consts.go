package shared

const (
	SessionUserIDKey    = "AuthenticatedUserID"
	SessionProfileIDKey = "ProfileID"
	SessionRoleIDKey    = "RoleID"

	DBRoleAdmin  int32 = 1
	DBRoleMember int32 = 2

	MaturityRatingPrefix = "RATING_"

	StorageFolderPermission = 0755
)
