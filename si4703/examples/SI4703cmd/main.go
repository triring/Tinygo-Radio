// TinyGo コード（si4703 初期化 → コマンドモード）
// tinygo build -target=m5stack -size=short -o SI4703cmd.uf2 .
// tinygo flash -target=m5stack -size=short -monitor .
// tinygo build -target=pico -size=short -o SI4703cmd.uf2 .
// tinygo flash -target=pico -size=short -monitor .

package main

import (
	"fmt"
	"machine"
	// "si4703" // ローカルのディレクトリに置かれたrda5807のパッケージをインポートする場合
	"github.com/triring/Tinygo-Radio/si4703" // githubで公開しているパッケージをインポートする場合
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
	"\tU             :(Up)   現在設定されている周波数から上方に向かって放送波を検索する。",
	"\tU [rssi]      :(Up)   閾値として電波強度[RSSI(0-63)]を設定。それ以上の信号を検出すると自動停止する。",
	"\tD             :(Down) 現在設定されている周波数から下方に向かって放送波を検索する。",
	"\tD [rssi]      :(Down) 閾値として電波強度[RSSI(0-63)]を設定。それ以上の信号を検出すると自動停止する。",
//	"\tS             :(scan) 電波を探しながら5秒ずつ次々と番組を切り替えていく",
	"\tV [volume]    :(Vol ) 音量を設定する。音量の設定範囲は、0 - 15",
	"\tV             :       引数がない場合は、現在、設定されている音量を表示する。",
	"\tM [state]     :(Mute) 無音の設定を行う。'M 0' 無音、'M 1' 有音",
	"\tM             :       引数がない場合は、現在、Muteの設定を表示する。",
	"\tR             :(RSSI) 信号強度を読み出す。信号の範囲 0-63",
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
	fmt.Printf("Si4703 command\n")
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
					freq := radio.GetFrequency()
					fmt.Printf("Freq   : %d\n", int(freq))
				} else if len(elements) == 2 {
					freq, err := strconv.ParseInt(strings.Trim(elements[1], " \n\r"), 0, 64)
					if err == nil {
						lower_frequency_limit, upper_frequency_limit, _, _ := radio.BandParameters()
						if (int(freq) < lower_frequency_limit || int(freq) > upper_frequency_limit) {
							fmt.Printf("The frequency setting range is from %d MHz to %d MHz.\n", lower_frequency_limit, upper_frequency_limit)
						} else {
							fmt.Printf("Freq   : %d\n", int(freq))
							radio.SetFrequency(int(freq))
						}
						//	radio.SetMute(false)
					}
				}
			case 'U': //	現在設定されている周波数から上方に向かって放送波を検索する。
				threshold := 32	// 受信電波強度(rssi)のデフォルト閾値
				if len(elements) == 1 {
					// 受信する電波強度の閾値をデフォルト値に設定する。
					if err := radio.SetSeekThreshold(uint8(threshold)); err != nil {
						fmt.Printf("SetSeekThreshold error: %v\r\n", err)
					}
					// Seek Up.
					fmt.Println("Seek Up...")
					freq, err := radio.Seek(si4703.SeekUp, si4703.SeekWrap)
					if err != nil {
						fmt.Printf("Seek error: %v\r\n", err)
					} else {
						fmt.Printf("Found %d.%03d MHz\r\n", freq/1000, freq%1000)
					}
					rssi, _ := radio.GetRSSI()
					fmt.Printf("RSSI : threshold =  %2d : %2d\r\n", rssi, threshold)
				} else if len(elements) == 2 {
					rssi_threshold, err := strconv.ParseInt(strings.Trim(elements[1], " \n\r"), 0, 64)
					if err == nil {
						// 受信する電波強度の閾値を40に設定する。
						if err := radio.SetSeekThreshold(uint8(rssi_threshold)); err != nil {
							fmt.Printf("SetSeekThreshold error: %v\r\n", err)
						}
						// Seek Up.
						fmt.Println("Seek Up...")
						freq, err := radio.Seek(si4703.SeekUp, si4703.SeekWrap)
						if err != nil {
							fmt.Printf("Seek error: %v\r\n", err)
						} else {
							fmt.Printf("Found %d.%03d MHz\r\n", freq/1000, freq%1000)
						}
						rssi, _ := radio.GetRSSI()
						fmt.Printf("RSSI : threshold =  %2d : %2d\r\n", rssi, rssi_threshold)
					}
				}
			case 'D': //	現在設定されている周波数から下方に向かって放送波を検索する。
				threshold := 32	// 受信電波強度(rssi)のデフォルト閾値
				if len(elements) == 1 {
					// 受信する電波強度の閾値をデフォルト値に設定する。
					if err := radio.SetSeekThreshold(uint8(threshold)); err != nil {
						fmt.Printf("SetSeekThreshold error: %v\r\n", err)
					}
					// Seek Down.
					fmt.Println("Seek Down...")
					freq, err := radio.Seek(si4703.SeekDown, si4703.SeekWrap)
					if err != nil {
						fmt.Printf("Seek error: %v\r\n", err)
					} else {
						fmt.Printf("Found %d.%03d MHz\r\n", freq/1000, freq%1000)
					}
					rssi, _ := radio.GetRSSI()
					fmt.Printf("RSSI : threshold =  %2d : %2d\r\n", rssi, threshold)
				} else if len(elements) == 2 {
					rssi_threshold, err := strconv.ParseInt(strings.Trim(elements[1], " \n\r"), 0, 64)
					if err == nil {
						// 受信する電波強度の閾値を40に設定する。
						if err := radio.SetSeekThreshold(uint8(rssi_threshold)); err != nil {
							fmt.Printf("SetSeekThreshold error: %v\r\n", err)
						}
						// Seek Down.
						fmt.Println("Seek Down...")
						freq, err := radio.Seek(si4703.SeekDown, si4703.SeekWrap)
						if err != nil {
							fmt.Printf("Seek error: %v\r\n", err)
						} else {
							fmt.Printf("Found %d.%03d MHz\r\n", freq/1000, freq%1000)
						}
						rssi, _ := radio.GetRSSI()
						fmt.Printf("RSSI : threshold =  %2d : %2d\r\n", rssi, rssi_threshold)
					}
				}
			case 'V': //	現在の音量の確認と、新しい音量の設定
				if len(elements) == 1 {
					volume := radio.GetVolume()
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
					mute := radio.GetMute()
					fmt.Printf("Mute   : %t\n", mute)
				} else if len(elements) == 2 {
					//	数値変換
					m, err := strconv.ParseInt(strings.Trim(elements[1], " \n\r"), 0, 64)
					//	fmt.Printf("volume : %d\n", vol)
					if err == nil {
						if m == 0 {
							radio.SetMute(true)
						} else {
							radio.SetMute(false)
						}
						mute := radio.GetMute()
						fmt.Printf("Mute   : %t\n", mute)
					}
				}
			case 'R': //	RSSI 信号強度を読み出す。信号の範囲 0-15"
				if len(elements) == 1 {
					rssi, _ := radio.GetRSSI()
					fmt.Printf("RSSI   : %2d\n", rssi)
				}
			case 'Q':
				radio.SetMute(true)
				execStatus = false // プログラムを終了する。
			}
		}

	}
	fmt.Printf("program terminated !\n")
	for {
		time.Sleep(time.Millisecond * 5000)
	}
}
