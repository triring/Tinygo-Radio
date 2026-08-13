package rda5807

/*
The RDA5807 is a highly integrated single-chip CMOS FM stereo radio tuner from RDA Microelectronics.
It supports the worldwide FM band (50–115 MHz) with flexible channel spacing (25/50/100/200 kHz).
Key features include a digital low-IF architecture, fully integrated synthesizer, AGC, RDS/RBDS support, autonomous seek tuning, soft mute, high-cut, bass boost, and programmable de-emphasis (50/75 µs).
It offers stereo/mono output, volume control, mute, and RSSI indication.
Operating at 1.8–3.3 V with low power consumption (typically under 20 mA), it requires minimal external components and uses a simple I2C interface, making it ideal for portable devices, MP3 players, and microcontroller projects.
*/

// TinyGo で RDA5807M を I2C 制御し、
// 日本向けWideFM (76–108MHz) バンドに設定してFM局の放送を受信するサンプル。
import (
	"fmt"
	//	"machine"
	"time"
	"tinygo.org/x/drivers"
)

// RDA5807M の I2C アドレス
const (
	rdaSeqAddr    = 0x10 // シーケンシャル書き込み用アドレス（初期化で使用）
	rdaRandomAddr = 0x11 // レジスタ指定書き込み・読み出し用アドレス
)

// Band Select.
const (
	Band_US_Europe   = 0b00000000 // 87 – 108 MHz (US/Europe)
	Band_Japan       = 0b00000100 // 76 –  91 MHz (Japan)
	Band_World_Wide  = 0b00001000 // 76 – 108 MHz (world wide)
	Band_East_Europe = 0b00001100 // 65 –  76 MHz （East Europe） or 50-65MHz
)

type BandData struct {
	BandName string
	Min      int
	Max      int
}

// 受信帯域の宣言と初期値の設定	受信地域名、最小受信周波数、最大受信周波数
var band_data [4]BandData = [4]BandData{
	{BandName: "US/Europe", Min: 87000, Max: 10800},
	{BandName: "Japan", Min: 76000, Max: 91000},
	{BandName: "World wide", Min: 76000, Max: 108000},
	{BandName: "East Europe", Min: 65000, Max: 76000},
}

// Device は RDA5807 を表す構造体
type Device struct {
	bus        drivers.I2C
	band       byte
	min_freq   int
	max_freq   int
	band_name  string
	mute_state bool
	// addr      uint8
	// cfg       Config
	// frequency int
}

// New はデバイスを生成する
// func New(bus i2c.Bus, addr uint8) Device {
// func New(bus drivers.I2C, addr uint8) Device {
func New(bus drivers.I2C) Device {

	return Device{
		bus: bus,
	}
}

// RDA5807のChip ID.を取得する。
func (d *Device) GetChipID() (uint16, error) {
	id, err := d.i2cReadReg(0x00)
	return id, err
}

// RDA5807のバンド設定情報を取得する。
func (d *Device) GetBandInfo() (byte, string, int, int) {
	return d.band, d.band_name, d.min_freq, d.max_freq
}

// inRange 指定した値の範囲にあるかを判別する。範囲内であればtrueを返す。
func inRange(min, value, max int) bool {
	return value >= min && value <= max
}

// 放送帯域の設定
func (d *Device) SetBroadcastBand(band byte) {
	d.band = band
	bs := d.band >> 2
	//	fmt.Printf("d.band = %b 0x%02x (%T)\n", d.band, d.band, d.band)
	d.band_name = band_data[bs].BandName
	d.min_freq = band_data[bs].Min
	d.max_freq = band_data[bs].Max
	// fmt.Printf("band [%s] freq %d - %d \n", d.band_name, d.min_freq, d.max_freq)
}

// Initialize the RDA5807
//
// RDA5807M 初期化シーケンス
// シーケンシャル書き込みでは、アドレス 0x10 に対して
//
//	[REG0x02_H, REG0x02_L, REG0x03_H, REG0x03_L, REG0x04_H, REG0x04_L, REG0x05_H, REG0x05_L]の順で書き込む。
//
// func (d *Device) InitRDA5807(i2c *machine.I2C, band byte) error {
func (d *Device) InitRDA5807(band byte) error {
	d.SetBroadcastBand(band) // 受信帯域の設定
	d.mute_state = false     // ミュート停止状態にしておく。
	// 以下は、RDA5807の内部レジスタの初期設定
	// REG 0x02: CONFIG
	//   bit 15: DHIZ (1 = Audio output high impedance disable)
	//   bit 14: DMUTE (1 = Normal operation / unmuted)
	//   bit 13: MONO (0 = Stereo)
	//   bit 12: BASS (0 = Bass boost disable)
	//   bit 11: RCLK_DIRECT (0)
	//   bit 10: SEEKUP (0)
	//   bit 9:  SEEK (0)
	//   bit 8:  SKMODE (0)
	//   bit 7:  CLK_MODE (000 = 32.768kHz)
	//   bit 4:  RDS_EN (0 = disable for stability)
	//   bit 1:  ENABLE (1 = Power up)
	//   bit 0:  SOFT_RESET (1 = Reset on init, then we send normal run config)

	// 0xC0 0x01 = 1100 0000 0000 0001
	reg02H := byte(0xC0)
	reg02L := byte(0x01)

	// REG 0x03: TUNING
	//   15:6 CHAN[9:0] = 0 : チャンネル番号（後で周波数設定時に書き換える）
	//    5 DIRECT_MODE = 0 : 通常モード
	//    4 TUNE        = 0 : チューニング開始フラグ（初期化時は必ず 0）
	//   3:2 BAND       = 10: 76–108MHz (WideFM含む)
	//   1:0 SPACE      = 00: 100kHz ステップ
	//
	// 初期化時は TUNE=0 にしておくことが重要。
	reg03H := byte(0x00)   // 0000 0000 : chan = 0000_0000
	reg03L := byte(d.band) // 0000 1000 : chan = 000,DIRECT_MODE = 0,TUNE=0, BAND=10, SPACE=00
	// 	reg03L := byte(0x08) // 0000 1000 : chan = 000,DIRECT_MODE = 0,TUNE=0, BAND=10, SPACE=00
	/*
	   Band Select.
	   00 = 87–108 MHz (US/Europe)
	   01 = 76–91 MHz (Japan)
	   10 = 76–108 MHz (world wide)
	   11 = 65 –76 MHz （East Europe） or 50-65MHz
	*/
	// REG 0x04: AUDIO / DEEMPHASIS など
	//   11 DE = 0 : 50us ディエンファシス（日本向け）
	//   他はデフォルト寄り
	reg04H := byte(0x0C)
	reg04L := byte(0x00)

	// REG 0x05: VOLUME / その他
	//   15:8  0x88 : 一般的なオーディオ設定
	//   3:0  VOLUME (0〜15)
	//
	// 初期化時の VOLUME は仮値。実際の音量は後で SetVolume() で設定し直す。
	reg05H := byte(0x88)
	reg05L := byte(0x8B)

	initData := []byte{
		reg02H, reg02L,
		reg03H, reg03L,
		reg04H, reg04L,
		reg05H, reg05L,
	}
	fmt.Printf("Initialization of RDA5807 is complete.\n")
	err := d.bus.Tx(rdaSeqAddr, initData, nil)
	// 初期設定の完了待ち
	time.Sleep(100 * time.Millisecond)
	return err
}

// 2バイトレジスタ書き込みヘルパ
func (d *Device) i2cWriteReg(reg byte, data uint16) error {
	high := byte(data >> 8)
	low := byte(data & 0xFF)
	buf := []byte{reg, high, low}
	return d.bus.Tx(rdaRandomAddr, buf, nil)
}

// 2バイトレジスタ読み出しヘルパ
func (d *Device) i2cReadReg(reg byte) (uint16, error) {
	write := []byte{reg}
	read := make([]byte, 2)
	if err := d.bus.Tx(rdaRandomAddr, write, read); err != nil {
		return 0, err
	}
	val := (uint16(read[0]) << 8) | uint16(read[1])
	return val, nil
}

// 現在の受信周波数取得 (KHz)
//
// ステータスレジスタ 0x0A には、現在のチャンネル情報が入っている。
//
//	val = read 0x0A (16bit)
//	frequency_b = ((val >> 8) & 0x03) << 8 | (val & 0x00FF)
//	frequency_khz = int(freqB * 100) + 76000
func (d *Device) GetFrequency() (int, error) {
	val, err := d.i2cReadReg(0x0A)
	if err != nil {
		return 0, err
	}
	//          	  fedc_ba98_7654_3210
	freqB := (val & 0b0000_0011_1111_1111)
	//	freqB := (((val >> 8) & 0x0003) << 8) | (val & 0x00FF)
	freqKHz := int(freqB*100) + d.min_freq
	return freqKHz, nil
}

// TuneFrequency は指定した周波数にチューニングする。
//
//	chan = int(freqKHz - 76000) / 100
//	h3 = chan >> 2
//	l3 = chan & 0b00000011
//	l3 = l3 << 6
//	l3 = l3 | 0b00011000  // BAND=10, SPACE=00, TUNE=1
//	fdata = (h3 << 8) + l3
//	send_data(0x03, fdata)
func (d *Device) SetFrequency(freqKHz int) error {
	if freqKHz < d.min_freq || freqKHz > d.max_freq {
		return nil
	}
	//	chanVal := int(freqKHz-76000) / 100
	chanVal := int(freqKHz-d.min_freq) / 100
	//	fmt.Printf("freq:%d,%d\n", freqKHz, chanVal)

	h3 := chanVal >> 2
	l3 := byte(chanVal & 0b00000011)
	l3 = l3 << 6
	// 0b00011000:
	//   BAND=10 (76–108MHz)
	//   SPACE=00 (100kHz)
	//   TUNE=1  (チューニング開始)
	//	fmt.Printf("%T,%b,%T,%b\n", l3, l3, d.band, d.band)
	l3 = l3 | byte(0b0001_0000) | d.band
	//	fmt.Printf("%T,%b,%T,%b\n", l3, l3, d.band, d.band)
	//	l3 = l3 | 0b00010000 | int(d.band)
	fdata := (uint16(h3) << 8) + uint16(l3)
	if err := d.i2cWriteReg(0x03, fdata); err != nil {
		return err
	}
	// DSP チューナーが内部でチューニングを完了するまで待つ。
	//	time.Sleep(200 * time.Millisecond)
	// チューニング完了待ち
	time.Sleep(50 * time.Millisecond)
	return nil
}

// 受信可能な最大周波数の取得
func (d *Device) GetUpperFrequencyLimit() int {
	return d.max_freq
}

// 受信可能な最小周波数の取得
func (d *Device) GetLowerFrequencyLimit() int {
	return d.min_freq
}

// RSSI の取得 (受信信号強度)
//
// 現在の信号強度(0-15)を読み出す
// REG 0x0B:
//
//	15:10 RSSI[5:0] : 受信信号強度
//	 9:0  その他ステータス
//
// RSSI は「上位6bit」にあるので、そこを取り出す。
func (d *Device) GetRSSI() (uint8, error) {
	val, err := d.i2cReadReg(0x0B)
	if err != nil {
		return 0, err
	}
	rssi := uint8((val >> 9) & 0x003F)
	return rssi, nil
}

// 音量設定 (0〜15)
//
//	vm = vol | 0b10000000
//	vdata = (0x88 << 8) + vm
//	send_data(0x05, vdata)
func (d *Device) SetVolume(vol uint8) error {
	if vol > 15 {
		vol = 15
	}
	vm := vol | 0b10000000
	vdata := (0x88 << 8) + uint16(vm)
	return d.i2cWriteReg(0x05, vdata)
}

// 現在の音量取得 (0〜15)
//
// REG 0x05 の下位4bitが VOLUME に相当する。
func (d *Device) GetVolume() (uint8, error) {
	val, err := d.i2cReadReg(0x05)
	if err != nil {
		return 0, err
	}
	vol := uint8(val & 0x000F)
	return vol, nil
}

// setMute 無音/有音切替 (0 or 1)
//
// REG 0x02 の14bitが Muteの設定
// 0 = Mute; 1 = Normal operation
func (d *Device) SetMute(state int) error {
	var mode uint16 = 0xC001
	if state != 0 {
		mode = 0xC001
	} else {
		mode = 0x8001
	}
	//	vdata := uint16(mode)<<8 + uint16(0x01)
	//	err := d.i2cWriteReg(0x02, vdata)
	err := d.i2cWriteReg(0x02, mode)
	// 設定完了待ち
	time.Sleep(20 * time.Millisecond)
	return err
}

// getMute Mute状態の取得
//
// REG 0x02 の14bitが Muteの設定
// 0 = Mute; 1 = Normal operation
func (d *Device) GetMute() (bool, error) {
	// --- 1. 現在の Reg 0x02 の設定値を取得 (Read-Modify-Write) ---
	reg02, err := d.i2cReadReg(0x02)
	if err != nil {
		return false, err
	}
	//	fmt.Printf("%d, %04x, %016b\n", reg02, reg02, reg02)
	//	fmt.Printf("%d, %04x, %016b\n", (uint16(reg02) & 0b0100_0000_0000_0000), (uint16(reg02) & 0b0100_0000_0000_0000), (uint16(reg02) & 0b0100_0000_0000_0000))
	if 0 != (uint16(reg02) & 0b0100_0000_0000_0000) {
		return false, nil
	} else {
		return true, nil
	}
}

// 使用禁止 機能確認をしたが、最初にセットした周波数から、まったく周波数が変化しない。
// 閾値設定が低く、ノイズに反応して、停止状態になるのかもしれない。
//
// SeekUp は上位周波数に向かって自動選局を開始し、選局完了（またはタイムアウト）まで待機します。
func (d *Device) SeekUp() error {
	return d.seek(true)
}

// 使用禁止 機能確認をしたが、最初にセットした周波数から、まったく周波数が変化しない。
// 閾値設定が低く、ノイズに反応して、停止状態になるのかもしれない。
//
// SeekDown は下位周波数に向かって自動選局を開始し、選局完了（またはタイムアウト）まで待機します。
func (d *Device) SeekDown() error {
	return d.seek(false)
}

// 使用禁止 機能確認をしたが、最初にセットした周波数から、まったく周波数が変化しない。
// 閾値設定が低く、ノイズに反応して、停止状態になるのかもしれない。
//
// seek は SeekUp / SeekDown の内部共通処理です。
func (d *Device) seek(up bool) error {
	// --- 1. 現在の Reg 0x02 の設定値を取得 (Read-Modify-Write) ---
	reg02, err := d.i2cReadReg(0x02)
	if err != nil {
		return err
	}

	// --- 2. シーク関連のビットを設定 ---
	// bit 10 (SEEKUP): 1 = UP, 0 = DOWN
	// bit 9  (SEEK)  : 1 = シーク開始
	// bit 8  (SKMODE): 0 = バンド端でループ（1にするとバンド端で停止）
	if up {
		reg02 |= uint16(1 << 10) // SEEKUP = 1
	} else {
		reg02 &^= uint16(1 << 10) // SEEKUP = 0
	}
	reg02 |= uint16(1 << 9)  // SEEK = 1 (シーク開始)
	reg02 &^= uint16(1 << 8) // SKMODE = 0 (ラップアラウンド有効)
	fmt.Printf("reg02 = %b\n", reg02)
	// 設定を書き込んでシークを開始
	if err := d.i2cWriteReg(0x02, reg02); err != nil {
		return err
	}

	// --- 3. STC (Seek/Tune Complete) フラグのポーリング待機 ---
	// レジスタ 0x0A の bit 14 (STC) が 1 になるのを待ちます。
	startTime := time.Now()
	seekCompleted := false

	for time.Since(startTime) < 3*time.Second { // 最大3秒タイムアウト
		time.Sleep(30 * time.Millisecond)
		fmt.Printf(".")
		status, err := d.i2cReadReg(0x0A)
		fmt.Printf("\nstatus = %b\n", status)
		if err != nil {
			continue
		}
		fmt.Printf("\n")
		// bit 14: STC (Seek/Tune Complete)
		if (status & (1 << 14)) != 0 {
			seekCompleted = true
			f, _ := d.GetFrequency()
			fmt.Printf("FM %3.2f\n", float64(f)/1000)
			break

		}
	}

	// --- 4. シーク完了後、SEEKビット(bit 9)をクリア ---
	// CLEAR処理を行わないと次回操作に影響が出る場合があるため、SEEKビットを0に戻します。
	reg02 &^= (1 << 9)
	_ = d.i2cWriteReg(0x02, reg02)

	if !seekCompleted {
		return fmt.Errorf("seek operation timed out")
	}
	return nil
}
