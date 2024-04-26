package main

import (
	"math"
	"strconv"
	"time"

	"github.com/gdamore/tcell/v2"
)

// Constants
const (
	thetaSpacing = 0.07
	phiSpacing   = 0.02
	R1           = 1
	R2           = 2
	K2           = 5
	scrWidth     = 200.0
	scrHeight    = 50.0
)

// Compute K1
var K1 = scrWidth * K2 * 3 / (8 * (R1 + R2))

// The potential illumination values
var illumination = []rune{' ', '.', ',', '-', '~', ':', ';', '!', '*', '#', '$', '@'}

func RecomputeK1(screenWidth float64) {
	K1 = screenWidth * K2 * 4 / (8 * (R1 + R2))
}

func main() {
	// Initialize screen

	A := 0.07
	B := 0.03
	screenData := make([][]rune, scrWidth)
	zBuffer := make([][]float64, scrWidth)
	for i := range screenData {
		screenData[i] = make([]rune, scrWidth)
		zBuffer[i] = make([]float64, scrWidth)
	}

	screen, err := tcell.NewScreen()
	if err != nil {
		panic(err)
	}
	if err := screen.Init(); err != nil {
		panic(err)
	}
	defer screen.Fini()
	quit := make(chan struct{})
	go func() {
		counter := 0
		for {
			for i := range screenData {
				for j := range screenData[i] {
					screenData[i][j] = ' '
					zBuffer[i][j] = 0
				}
			}
			CalculateFrame(A, B, screenData, zBuffer)
			A += 0.07
			B += 0.03
			// Set up the screen update interval
			time.Sleep(time.Second / 60)
			screen.Clear()

			yOffset := 20

			// Drawing operations
			style := tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorRed)
			for i := yOffset; i < scrWidth; i++ {
				for j, r := range screenData[i] {
					screen.SetContent(j, i-yOffset, r, nil, style)
				}
			}
			w, h := screen.Size()
			wstr := strconv.Itoa(w)
			hstr := strconv.Itoa(h)
			for i, c := range wstr {
				screen.SetContent(i, 0, c, nil, style)
			}
			for i, c := range hstr {
				screen.SetContent(i, 1, c, nil, style)
			}

			screen.Show()
			select {
			case <-quit:
				return
			default:
			}
			counter++
		}
	}()

	// Wait for a key press to exit
	for {
		ev := screen.PollEvent()
		switch ev.(type) {
		case *tcell.EventKey:
			close(quit)
			return
		}
	}
}

func CalculateFrame(A, B float64, screenData [][]rune, zBuffer [][]float64) {
	var (
		cosA = math.Cos(A)
		sinA = math.Sin(A)
		cosB = math.Cos(B)
		sinB = math.Sin(B)
	)
	for theta := 0.0; theta < 2*math.Pi; theta += thetaSpacing {
		// Precompute sin & cos of theta
		costheta := math.Cos(theta)
		sintheta := math.Sin(theta)

		for phi := 0.0; phi < 2*math.Pi; phi += phiSpacing {
			// Precompute sin & cos of phi
			cosphi := math.Cos(phi)
			sinphi := math.Sin(phi)

			// X/Y-coordinate of the circle before revolving
			circleX := R2 + R1*costheta
			circleY := R1 * sintheta

			// Final x/y-coordinates after rotations
			x := circleX*(cosB*cosphi+sinA*sinB*sinphi) - circleY*cosA*sinB
			y := circleX*(sinB*cosphi-sinA*cosB*sinphi) + circleY*cosA*cosB
			ooz := 1 / (K2 + cosA*circleX*sinphi + circleY*sinA + 5)

			// x and y projection
			xp := int(scrWidth/2.0 + K1*ooz*x)
			yp := int(scrWidth/2.0 - K1*ooz*y) // NOTE: Negated

			// Luminance calculation
			L := cosphi*costheta*sinB - cosA*costheta*sinphi - sinA*sintheta + cosB*(cosA*sintheta-costheta*sinA*sinphi)

			// L is from -sqrt(2) to +sqrt(2). Negative values are facing away from the viewer
			if L > 0 {
				if ooz > zBuffer[xp][yp] {
					zBuffer[xp][yp] = ooz
					luminanceIndex := int(L * 8)
					screenData[xp][yp] = illumination[luminanceIndex]
				}
			}
		}
	}
}
