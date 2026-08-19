// TinyGo コード（rda5807 初期化 → FM受信）
// tinygo build -target=m5stack -size=short -o Sample.uf2 .
// tinygo flash -target=m5stack -size=short -monitor .
// tinygo build -target=pico -size=short -o Sample.uf2 .
// tinygo flash -target=pico -size=short -monitor .

package main

import (
	"fmt"
	"machine"
	// "si4703" // ローカルのディレクトリに置かれたrda5807のパッケージをインポートする場合
	"time"
	"github.com/triring/Tinygo-Radio/si4703" // githubで公開しているパッケージをインポートする場合
)

func main() {
	fmt.Println("RP2040 + Si4703")
	// --------------------------------------------------
	// RP2040
	//
	// I2C0
	//   SDA = GPIO4
	//   SCL = GPIO5
	//
	// Si4703 RESET
	//   GPIO3
	// --------------------------------------------------

	i2c := machine.I2C0

	err := i2c.Configure(machine.I2CConfig{
		Frequency: 400 * machine.KHz,
		SDA:       machine.GPIO4,
		SCL:       machine.GPIO5,
	})

	if err != nil {
		fmt.Printf("I2C configure error: %v\r\n", err)
		for {
			time.Sleep(time.Second)
		}
	}

	radio := si4703.New(
		i2c,
		machine.GPIO3,
	)

	// Si4703初期化。
	err = radio.Init()
	if err != nil {
		fmt.Printf("Si4703 init error: %v\r\n", err)
		for {
			time.Sleep(time.Second)
		}
	}

	fmt.Printf("DEVICEID = 0x%04X\r\n", radio.GetDeviceID())
	fmt.Printf("CHIPID   = 0x%04X\r\n", radio.GetChipID())

	// --------------------------------------------------
	// 日本向け設定
	// 76～108MHz
	// 100kHz spacing
	// --------------------------------------------------

	if err := radio.SetBand(si4703.BandJapanWide); err != nil {
		fmt.Printf("SetBand error: %v\r\n", err)
	}

	if err := radio.SetSpace(si4703.Space100kHz); err != nil {
		fmt.Printf("SetSpace error: %v\r\n", err)
	}

	// SeekのRSSI閾値。
	//
	// 0は最も弱い局も候補になります。
	// 実際の環境では10～30程度から試すとよいでしょう。
	if err := radio.SetSeekThreshold(40); err != nil {
		fmt.Printf("SetSeekThreshold error: %v\r\n", err)
	}
	// 音量。
	if err := radio.SetVolume(8); err != nil {
		fmt.Printf("SetVolume error: %v\r\n", err)
	}

	// Unmute。
	if err := radio.SetMute(false); err != nil {
		fmt.Printf("SetMute error: %v\r\n", err)
	}

	// 受信地域の設定から、受信可能な周波数範囲を取得する。
	start, end, spacing, _ := radio.BandParameters()
	fmt.Printf("Band Frequency %d - %d MHz\n", start/1000, end/1000)
	// バンドの下限周波数にチューニング
	if err := radio.SetFrequency(start); err != nil {
		fmt.Printf("SetFrequency error: %v\r\n", err)
	}

	var (
		max_rssi          uint8 = 0
		station_frequency int   = 0
		freq              int   = 0
		rssi              uint8 = 0
		i                 int   = 0
	)
	radio.SetMute(true)
	for freq = start; freq <= end; freq += spacing {
		radio.SetFrequency(freq)
		rssi, _ = radio.GetRSSI() // 現在の信号強度(0-63)を読み出す
		time.Sleep(50 * time.Millisecond)
		if rssi > max_rssi { // 最も信号強度の強い放送周波数を探す。
			max_rssi = rssi
			station_frequency = freq
		}
		fmt.Printf("%3d.%03d, %2d :", freq/1000, freq%1000, rssi)
		//	for i = 0; i < int(rssi)*int(rssi)/8; i++ {
		for i = 0; i < int(rssi); i++ {
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
		freq, _ = radio.GetRealFrequency()
		fmt.Printf("Freq : %3d.%03d, RSSI : %2d\n", freq/1000, freq%1000, rssi)
		time.Sleep(time.Second * 5)
	}

	/*
		for {
			frequency, err := radio.GetRealFrequency()
			if err != nil {
				fmt.Printf("Frequency error: %v\r\n", err)
			} else {
				fmt.Printf("Frequency=%3d.%03d MHz\r\n", frequency/1000, frequency%1000)
			}
			time.Sleep(5 * time.Second)
		}
	*/
}
