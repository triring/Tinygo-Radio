// TinyGo コード（TEA5767 初期化 → 76.8MHz 受信）
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
	var (
		rssi byte = 0
		freq int
		i    byte
	)
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
	radio.InitTEA5767() // TEA5767の初期化を行う。
	time.Sleep(time.Second * 1)
	var (
		FREQ_MIN int = 76000
		//	FREQ_MAX int = 99000
		FREQ_MAX int = 87500 // 受信帯域をUS/EUモードに切り替えると、
		// フリーズ（I2C通信が設定できなくなる）してしまうバグがあるので、
		// 受信帯域をJPモードに制限する。
	)
	radio.SetFrequency(FREQ_MIN)
	radio.SetMute(true)
	// WideFMの周波数範囲は、76.0MHzから99.0MHzまで
	for freq = FREQ_MIN; freq <= FREQ_MAX; freq += 100 {
		radio.SetFrequency(freq)
		rssi = radio.GetRSSI()
		fmt.Printf("%3.1f,%2d :", float32(freq)/1000.0, rssi)
		for i = 0; i < (rssi*rssi)/4; i++ {
			fmt.Printf("w")
		}
		fmt.Printf("\n")
	}
	for {
		time.Sleep(time.Second * 5)
	}
}

/*
> tinygo flash -target=pico -size=short -monitor .
   code    data     bss |   flash     ram
  66108    1556    5576 |   67664    7132
Connected to COM4. Press Ctrl-C to exit.
76.0, 7 :wwwwwwwwwwww
76.1, 5 :wwwwww
76.2, 7 :wwwwwwwwwwww
76.3, 5 :wwwwww
76.4, 7 :wwwwwwwwwwww
76.5, 9 :wwwwwwwwwwwwwwwwwwww
76.6, 7 :wwwwwwwwwwww
76.7,10 :wwwwwwwwwwwwwwwwwwwwwwwww
76.8, 6 :wwwwwwwww
76.9, 5 :wwwwww
77.0, 8 :wwwwwwwwwwwwwwww
77.1, 6 :wwwwwwwww
77.2, 6 :wwwwwwwww
77.3, 6 :wwwwwwwww
77.4, 7 :wwwwwwwwwwww
77.5, 5 :wwwwww
77.6, 7 :wwwwwwwwwwww
77.7, 4 :wwww
77.8, 6 :wwwwwwwww
77.9, 5 :wwwwww
78.0, 5 :wwwwww
78.1, 3 :ww
78.2, 7 :wwwwwwwwwwww
78.3, 5 :wwwwww
78.4, 6 :wwwwwwwww
78.5, 5 :wwwwww
78.6, 7 :wwwwwwwwwwww
78.7, 5 :wwwwww
78.8, 5 :wwwwww
78.9, 5 :wwwwww
79.0, 7 :wwwwwwwwwwww
79.1, 5 :wwwwww
79.2, 6 :wwwwwwwww
79.3, 6 :wwwwwwwww
79.4, 6 :wwwwwwwww
79.5, 5 :wwwwww
79.6, 3 :ww
79.7, 5 :wwwwww
79.8, 5 :wwwwww
79.9, 6 :wwwwwwwww
80.0, 6 :wwwwwwwww

*/
