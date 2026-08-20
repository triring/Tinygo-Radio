// TinyGo コード（si4703 初期化 → FM受信）
// tinygo build -target=m5stack -size=short -o FMstation.uf2 .
// tinygo flash -target=m5stack -size=short -monitor .
// tinygo build -target=pico -size=short -o FMstation.uf2 .
// tinygo flash -target=pico -size=short -monitor .

package main

import (
	"fmt"
	"machine"
	// "si4703" // ローカルのディレクトリに置かれたsi4703のパッケージをインポートする場合
	"github.com/triring/Tinygo-Radio/si4703" // githubで公開しているパッケージをインポートする場合
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
	station := [...]int{76800, 79000, 88700, 91400}

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

	// 配列stationの先頭に定義された周波数にチューニングする。
	if err := radio.SetFrequency(station[0]); err != nil {
		fmt.Printf("SetFrequency error: %v\r\n", err)
	}
	for {
		// 配列に登録されている放送局の周波数を順次、チューニングして5秒づつ受信を繰り返す。
		for i, f := range station {
			radio.SetFrequency(f)
			freq, _ = radio.GetRealFrequency() // 現在、受信中の周波数を取得
			rssi, _ = radio.GetRSSI()          // 電波強度を取得
			fmt.Printf("%d. FM %3d.%03dMHz 受信中, RSSI : %2d\n", i, freq/1000, freq%1000, rssi)
			time.Sleep(time.Second * 5)
		}
		time.Sleep(5 * time.Second)
	}
}

/*
> tinygo flash -target=pico -size=short -monitor .
   code    data     bss |   flash     ram
  65740    1604    5576 |   67344    7180
Connected to COM4. Press Ctrl-C to exit.
DEVICEID = 0x1242
CHIPID   = 0x1253
0. FM  76.800MHz 受信中, RSSI : 48
1. FM  79.000MHz 受信中, RSSI : 23
2. FM  88.700MHz 受信中, RSSI : 47
3. FM  91.400MHz 受信中, RSSI : 40
0. FM  76.800MHz 受信中, RSSI : 52
1. FM  79.000MHz 受信中, RSSI : 35
2. FM  88.700MHz 受信中, RSSI : 47
3. FM  91.400MHz 受信中, RSSI : 44
*/
