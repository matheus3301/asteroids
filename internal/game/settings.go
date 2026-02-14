package game

import "github.com/hajimehoshi/ebiten/v2"

type resolution struct {
	Width, Height int
	Label         string
}

var resolutions = []resolution{
	{800, 600, "800X600"},
	{1280, 720, "1280X720"},
	{1920, 1080, "1920X1080"},
}

type settings struct {
	resolutionIndex int
	fullscreen      bool
}

func (s *settings) apply() {
	r := resolutions[s.resolutionIndex]
	ebiten.SetWindowSize(r.Width, r.Height)
	ebiten.SetFullscreen(s.fullscreen)
	if !s.fullscreen {
		mw, mh := ebiten.Monitor().Size()
		ebiten.SetWindowPosition((mw-r.Width)/2, (mh-r.Height)/2)
	}
}
