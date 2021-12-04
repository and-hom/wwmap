package toggles

import (
	"context"
	"errors"
	"github.com/and-hom/wwmap/lib/dao"
	"github.com/and-hom/wwmap/lib/handler"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strconv"
)

const FALLBACK int = 0

const (
	FEATURE_SHOW_CAMPS        = 0
	FEATURE_SHOW_UNPUBLISHED  = 1
	FEATURE_SHOW_SLOPE        = 2
	FEATURE_ALTITUED_COVERAGE = 3
)

const NEEDS_AUTH = 0x110

type Toggles interface {
	GetShowCamps(ctx context.Context) (bool, context.Context)
	GetShowUnpublished(ctx context.Context) (bool, context.Context)
	GetShowSlope(ctx context.Context) (bool, context.Context)
	GetAltitudeCoverage(ctx context.Context) (bool, context.Context)
}

func Create(
	value string,
	checkers []ToggleAvailableChecker,
	req *http.Request,
	userDao dao.UserDao,
) (Toggles, error) {
	i, err := strconv.ParseInt(value, 2, 32)
	if err != nil {
		return nil, err
	}
	return &bitmaskToggles{
		Value:    int(i),
		checkers: checkers,
		getUser: func(ctx context.Context) (*dao.User, context.Context, bool, error) {
			userFromContext := handler.GetUserFromContext(ctx)
			if userFromContext != nil {
				return userFromContext, ctx, true, nil
			}
			if userDao == nil {
				return nil, ctx, false, errors.New("User dao is null")
			}
			if req == nil {
				return nil, ctx, false, errors.New("Request is null")
			}
			user, _, authorized, authErr := handler.GetUser(req, userDao)
			if authErr == nil && authorized {
				ctx = handler.WithUser(ctx, user)
			}
			return user, ctx, authorized, authErr
		},
	}, err
}

func Fallback() Toggles {
	return &bitmaskToggles{
		FALLBACK,
		[]ToggleAvailableChecker{},
		func(ctx context.Context) (*dao.User, context.Context, bool, error) {
			return nil, ctx, false, nil
		},
	}
}

type bitmaskToggles struct {
	Value    int
	checkers []ToggleAvailableChecker
	getUser  func(ctx context.Context) (*dao.User, context.Context, bool, error)
}

func (this *bitmaskToggles) get(idx int, ctx context.Context) (bool, context.Context) {
	if (this.Value & (1 << idx)) == 0 {
		return false, ctx
	}

	if this.checkers == nil || len(this.checkers) <= idx || this.checkers[idx] == nil {
		return true, ctx
	}

	user, ctx, authorized, err := this.getUser(ctx)
	if err != nil {
		log.Error("Authorization error", err)
		return false, ctx
	}
	if !authorized {
		return false, ctx
	}
	return this.checkers[idx].Available(user), ctx
}

func (this *bitmaskToggles) GetShowCamps(ctx context.Context) (bool, context.Context) {
	return this.get(FEATURE_SHOW_CAMPS, ctx)
}

func (this *bitmaskToggles) GetShowUnpublished(ctx context.Context) (bool, context.Context) {
	return this.get(FEATURE_SHOW_UNPUBLISHED, ctx)
}

func (this *bitmaskToggles) GetShowSlope(ctx context.Context) (bool, context.Context) {
	return this.get(FEATURE_SHOW_SLOPE, ctx)
}

func (this *bitmaskToggles) GetAltitudeCoverage(ctx context.Context) (bool, context.Context) {
	return this.get(FEATURE_ALTITUED_COVERAGE, ctx)
}
