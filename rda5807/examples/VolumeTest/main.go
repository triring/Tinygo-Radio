// TinyGo コード（rda5807 初期化 → 76.8MHz 受信）
// tinygo build -target=m5stack -size=short -o VolumeTest.uf2 .
// tinygo flash -target=m5stack -size=short -monitor .
// tinygo flash -target=pico -size=short -monitor .

package main

import (
	"fmt"
	"machine"
	//	"rda5807" // ローカルのディレクトリに置かれたrda5807のパッケージをインポートする場合
	"github.com/triring/Tinygo-Radio/rda5807" // githubで公開しているパッケージをインポートする場合
	"time"
)

func main() {

	var (
		freq int
	)

	// 地元のFM局の送信周波数を変数freq に設定する。
	// 周波数は、KHzで設定すること
	// 設定可能範囲	76000 – 108000 KHz

	freq = 76800
	// freq := 79000
	// freq := 88700
	// freq := 91400

	// 音量設定データ
	// 設定可能範囲	0-15
	volume := [...]int{0, 3, 7, 15, 7, 3}

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
	// 受信帯域を設定する。
	radio.InitRDA5807(rda5807.Band_World_Wide)
	// 周波数を設定
	radio.SetFrequency(freq)
	time.Sleep(500 * time.Millisecond)
	for {
		for i, v := range volume {
			radio.SetVolume(uint8(v))
			vol, _ := radio.GetVolume() //	音量の取得
			fmt.Printf("%d. FM %5.1fMHz 受信中, Volume : %2d\n", i, (float64(freq) / 1000.0), vol)
			time.Sleep(time.Second * 5)
		}
	}
}

/*
> tinygo flash -target=pico -size=short -monitor .
   code    data     bss |   flash     ram
  65160    1564    5576 |   66724    7140
Connected to COM4. Press Ctrl-C to exit.
0. FM  76.8MHz 受信中, Volume :  0
1. FM  76.8MHz 受信中, Volume :  3
2. FM  76.8MHz 受信中, Volume :  7
3. FM  76.8MHz 受信中, Volume : 15
4. FM  76.8MHz 受信中, Volume :  7
5. FM  76.8MHz 受信中, Volume :  3
0. FM  76.8MHz 受信中, Volume :  0
1. FM  76.8MHz 受信中, Volume :  3
2. FM  76.8MHz 受信中, Volume :  7
3. FM  76.8MHz 受信中, Volume : 15
4. FM  76.8MHz 受信中, Volume :  7
5. FM  76.8MHz 受信中, Volume :  3
0. FM  76.8MHz 受信中, Volume :  0
*/
