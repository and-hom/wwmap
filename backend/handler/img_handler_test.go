package handler_test

import (
	"testing"
	"github.com/and-hom/wwmap/backend/handler"
	"image"
	"github.com/stretchr/testify/assert"
)

func TestResizeEqImage(t *testing.T) {
	rect, small := handler.PreviewRect(image.Rect(0, 0, 2, 2), 2, 2)
	assert.False(t, small)
	assert.Equal(t, image.Rect(0, 0, 2, 2), rect)
}

func TestResizeBigImage(t *testing.T) {
	rect, small := handler.PreviewRect(image.Rect(0, 0, 8, 4), 2, 2)
	assert.False(t, small)
	assert.Equal(t, image.Rect(0, 0, 2, 1), rect)
}

func TestResizeSmallImage(t *testing.T) {
	rect, small := handler.PreviewRect(image.Rect(0, 0, 8, 4), 12, 12)
	assert.True(t, small)
	assert.Equal(t, image.Rect(0, 0, 12, 6), rect)
}

func TestResizeDifferentProportions(t *testing.T) {
	rect, small := handler.PreviewRect(image.Rect(0, 0, 8, 4), 4, 8)
	assert.False(t, small)
	assert.Equal(t, image.Rect(0, 0, 4, 2), rect)
}