package interfaces

import "github.com/hajimehoshi/ebiten/v2"

type TestInterface interface {
	TestMethod(screen *ebiten.Image)
}