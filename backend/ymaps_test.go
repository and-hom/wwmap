package main

import (
	. "github.com/franela/goblin"
	. "github.com/onsi/gomega"
	"testing"
)

func TestIntegration(t *testing.T) {
	g := Goblin(t)

	//special hook for gomega
	RegisterFailHandler(func(m string, _ ...int) {
		g.Fail(m)
	})

	g.Describe("Coords translation", func() {
		g.It("To tile z=0", func() {
			p, x, y := toTileCoords(0, Point{-90.0, 50.0})
			Expect(x).To(BeEquivalentTo(1))
			Expect(y).To(BeEquivalentTo(1))
			Expect(p.x).To(BeEquivalentTo(64))
			Expect(p.y).To(BeEquivalentTo(56))
		})
		g.It("To tile z=1 North", func() {
			p, x, y := toTileCoords(1, Point{0.1, 20.0})
			Expect(x).To(BeEquivalentTo(2))
			Expect(y).To(BeEquivalentTo(1))
			Expect(p.x).To(BeEquivalentTo(0))
			Expect(p.y).To(BeEquivalentTo(199))
		})
		g.It("To tile z=2 South", func() {
			p, x, y := toTileCoords(2, Point{-70.0, -20.0})
			Expect(x).To(BeEquivalentTo(2))
			Expect(y).To(BeEquivalentTo(3))
			Expect(p.x).To(BeEquivalentTo(56))
			Expect(p.y).To(BeEquivalentTo(113))
		})
	})

}

