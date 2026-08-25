// TinyGo コード（rda5807 初期化 → 76.8MHz 受信）
// tinygo build -target=m5stack -size=short -o Sample.uf2 .
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
	// 周波数設定:KHzで設定すること
	// 設定可能範囲	76000 – 108000 KHz

	// freq := 76800
	// freq := 79000
	freq := 88700
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
	// 受信帯域を設定する。
//	radio.InitRDA5807(rda5807.Band_World_Wide)
	radio.InitRDA5807(rda5807.Band_Japan)
//	radio.InitRDA5807_ORG(rda5807.Band_Japan)
	time.Sleep(500 * time.Millisecond)
	// 周波数を設定
	radio.SetFrequency(freq)
	time.Sleep(500 * time.Millisecond)
	radio.SetVolume(8)
	v, _ := radio.GetVolume()
	fmt.Printf("Volume %d\n", v)

	// プログラムが終了しないように無限ループで待機（ラジオは鳴り続けます）
	for {
		f, _ := radio.GetFrequency() // 現在、受信中の周波数を取得
		fmt.Printf("FM %3.1fMHz 受信中\n", (float64(f) / 1000.0))
		time.Sleep(time.Second * 3)
	}

}
