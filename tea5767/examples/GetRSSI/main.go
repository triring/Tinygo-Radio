// TinyGo コード（TEA5767 初期化 → FM 受信）
// tinygo build -target=m5stack -size=short -o fm_radio.uf2 ./main.go
// tinygo flash -target=m5stack -size=short -monitor ./main.go
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
	//	周波数設定
	//	freq := 76800   // KHzで設定すること
	freq := 79000 // KHzで設定すること
	//  freq := 88700 // KHzで設定すること
	//	freq := 91400   // KHzで設定すること

	machine.I2C0.Configure(machine.I2CConfig{
		Frequency: 100 * machine.KHz,
		SDA:       machine.GPIO12, // for zero-kb02(raspi pico)
		SCL:       machine.GPIO13, // for zero-kb02(raspi pico)
	})

	radio := tea5767.New(machine.I2C0, tea5767.DefaultAddress)
	radio.Configure(tea5767.Config{
		JapanBand: true,
		Mute:      false,
	})

	radio.SetFrequency(freq)

	for {
		rssi := radio.GetRSSI()
		fmt.Printf("FM %3.1fMHz 受信中,RSSI: %2d\n", (float64(freq) / 1000.0), rssi)
		time.Sleep(time.Second * 3)
	}
}

/*
> tinygo flash -target=pico -size=short -monitor .
   code    data     bss |   flash     ram
  65028    1556    5576 |   66584    7132
Connected to COM4. Press Ctrl-C to exit.
FM 79.0MHz 受信中,RSSI:  8
FM 79.0MHz 受信中,RSSI:  8
FM 79.0MHz 受信中,RSSI:  8
FM 79.0MHz 受信中,RSSI:  8
FM 79.0MHz 受信中,RSSI:  8
FM 79.0MHz 受信中,RSSI:  8
FM 79.0MHz 受信中,RSSI:  8
FM 79.0MHz 受信中,RSSI:  8
FM 79.0MHz 受信中,RSSI:  8
FM 79.0MHz 受信中,RSSI:  8
FM 79.0MHz 受信中,RSSI:  8
*/
