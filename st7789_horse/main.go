// ST7789 の液晶に 240x240 の画像を1枚表示する。
// 画像は tools/main.go で RGB565BE の生バイトに変換したものを埋め込む。
//
//	go run upstream/tools/main.go st7789_horse/horse.png st7789_horse/horse.raw
package main

import (
	_ "embed"
	"machine"

	"tinygo.org/x/drivers/pixel"
	"tinygo.org/x/drivers/st7789"
)

//go:embed horse.raw
var imgData []byte

func main() {
	machine.SPI1.Configure(machine.SPIConfig{
		Frequency: 16000000,
		Mode:      0,
	})

	display := st7789.New(machine.SPI1,
		machine.GPIO9,  // RESET
		machine.GPIO12, // DC
		machine.GPIO13, // CS
		machine.GPIO14) // LITE

	display.Configure(st7789.Config{
		Rotation: st7789.ROTATION_90,
		Height:   240,
		Width:    240,
	})

	img := pixel.NewImageFromBytes[pixel.RGB565BE](240, 240, imgData)
	display.DrawBitmap(0, 0, img)

	// 描画後は何もしない。画面はそのまま保持される。
	select {}
}
