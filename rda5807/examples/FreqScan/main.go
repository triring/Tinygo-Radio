// TinyGo コード（rda5807 初期化 → 76.8MHz 受信）
// tinygo build -target=m5stack -size=short -o FreqScan.uf2 .
// tinygo flash -target=m5stack -size=short -monitor .
// tinygo flash -target=pico -size=short -monitor .

// ソフトウェアスキャンに対応

package main

import (
	"fmt"
	"machine"
	// "rda5807" // ローカルのディレクトリに置かれたrda5807のパッケージをインポートする場合
	"github.com/triring/Tinygo-Radio/rda5807" // githubで公開しているパッケージをインポートする場合
	"time"
)

func main() {
	var (
		rssi byte = 0
		freq int
		i    byte
	)

	// I2Cの定義と設定
	i2c := machine.I2C0
	i2c.Configure(machine.I2CConfig{
		SDA:       machine.GPIO12, // for zero-kb02(raspi pico)
		SCL:       machine.GPIO13, // for zero-kb02(raspi pico)
		Frequency: 400 * machine.KHz,
	})
	time.Sleep(500 * time.Millisecond)
	// rda5807のオブジェクト生成
	// 第1引数: 使用するI2Cチャンネル
	radio := rda5807.New(i2c)
	time.Sleep(500 * time.Millisecond)
	// rda5807を初期化
	radio.InitRDA5807(rda5807.Band_World_Wide)
	time.Sleep(500 * time.Millisecond)
	// 受信地域の設定から、受信可能な周波数範囲を取得する。
	band, band_name, min_freq, max_freq := radio.GetBandInfo()
	fmt.Printf("%08b | Band : [%s] %d - %d MHz\n", band, band_name, min_freq/1000, max_freq/1000)
	// 周波数を設定
	radio.SetFrequency(min_freq)

	var max_rssi uint8 = 0
	var station_frequency int = 0
	radio.SetMute(true)
	for freq = min_freq; freq <= max_freq; freq += 100 {
		radio.SetFrequency(freq)
		rssi, _ = radio.GetRSSI() // 現在の信号強度(0-63)を読み出す
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
	radio.SetVolume(12)
	for {
		rssi, _ = radio.GetRSSI()
		fmt.Printf("Freq : %3.1f, RSSI : %2d\n", float32(station_frequency)/1000.0, rssi)
		time.Sleep(time.Second * 5)
	}
}

/*
> tinygo flash -target=pico -size=short -monitor .
   code    data     bss |   flash     ram
  64336    1564    5576 |   65900    7140
Connected to COM4. Press Ctrl-C to exit.
Initialization of RDA5807 is complete.
00001000 | Band : [World wide] 76 - 108 MHz
76.0, 2 :w
76.1, 2 :w
76.2, 2 :w
76.3, 2 :w
76.4, 2 :w
76.5, 2 :w
76.6, 2 :w
76.7, 7 :wwwwwwwwwwww
76.8,13 :wwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwww
76.9, 9 :wwwwwwwwwwwwwwwwwwww
77.0, 2 :w
77.1, 2 :w
77.2, 2 :w
77.3, 2 :w
77.4, 2 :w
77.5, 2 :w
77.6, 2 :w
77.7, 2 :w
77.8, 2 :w
77.9, 2 :w
78.0, 2 :w
78.1, 2 :w
78.2, 2 :w
78.3, 2 :w
78.4, 2 :w
78.5, 2 :w
78.6, 2 :w
78.7, 2 :w
78.8, 2 :w
78.9, 2 :w
79.0, 2 :w
*/
