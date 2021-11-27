package toggles

import (
	"github.com/and-hom/wwmap/lib/dao"
)

type ToggleAvailableChecker interface {
	Available(user *dao.User) bool
}

func ExperimentalFlagChecker() ToggleAvailableChecker {
	return &experimentalFlagChecker{}
}

type experimentalFlagChecker struct {
}

func (this *experimentalFlagChecker) Available(user *dao.User) bool {
	return user.ExperimentalFeaures
}

func RoleChecker() ToggleAvailableChecker {
	return &roleChecker{}
}

type roleChecker struct {
}

func (this *roleChecker) Available(user *dao.User) bool {
	return user.Role == dao.ADMIN || user.Role == dao.EDITOR
}
