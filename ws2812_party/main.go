// Gopher くんの目のフルカラーLED(WS2812)を激しく光らせる。
// 基板のスイッチでパターン・速度・点灯/消灯を切り替えられる。
//
//	Up    : 次のパターン
//	Down  : 前のパターン
//	Right : 速くする
//	Left  : 遅くする
//	B     : 消灯 / 再開
package main

import (
	"image/color"
	"machine"
	"time"

	"tinygo.org/x/drivers/ws2812"
)

// 明るさの上限。WS2812 の最大値。まぶしすぎる場合はここを下げる。
const maxLevel = 255

// パターンは「フレーム番号から左右の色を決める関数」として書く。
// 状態を持たないので、ボタン入力の待ち時間を挟んでも破綻しない。
type pattern struct {
	frame func(n int) (left, right color.RGBA)
	// 1フレームあたりの基準の長さ。速度倍率はこれに掛かる。
	interval time.Duration
}

var patterns = []pattern{
	{frame: strobe, interval: 15 * time.Millisecond},
	{frame: rainbow, interval: 4 * time.Millisecond},
	{frame: flicker, interval: 20 * time.Millisecond},
	{frame: police, interval: 50 * time.Millisecond},
}

// 速度倍率(%)。小さいほど1フレームが短くなり、速く光る。
var speeds = []int{200, 150, 100, 70, 40}

var (
	neo  machine.Pin = machine.D29
	leds [2]color.RGBA
	ws   ws2812.Device

	btnUp    = machine.GPIO3
	btnDown  = machine.GPIO6
	btnLeft  = machine.GPIO4
	btnRight = machine.GPIO5
	btnB     = machine.GPIO15

	current = 0 // 再生中のパターン
	speed   = 2 // speeds のインデックス。初期値は等倍
	paused  = false
)

var off = color.RGBA{A: 255}

func main() {
	neo.Configure(machine.PinConfig{Mode: machine.PinOutput})
	ws = ws2812.NewWS2812(neo)

	buttons := []machine.Pin{btnUp, btnDown, btnLeft, btnRight, btnB}
	for _, b := range buttons {
		b.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	}
	// プルアップなので、押されていない状態が true。
	pressed := make([]bool, len(buttons))

	frame := 0
	for {
		for i, b := range buttons {
			now := !b.Get()
			// 押した瞬間だけ反応させる。押しっぱなしでは連続実行しない。
			if now && !pressed[i] {
				onPress(buttons[i])
				frame = 0
			}
			pressed[i] = now
		}

		p := patterns[current]
		if paused {
			write(off, off)
		} else {
			write(p.frame(frame))
		}
		frame++

		time.Sleep(p.interval * time.Duration(speeds[speed]) / 100)
	}
}

func onPress(pin machine.Pin) {
	switch pin {
	case btnUp:
		current = (current + 1) % len(patterns)
	case btnDown:
		current = (current - 1 + len(patterns)) % len(patterns)
	case btnRight:
		if speed < len(speeds)-1 {
			speed++
		}
	case btnLeft:
		if speed > 0 {
			speed--
		}
	case btnB:
		paused = !paused
	}
}

// 白の全点灯と消灯を繰り返す。
func strobe(n int) (color.RGBA, color.RGBA) {
	if n%2 == 0 {
		white := color.RGBA{R: maxLevel, G: maxLevel, B: maxLevel, A: 255}
		return white, white
	}
	return off, off
}

// 色相を高速に回す。左右の目は色相を半周ずらす。
func rainbow(n int) (color.RGBA, color.RGBA) {
	deg := uint16(n * 20 % 360)
	return hsv(deg), hsv((deg + 180) % 360)
}

// 左右それぞれランダムな色に切り替える。
func flicker(n int) (color.RGBA, color.RGBA) {
	return hsv(uint16(random() % 360)), hsv(uint16(random() % 360))
}

// 赤と青を交互に点ける。
func police(n int) (color.RGBA, color.RGBA) {
	if n%2 == 0 {
		return color.RGBA{R: maxLevel, A: 255}, off
	}
	return off, color.RGBA{B: maxLevel, A: 255}
}

func write(left, right color.RGBA) {
	leds[0], leds[1] = left, right
	ws.WriteColors(leds[:])
}

// 色相(0〜359度)を彩度・明度が最大の RGB に変換する。
func hsv(deg uint16) color.RGBA {
	region := deg / 60
	// 区間内の進み具合を 0〜maxLevel に写す。
	rise := uint8(uint32(deg%60) * maxLevel / 60)
	fall := maxLevel - rise

	switch region {
	case 0:
		return color.RGBA{R: maxLevel, G: rise, A: 255}
	case 1:
		return color.RGBA{R: fall, G: maxLevel, A: 255}
	case 2:
		return color.RGBA{G: maxLevel, B: rise, A: 255}
	case 3:
		return color.RGBA{G: fall, B: maxLevel, A: 255}
	case 4:
		return color.RGBA{R: rise, B: maxLevel, A: 255}
	default:
		return color.RGBA{R: maxLevel, B: fall, A: 255}
	}
}

// xorshift。math/rand を持ち込まずに済ませる。
var seed uint32 = 2463534242

func random() uint32 {
	seed ^= seed << 13
	seed ^= seed >> 17
	seed ^= seed << 5
	return seed
}
