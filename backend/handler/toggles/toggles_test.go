package toggles_test

import (
	"context"
	"github.com/and-hom/wwmap/backend/handler/toggles"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParse(t *testing.T) {
	_toggles, err := toggles.Create("0010", nil, nil, nil)
	assert.Nil(t, err)
	showCamps, _ := _toggles.GetShowCamps(context.Background())
	assert.False(t, showCamps)
	showUnpublished, _ := _toggles.GetShowUnpublished(context.Background())
	assert.True(t, showUnpublished)
	showSlope, _ := _toggles.GetShowSlope(context.Background())
	assert.False(t, showSlope)
}
