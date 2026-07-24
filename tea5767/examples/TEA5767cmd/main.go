// TinyGo コード（TEA5767 初期化 → 76.8MHz 受信）
// tinygo build -target=m5stack -size=short -o fm_radio.uf2 .
// tinygo flash -target=m5stack -size=short -monitor .
// tinygo flash -target=pico -size=short -monitor .
// ソフトウェアスキャンに対応

package main

import (
	//	"bufio"
	"fmt"
	//	"os"
	//	"strconv"
	"machine"
	"strconv"
	"strings"
	//	"tea5767" // ローカルのディレクトリに置かれたtea5767のパッケージをインポートする場合
	"github.com/triring/Tinygo-Radio/tea5767" // githubで公開しているパッケージをインポートする場合
	"time"
)

var (
	FREQ_MIN int = 76000
	//	FREQ_MAX int = 99000
	FREQ_MAX int = 87500 // 受信帯域をUS/EUモードに切り替えると、
	// フリーズ（I2C通信が設定できなくなる）してしまうバグがあるので、
	// 受信帯域をJPモードに制限する。
)

var HelpText = [...]string{
	"Command list",
	"\tH           :(Help) コマンドの使用方法を表示する。",
	"\tF [address] :(Freq) 受信周波数を設定する。なお、周波数の設定は、KHzで行う。",
	"\t            :       例 77.7MHz の場合は、F 77700と入力する。",
	"\tF           :       引数がない場合は、現在、受信している周波数を表示する。",
	"\tM [state]   :(Mute) 無音の設定を行う。'M 0' 無音、'M 1' 有音",
	"\tM           :       引数がない場合は、現在、Muteの設定を表示する。",
	"\tR           :(RSSI) 信号強度を読み出す。信号の範囲 0-15",
	"\tQ           :(Quit) このプログラムを終了する。",
}

// inRange 指定した値の範囲にあるかを判別する。範囲内であればtrueを返す。
func inRange(min, value, max uint8) bool {
	return value >= min && value <= max
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
	var readbuffer string
	fmt.Printf("TEA5767 command\n")
	time.Sleep(time.Millisecond * 2000)
	machine.I2C0.Configure(machine.I2CConfig{
		Frequency: 100 * machine.KHz,
		SDA:       machine.GPIO12, // for zero-kb02(raspi pico)
		SCL:       machine.GPIO13, // for zero-kb02(raspi pico)
	})
	time.Sleep(50 * time.Millisecond)
	radio := tea5767.New(machine.I2C0, tea5767.DefaultAddress)

	radio.Configure(tea5767.Config{
		JapanBand: true,
		Mute:      false,
	})

	radio.TuneFrequency(76000) // KHzで設定すること
	radio.SetMute(false)       // 最初は出力を切っておく。
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
			/*
				実装予定
				Xコマンド	レジスタ、カウンタ、フラグ類の検査と変更
			*/
			case 'H': //	ヘルプの表示(help)
				if len(elements) == 1 {
					for i := 0; i < len(HelpText); i++ {
						fmt.Printf("%s\n", HelpText[i])
					}
				}
			case 'F': //	メモリの指定されたアドレスに値を書き込む。
				if len(elements) == 1 {
					freq := radio.GetFrequency()
					fmt.Printf("FM %3.1f\n", float64(freq)/1000.0)
				} else if len(elements) == 2 {
					freq, err := strconv.ParseInt(strings.Trim(elements[1], " \n\r"), 0, 64)
					if err == nil {
						fmt.Printf("freq   : %d\n", int(freq))
						radio.TuneFrequency(int(freq))
						radio.SetMute(false)
					}
				}
			case 'M': //	音量設定	TEA5767は、音量設定がなく、無音と有音のみなので、０か１だけしか設定できない。
				if len(elements) == 1 {
					fmt.Printf("Mute=%t\n", radio.GetMute())
				} else if len(elements) > 1 {
					//	数値変換
					m, err := strconv.ParseInt(strings.Trim(elements[1], " \n\r"), 0, 64)
					//	fmt.Printf("volume : %d\n", vol)
					if err == nil {
						if m == 0 {
							radio.SetMute(true)
						} else {
							radio.SetMute(false)
						}
						fmt.Printf("Mute=%t\n", radio.GetMute())
					}
				}

			case 'R': //	RSSI 信号強度を読み出す。信号の範囲 0-15"
				if len(elements) == 1 {
					fmt.Printf("RSSI=%2d\n", radio.GetRSSI())
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
