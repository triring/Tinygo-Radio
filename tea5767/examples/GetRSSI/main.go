// TinyGo コード（TEA5767 初期化 → FM 受信）
// tinygo build -target=m5stack -size=short -o GetRSSI.uf2 .
// tinygo flash -target=m5stack -size=short -monitor .
// tinygo flash -target=pico -size=short -monitor .

package main

import (
	"fmt"
	"machine"
	//	"tea5767" // ローカルのディレクトリに置かれたtea5767のパッケージをインポートする場合
	"github.com/triring/Tinygo-Radio/tea5767" // githubで公開しているパッケージをインポートする場合
	"time"
)

func main() {
	// 周波数設定:KHzで設定すること
	// 設定可能範囲	76000 – 90000 KHz

	// freq := 76800
	// freq := 79000
	freq := 88700

	machine.I2C0.Configure(machine.I2CConfig{
		Frequency: 100 * machine.KHz,
		SDA:       machine.GPIO12, // for zero-kb02(raspi pico)
		SCL:       machine.GPIO13, // for zero-kb02(raspi pico)
	})
	// tea5767のオブジェクト生成と初期設定
	radio := tea5767.New(machine.I2C0, tea5767.DefaultAddress)
	radio.Configure(tea5767.Config{
		JapanBand: true,
		Mute:      false,
	})

	// 周波数を設定
	radio.SetFrequency(freq)
	time.Sleep(200 * time.Millisecond)
	for {
		rssi := radio.GetRSSI()
		fmt.Printf("FM %5.1fMHz 受信中,RSSI: %2d\n", (float64(freq) / 1000.0), rssi)
		time.Sleep(time.Second * 5)
	}
}

/*
> tinygo flash -target=pico -size=short -monitor .
   code    data     bss |   flash     ram
  65028    1556    5576 |   66584    7132
Connected to COM4. Press Ctrl-C to exit.
FM  88.7MHz 受信中,RSSI: 11
FM  88.7MHz 受信中,RSSI: 11
FM  88.7MHz 受信中,RSSI: 11
FM  88.7MHz 受信中,RSSI: 11
FM  88.7MHz 受信中,RSSI: 11
FM  88.7MHz 受信中,RSSI: 11
FM  88.7MHz 受信中,RSSI: 11
FM  88.7MHz 受信中,RSSI: 11
FM  88.7MHz 受信中,RSSI: 11
FM  88.7MHz 受信中,RSSI: 11
FM  88.7MHz 受信中,RSSI: 11
FM  88.7MHz 受信中,RSSI: 11
*/
