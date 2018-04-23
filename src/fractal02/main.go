package main

import (
	"fmt"
	"log"
	"os"

	"image"
	"image/color"
	"math/cmplx"

	"github.com/veandco/go-sdl2/sdl"
)

func timeLeft(nextTime uint32) uint32 {
	var now = sdl.GetTicks()

	if nextTime <= now {
		return 0
	} else {
		return nextTime - now
	}
}

const (
	winWidth     = 940
	winHeight    = 720
	tickInterval = 30
)

var (
	nextTime uint32

	err      error
	window   *sdl.Window
	renderer *sdl.Renderer
	texture  *sdl.Texture

	quit      bool
	event     sdl.Event
	locationX = 0.0
	locationY = 0.0
	step      = 0.5
	deviation = 2.0
)

func mandelbrot(z complex128) color.Color {
	const iterations = 200
	const contrast = 25

	var v complex128
	for n := uint8(0); n < iterations; n++ {
		v = v*v + z
		if cmplx.Abs(v) > 2 {
			return color.RGBA{0, 0, 155 - contrast*n, 255}
		}
	}
	return color.Black
}

func createFractal(minX float64, maxX float64, minY float64, maxY float64) *image.RGBA {
	fmt.Println(minX, maxX, minY, maxY)
	xmin, xmax, ymin, ymax := minX, maxX, minY, maxY

	img := image.NewRGBA(image.Rect(0, 0, winWidth, winHeight))
	for py := 0; py < winHeight; py++ {
		y := float64(py)/winHeight*(ymax-ymin) + ymin
		for px := 0; px < winWidth; px++ {
			x := float64(px)/winWidth*(xmax-xmin) + xmin
			z := complex(x, y)
			// image point (px, py) represents complex value z
			img.Set(px, py, mandelbrot(z))
		}
	}

	return img // note: ignoring errors
}

func loadMyFractal(minX float64, maxX float64, minY float64, maxY float64) (*sdl.Texture, error) {
	var newSurface *sdl.Surface
	var newTexture *sdl.Texture

	myImage := createFractal(minX, maxX, minY, maxY)

	newSurface, err = createSurfaceFromImage(myImage)
	if err != nil {
		return nil, err
	}

	newTexture, err = createTextureFromSurface(newSurface)
	if err != nil {
		return nil, err
	}

	return newTexture, nil
}

func createSurfaceFromImage(i image.Image) (*sdl.Surface, error) {
	rgba := image.NewRGBA(i.Bounds())
	w, h := i.Bounds().Max.X, i.Bounds().Max.Y

	s, err := sdl.CreateRGBSurface(0, int32(w), int32(h), 32, 0, 0, 0, 0)
	if err != nil {
		return s, err
	}
	rgba.Pix = s.Pixels()

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			c := i.At(x, y)
			rgba.Set(x, y, c)
		}
	}

	return s, err
}

func createTextureFromSurface(mySurface *sdl.Surface) (*sdl.Texture, error) {
	var newTexture *sdl.Texture

	newTexture, err = renderer.CreateTextureFromSurface(mySurface)
	if err != nil {
		return nil, err
	}

	return newTexture, nil
}

func run() int {
	if err = sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize SDL: %s\n", err)
		return 1
	}
	defer sdl.Quit()

	if window, err = sdl.CreateWindow("colorsort", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, winWidth, winHeight, sdl.WINDOW_SHOWN); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create window: %s\n", err)
		return 2
	}
	defer window.Destroy()

	if renderer, err = sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create renderer: %s\n", err)
		return 3 // don't use os.Exit(3); otherwise, previous deferred calls will never run
	}
	defer renderer.Destroy()

	if texture, err = loadMyFractal(locationX-deviation, locationX+deviation, locationY-deviation, locationY+deviation); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create texture: %s\n", err)
		return 4
	}
	defer texture.Destroy()

	quit = false
	for !quit {
		for event = sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch t := event.(type) {
			case *sdl.QuitEvent:
				quit = true
			case *sdl.KeyboardEvent:
				switch t.Keysym.Sym {
				case sdl.K_UP:
					locationY -= step * deviation
				case sdl.K_DOWN:
					locationY += step * deviation
				case sdl.K_LEFT:
					locationX -= step * deviation
				case sdl.K_RIGHT:
					locationX += step * deviation
				case sdl.K_RETURN:
					deviation *= step
				}
				texture, err = loadMyFractal(locationX-deviation, locationX+deviation, locationY-deviation, locationY+deviation)
				if err != nil {
					log.Fatal("Error creating Texture:", err)
				}
			}
		}

		renderer.Clear()
		renderer.Copy(texture, nil, nil)
		renderer.Present()

		sdl.Delay(timeLeft(nextTime))
		nextTime += tickInterval
	}
	return 1
}

func main() {
	os.Exit(run())
}
