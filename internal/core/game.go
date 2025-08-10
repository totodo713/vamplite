package core

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type Game struct {
	// ゲーム状態
}

func NewGame() *Game {
	return &Game{}
}

func (g *Game) Update() error {
	// ゲーム更新ロジック
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	// 描画ロジック
	ebitenutil.DebugPrint(screen, "マッスルドリーマー開発中...")
	screen.Fill(color.RGBA{50, 50, 100, 255})
}

func (g *Game) Layout(_, _ int) (screenWidth, screenHeight int) {
	return 1280, 720
}

func (g *Game) Run() error {
	ebiten.SetWindowSize(1280, 720)
	ebiten.SetWindowTitle("マッスルドリーマー〜観光編〜")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	return ebiten.RunGame(g)
}
