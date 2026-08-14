// TinyGo コード（TEA5767 初期化 → 76.8MHz 受信）
// tinygo build -target=m5stack -size=short -o FMstation.uf2 .
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
	var (
		rssi byte
		freq int
	)	
	// 地元で受信可能なFM局の周波数を以下の配列に登録して下さい。
	// 登録された周波数を順番に受信していきます。
	// 周波数設定:KHzで設定すること
	// 設定可能範囲	76000 – 108000 KHz
	station := [...]int{76800, 79000, 88700}

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
	time.Sleep(500 * time.Millisecond)
	for {
		// 配列に登録されている放送局の周波数を順次、チューニングして5秒づつ受信を繰り返す。
		for i, f := range station {
			radio.SetFrequency(f)
			freq = radio.GetFrequency() //	受信中の周波数を取得
			rssi = radio.GetRSSI()      //	電波強度を取得
			fmt.Printf("%d. FM %5.1fMHz 受信中, RSSI : %2d\n", i, (float64(freq) / 1000.0), rssi)
			time.Sleep(time.Second * 5)
		}
	}

}

/*
> tinygo flash -target=pico -size=short -monitor .
   code    data     bss |   flash     ram
  65332    1556    5576 |   66888    7132
Connected to COM4. Press Ctrl-C to exit.
0. FM  76.8MHz 受信中, RSSI : 12
1. FM  79.0MHz 受信中, RSSI :  8
2. FM  88.7MHz 受信中, RSSI :  5
0. FM  76.8MHz 受信中, RSSI :  9
1. FM  79.0MHz 受信中, RSSI :  8
2. FM  88.7MHz 受信中, RSSI :  5
0. FM  76.8MHz 受信中, RSSI : 10
1. FM  79.0MHz 受信中, RSSI :  5
2. FM  88.7MHz 受信中, RSSI : 12
0. FM  76.8MHz 受信中, RSSI : 12
1. FM  79.0MHz 受信中, RSSI :  6
2. FM  88.7MHz 受信中, RSSI :  4
*/
