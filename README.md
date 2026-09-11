# gopher-board-workshop（作業用リポジトリ）

[sat0ken/gopher-board-workshop](https://github.com/sat0ken/gopher-board-workshop) を進めるための作業場。
upstream の中身はコピーせず、参考用に `upstream/`（.gitignore 済み）へ clone して使う。
自分で書いたコードや設定だけをこのリポジトリで管理する。

## セットアップ状況（macOS 26.2 / arm64）

| 項目 | 状態 |
| --- | --- |
| Go | 1.27.1（Homebrew。TinyGo 0.42 は Go 1.25〜1.27 が必須） |
| TinyGo | 0.42.0（`~/.local/opt/tinygo`、`~/.local/bin/tinygo` の wrapper 経由） |
| ターゲット | `waveshare-rp2040-zero` |

### TinyGo を Homebrew で入れなかった理由

`brew install tinygo` は macOS 26 で以下のエラーになる。

```
Error: Your Command Line Tools (CLT) does not support macOS 26.
```

Xcode 本体はあるが CLT の receipt がなく、復旧には `sudo xcode-select --install`（GUI 操作）が要る。
公式リリースの tarball は LLVM を同梱した自己完結型なので CLT 不要で動く。そのため tarball を採用した。

再インストールする場合:

```sh
curl -fsSL -o /tmp/tinygo.tar.gz \
  https://github.com/tinygo-org/tinygo/releases/download/v0.42.0/tinygo0.42.0.darwin-arm64.tar.gz
mkdir -p ~/.local/opt && tar xzf /tmp/tinygo.tar.gz -C ~/.local/opt
```

`~/.local/bin/tinygo` は **symlink ではなく wrapper スクリプト**にすること。
TinyGo は実体パスから TINYGOROOT を自動検出するため、symlink だと
`could not autodetect root directory` で失敗する。

```sh
printf '#!/bin/sh\nexec "$HOME/.local/opt/tinygo/bin/tinygo" "$@"\n' > ~/.local/bin/tinygo
chmod +x ~/.local/bin/tinygo
```

## 使い方

```sh
make upstream   # 参考リポジトリを取得
make doctor     # go / tinygo / シリアルポートの確認
make 00         # 00_blink を書き込み（01, 02, ... 08-multi も同様）
make monitor    # tinygo monitor
```

書き込みできないときは、基板の BOOT ボタンを押しながら USB を挿して
ブートローダーモード（RPI-RP2 がマウントされる）にする。

## 進捗

- [x] 00. Lチカ
- [ ] 01. デジタル入力とシリアル通信
- [ ] 02. アナログ入力（光センサー）
- [ ] 03. アナログ出力（PWM）
- [x] 04. フルカラーLED（WS2812）
- [x] 05. ブザー
- [ ] 06. 温度・気圧センサー（BMP280）
- [ ] 07. 赤外線リモコン（送信 / 受信）
- [x] 08-1. 液晶に文字
- [ ] 08-2. 液晶を塗りつぶし
- [x] 08-3. 液晶に画像
- [x] koebiten でゲーム

## 自作プログラム

| ディレクトリ | 内容 | make |
| --- | --- | --- |
| `ws2812_party/` | 目のLEDを4パターンで光らせる。Up/Downでパターン、Left/Rightで速度、Bで消灯 | `make party` |
| `st7789_horse/` | 液晶に画像を1枚表示 | `make horse IMG=foo.png` |
| `st7789_profile/` | プロフィール / X のQRコード / 画像 の3画面をUp/Downで切り替え | `make profile URL=https://x.com/...` |
| `tools/qrgen/` | ビルド時にQRコードをビットマップへ変換するホスト側ツール | — |

### メモリの目安

`//go:embed` した `[]byte` は TinyGo では RAM に展開される。240x240 の画像1枚で 115KB 使うため、
RAM 264KB のうち画像は実質1枚が上限。複数枚載せるなら `string` で embed してフラッシュ側に置く。

日本語表示に使う東雲フォント（`tinyfont/shnm`）は JIS X 0208 の全グリフを持つのでフラッシュを約 205KB 使う。
