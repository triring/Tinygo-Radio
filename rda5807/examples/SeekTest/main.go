// TinyGo コード（rda5807 初期化 → 76.8MHz 受信）
// tinygo build -target=m5stack -size=short -o SeekTest.uf2 .
// tinygo flash -target=m5stack -size=short -monitor .
// tinygo build -target=pico -size=short -o SeekTest.uf2 .
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
	threshold := 35 // 電波強度の閾値
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
	time.Sleep(500 * time.Millisecond)
	// CHIPIDの確認
	chipid, _ := radio.GetChipID()
	fmt.Printf("CHIPID   = 0x%04X\r\n", chipid)

	// 音量。
	if err := radio.SetVolume(12); err != nil {
		fmt.Printf("SetVolume error: %v\r\n", err)
	}

	// Unmute。
	if err := radio.SetMute(false); err != nil {
		fmt.Printf("SetMute error: %v\r\n", err)
	}

	lower_limit := radio.GetLowerFrequencyLimit()
	upper_limit := radio.GetUpperFrequencyLimit()
	// 受信下限周波数に設定する。
	if err := radio.SetFrequency(lower_limit); err != nil {
		fmt.Printf("SetFrequency error: %v\r\n", err)
	}
	for {
		for {
			// Seek Up.
			fmt.Println("Seek Up...")
			frequency, err := radio.SeekUp(uint8(threshold))
			if err != nil {
				fmt.Printf("Seek error: %v\r\n", err)
				fmt.Printf("Upper frequency limit: %v\r\n", upper_limit)
				break
			} else {
				rssi, _ := radio.GetRSSI()
				volume, _ := radio.GetVolume()
				frequency, err = radio.GetFrequency()
				fmt.Printf(
					"Frequency= %3d.%03d MHz, RSSI : Threshold = %d : %d dBuV, Volume=%d\r\n",
					frequency/1000,
					frequency%1000,
					rssi,
					threshold,
					volume,
				)
			}
			time.Sleep(5 * time.Second)
		}
		for {
			// Seek Down.
			fmt.Println("Seek Down...")
			frequency, err := radio.SeekDown(uint8(threshold))
			if err != nil {
				fmt.Printf("Seek error: %v\r\n", err)
				fmt.Printf("Lower frequency limit: %v\r\n", lower_limit)
				break
			} else {
				rssi, _ := radio.GetRSSI()
				volume, _ := radio.GetVolume()
				frequency, err = radio.GetFrequency()
				fmt.Printf(
					"Frequency= %3d.%03d MHz, RSSI : Threshold = %d : %d dBuV, Volume=%d\r\n",
					frequency/1000,
					frequency%1000,
					rssi,
					threshold,
					volume,
				)
			}
			time.Sleep(5 * time.Second)
		}
	}
}

/*
> tinygo flash -target=pico -size=short -monitor .
   code    data     bss |   flash     ram
  65240    1564    5576 |   66804    7140
Connected to COM4. Press Ctrl-C to exit.
FM  76.8MHz 受信中 : Unmute
FM  76.8MHz 受信中 :   Mute
FM  76.8MHz 受信中 : Unmute
FM  76.8MHz 受信中 :   Mute
FM  76.8MHz 受信中 : Unmute
FM  76.8MHz 受信中 :   Mute
FM  76.8MHz 受信中 : Unmute
FM  76.8MHz 受信中 :   Mute
*/
