// QR コードをビルド時に生成し、マイコンに埋め込む用のビットマップとして書き出す。
// 出力形式は 1 バイト目がモジュール数 N、以降が N*N ビットを詰めたもの（1=黒）。
//
//	go run ./tools/qrgen "https://x.com/example" st7789_profile/qr.bin
package main

import (
	"fmt"
	"os"

	qrcode "github.com/skip2/go-qrcode"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("usage: qrgen <text> <output.bin>")
		os.Exit(1)
	}

	q, err := qrcode.New(os.Args[1], qrcode.Medium)
	if err != nil {
		panic(err)
	}
	// 余白は液晶側で付けるので、ここでは付けない。
	q.DisableBorder = true

	matrix := q.Bitmap()
	n := len(matrix)
	if n > 255 {
		panic("QR が大きすぎる")
	}

	out := []byte{byte(n)}
	bits := make([]byte, (n*n+7)/8)
	for y := 0; y < n; y++ {
		for x := 0; x < n; x++ {
			if matrix[y][x] {
				i := y*n + x
				bits[i/8] |= 1 << (7 - uint(i%8))
			}
		}
	}
	out = append(out, bits...)

	if err := os.WriteFile(os.Args[2], out, 0o644); err != nil {
		panic(err)
	}
	fmt.Printf("done: %s (%dx%d modules -> %d bytes)\n", os.Args[2], n, n, len(out))
}
