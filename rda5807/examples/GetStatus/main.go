// TinyGo コード（rda5807 初期化 → 76.8MHz 受信）
// tinygo build -target=m5stack -size=short -o GetStatus.uf2 .
// tinygo flash -target=m5stack -size=short -monitor .
// tinygo flash -target=pico -size=short -monitor .

// ソフトウェアスキャンに対応

package main

import (
	"fmt"
	"machine"
	//	"rda5807" // ローカルのディレクトリに置かれたrda5807のパッケージをインポートする場合
	"github.com/triring/Tinygo-Radio/rda5807" // githubで公開しているパッケージをインポートする場合
	"time"
)

func main() {
	// 周波数設定:KHzで設定すること
	// 設定可能範囲	76000 – 108000 KHz

	freq := 76800
	// freq := 79000
	// freq := 88700
	// freq := 91400

	// I2Cの定義と設定
	i2c := machine.I2C0
	i2c.Configure(machine.I2CConfig{
		SDA:       machine.GPIO12, // for zero-kb02(raspi pico)
		SCL:       machine.GPIO13, // for zero-kb02(raspi pico)
		Frequency: 400 * machine.KHz,
	})
	// rda5807のオブジェクト生成
	// 第1引数: 使用するI2Cチャンネル
	radio := rda5807.New(i2c)
	// rda5807を初期化
	radio.InitRDA5807(rda5807.Band_World_Wide)
	// 周波数を設定
	radio.SetFrequency(freq)
	time.Sleep(time.Millisecond * 800)
	id, err := radio.GetChipID()
	if err != nil {
		fmt.Printf("GetChipID() error %s\n", err)
	}
	fmt.Printf("ChipID : 0x%02x\n", id)
	band, band_name, min_freq, max_freq := radio.GetBandInfo()
	fmt.Printf("%08b | Band : [%s] %d - %d MHz\n", band, band_name, min_freq/1000, max_freq/1000)

	f, err := radio.GetFrequency()
	if err != nil {
		fmt.Printf("GetFrequency() error %s\n", err)
	}
	fmt.Printf("Frequency : %5.2f MHz\n", float64(f)/1000.0)

	v, err := radio.GetVolume()
	if err != nil {
		fmt.Printf("GetVolume() error %s\n", err)
	}
	fmt.Printf("Volume    : %d\n", v)

	r, err := radio.GetRSSI()
	if err != nil {
		fmt.Printf("GetRSSI() error %s\n", err)
	}
	fmt.Printf("RSSI      : %d\n", r)

	for {
		f, _ := radio.GetFrequency() // 現在、受信中の周波数を取得
		v, _ := radio.GetVolume()    // 現在の音量を取得
		r, _ := radio.GetRSSI()      // 現在の電波強度を取得
		fmt.Printf("FM %5.1fMHz 受信中,Vol : %2d, RSSI : %2d\n", (float64(f) / 1000.0), v, r)
		time.Sleep(time.Second * 5)
	}

}

/*
> tinygo flash -target=pico -size=short -monitor .
   code    data     bss |   flash     ram
  65896    1596    5576 |   67492    7172
Connected to COM4. Press Ctrl-C to exit.
ChipID : 0x5804
00001000 | Band : [World wide] 76 - 108 MHz
Frequency : 76.80 MHz
Volume    : 11
RSSI      : 61
FM  76.8MHz 受信中,Vol : 11, RSSI : 61
FM  76.8MHz 受信中,Vol : 11, RSSI : 61
FM  76.8MHz 受信中,Vol : 11, RSSI : 61
FM  76.8MHz 受信中,Vol : 11, RSSI : 61
FM  76.8MHz 受信中,Vol : 11, RSSI : 61
FM  76.8MHz 受信中,Vol : 11, RSSI : 60
FM  76.8MHz 受信中,Vol : 11, RSSI : 60
FM  76.8MHz 受信中,Vol : 11, RSSI : 58
FM  76.8MHz 受信中,Vol : 11, RSSI : 58
FM  76.8MHz 受信中,Vol : 11, RSSI : 59
FM  76.8MHz 受信中,Vol : 11, RSSI : 60
*/
