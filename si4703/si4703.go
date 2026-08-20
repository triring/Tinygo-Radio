package si4703

// package si4703 // ローカルでのテスト用
module github.com/triring/Tinygo-Radio/si4703 // githubでの公開用

import (
	"fmt"
	"machine"
	"time"
)

// I2Cアドレス。
const i2cAddress uint16 = 0x10

// Si4703のレジスタ番号。
const (
	regDeviceID   = 0x00
	regChipID     = 0x01
	regPowerCFG   = 0x02
	regChannel    = 0x03
	regSysConfig1 = 0x04
	regSysConfig2 = 0x05
	regSysConfig3 = 0x06
	regTest1      = 0x07

	regStatusRSSI = 0x0A
	regReadChan   = 0x0B
)

// POWERCFGのビット定義。
const (
	powerCFGDSMUTE  uint16 = 1 << 15
	powerCFGDMUTE   uint16 = 1 << 14
	powerCFGMono    uint16 = 1 << 13
	powerCFGRDSM    uint16 = 1 << 11
	powerCFGSKMODE  uint16 = 1 << 10
	powerCFGSeekUp  uint16 = 1 << 9
	powerCFGSeek    uint16 = 1 << 8
	powerCFGDisable uint16 = 1 << 6
	powerCFGEnable  uint16 = 1 << 0
)

// CHANNELのビット定義。
const (
	channelTune uint16 = 1 << 15
	channelMask uint16 = 0x03FF
)

// SYSCONFIG1のビット定義。
const (
	sysConfig1RDS    uint16 = 1 << 12
	sysConfig1DE     uint16 = 1 << 11
	sysConfig1AGCD   uint16 = 1 << 10
	sysConfig1STCIEN uint16 = 1 << 14
)

// SYSCONFIG2のビット定義。
const (
	sysConfig2VolumeMask uint16 = 0x000F
	sysConfig2SpaceMask  uint16 = 0x0030
	sysConfig2BandMask   uint16 = 0x00C0
	sysConfig2SeekTHMask uint16 = 0xFF00
)

// SYSCONFIG3のビット定義。
const (
	sysConfig3VolumeExt uint16 = 1 << 8
	sysConfig3SKSNRMask uint16 = 0x000F << 4
	sysConfig3SKCNTMask uint16 = 0x000F
)

// STATUSRSSIのビット定義。
const (
	statusRSSIRDSR     uint16 = 1 << 15
	statusRSSISTC      uint16 = 1 << 14
	statusRSSISFBL     uint16 = 1 << 13
	statusRSSIAFCRL    uint16 = 1 << 12
	statusRSSIRDSS     uint16 = 1 << 11
	statusRSSIST       uint16 = 1 << 8
	statusRSSIRSSIMask uint16 = 0x00FF
)

// BANDの値。
const (
	BandUSEurope  uint8 = 0
	BandJapanWide uint8 = 1
	BandJapan     uint8 = 2
)

// SPACEの値。
const (
	Space200kHz uint8 = 0
	Space100kHz uint8 = 1
	Space50kHz  uint8 = 2
)

// Seek方向。
const (
	SeekDown = false
	SeekUp   = true
)

// Seek終了モード。
const (
	SeekWrap = false
	SeekStop = true
)

// Si4703はFMラジオICを制御するための構造体です。
type Si4703 struct {
	i2c   *machine.I2C
	reset machine.Pin

	// Si4703のレジスタ内容を保持するshadow register。
	reg [16]uint16

	// 現在設定しているFMバンド。
	band uint8

	// 現在設定しているチャンネル間隔。
	space uint8

	// ドライバが最後に設定した周波数。
	frequency int

	// ドライバが最後に設定した音量。
	volume uint8
}

// NewはSi4703ドライバを生成します。
func New(i2c *machine.I2C, reset machine.Pin) *Si4703 {
	return &Si4703{
		i2c:   i2c,
		reset: reset,
		band:  BandJapanWide,
		space: Space100kHz,
	}
}

// ResetはSi4703をハードウェアリセットします。
func (s *Si4703) Reset() {
	s.reset.Configure(machine.PinConfig{
		Mode: machine.PinOutput,
	})

	s.reset.Low()
	time.Sleep(1 * time.Millisecond)

	s.reset.High()
	time.Sleep(1 * time.Millisecond)
}

// readRegistersはSi4703の全16レジスタを読み出します。
//
// Si4703の読み出し順序は0x0A～0x0F、続いて0x00～0x09です。
func (s *Si4703) readRegisters() error {
	var buf [32]byte

	if err := s.i2c.Tx(i2cAddress, nil, buf[:]); err != nil {
		return err
	}

	index := 0

	for reg := 0x0A; reg <= 0x0F; reg++ {
		s.reg[reg] = uint16(buf[index])<<8 |
			uint16(buf[index+1])
		index += 2
	}

	for reg := 0x00; reg <= 0x09; reg++ {
		s.reg[reg] = uint16(buf[index])<<8 |
			uint16(buf[index+1])
		index += 2
	}

	return nil
}

// writeRegistersはレジスタ0x02～0x07をSi4703へ書き込みます。
//
// Si4703ではレジスタ番号を送信せず、0x02から0x07までを
// 連続して送信します。
func (s *Si4703) writeRegisters() error {
	var buf [12]byte

	index := 0

	for reg := 0x02; reg <= 0x07; reg++ {
		buf[index] = byte(s.reg[reg] >> 8)
		buf[index+1] = byte(s.reg[reg])
		index += 2
	}

	return s.i2c.Tx(i2cAddress, buf[:], nil)
}

// readStatusはSTATUSRSSIレジスタだけを読み出します。
func (s *Si4703) readStatus() error {
	var buf [2]byte

	if err := s.i2c.Tx(i2cAddress, nil, buf[:]); err != nil {
		return err
	}

	s.reg[regStatusRSSI] =
		uint16(buf[0])<<8 | uint16(buf[1])

	return nil
}

// InitはSi4703をリセットしてFM受信機として初期化します。
//
// 初期設定は日本向け76～108MHz、100kHz間隔、
// 50us de-emphasis、AGC有効、音量0です。
func (s *Si4703) Init() error {
	s.Reset()

	if err := s.readRegisters(); err != nil {
		return fmt.Errorf("Si4703 register read failed: %v", err)
	}

	/*
		Si4703内部の32.768kHz水晶発振器を有効にする。

		ここではreserved bitを直接0x8100などに書き換えず、
		読み出した値をベースにXOSCENだけを変更する。
	*/
	s.reg[regTest1] |= 1 << 15
	s.reg[regTest1] &^= 1 << 14

	if err := s.writeRegisters(); err != nil {
		return fmt.Errorf("enable crystal oscillator failed: %v", err)
	}

	// 発振器の安定待ち。
	time.Sleep(500 * time.Millisecond)

	if err := s.readRegisters(); err != nil {
		return fmt.Errorf("register read after oscillator start failed: %v", err)
	}

	// Power configuration。
	s.reg[regPowerCFG] |= powerCFGEnable
	s.reg[regPowerCFG] &^= powerCFGDisable

	// 電源投入時はMute状態から開始する。
	s.reg[regPowerCFG] &^= powerCFGDMUTE

	// ステレオ受信。
	s.reg[regPowerCFG] &^= powerCFGMono

	// Softmute有効。
	s.reg[regPowerCFG] |= powerCFGDSMUTE

	// Seekはバンド端でラップする。
	s.reg[regPowerCFG] &^= powerCFGSKMODE

	// Seek方向は上。
	s.reg[regPowerCFG] |= powerCFGSeekUp

	// Seek/Tuneは停止状態。
	s.reg[regPowerCFG] &^= powerCFGSeek

	// SYSCONFIG1。
	s.reg[regSysConfig1] &^= sysConfig1RDS
	s.reg[regSysConfig1] |= sysConfig1DE
	s.reg[regSysConfig1] &^= sysConfig1STCIEN

	// AGCを有効化。
	s.reg[regSysConfig1] &^= sysConfig1AGCD

	// 日本向け76～108MHz。
	s.reg[regSysConfig2] &^= sysConfig2BandMask
	s.reg[regSysConfig2] |= uint16(BandJapanWide) << 6

	// 100kHz間隔。
	s.reg[regSysConfig2] &^= sysConfig2SpaceMask
	s.reg[regSysConfig2] |= uint16(Space100kHz) << 4

	// Seek RSSI threshold = 0。
	s.reg[regSysConfig2] &^= sysConfig2SeekTHMask

	// 音量0。
	s.reg[regSysConfig2] &^= sysConfig2VolumeMask

	// Extended volume rangeを無効化。
	s.reg[regSysConfig3] &^= sysConfig3VolumeExt

	if err := s.writeRegisters(); err != nil {
		return fmt.Errorf("Si4703 power up failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	if err := s.readRegisters(); err != nil {
		return fmt.Errorf("Si4703 initialization read failed: %v", err)
	}

	s.band = BandJapanWide
	s.space = Space100kHz
	s.volume = 0
	s.frequency = 76000

	return nil
}

// PowerDownはSi4703を低消費電力状態に移行します。
func (s *Si4703) PowerDown() error {
	if err := s.readRegisters(); err != nil {
		return err
	}

	s.reg[regPowerCFG] |= powerCFGEnable
	s.reg[regPowerCFG] |= powerCFGDisable

	// Audio outputをHigh-Zにする。
	s.reg[regTest1] |= 1 << 14

	return s.writeRegisters()
}

// SetChannelはFMチャンネルを直接指定して受信周波数を設定します。
//
// チャンネル番号は現在のBANDとSPACEの設定に従います。
func (s *Si4703) SetChannel(channel uint16) error {
	if channel > 0x03FF {
		return fmt.Errorf("channel out of range: %d", channel)
	}

	if err := s.readRegisters(); err != nil {
		return err
	}

	s.reg[regChannel] &^= channelMask
	s.reg[regChannel] |= channel
	s.reg[regChannel] |= channelTune

	if err := s.writeRegisters(); err != nil {
		return err
	}

	if err := s.waitSTC(true, 2*time.Second); err != nil {
		return err
	}

	// TUNEを解除する。
	if err := s.readRegisters(); err != nil {
		return err
	}

	s.reg[regChannel] &^= channelTune

	if err := s.writeRegisters(); err != nil {
		return err
	}

	if err := s.waitSTC(false, 2*time.Second); err != nil {
		return err
	}

	return nil
}

// SetFrequencyは指定した周波数へチューニングします。
//
// 周波数はkHz単位で指定します。
// 日本向け設定では76000～108000kHzを指定できます。
func (s *Si4703) SetFrequency(freqKHz int) error {
	start, end, spacing, err := s.BandParameters()

	if err != nil {
		return err
	}

	if freqKHz < start || freqKHz > end {
		return fmt.Errorf(
			"frequency out of range: %d kHz",
			freqKHz,
		)
	}

	if (freqKHz-start)%spacing != 0 {
		return fmt.Errorf(
			"frequency %d kHz is not aligned to %d kHz spacing",
			freqKHz,
			spacing,
		)
	}

	channel := uint16((freqKHz - start) / spacing)

	if err := s.SetChannel(channel); err != nil {
		return err
	}

	s.frequency = freqKHz

	return nil
}

// GetFrequencyは最後に設定した周波数を返します。
//
// Si4703の現在の実チャンネルを読み出す場合は
// GetRealFrequencyを使用してください。
func (s *Si4703) GetFrequency() int {
	return s.frequency
}

// GetRealChannelはSi4703のREADCHANレジスタから現在のチャンネルを取得します。
func (s *Si4703) GetRealChannel() (uint16, error) {
	if err := s.readRegisters(); err != nil {
		return 0, err
	}

	return s.reg[regReadChan] & channelMask, nil
}

// GetRealFrequencyはREADCHANレジスタから現在の実際の受信周波数を取得します。
//
// Seek実行後など、Si4703が実際に選択した周波数を取得したい場合に使用します。
func (s *Si4703) GetRealFrequency() (int, error) {
	channel, err := s.GetRealChannel()
	if err != nil {
		return 0, err
	}

	start, _, spacing, err := s.BandParameters()
	if err != nil {
		return 0, err
	}

	freq := start + int(channel)*spacing
	s.frequency = freq

	return freq, nil
}

// FrequencyUpは現在の周波数を1チャンネル上へ移動します。
func (s *Si4703) FrequencyUp() error {
	_, end, spacing, err := s.BandParameters()

	if err != nil {
		return err
	}

	if s.frequency+spacing > end {
		return s.SetFrequency(s.bandStart())
	}

	return s.SetFrequency(s.frequency + spacing)
}

// FrequencyDownは現在の周波数を1チャンネル下へ移動します。
func (s *Si4703) FrequencyDown() error {
	start, _, spacing, err := s.BandParameters()

	if err != nil {
		return err
	}

	if s.frequency < start+spacing {
		return s.SetFrequency(s.bandEnd())
	}

	return s.SetFrequency(s.frequency - spacing)
}

// SetVolumeは音量を0～15の範囲で設定します。
func (s *Si4703) SetVolume(volume uint8) error {
	if volume > 15 {
		return fmt.Errorf("volume must be 0..15")
	}

	if err := s.readRegisters(); err != nil {
		return err
	}

	s.reg[regSysConfig2] &^= sysConfig2VolumeMask
	s.reg[regSysConfig2] |= uint16(volume)

	if err := s.writeRegisters(); err != nil {
		return err
	}

	s.volume = volume

	return nil
}

// GetVolumeは最後に設定した音量を返します。
func (s *Si4703) GetVolume() uint8 {
	return s.volume
}

// SetMuteは音声出力をMuteまたはUnmuteします。
//
// mute=trueでMute、mute=falseでUnmuteです。
// 音量設定値そのものは変更しません。
func (s *Si4703) SetMute(mute bool) error {
	if err := s.readRegisters(); err != nil {
		return err
	}
	if mute {
		s.reg[regPowerCFG] &^= powerCFGDMUTE
	} else {
		s.reg[regPowerCFG] |= powerCFGDMUTE
	}
	return s.writeRegisters()
}

// GetMuteは音声出力の状態を返します。MuteまたはUnmuteします。
//
// Mute状態の時は true、Unmute状態の時は falseを返します。
func (s *Si4703) GetMute() bool {
	s.readRegisters()
	//	fmt.Printf("MuteFlag : %d, 0x%x, 0b%0b\n\r",s.reg[regPowerCFG],s.reg[regPowerCFG],s.reg[regPowerCFG])
	//	fmt.Printf("s.reg[regPowerCFG] & powerCFGDMUTE : %02x,%0b\n\r\n\r",s.reg[regPowerCFG] & powerCFGDMUTE,s.reg[regPowerCFG] & powerCFGDMUTE)
	if 0 == s.reg[regPowerCFG]&powerCFGDMUTE {
		return true
	} else {
		return false
	}
}

// SetSoftMuteは受信信号が弱い場合のSoftmute機能を有効または無効にします。
//
// enable=trueでSoftmute有効、falseで無効です。
func (s *Si4703) SetSoftMute(enable bool) error {
	if err := s.readRegisters(); err != nil {
		return err
	}

	if enable {
		s.reg[regPowerCFG] |= powerCFGDSMUTE
	} else {
		s.reg[regPowerCFG] &^= powerCFGDSMUTE
	}

	return s.writeRegisters()
}

// SetSoftMuteAttackはSoftmuteのAttack/Recovery速度を設定します。
//
// 値は0～3です。
// 0が最速、3が最も遅い設定です。
func (s *Si4703) SetSoftMuteAttack(value uint8) error {
	if value > 3 {
		return fmt.Errorf("soft mute attack must be 0..3")
	}

	if err := s.readRegisters(); err != nil {
		return err
	}

	s.reg[regSysConfig3] &^= 0x3000
	s.reg[regSysConfig3] |= uint16(value) << 12

	return s.writeRegisters()
}

// SetSoftMuteAttenuationはSoftmute時の減衰量を設定します。
//
// 値は0～3です。
// 0=16dB、1=14dB、2=12dB、3=10dBです。
func (s *Si4703) SetSoftMuteAttenuation(value uint8) error {
	if value > 3 {
		return fmt.Errorf("soft mute attenuation must be 0..3")
	}

	if err := s.readRegisters(); err != nil {
		return err
	}

	s.reg[regSysConfig3] &^= 0x0C00
	s.reg[regSysConfig3] |= uint16(value) << 10

	return s.writeRegisters()
}

// SetExtendedVolumeRangeはExtended Volume Rangeを設定します。
//
// enable=trueの場合、音量範囲を30dB低い領域へ移動します。
func (s *Si4703) SetExtendedVolumeRange(enable bool) error {
	if err := s.readRegisters(); err != nil {
		return err
	}

	if enable {
		s.reg[regSysConfig3] |= sysConfig3VolumeExt
	} else {
		s.reg[regSysConfig3] &^= sysConfig3VolumeExt
	}

	return s.writeRegisters()
}

// SetMonoはモノラル受信を設定します。
//
// mono=trueで強制モノラル、falseでステレオ受信を有効にします。
func (s *Si4703) SetMono(mono bool) error {
	if err := s.readRegisters(); err != nil {
		return err
	}

	if mono {
		s.reg[regPowerCFG] |= powerCFGMono
	} else {
		s.reg[regPowerCFG] &^= powerCFGMono
	}

	return s.writeRegisters()
}

// IsStereoは現在の受信信号がステレオと判定されているか取得します。
func (s *Si4703) IsStereo() (bool, error) {
	if err := s.readStatus(); err != nil {
		return false, err
	}

	return (s.reg[regStatusRSSI] & statusRSSIST) != 0, nil
}

// GetRSSIは現在の受信信号強度を取得します。
//
// Si4703のRSSIはdBμV単位で、およそ0～75dBμVです。
func (s *Si4703) GetRSSI() (uint8, error) {
	if err := s.readStatus(); err != nil {
		return 0, err
	}

	return uint8(s.reg[regStatusRSSI] & statusRSSIRSSIMask), nil
}

// SetSeekThresholdはSeek時のRSSI閾値を設定します。
//
// 値は0～127です。値を大きくすると、より強い局だけを
// Seekの候補として扱うようになります。
func (s *Si4703) SetSeekThreshold(value uint8) error {
	if value > 127 {
		return fmt.Errorf("seek threshold must be 0..127")
	}

	if err := s.readRegisters(); err != nil {
		return err
	}

	s.reg[regSysConfig2] &^= sysConfig2SeekTHMask
	s.reg[regSysConfig2] |= uint16(value) << 8

	return s.writeRegisters()
}

// Seekは指定方向に放送局を検索します。
//
// up=trueなら上方向、up=falseなら下方向へ検索します。
// wrap=trueならバンド端で反対側へ回り込みます。
// 見つかった局の周波数を返します。
func (s *Si4703) Seek(up bool, wrap bool) (int, error) {
	if err := s.readRegisters(); err != nil {
		return 0, err
	}

	// SEEK開始前にTUNEを解除する。
	s.reg[regChannel] &^= channelTune

	// Seekモード。
	if wrap {
		s.reg[regPowerCFG] &^= powerCFGSKMODE
	} else {
		s.reg[regPowerCFG] |= powerCFGSKMODE
	}

	// Seek方向。
	if up {
		s.reg[regPowerCFG] |= powerCFGSeekUp
	} else {
		s.reg[regPowerCFG] &^= powerCFGSeekUp
	}

	// SEEK開始。
	s.reg[regPowerCFG] |= powerCFGSeek

	if err := s.writeRegisters(); err != nil {
		return 0, err
	}

	// STC=1を待つ。
	if err := s.waitSTC(true, 10*time.Second); err != nil {
		return 0, err
	}

	// SF/BLを確認。
	if err := s.readStatus(); err != nil {
		return 0, err
	}

	if (s.reg[regStatusRSSI] & statusRSSISFBL) != 0 {
		// SEEKを解除してSF/BLをクリアする。
		if err := s.readRegisters(); err != nil {
			return 0, err
		}

		s.reg[regPowerCFG] &^= powerCFGSeek

		if err := s.writeRegisters(); err != nil {
			return 0, err
		}

		_ = s.waitSTC(false, 2*time.Second)

		return 0, fmt.Errorf("seek failed or band limit reached")
	}

	// SEEK解除。
	if err := s.readRegisters(); err != nil {
		return 0, err
	}

	s.reg[regPowerCFG] &^= powerCFGSeek

	if err := s.writeRegisters(); err != nil {
		return 0, err
	}

	if err := s.waitSTC(false, 2*time.Second); err != nil {
		return 0, err
	}

	// 実際に選択されたチャンネルを取得する。
	return s.GetRealFrequency()
}

// waitSTCはSTCビットが指定状態になるまで待機します。
func (s *Si4703) waitSTC(expected bool, timeout time.Duration) error {
	start := time.Now()

	for {
		if err := s.readStatus(); err != nil {
			return err
		}

		stc := (s.reg[regStatusRSSI] & statusRSSISTC) != 0

		if stc == expected {
			return nil
		}

		if time.Since(start) >= timeout {
			return fmt.Errorf("timeout waiting for STC=%v", expected)
		}

		time.Sleep(5 * time.Millisecond)
	}
}

// SetBandはFM受信バンドを設定します。
//
// band=0は87.5～108MHz、band=1は76～108MHz、
// band=2は76～90MHzです。
func (s *Si4703) SetBand(band uint8) error {
	if band > 2 {
		return fmt.Errorf("invalid band: %d", band)
	}

	if err := s.readRegisters(); err != nil {
		return err
	}

	s.reg[regSysConfig2] &^= sysConfig2BandMask
	s.reg[regSysConfig2] |= uint16(band) << 6

	s.band = band

	return s.writeRegisters()
}

// SetSpaceはFMチャンネル間隔を設定します。
//
// space=0は200kHz、space=1は100kHz、space=2は50kHzです。
func (s *Si4703) SetSpace(space uint8) error {
	if space > 2 {
		return fmt.Errorf("invalid channel spacing: %d", space)
	}

	if err := s.readRegisters(); err != nil {
		return err
	}

	s.reg[regSysConfig2] &^= sysConfig2SpaceMask
	s.reg[regSysConfig2] |= uint16(space) << 4

	s.space = space

	return s.writeRegisters()
}

// GetDeviceIDはDEVICEIDレジスタの値を返します。
func (s *Si4703) GetDeviceID() uint16 {
	return s.reg[regDeviceID]
}

// GetChipIDはCHIPIDレジスタの値を返します。
func (s *Si4703) GetChipID() uint16 {
	return s.reg[regChipID]
}

// BandParametersは現在のBANDとSPACEから周波数パラメータを取得します。
func (s *Si4703) BandParameters() (int, int, int, error) {
	var start int
	var end int

	switch s.band {
	case BandUSEurope:
		start = 87500
		end = 108000

	case BandJapanWide:
		start = 76000
		end = 108000

	case BandJapan:
		start = 76000
		end = 90000

	default:
		return 0, 0, 0, fmt.Errorf("invalid band")
	}

	var spacing int

	switch s.space {
	case Space200kHz:
		spacing = 200

	case Space100kHz:
		spacing = 100

	case Space50kHz:
		spacing = 50

	default:
		return 0, 0, 0, fmt.Errorf("invalid channel spacing")
	}

	return start, end, spacing, nil
}

// bandStartは現在のFMバンドの開始周波数を返します。
func (s *Si4703) bandStart() int {
	start, _, _, _ := s.BandParameters()
	return start
}

// bandEndは現在のFMバンドの終了周波数を返します。
func (s *Si4703) bandEnd() int {
	_, end, _, _ := s.BandParameters()
	return end
}
