// TinyGo コード（si4703 初期化 → FM受信）
// tinygo build -target=m5stack -size=short -o GetRSSI.uf2 .
// tinygo flash -target=m5stack -size=short -monitor .
// tinygo build -target=pico -size=short -o GetRSSI.uf2 .
// tinygo flash -target=pico -size=short -monitor .

package main

import (
	"fmt"
	"machine"
	//	"si4703" // ローカルのディレクトリに置かれたsi4703のパッケージをインポートする場合
	"github.com/triring/Tinygo-Radio/si4703" // githubで公開しているパッケージをインポートする場合
	"time"
)

func main() {
	// 地元で受信可能なFM局の周波数を変数freqに定義して下さい。
	// * 周波数設定:KHzで設定すること
	// * 設定可能範囲	76000 – 108000 KHz
	freq := 91400
	/* その他の候補
	freq := 76800
	freq := 79000
	freq := 88700
	*/

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

	// freqに定義された周波数にチューニング
	if err := radio.SetFrequency(freq); err != nil {
		fmt.Printf("SetFrequency error: %v\r\n", err)
	}
	for {
		frequency, err := radio.GetRealFrequency()
		rssi, err := radio.GetRSSI()
		if err != nil {
			fmt.Printf("Frequency error: %v\r\n", err)
		} else {
			fmt.Printf("Frequency: %3d.%03d MHz, RSSI: %2d dBμV\r\n", frequency/1000, frequency%1000, rssi)
		}
		time.Sleep(5 * time.Second)
	}
}
