TARGET := waveshare-rp2040-zero
FLASH  := tinygo flash --target $(TARGET) --size short

.PHONY: upstream doctor monitor
## 参考用の upstream リポジトリを取得（.gitignore 済み）
upstream:
	@test -d upstream || git clone https://github.com/sat0ken/gopher-board-workshop.git upstream

## 環境チェック
doctor:
	@go version
	@tinygo version
	@echo "--- serial ---"; ls /dev/cu.usb* 2>/dev/null || echo "(デバイス未接続)"

## シリアルモニタ
monitor:
	tinygo monitor

# --- ワークショップ課題の書き込み ---
.PHONY: 00 01 02 03 04 05 06 07-send 07-recv 08-txt 08-bmp 08-img 08-multi
00: ; $(FLASH) ./upstream/00_blink/main.go
01: ; $(FLASH) ./upstream/01_switch/main.go
02: ; $(FLASH) ./upstream/02_analog_input/main.go
03: ; $(FLASH) ./upstream/03_pwm/main.go
04: ; cd upstream/04_ws2812 && $(FLASH) main.go
05: ; $(FLASH) ./upstream/05_buzzer/main.go
06: ; cd upstream/06_bmp280 && $(FLASH) main.go
07-send: ; $(FLASH) ./upstream/07_ir_send/main.go
07-recv: ; cd upstream/07_ir_recieve && $(FLASH) main.go
08-txt:   ; cd upstream/08_st7789_txt && $(FLASH) main.go
08-bmp:   ; cd upstream/08_st7789_bmp && $(FLASH) main.go
08-img:   ; cd upstream/08_st7789_img && $(FLASH) main.go
08-multi: ; cd upstream/08_st7789_multi_img && $(FLASH) main.go

# --- koebiten（マイコン向け ebiten 移植のゲームエンジン） ---
.PHONY: koebiten game
koebiten:
	@test -d koebiten || git clone --depth 1 https://github.com/sago35/koebiten.git koebiten

## GAME=blocks などで単体のゲームを書き込む。省略時は全ゲーム入りのランチャー
GAME ?= all
game: koebiten
	cd koebiten && tinygo flash --target ./targets/gopher-board-spi.json --size short ./games/$(GAME)

# --- 自作プログラム ---
.PHONY: party horse
## フルカラーLED をボタンで操作する
party:
	cd ws2812_party && $(FLASH) .

## 液晶に画像を表示する。IMG=path/to/foo.png で画像を差し替える
horse:
ifdef IMG
	cd upstream/tools && go run main.go "$(CURDIR)/$(IMG)" "$(CURDIR)/st7789_horse/horse.raw"
endif
	cd st7789_horse && $(FLASH) .
