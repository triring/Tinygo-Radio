// TinyGo コード（TEA5767 初期化 -> 受信）
// tinygo build -target=m5stack -size=short -o FreqScan.uf2 ./main.go
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
	radio.SetMute(false)
	// WideFMの周波数範囲は、76.0MHzから99.0MHzまで
	var max_rssi uint8 = 0
	var station_frequency int = 0
	for freq = FREQ_MIN; freq <= FREQ_MAX; freq += 100 {
		radio.SetFrequency(freq)
		rssi = radio.GetRSSI() // 現在の信号強度(0-63)を読み出す
		time.Sleep(50 * time.Millisecond)
		if rssi > max_rssi { // 最も信号強度の強い放送周波数を探す。
			max_rssi = rssi
			station_frequency = freq
		}
		fmt.Printf("%5.1f, %2d :", float32(freq)/1000.0, rssi)
		for i = 0; i < (rssi*rssi)/4; i++ {
			fmt.Printf("w")
		}
		fmt.Printf("\n")
	}
	// 最も信号強度の強い放送周波数にチューニングする。
	radio.SetFrequency(station_frequency)
	radio.SetMute(false)
	for {
		rssi = radio.GetRSSI()
		fmt.Printf("Freq : %3.1f, RSSI : %2d\n", float32(station_frequency)/1000.0, rssi)
		time.Sleep(time.Second * 5)
	}
}

/*
> tinygo flash -target=pico -size=short -monitor .
   code    data     bss |   flash     ram
  66252    1556    5576 |   67808    7132
Connected to COM4. Press Ctrl-C to exit.
 76.0,  4 :wwww
 76.1,  3 :ww
 76.2,  3 :ww
 76.3,  3 :ww
 76.4,  4 :wwww
 76.5,  3 :ww
 76.6,  4 :wwww
 76.7,  7 :wwwwwwwwwwww
 76.8, 11 :wwwwwwwwwwwwwwwwwwwwwwwwwwwwww
 76.9,  4 :wwww
 77.0,  4 :wwww
 77.1,  4 :wwww
 77.2,  4 :wwww
 77.3,  3 :ww
 77.4,  5 :wwwwww
 77.5,  3 :ww
 77.6,  3 :ww
 77.7,  3 :ww
 77.8,  2 :w
 77.9,  4 :wwww
 78.0,  3 :ww
 78.1,  2 :w
 78.2,  3 :ww
 78.3,  5 :wwwwww
 78.4,  2 :w
 78.5,  3 :ww
 78.6,  4 :wwww
 78.7,  3 :ww
 78.8,  2 :w
 78.9,  2 :w
 79.0,  6 :wwwwwwwww
 79.1,  3 :ww
 79.2,  5 :wwwwww
 79.3,  2 :w
 79.4,  3 :ww
 79.5,  3 :ww
 79.6,  3 :ww
 79.7,  3 :ww
 79.8,  2 :w
 79.9,  4 :wwww
 80.0,  2 :w

*/
