// TinyGo コード（rda5807 初期化 → 76.8MHz 受信）
// tinygo build -target=m5stack -size=short -o GetRSSI.uf2 .
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
	time.Sleep(500 * time.Millisecond)
	// rda5807のオブジェクト生成
	// 第1引数: 使用するI2Cチャンネル
	radio := rda5807.New(i2c)
	time.Sleep(500 * time.Millisecond)
	// rda5807を初期化
	radio.InitRDA5807(rda5807.Band_World_Wide)
	time.Sleep(500 * time.Millisecond)
	// 周波数を設定
	radio.SetFrequency(freq)
	time.Sleep(500 * time.Millisecond)

	for {
		rssi, err := radio.GetRSSI()
		if err != nil {
			fmt.Printf("GetRSSI() error %s\n", err)
		}
		fmt.Printf("FM %5.1fMHz 受信中,RSSI: %2d\n", (float64(freq) / 1000.0), rssi)
		time.Sleep(time.Second * 5)
	}
}

/*
> tinygo flash -target=pico -size=short -monitor .
   code    data     bss |   flash     ram
  63648    1596    5576 |   65244    7172
Connected to COM4. Press Ctrl-C to exit.
Initialization of RDA5807 is complete.
FM 76.8MHz 受信中,RSSI: 59
FM 76.8MHz 受信中,RSSI: 60
FM 76.8MHz 受信中,RSSI: 61
FM 76.8MHz 受信中,RSSI: 60
FM 76.8MHz 受信中,RSSI: 60
FM 76.8MHz 受信中,RSSI: 60
FM 76.8MHz 受信中,RSSI: 60
FM 76.8MHz 受信中,RSSI: 60
FM 76.8MHz 受信中,RSSI: 58
FM 76.8MHz 受信中,RSSI: 58
FM 76.8MHz 受信中,RSSI: 58
FM 76.8MHz 受信中,RSSI: 58
FM 76.8MHz 受信中,RSSI: 59
*/
