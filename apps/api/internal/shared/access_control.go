package shared

type AccessControlService struct{}

func NewAccessControlService() AccessControlService {
	return AccessControlService{}
}

func (a AccessControlService) CanAccess(hasParentalControls bool, isKidsFriendly bool) bool {
	if hasParentalControls {
		return isKidsFriendly
	}
	return true
}
