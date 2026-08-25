// Package rda5807 provides a TinyGo driver for the RDA5807M FM tuner.
//
// The driver uses the RDA5807M sequential/random I2C register interfaces and
// is intended for FM reception, including the 76-108 MHz band used in Japan.
//
// RDS/RBDS is intentionally not exposed by this driver.
package rda5807

import (
	"fmt"
	"time"
	"tinygo.org/x/drivers"
)

const (
	// I2C addresses used by the RDA5807M. They are 7-bit I2C addresses.
	// 0x10 is used for sequential register writes.
	rdaSeqAddr uint16 = 0x10

	// rdaRandomAddr is used when a register address is explicitly specified.
	rdaRandomAddr uint16 = 0x11
)

// Register addresses.
const (
	// regChipID is the chip ID register.
	regChipID byte = 0x00

	// regConfig is the main configuration register.
	regConfig byte = 0x02

	// regTune contains the channel, band, spacing and tune control.
	regTune byte = 0x03

	// regAudio contains de-emphasis, soft-mute and AFC settings.
	regAudio byte = 0x04

	// regVolume contains seek threshold and DAC volume.
	regVolume byte = 0x05

	// regStatus is the seek/tune status and received channel register.
	regStatus byte = 0x0A

	// regRSSI contains the RSSI and other receiver status bits.
	regRSSI byte = 0x0B
)

// REG 0x02: CONFIG bit definitions.
const (
	// regConfigDHIZ enables the normal audio output path.
	regConfigDHIZ uint16 = 1 << 15

	// regConfigDMUTE disables audio mute when set.
	regConfigDMUTE uint16 = 1 << 14

	// regConfigMono forces mono reception when set.
	regConfigMono uint16 = 1 << 13

	// regConfigBass enables bass boost when set.
	regConfigBass uint16 = 1 << 12

	// regConfigRCLKNonCal selects the RCLK non-calibration mode.
	regConfigRCLKNonCal uint16 = 1 << 11

	// regConfigRCLKDirect selects direct RCLK input mode.
	regConfigRCLKDirect uint16 = 1 << 10

	// regConfigSeekUp selects seek direction: 1=up, 0=down.
	regConfigSeekUp uint16 = 1 << 9

	// regConfigSeek starts the seek operation when set.
	regConfigSeek uint16 = 1 << 8

	// regConfigSeekMode selects seek behavior at the band limit.
	// 0=wrap around, 1=stop at the band limit.
	regConfigSeekMode uint16 = 1 << 7

	// regConfigRDS enables RDS/RBDS. This driver intentionally leaves it off.
	regConfigRDS uint16 = 1 << 3

	// regConfigNewMethod enables the newer demodulation method.
	regConfigNewMethod uint16 = 1 << 2

	// regConfigSoftReset performs a software reset when set.
	regConfigSoftReset uint16 = 1 << 1

	// regConfigEnable powers up the FM receiver when set.
	regConfigEnable uint16 = 1 << 0
)

// REG 0x03: TUNE bit definitions.
const (
	// regTuneTune starts a frequency tuning operation.
	regTuneTune uint16 = 1 << 4

	// regTuneBandMask selects the FM band.
	regTuneBandMask uint16 = 0x000C

	// regTuneSpaceMask selects the channel spacing.
	regTuneSpaceMask uint16 = 0x0003

	// regTuneChanMask selects the 10-bit channel number.
	regTuneChanMask uint16 = 0xFFC0
)

// REG 0x04: AUDIO bit definitions.
const (
	// regAudioDeemphasis selects 50 us de-emphasis when set and
	// 75 us when clear. Japan uses 50 us.
	regAudioDeemphasis uint16 = 1 << 11

	// regAudioSoftMute enables soft-mute processing.
	regAudioSoftMute uint16 = 1 << 9

	// regAudioAFCD disables AFC when set.
	regAudioAFCD uint16 = 1 << 8
)

// REG 0x05: VOLUME bit definitions.
const (
	// regVolumeSeekThresholdMask contains SEEKTH[3:0].
	regVolumeSeekThresholdMask uint16 = 0x0F00

	// regVolumeMask contains VOLUME[3:0].
	regVolumeMask uint16 = 0x000F
)

// REG 0x0A: STATUS bit definitions.
const (
	// regStatusSTC indicates that seek/tune has completed.
	regStatusSTC uint16 = 1 << 14

	// regStatusSF indicates that seek failed to find a valid station.
	regStatusSF uint16 = 1 << 13

	// regStatusStereo indicates stereo reception.
	regStatusStereo uint16 = 1 << 10

	// regStatusReadChannelMask contains READCHAN[9:0].
	regStatusReadChannelMask uint16 = 0x03FF
)

// REG 0x0B: RSSI bit definitions.
const (
	// regRSSIMask contains RSSI[6:0].
	regRSSIMask uint16 = 0xFE00
)

// Band selection values stored in REG 0x03 bits 3:2.
//
// The values are deliberately kept compatible with the original driver.
const (
	// BandUSEurope selects 87-108 MHz with 100 kHz spacing.
	Band_US_Europe byte = 0x00

	// BandJapan selects 76-91 MHz with 100 kHz spacing.
	Band_Japan byte = 0x04

	// BandWorldWide selects 76-108 MHz with 100 kHz spacing.
	Band_World_Wide byte = 0x08

	// BandEastEurope selects the 65-76 MHz mode.
	Band_East_Europe byte = 0x0C
)

// Channel spacing values for REG 0x03 bits 1:0.
const (
	// Space100kHz selects 100 kHz channel spacing.
	Space100kHz byte = 0x00

	// Space200kHz selects 200 kHz channel spacing.
	Space200kHz byte = 0x01

	// Space50kHz selects 50 kHz channel spacing.
	Space50kHz byte = 0x02

	// Space25kHz selects 25 kHz channel spacing.
	Space25kHz byte = 0x03
)

// Default timing values.
const (
	// initDelay is the time required after the initial register write
	// before normal operation is attempted.
	initDelay = 100 * time.Millisecond

	// tuneTimeout is the maximum time to wait for a manual tune operation.
	tuneTimeout = 500 * time.Millisecond

	// seekTimeout is the maximum time to wait for a seek operation.
	seekTimeout = 3 * time.Second

	// statusPollInterval is the interval between status register reads.
	statusPollInterval = 20 * time.Millisecond
)

// BandData describes an FM band supported by the driver.
//
// Frequencies are expressed in kHz.
type BandData struct {
	// BandName is a human-readable name for the band.
	BandName string

	// Min is the lowest frequency in kHz.
	Min int

	// Max is the highest frequency in kHz.
	Max int
}

var bandData = [4]BandData{
	{BandName: "US/Europe", Min: 87000, Max: 108000},
	{BandName: "Japan", Min: 76000, Max: 91000},
	{BandName: "World wide", Min: 76000, Max: 108000},
	{BandName: "East Europe", Min: 65000, Max: 76000},
}

// Device represents an RDA5807M FM tuner connected through I2C.
type Device struct {
	bus drivers.I2C

	band     byte
	space    byte
	minFreq  int
	maxFreq  int
	bandName string
}

// New creates an RDA5807M device using the supplied I2C bus.
//
// The device is not initialized by New. Call InitRDA5807 before using
// the tuner.
func New(bus drivers.I2C) Device {
	return Device{
		bus:   bus,
		band:  Band_World_Wide,
		space: Space100kHz,
	}
}

// GetChipID reads and returns the RDA5807M chip ID.
//
// The expected chip ID is 0x58 in the upper byte of register 0x00.
func (d *Device) GetChipID() (uint16, error) {
	return d.i2cReadReg(regChipID)
}

// GetBandInfo returns the currently selected band, its name, and its
// minimum and maximum frequencies in kHz.
func (d *Device) GetBandInfo() (byte, string, int, int) {
	return d.band, d.bandName, d.minFreq, d.maxFreq
}

// SetBroadcastBand selects the FM band used by the driver.
//
// The band argument should be one of BandUSEurope, BandJapan,
// BandWorldWide, or BandEastEurope. The value is not written to the
// chip until InitRDA5807 or SetFrequency is called.
// band 0:US/Europe
// band 1:Japan
// band 2:World wide
// band 3:East Europe
func (d *Device) SetBroadcastBand(band byte) {
	index := int((band >> 2) & 0x03)
	d.band = band & 0x0C
	d.bandName = bandData[index].BandName
	d.minFreq = bandData[index].Min
	d.maxFreq = bandData[index].Max
}

// SetChannelSpacing selects the channel spacing used for frequency calculation and tuning.
//
// Supported values are Space100kHz, Space200kHz, Space50kHz, and
// Space25kHz. The setting is applied to the tuner by InitRDA5807.
// For the current driver, 100 kHz spacing is recommended for normal
// Japanese FM/Wide FM use.
func (d *Device) SetChannelSpacing(space byte) error {
	if space > Space25kHz {
		return fmt.Errorf("invalid channel spacing: %d", space)
	}
	d.space = space
	return nil
}

// InitRDA5807 initializes the tuner and enables FM reception.
//
// The supplied band is selected before initialization. RDS/RBDS is
// deliberately disabled. For Japanese reception, use BandWorldWide
// to cover 76-108 MHz or BandJapan for 76-91 MHz.
func (d *Device) InitRDA5807(band byte) error {
	d.SetBroadcastBand(band)

	if d.space > Space25kHz {
		d.space = Space100kHz
	}

	// REG 0x02:
	// DHIZ=1, DMUTE=1, SEEKUP=0, SEEK=0, SKMODE=0,
	// CLK_MODE=000 (32.768 kHz), RDS_EN=0, ENABLE=1.
	//
	// SOFT_RESET is intentionally not included here. The chip is
	// brought directly into normal operation by the power-up write.
	reg02 := regConfigDHIZ | regConfigDMUTE | regConfigEnable

	// REG 0x03:
	// CHAN=0, DIRECT_MODE=0, TUNE=0, selected BAND, selected SPACE.
	// 初期化時は TUNE=0 にしておくことが重要。
	reg03 := uint16(d.band) | uint16(d.space)

	// REG 0x04:
	// 50 us de-emphasis for Japan, soft-mute enabled, AFC enabled.
	// Reserved bits are kept at zero.
	reg04 := regAudioDeemphasis | regAudioSoftMute
	// reg04 := regAudioDeemphasis | regAudioSoftMute | regAudioAFCD


	// REG 0x05:
	// INT_MODE=1, SEEKTH=1000 (datasheet default), volume=8.
	reg05 := uint16(0x8000) | uint16(0x0800) | 0x008f
	// LNA（Low Noise Amplifier：低雑音増幅器） は、アンテナで受信した微弱な高周波（RF）信号を、自身のノイズ（雑音）を最小限に抑えながら増幅する重要な回路・パーツ

	data := []byte{
		byte(reg02 >> 8), byte(reg02),
		byte(reg03 >> 8), byte(reg03),
		byte(reg04 >> 8), byte(reg04),
		byte(reg05 >> 8), byte(reg05),
	}

	if err := d.bus.Tx(rdaSeqAddr, data, nil); err != nil {
		return err
	}
	/* for debug
	fmt.Printf("reg02: 0x%04x, 0b%016b\n", reg02, reg02)
	fmt.Printf("reg03: 0x%04x, 0b%016b\n", reg03, reg03)
	fmt.Printf("reg04: 0x%04x, 0b%016b\n", reg04, reg04)
	fmt.Printf("reg05: 0x%04x, 0b%016b\n", reg05, reg05)
	*/
	time.Sleep(initDelay)
	return nil
}

// SetFrequency tunes the receiver to freqKHz.
//
// The frequency must be within the currently selected band and aligned
// to the selected channel spacing. The method waits for the STC flag
// and returns an error if tuning does not complete in time.
func (d *Device) SetFrequency(freqKHz int) error {
	if d.minFreq == 0 || d.maxFreq == 0 {
		return fmt.Errorf("broadcast band is not initialized")
	}
	if freqKHz < d.minFreq || freqKHz > d.maxFreq {
		return fmt.Errorf("frequency %d kHz is outside %d-%d kHz",
			freqKHz, d.minFreq, d.maxFreq)
	}

	step := d.channelSpacingKHz()
	if (freqKHz-d.minFreq)%step != 0 {
		return fmt.Errorf("frequency %d kHz is not aligned to %d kHz spacing",
			freqKHz, step)
	}

	channel := uint16((freqKHz - d.minFreq) / step)
	if channel > 0x03FF {
		return fmt.Errorf("frequency channel out of range: %d", channel)
	}

	reg03 := (channel << 6) | uint16(d.band) | uint16(d.space)
	reg03 |= regTuneTune

	if err := d.i2cWriteReg(regTune, reg03); err != nil {
		return err
	}

	// TUNE is cleared automatically by the device after completion.
	// Poll STC rather than relying only on a fixed delay.
	if err := d.waitForSTC(tuneTimeout); err != nil {
		return err
	}
	return nil
}

// GetFrequency returns the currently received frequency in kHz.
//
// The value is calculated from READCHAN in register 0x0A and the
// current band and channel spacing.
func (d *Device) GetFrequency() (int, error) {
	status, err := d.i2cReadReg(regStatus)
	if err != nil {
		return 0, err
	}

	channel := status & regStatusReadChannelMask
	return d.minFreq + int(channel)*d.channelSpacingKHz(), nil
}

// GetUpperFrequencyLimit returns the upper limit of the selected band
// in kHz.
func (d *Device) GetUpperFrequencyLimit() int {
	return d.maxFreq
}

// GetLowerFrequencyLimit returns the lower limit of the selected band
// in kHz.
func (d *Device) GetLowerFrequencyLimit() int {
	return d.minFreq
}

// GetRSSI returns the received signal strength indicator.
//
// The RDA5807M reports RSSI as a 7-bit value from 0 to 127 in
// register 0x0B bits 15:9.
// RSSI の取得 (受信信号強度)
// 現在の信号強度(0-63)を読み出す
//	15:10 RSSI[5:0] : 受信信号強度
//	 9:0  その他ステータス
//
// RSSI は「上位6bit」にあるので、そこを取り出す。
func (d *Device) GetRSSI() (uint8, error) {
	value, err := d.i2cReadReg(regRSSI)
	if err != nil {
		return 0, err
	}
//	rssi := uint8((val >> 9) & 0x003F)
//	return rssi, nil
	return uint8((value & regRSSIMask) >> 9), nil
}

// SetVolume sets the DAC volume from 0 (minimum) to 15 (maximum).
//
// Values greater than 15 are rejected rather than silently truncated.
func (d *Device) SetVolume(volume uint8) error {
	if volume > 15 {
		return fmt.Errorf("volume out of range: %d", volume)
	}

	reg05, err := d.i2cReadReg(regVolume)
	if err != nil {
		return err
	}

	reg05 = (reg05 &^ regVolumeMask) | uint16(volume)
	return d.i2cWriteReg(regVolume, reg05)
}

// GetVolume returns the current DAC volume from 0 to 15.
func (d *Device) GetVolume() (uint8, error) {
	value, err := d.i2cReadReg(regVolume)
	if err != nil {
		return 0, err
	}

	return uint8(value & regVolumeMask), nil
}

// SetMute enables or disables audio mute.
//
// A true value mutes the audio output. A false value restores normal
// audio operation. The method preserves the other REG 0x02 settings.
func (d *Device) SetMute(mute bool) error {
	reg02, err := d.i2cReadReg(regConfig)
	if err != nil {
		return err
	}

	if mute {
		reg02 &^= regConfigDMUTE
	} else {
		reg02 |= regConfigDMUTE
	}

	return d.i2cWriteReg(regConfig, reg02)
}

// GetMute reports whether audio mute is currently enabled.
func (d *Device) GetMute() (bool, error) {
	reg02, err := d.i2cReadReg(regConfig)
	if err != nil {
		return false, err
	}

	return (reg02 & regConfigDMUTE) == 0, nil
}

// SeekUp starts automatic station seeking toward higher frequencies.
//
// The method waits until STC becomes set. If the tuner reports SF,
// SeekUp returns an error. Seek wraps from the upper band limit to the
// lower limit because SKMODE is cleared.
func (d *Device) SeekUp(rssi_threshold uint8) (freq int, err error) {
	for {
		f,     _ := d.GetFrequency()
		limit := d.GetUpperFrequencyLimit() 
		if f >= limit {
			return 0, fmt.Errorf("The set frequency has reached the upper limit: %d", limit)
		//	return 0, fmt.Errorf("The set frequency has reached the lower limit: %d", limit)
		}
		step := d.channelSpacingKHz()
		d.SetFrequency(f + step)
		time.Sleep(initDelay)
		rssi, _ := d.GetRSSI()
		if rssi > rssi_threshold {
			f,     _ := d.GetFrequency()
			return f, nil
		}
	}
}

// SeekDown starts automatic station seeking toward lower frequencies.
//
// The method waits until STC becomes set. If the tuner reports SF,
// SeekDown returns an error. Seek wraps from the lower band limit to the
// upper limit because SKMODE is cleared.
func (d *Device) SeekDown(rssi_threshold uint8) (freq int, err error) {
	for {
		f,     _ := d.GetFrequency()
		limit := d.GetLowerFrequencyLimit() 
		if limit >= f {
		//	return 0, fmt.Errorf("The set frequency has reached the upper limit: %d", limit)
			return 0, fmt.Errorf("The set frequency has reached the lower limit: %d", limit)
		}
		step := d.channelSpacingKHz()
		d.SetFrequency(f - step)
		time.Sleep(initDelay)
		rssi, _ := d.GetRSSI()
		if rssi > rssi_threshold {
			f,     _ := d.GetFrequency()
			return f, nil
		}
	}
}

// waitForSTC waits until the seek/tune complete flag is set.
func (d *Device) waitForSTC(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		time.Sleep(statusPollInterval)

		status, err := d.i2cReadReg(regStatus)
		if err != nil {
			return err
		}

		if status&regStatusSTC != 0 {
			return nil
		}
	}

	return fmt.Errorf("tune timeout")
}

// channelSpacingKHz returns the channel spacing represented by the
// currently selected SPACE bits.
func (d *Device) channelSpacingKHz() int {
	switch d.space & 0x03 {
	case Space200kHz:
		return 200
	case Space50kHz:
		return 50
	case Space25kHz:
		return 25
	default:
		return 100
	}
}

// i2cWriteReg writes one 16-bit register through the random-access I2C
// interface.
func (d *Device) i2cWriteReg(reg byte, data uint16) error {
	buf := []byte{
		reg,
		byte(data >> 8),
		byte(data),
	}
	return d.bus.Tx(rdaRandomAddr, buf, nil)
}

// i2cReadReg reads one 16-bit register through the random-access I2C
// interface.
func (d *Device) i2cReadReg(reg byte) (uint16, error) {
	write := []byte{reg}
	read := make([]byte, 2)

	if err := d.bus.Tx(rdaRandomAddr, write, read); err != nil {
		return 0, err
	}

	return (uint16(read[0]) << 8) | uint16(read[1]), nil
}
