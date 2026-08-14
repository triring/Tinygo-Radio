// TinyGo コード（rda5807 初期化 → 76.8MHz 受信）
// tinygo build -target=m5stack -size=short -o MuteTest.uf2 .
// tinygo flash -target=m5stack -size=short -monitor .
// tinygo flash -target=pico -size=short -monitor .

// ソフトウェアスキャンに対応

package main

import (
	"fmt"
	"machine"
	//	"rda5807" // ローカルのディレクトリに置かれたrda5807のパッケージをインポートする場合
	"github.com/triring/Tinygo-Radio/rda5807" // githubで公開しているパッケージをインポートする場合
	"time"
)

func main() {
	// 地元のFM局の送信周波数を変数freq に設定する。
	// 周波数は、KHzで設定すること
	// 設定可能範囲	76000 – 108000 KHz

	freq := 76800
	// freq := 79000
	// freq := 88700
	// freq := 91400

	// I2Cの定義と設定
	i2c := machine.I2C0
	i2c.Configure(machine.I2CConfig{
		SDA:       machine.GPIO12, // for zero-kb02(raspi pico)
		SCL:       machine.GPIO13, // for zero-kb02(raspi pico)
		Frequency: 400 * machine.KHz,
	})
	// rda5807のオブジェクト生成
	// 第1引数: 使用するI2Cチャンネル
	radio := rda5807.New(i2c)
	// rda5807を初期化
	radio.InitRDA5807(rda5807.Band_World_Wide)
	// 周波数を設定
	radio.SetFrequency(freq)
	time.Sleep(500 * time.Millisecond)

	// 無限ループでmute とUnmuteを5秒毎に繰り返す。
	for {
		f, _ := radio.GetFrequency() // 現在、受信中の周波数を取得
		fmt.Printf("FM %5.1fMHz 受信中 : Unmute\n", (float64(f) / 1000.0))
		radio.SetMute(1)
		for i := 1; i <= 10; i++ {
			fmt.Printf(" %2d", i)
			time.Sleep(time.Second * 1)
		}
		fmt.Printf("FM %5.1fMHz 受信中 :   Mute\n", (float64(f) / 1000.0))
		radio.SetMute(0)
		for i := 1; i <= 10; i++ {
			fmt.Printf(" %2d", i)
			time.Sleep(time.Second * 1)
		}
	}
}

/*
> tinygo flash -target=pico -size=short -monitor .
   code    data     bss |   flash     ram
  65240    1564    5576 |   66804    7140
Connected to COM4. Press Ctrl-C to exit.
FM  76.8MHz 受信中 : Unmute
FM  76.8MHz 受信中 :   Mute
FM  76.8MHz 受信中 : Unmute
FM  76.8MHz 受信中 :   Mute
FM  76.8MHz 受信中 : Unmute
FM  76.8MHz 受信中 :   Mute
FM  76.8MHz 受信中 : Unmute
FM  76.8MHz 受信中 :   Mute
*/
