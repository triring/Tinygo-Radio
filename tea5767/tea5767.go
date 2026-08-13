package tea5767

/*
The TEA5767 is low-voltage, single-chip stereo FM receiver ICs designed for compact and portable devices.)
They integrate IF selectivity, demodulation, and PLL-based digital tuning, eliminating external filters and alignment.
The tuners support software-controlled tuning across European, U.S., and Japanese FM bands.
The power supply voltage range is 2.5 to 5 V, and the current consumption is approximately 13 mA, allowing for battery operation.
*/

import (
	//  "fmt"
	"machine"
	"time"
	"tinygo.org/x/drivers"
)

const DefaultAddress = 0x60

// Config は初期化設定
type Config struct {
	JapanBand bool // JPバンドを使うか
	Mute      bool // 初期ミュート
}

// Device は TEA5767 を表す
type Device struct {
	bus       drivers.I2C
	addr      uint8
	cfg       Config
	frequency int
}

// New はデバイスを生成する
// func New(bus i2c.Bus, addr uint8) Device {
func New(bus drivers.I2C, addr uint8) Device {
	return Device{
		bus:  bus,
		addr: addr,
		cfg: Config{
			JapanBand: true,
			Mute:      true,
		},
	}
}

// Configure は初期化設定を反映する
func (d *Device) Configure(cfg Config) {
	d.cfg = cfg
	// 初期化時は PLL を 0 にして設定だけ送る
	muteBit := byte(0)
	if d.cfg.Mute {
		muteBit = 0b10000000
	}
	jpBit := byte(0)
	if d.cfg.JapanBand {
		jpBit = 0b00100000
	}
	data := []byte{
		0x00,                 // PLL High
		0x00,                 // PLL Low
		muteBit | 0b00010000, // search off
		jpBit | 0b00010000,   // JP band + soft mute off
		0x00,
	}
	d.bus.Tx(uint16(d.addr), data, nil)
	// 書き出し処理が終わるまでの時間待ち
	time.Sleep(100 * time.Millisecond)
}

// inRange 指定した値の範囲にあるかを判別する。範囲内であればtrueを返す。
func inRange(min, value, max int) bool {
	return value >= min && value <= max
}

// Initialize the TEA5767.
//
// TEA5767を初期化する。起動後に
func (d *Device) InitTEA5767() {
	// 起動時はスタンバイ状態(STBY=1) ＆ ミュート(MUTE=1)で安全に初期化
	initData := []byte{0x80, 0x00, 0xB0, 0x50, 0x00}
	machine.I2C0.Tx(DefaultAddress, initData, nil)
	time.Sleep(100 * time.Millisecond)
}

// SetFrequency は MHz を指定して周波数を設定する
func (d *Device) GetFrequency() int {
	return d.frequency
}

// tuneRaw は指定した設定でTEA5767にデータを送信します
func tuneRaw(freqHz float64, hlsi int, bl byte, mute bool) {
	// HLSI (High/Low Side Injection) の判定
	sign := 1.0
	byte3 := byte(0b10100000) // HLSI=0 (Low Side) のベース
	if hlsi == 1 {
		sign = 1.0
		byte3 = 0b10110000 // HLSI=1 (High Side) に変更
	} else {
		sign = -1.0
	}

	// PLL値の計算 (High/Low サイドを動的に切り替え)
	pllFloat := 4.0 * (freqHz + sign*225000.0) / 32768.0
	pll := uint16(pllFloat)

	// 1バイト目のMUTE処理
	byte1 := byte((pll >> 8) & 0x3F)
	if mute {
		byte1 |= 0x80 // bit 7 を 1 にしてミュート
	}

	data := []byte{
		byte1,
		byte(pll & 0xFF),
		byte3,
		bl,
		0x00,
	}
	machine.I2C0.Tx(DefaultAddress, data, nil)
	time.Sleep(100 * time.Millisecond) // チューニングと内部安定化のための待機
}

// GetRSSI は現在の信号強度(0-15)を読み出します
func GetRSSI() uint8 {
	var readBuf [5]byte
	machine.I2C0.Tx(DefaultAddress, nil, readBuf[:])
	return readBuf[3] >> 4
}

// SetFrequency は指定した周波数にチューニングする。
//
// High/Low Sideの最適化をしてからチューニングを行う。
func (d *Device) SetFrequency(freqKHz int) (uint8, string) {
	if false == inRange(76000, freqKHz, 99000) {
		// 周波数範囲外であれば、エラーとして、falseを返す。
		return 0, "Outside frequency range."
	}
	d.frequency = freqKHz
	freqHz := float64(freqKHz * 1000)
	// 1. 周波数に応じたバンド(BL)の設定
	var bl byte
	if freqHz < 87500000.0 {
		bl = 0b00110000 // JPモード (BL=1, XTAL=1)
	} else {
		bl = 0b00010000 // US/EUモード (BL=0, XTAL=1)
	}

	// 2. High Side (+0.45MHz) のノイズレベルをチェック (ミュート状態)
	tuneRaw(freqHz+450000.0, 1, bl, true)
	signalHigh := d.GetRSSI()

	// 3. Low Side (-0.45MHz) のノイズレベルをチェック (ミュート状態)
	tuneRaw(freqHz-450000.0, 0, bl, true)
	signalLow := d.GetRSSI()

	// 4. ノイズが少ない(電波が弱い)方を採用する
	hlsi := 1
	injectionMode := "High Side"
	if signalHigh < signalLow {
		hlsi = 1 // +0.45MHzの方が空いている
	} else {
		hlsi = 0 // -0.45MHzの方が空いている
		injectionMode = "Low Side"
	}

	// 5. 決定した最適な設定で本命の周波数にチューニング (ミュート解除)
	//	tuneRaw(freqHz, hlsi, bl, d.cfg.Mute)
	//	tuneRaw(freqHz, hlsi, bl, true)
	tuneRaw(freqHz, hlsi, bl, d.cfg.Mute)

	rssi := d.GetRSSI() // 現在の信号強度(0-15)を取得する。
	//	for debug
	//	fmt.Printf(" -> Optimized with %s (High: %d, Low: %d, Current rssi: %d)\n", injectionMode, signalHigh, signalLow, rssi)
	return rssi, injectionMode
}

// SetMute はミュート ON/OFF
func (d *Device) SetMute(state bool) {
	d.cfg.Mute = state
	d.SetFrequency(d.frequency) // PLL 0 で制御バイトだけ送る
}

// GetMute はミュートの状態を返す。
func (d *Device) GetMute() bool {
	return d.cfg.Mute
}

// -----------------------------
// Status 取得
// -----------------------------
func (d *Device) GetStatus() (rssi uint8, ifCounter uint8, stereo bool) {
	buf := make([]byte, 5)
	d.bus.Tx(uint16(d.addr), nil, buf)
	// 読み込み処理が終わるまでの時間待ち
	time.Sleep(100 * time.Millisecond)

	ifCounter = buf[0] & 0x7F
	// rssi = buf[1] & 0x7F
	// 4バイト目(index:3)の上位4ビットがRSSI
	rssi = buf[3] >> 4
	stereo = (buf[2] & 0x80) != 0
	return
}

// RSSI の取得
//
// 現在の信号強度(0-15)を読み出す
func (d *Device) GetRSSI() (rssi uint8) {
	buf := make([]byte, 5)
	d.bus.Tx(uint16(d.addr), nil, buf)
	// 読み込み処理が終わるまでの時間待ち
	time.Sleep(100 * time.Millisecond)
	rssi = buf[3] >> 4
	return
}

// -----------------------------
// 自動スキャン（上方向）
// -----------------------------
func (d *Device) ScanUp() float64 {
	return d.scan(true)
}

// -----------------------------
// 自動スキャン（下方向）
// -----------------------------
func (d *Device) ScanDown() float64 {
	return d.scan(false)
}

// 内部スキャン処理
func (d *Device) scan(up bool) float64 {
	// 現在のステータスを読む
	buf := make([]byte, 5)
	d.bus.Tx(uint16(d.addr), nil, buf)
	// 読み込み処理が終わるまでの時間待ち
	time.Sleep(100 * time.Millisecond)

	// 現在の PLL を取得
	pll := uint16(buf[0]&0x3F)<<8 | uint16(buf[1])

	// スキャン方向
	searchMode := byte(0)
	if up {
		searchMode = 0b01000000
	} else {
		searchMode = 0b10000000
	}

	jpBit := byte(0)
	if d.cfg.JapanBand {
		jpBit = 0b00100000
	}

	muteBit := byte(0)
	if d.cfg.Mute {
		muteBit = 0b10000000
	}

	// スキャン開始
	data := []byte{
		byte(pll >> 8),
		byte(pll & 0xFF),
		muteBit | searchMode | 0b00110000,
		0b10110000, // HLSI=1
		jpBit | 0b00010000,
		0x00,
	}
	d.bus.Tx(uint16(d.addr), data, nil)
	// 書き込み処理が終わるまでの時間待ち
	time.Sleep(50 * time.Millisecond)

	// スキャン結果を待つ
	for {
		d.bus.Tx(uint16(d.addr), nil, buf)
		// 読み込み処理が終わるまでの時間待ち
		time.Sleep(100 * time.Millisecond)
		ifCounter := buf[0] & 0x7F

		// IF Counter が 0x00〜0x1F なら局を発見
		if ifCounter < 0x20 {
			break
		}
	}

	// 最終周波数を計算して返す
	finalPLL := uint16(buf[0]&0x3F)<<8 | uint16(buf[1])
	freq := (float64(finalPLL)*32768.0/4.0 - 225000.0) / 1000000.0

	return freq
}
