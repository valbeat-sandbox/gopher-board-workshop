// 液晶に X のプロフィールと、そのリンクの QR コードを表示する。
// Up / Down ボタンで画面を切り替える。
//
// QR はビルド時に生成したビットマップを埋め込む。マイコン上で
// エンコードするとバイナリが大きくなるため、生成はホスト側で済ませる。
//
//	go run ./tools/qrgen "<プロフィールのURL>" st7789_profile/qr.bin
package main

import (
	_ "embed"
	"image/color"
	"machine"
	"time"

	"tinygo.org/x/drivers/st7789"
	"tinygo.org/x/tinyfont"
	"tinygo.org/x/tinyfont/freesans"
)

// 表示内容。変えたいときはここだけ触る。
// profileURL を変えたら qrgen を実行し直して qr.bin を作り直すこと。
const (
	profileName   = "Takuma Kajikawa"
	profileHandle = "@kajitack"
	profileURL    = "https://x.com/kajitack"
	profileBio1   = "Software Engineer"
	profileBio2   = "Go / Kubernetes"
)

//go:embed qr.bin
var qrData []byte

const screenSize = 240

var (
	white = color.RGBA{R: 255, G: 255, B: 255, A: 255}
	black = color.RGBA{A: 255}
	cyan  = color.RGBA{G: 220, B: 255, A: 255}
	gray  = color.RGBA{R: 150, G: 150, B: 150, A: 255}
)

var (
	btnUp   = machine.GPIO3
	btnDown = machine.GPIO6
)

var display st7789.Device

func main() {
	machine.SPI1.Configure(machine.SPIConfig{
		Frequency: 16000000,
		Mode:      0,
	})

	display = st7789.New(machine.SPI1,
		machine.GPIO9,  // RESET
		machine.GPIO12, // DC
		machine.GPIO13, // CS
		machine.GPIO14) // LITE

	display.Configure(st7789.Config{
		Rotation: st7789.ROTATION_90,
		Height:   screenSize,
		Width:    screenSize,
	})

	buttons := []machine.Pin{btnUp, btnDown}
	for _, b := range buttons {
		b.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	}
	pressed := make([]bool, len(buttons))

	screens := []func(){drawProfile, drawQR}
	current := 0
	screens[current]()

	for {
		for i, b := range buttons {
			now := !b.Get()
			// 押した瞬間だけ切り替える。描画は重いので連打させない。
			if now && !pressed[i] {
				if b == btnUp {
					current = (current + 1) % len(screens)
				} else {
					current = (current - 1 + len(screens)) % len(screens)
				}
				screens[current]()
			}
			pressed[i] = now
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func drawProfile() {
	display.FillScreen(black)

	tinyfont.WriteLine(&display, &freesans.Bold12pt7b, 10, 60, profileName, white)
	tinyfont.WriteLine(&display, &freesans.Bold12pt7b, 10, 95, profileHandle, cyan)
	tinyfont.WriteLine(&display, &freesans.Regular9pt7b, 10, 140, profileBio1, white)
	tinyfont.WriteLine(&display, &freesans.Regular9pt7b, 10, 165, profileBio2, white)
	tinyfont.WriteLine(&display, &freesans.Regular9pt7b, 10, 220, "Press Up for QR", gray)
}

// 埋め込んだビットマップを画面いっぱいに拡大して描く。
func drawQR() {
	display.FillScreen(white)

	n := int(qrData[0])
	bits := qrData[1:]

	// モジュールが整数倍になる拡大率を選び、余りを余白として左右に振る。
	scale := screenSize / (n + 2) // +2 は QR の静穏帯（余白）の分
	origin := int16((screenSize - n*scale) / 2)

	for y := 0; y < n; y++ {
		for x := 0; x < n; x++ {
			i := y*n + x
			if bits[i/8]&(1<<(7-uint(i%8))) == 0 {
				continue
			}
			display.FillRectangle(
				origin+int16(x*scale),
				origin+int16(y*scale),
				int16(scale), int16(scale), black)
		}
	}
}
