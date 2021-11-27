package toggles

import (
	"github.com/and-hom/wwmap/lib/dao"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func ParseFeatureTogglesOrFallback(req *http.Request, userDao dao.UserDao) Toggles {
	featueToggles, err := Create(
		req.FormValue("toggles"),
		[]ToggleAvailableChecker{
			nil,
			RoleChecker(),
			ExperimentalFlagChecker(),
		},
		req,
		userDao,
	)
	if err != nil {
		log.Error("failed to parse toggles", err)
		featueToggles = Fallback()
	}
	return featueToggles
}
