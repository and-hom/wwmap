package main

//go:generate go-bindata -pkg $GOPACKAGE -o pattern-data.go ./pattern

import (
	"bytes"
	log "github.com/Sirupsen/logrus"
	_ "golang.org/x/image/bmp"
	"image"
	"math"
	"regexp"
	"strconv"
)

const MAX_DELTA_2 = 90

type PatternMatcher interface {
	Match(src image.Image, xLimit int) map[int]int
}

func NewPatternMatcher() (PatternMatcher, error) {
	re := regexp.MustCompile("^(-?\\d+)\\.\\w{3}$")

	patterns := make(map[int]image.Image)
	for _, name := range AssetNames() {
		if matches := re.FindStringSubmatch(name); matches != nil {
			id, err := strconv.Atoi(matches[1])
			if err != nil {
				log.Errorf("Failed to parse pattern id %s %v", matches[1], err)
				continue
			}
			patternImageBytes := MustAsset(matches[0])
			pImg, _, err := image.Decode(bytes.NewReader(patternImageBytes))
			if err != nil {
				log.Errorf("Failed to load pattern image %s %v", matches[1], err)
				continue
			}
			patterns[id] = pImg

			log.Infof("Pattern %d loaded", id)
		}
	}

	return &patternMatcher{
		patterns: patterns,
	}, nil
}

type patternMatcher struct {
	patterns map[int]image.Image
}

func (this *patternMatcher) Match(src image.Image, xLimit int) map[int]int {
	result := make(map[int]int)

	for pNum, pattern := range this.patterns {
		log.Debugf("Matching pattern %d", pNum)
		patternBounds := pattern.Bounds()
		pW := patternBounds.Dx()
		pH := patternBounds.Dy()
		for y := 0; y < (src.Bounds().Dy() - pH + 1); y++ {
			for x := 0; x < (xLimit - pW + 1); x++ {
				if this.matches(pNum, src, x, y) {
					result[pNum] = y
				}
			}
		}
	}

	return result
}

func (this *patternMatcher) matches(pNum int, src image.Image, x int, y int) bool {
	pW := this.patterns[pNum].Bounds().Dx()
	pH := this.patterns[pNum].Bounds().Dy()

	for j := 0; j < pH; j++ {
		for i := 0; i < pW; i++ {
			rO, gO, bO, _ := src.At(x+i, y+j).RGBA()
			rP, gP, bP, _ := this.patterns[pNum].At(i, j).RGBA()

			delta := math.Pow((float64)(rO-rP), 2) +
				math.Pow((float64)(gO-gP), 2) +
				math.Pow((float64)(bO-bP), 2)

			if delta > MAX_DELTA_2 {
				return false
			}
		}
	}

	// Check white fields left and right
	for i := 1; i < 3; i++ {
		for j := 0; j < pH; j++ {
			r, g, b, _ := src.At(x-i, y+j).RGBA()
			if !isWhite(r, g, b) {
				return false
			}
			r, g, b, _ = src.At(x+pW+i-1, y+j).RGBA()
			if !isWhite(r, g, b) {
				return false
			}
		}
	}

	return true
}
