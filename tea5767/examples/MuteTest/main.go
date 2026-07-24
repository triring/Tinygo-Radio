// TinyGo コード（TEA5767 初期化 → 76.8MHz 受信）
// tinygo build -target=m5stack -size=short -o fm_radio.uf2 .
// tinygo flash -target=m5stack -size=short -monitor .
// tinygo flash -target=pico -size=short -monitor .

// ソフトウェアスキャンに対応

package main

import (
	"fmt"
	"machine"
	//	"tea5767" // ローカルのディレクトリに置かれたtea5767のパッケージをインポートする場合
	"github.com/triring/Tinygo-Radio/tea5767" // githubで公開しているパッケージをインポートする場合
	"time"
)

func main() {
	//	周波数設定
	freq := 76800 // KHzで設定すること
	//  freq := 88700 // KHzで設定すること
	//	freq := 91400   // KHzで設定すること

	machine.I2C0.Configure(machine.I2CConfig{
		Frequency: 100 * machine.KHz,
		SDA:       machine.GPIO12, // for zero-kb02(raspi pico)
		SCL:       machine.GPIO13, // for zero-kb02(raspi pico)
	})

	// tea5767を初期化
	// 第1引数: I2Cチャンネル
	// 第2引数: tea5767のアドレス
	radio := tea5767.New(machine.I2C0, tea5767.DefaultAddress)
	radio.Configure(tea5767.Config{
		JapanBand: true,
		Mute:      false,
	})
	radio.InitTEA5767() // TEA5767の初期化を行う。
	time.Sleep(time.Second * 1)
	radio.TuneFrequency(freq)
	time.Sleep(time.Second * 3)

	// プログラムが終了しないように無限ループで待機（ラジオは鳴り続けます）
	for {
		f := radio.GetFrequency() // 現在、受信中の周波数を取得
		m := radio.GetMute()
		fmt.Printf("FM %3.1fMHz 受信中, Mute=%5t ", (float64(f) / 1000.0), m)
		for i := 1; i <= 10; i++ {
			fmt.Printf(" %2d", i)
			time.Sleep(time.Second * 1)
		}
		fmt.Printf("\n")
		radio.SetMute(!m)
	}

}

/*

  65416    1556    5576 |   66972    7132
Connected to COM4. Press Ctrl-C to exit.
FM 76.8MHz 受信中, Mute=false   1  2  3  4  5  6  7  8  9 10
FM 76.8MHz 受信中, Mute= true   1  2  3  4  5  6  7  8  9 10
FM 76.8MHz 受信中, Mute=false   1  2  3  4  5  6  7  8  9 10
FM 76.8MHz 受信中, Mute= true   1  2  3  4  5  6  7  8  9 10
FM 76.8MHz 受信中, Mute=false   1  2  3  4  5  6  7  8  9 10
FM 76.8MHz 受信中, Mute= true   1  2  3  4  5  6  7  8  9 10
FM 76.8MHz 受信中, Mute=false   1  2  3  4  5  6  7  8  9 10
FM 76.8MHz 受信中, Mute= true   1  2  3  4  5  6
*/
