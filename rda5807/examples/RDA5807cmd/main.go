// TinyGo コード（rda5807 初期化 → FM 受信）
// tinygo build -target=m5stack -size=short -o DRA5807.uf2 .
// tinygo flash -target=m5stack -size=short -monitor .
// tinygo flash -target=pico -size=short -monitor .

package main

import (
	"fmt"
	"machine"
	//	"rda5807" // ローカルのディレクトリに置かれたrda5807のパッケージをインポートする場合
	"github.com/triring/Tinygo-Radio/rda5807" // githubで公開しているパッケージをインポートする場合
	"strconv"
	"strings"
	"time"
)

var HelpText = [...]string{
	"Command list",
	"\tH             :(Help) コマンドの使用方法を表示する。",
	"\tF [frequency] :(Freq) 受信周波数を設定する。なお、周波数の設定は、KHzで行う。",
	"\t              :       例 77.7MHz の場合は、F 77700と入力する。",
	"\tF             :       引数がない場合は、現在、受信している周波数を表示する。",
	"\tV [volume]    :(Vol ) 音量を設定する。音量の設定範囲は、0 - 15",
	"\tV             :       引数がない場合は、現在、設定されている音量を表示する。",
	"\tM [state]     :(Mute) 無音の設定を行う。'M 0' 無音、'M 1' 有音",
	"\tM             :       引数がない場合は、現在、Muteの設定を表示する。",
	"\tR             :(RSSI) 信号強度を読み出す。信号の範囲 0-15",
	"\tQ             :(Quit) このプログラムを終了する。",
}

// TrimLastChar は文字列の最後のルーンを削除します
func TrimLastChar(s string) string {
	if s == "" {
		return ""
	}
	// 文字列をルーンのスライスに変換する
	runes := []rune(s)
	// スライスの最後の要素を除外して、新しい文字列として返す
	return string(runes[:len(runes)-1])
}

// getCommand はコンソールから文字列を受け取り、
// 不要な前後の空白を取り除いた文字列を返す。
func getCommand() string {
	enter_flag := false // 改行コードのチェック用フラグ
	var readbuffer string
	for { // キー入力待ち
		// PCからの受信データをチェック
		if machine.Serial.Buffered() > 0 {
			c, err := machine.Serial.ReadByte()
			if err == nil {
				if c < 32 {
					switch c {
					case '\r':
						enter_flag = true // machine.Serial.WriteByte('\r')
					case '\n':
						enter_flag = true // machine.Serial.WriteByte('\n')
					case '\b':
						if len(readbuffer) > 0 { // バックスペースで、最後尾の１文字を削除
							fmt.Printf("%c", '\b') // 表示部分の最後の1文字を消去
							fmt.Printf("%c", ' ')
							fmt.Printf("%c", '\b')
							/*
								machine.Serial.WriteByte('\b') // 表示部分の最後の1文字を消去
								machine.Serial.WriteByte(' ')
								machine.Serial.WriteByte('\b')
							*/
							readbuffer = TrimLastChar(readbuffer) // すでに取り込んでいる文字列データの最後の1文字を消去
						}
					default:
						// println(c)	  -- >  BS=8, Enter=13
						// Convert nonprintable control characters to
						// ^A, ^B, etc.
						fmt.Printf("%c%c%c", '^', c, '@')
						/*
							machine.Serial.WriteByte('^')
							machine.Serial.WriteByte(c + '@')
						*/
					}
				} else if c >= 127 {
					// Anything equal or above ASCII 127, print ^?.
					fmt.Printf("%c%c", '^', '?')
					/*
						machine.Serial.WriteByte('^')
						machine.Serial.WriteByte('?')
					*/
				} else {
					// Echo the printable character back to the
					// host computer.
					fmt.Printf("%c", c)
					/*
						machine.Serial.WriteByte(c)
					*/
					// 読み込んだ文字をエコーバックし、文字列バッファーに保存する。
					readbuffer = readbuffer + string(c)
				}
			}
		}
		// This assumes that the input is coming from a keyboard
		// so checking 120 times per second is sufficient. But if
		// the data comes from another processor, the port can
		// theoretically receive as much as 11000 bytes/second
		// (115200 baud). This delay can be removed and the
		// Serial.Read() method can be used to retrieve
		// multiple bytes from the receive buffer for each
		// iteration.
		if true == enter_flag {
			// 改行コードを検出したら、ループを抜け、取り込んだ文字列を返す。
			fmt.Printf("\n")
			break
		}
	}
	return readbuffer
}

func main() {
	time.Sleep(time.Millisecond * 1000)
	var readbuffer string
	fmt.Printf("RDA5807 command\n")

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
	// rda5807を初期化し、受信帯域を設定する。
	radio.InitRDA5807(rda5807.Band_World_Wide)
	time.Sleep(500 * time.Millisecond)
	// 周波数を設定
	radio.SetFrequency(radio.GetLowerFrequencyLimit())
	time.Sleep(500 * time.Millisecond)
	radio.SetVolume(7)
	// 初期設定完了

	execStatus := true
	for execStatus {
		fmt.Printf("> ")
		readbuffer = getCommand() // コンソールから文字列を受け取る。
		// コマンドの解析を開始
		if len(readbuffer) > 0 {
			line := strings.Replace(readbuffer, "\t", " ", -1) // タブをスペースに置換えて、区切り文字として使えるようにする。
			line = strings.Replace(line, ",", " ", -1)
			line = strings.ToUpper(line)
			line = strings.Trim(line, " \n\r")
			elements := strings.Split(line, " ")
			firstWord := line[0]
			// コマンド解析の開始
			switch firstWord {
			case 'H': //	ヘルプの表示(help)
				if len(elements) == 1 {
					for i := 0; i < len(HelpText); i++ {
						fmt.Printf("%s\n", HelpText[i])
					}
				}
			case 'F': //	現在の周波数の確認と、新しい周波数の設定
				if len(elements) == 1 {
					freq, _ := radio.GetFrequency()
					fmt.Printf("Freq   : %d\n", int(freq))
				} else if len(elements) == 2 {
					freq, err := strconv.ParseInt(strings.Trim(elements[1], " \n\r"), 0, 64)
					if err == nil {
						if freq < int64(radio.GetLowerFrequencyLimit()) || freq > int64(radio.GetUpperFrequencyLimit()) {
							fmt.Printf("The frequency setting range is from %d MHz to %d MHz.\n", radio.GetLowerFrequencyLimit(), radio.GetUpperFrequencyLimit())
						} else {
							fmt.Printf("Freq   : %d\n", int(freq))
							radio.SetFrequency(int(freq))
						}
						//	radio.SetMute(1)
					}
				}
			case 'V': //	現在の音量の確認と、新しい音量の設定
				if len(elements) == 1 {
					volume, _ := radio.GetVolume()
					fmt.Printf("Volume : %d\n", volume)
				} else if len(elements) == 2 {
					volume, err := strconv.ParseInt(strings.Trim(elements[1], " \n\r"), 0, 64)
					if err == nil {
						if volume < 0 || volume > 0b1111 {
							fmt.Printf("The volume setting range is from %d to %d.\n", 0, 15)
						} else {
							fmt.Printf("Volume : %d\n", volume)
							radio.SetVolume(uint8(volume))
						}
					}
				}
			case 'M': //	無音有音の設定できない。
				if len(elements) == 1 {
					mute, _ := radio.GetMute()
					fmt.Printf("Mute   : %t\n", mute)
				} else if len(elements) > 1 {
					//	数値変換
					m, err := strconv.ParseInt(strings.Trim(elements[1], " \n\r"), 0, 64)
					//	fmt.Printf("volume : %d\n", vol)
					if err == nil {
						if m == 0 {
							radio.SetMute(0)
						} else {
							radio.SetMute(1)
						}
						mute, _ := radio.GetMute()
						fmt.Printf("Mute   : %t\n", mute)
					}
				}
			case 'R': //	RSSI 信号強度を読み出す。信号の範囲 0-15"
				if len(elements) == 1 {
					rssi, _ := radio.GetRSSI()
					fmt.Printf("RSSI   : %2d\n", rssi)
				}
			case 'Q':
				radio.SetMute(0)
				execStatus = false // プログラムを終了する。
			}
		}

	}
	fmt.Printf("program terminated !\n")
	for {
		time.Sleep(time.Millisecond * 5000)
	}
}

/*
> tinygo flash -target=pico -size=short -monitor .
   code    data     bss |   flash     ram
  74960    1612    5576 |   76572    7188
Connected to COM23. Press Ctrl-C to exit.
RDA5807 command
Initialization of RDA5807 is complete.
Command list
        H           :(Help) コマンドの使用方法を表示する。
        F [address] :(Freq) 受信周波数を設定する。なお、周波数の設定は、KHzで行う。
                    :       例 77.7MHz の場合は、F 77700と入力する。
        F           :       引数がない場合は、現在、受信している周波数を表示する。
        V [volume]  :(Vol ) 音量を設定する。音量の設定範囲は、0 - 15
        V           :       引数がない場合は、現在、設定されている音量を表示する。
        M [state]   :(Mute) 無音の設定を行う。'M 0' 無音、'M 1' 有音
        M           :       引数がない場合は、現在、Muteの設定を表示する。
        R           :(RSSI) 信号強度を読み出す。信号の範囲 0-15
        Q           :(Quit) このプログラムを終了する。
> F
Freq   : 76000
> F 88700
Freq   : 88700
> V
Volume : 12
> V 15
Volume : 15
> R
RSSI   : 52
> M 0
Mute   : true
> M 1
Mute   : false
*/
