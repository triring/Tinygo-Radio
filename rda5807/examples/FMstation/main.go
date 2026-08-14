// TinyGo コード（rda5807 初期化 → 76.8MHz 受信）
// tinygo build -target=m5stack -size=short -o FMstation.uf2 .
// tinygo flash -target=m5stack -size=short -monitor .
// tinygo flash -target=pico -size=short -monitor .

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
		rssi byte
		freq int
	)
	// 地元で受信可能なFM局の周波数を以下の配列に登録して下さい。
	// 登録された周波数を順番に受信していきます。
	// 周波数設定:KHzで設定すること
	// 設定可能範囲	76000 – 108000 KHz
	station := [...]int{76800, 79000, 88700, 91400}

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
	// 受信帯域を設定する。
	radio.InitRDA5807(rda5807.Band_World_Wide)
	time.Sleep(500 * time.Millisecond)
	// 周波数を設定
	time.Sleep(500 * time.Millisecond)
	radio.SetVolume(12)
	v, _ := radio.GetVolume()
	fmt.Printf("Volume %d\n", v)

	for {
		for i, f := range station {
			radio.SetFrequency(f)
			freq, _ = radio.GetFrequency() //	受信中の周波数を取得
			rssi, _ = radio.GetRSSI()      //	電波強度を取得
			fmt.Printf("%d. FM %3.1fMHz 受信中, RSSI : %2d\n", i, (float64(freq) / 1000.0), rssi)
			time.Sleep(time.Second * 5)
		}
	}
}

/*
> tinygo flash -target=pico -size=short -monitor .
   code    data     bss |   flash     ram
  65564    1596    5576 |   67160    7172
Connected to COM23. Press Ctrl-C to exit.
Initialization of RDA5807 is complete.
Volume 8
0. FM 76.8MHz 受信中, RSSI : 16
1. FM 79.0MHz 受信中, RSSI : 13
2. FM 88.7MHz 受信中, RSSI : 16
3. FM 91.4MHz 受信中, RSSI : 19
0. FM 76.8MHz 受信中, RSSI : 18
1. FM 79.0MHz 受信中, RSSI : 14
2. FM 88.7MHz 受信中, RSSI : 16
3. FM 91.4MHz 受信中, RSSI : 19
*/
