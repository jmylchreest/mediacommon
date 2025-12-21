package mp4

import (
	amp4 "github.com/abema/go-mp4"

	"github.com/bluenviron/mediacommon/v2/pkg/formats/mp4"
)

func init() {
	// Register ec-3 as an AudioSampleEntry type (like ac-3)
	amp4.AddAnyTypeBoxDef(&amp4.AudioSampleEntry{}, amp4.StrToBoxType("ec-3"))
	// Register dec3 box type
	amp4.AddBoxDef(&Dec3{})
}

// Dec3 is the E-AC-3 decoder configuration box (EC3SpecificBox).
// This implements the structure defined in ETSI TS 102 366.
//
// The dec3 box has variable length depending on NumDepSub:
// - When NumDepSub = 0: 5 bytes (40 bits) - 1 reserved bit after NumDepSub
// - When NumDepSub > 0: 6 bytes (48 bits) - 9-bit ChanLoc after NumDepSub
//
// We use raw bytes for parsing to handle this variable structure.
type Dec3 struct {
	amp4.Box

	// Raw payload - we parse/marshal manually due to variable length
	Payload []byte `mp4:"0,size=8,len=dynamic"`
}

// GetType returns the box type for dec3.
func (*Dec3) GetType() amp4.BoxType {
	return amp4.StrToBoxType("dec3")
}

// GetFieldLength returns the length for dynamic fields.
func (d *Dec3) GetFieldLength(name string, ctx amp4.Context) uint {
	switch name {
	case "Payload":
		return amp4.LengthUnlimited
	}
	panic("invalid field name: " + name)
}

// ToCodec converts the Dec3 box to a CodecEAC3.
func (d *Dec3) ToCodec(sampleRate, channelCount int) *mp4.CodecEAC3 {
	if len(d.Payload) < 5 {
		// Not enough data, return minimal codec
		return &mp4.CodecEAC3{
			SampleRate:   sampleRate,
			ChannelCount: channelCount,
		}
	}

	// Parse the dec3 payload
	// Byte 0-1: data_rate (13 bits) + num_ind_sub (3 bits)
	dataRate := (uint16(d.Payload[0]) << 5) | (uint16(d.Payload[1]) >> 3)
	numIndSub := d.Payload[1] & 0x07

	// Byte 2-3: fscod (2) + bsid (5) + reserved (1) + asvc (1) + bsmod (3) + acmod (3) + lfeon (1)
	fscod := d.Payload[2] >> 6
	bsid := (d.Payload[2] >> 1) & 0x1F
	asvc := (d.Payload[3] >> 7) & 0x01
	bsmod := (d.Payload[3] >> 4) & 0x07
	acmod := (d.Payload[3] >> 1) & 0x07
	lfeon := d.Payload[3] & 0x01

	// Byte 4: reserved (3) + num_dep_sub (4) + (reserved (1) OR start of chan_loc)
	numDepSub := (d.Payload[4] >> 1) & 0x0F

	var chanLoc uint16
	if numDepSub > 0 && len(d.Payload) >= 6 {
		// chan_loc is 9 bits: 1 bit from byte 4 + 8 bits from byte 5
		chanLoc = (uint16(d.Payload[4]&0x01) << 8) | uint16(d.Payload[5])
	}

	return &mp4.CodecEAC3{
		SampleRate:   sampleRate,
		ChannelCount: channelCount,
		DataRate:     dataRate,
		NumIndSub:    numIndSub,
		Fscod:        fscod,
		Bsid:         bsid,
		Asvc:         asvc != 0,
		Bsmod:        bsmod,
		Acmod:        acmod,
		LfeOn:        lfeon != 0,
		NumDepSub:    numDepSub,
		ChanLoc:      chanLoc,
	}
}

// FromCodec creates a Dec3 box from a CodecEAC3.
func FromCodec(codec *mp4.CodecEAC3) *Dec3 {
	var asvc, lfeon uint8
	if codec.Asvc {
		asvc = 1
	}
	if codec.LfeOn {
		lfeon = 1
	}

	// Build the payload
	// Byte 0-1: data_rate (13 bits) + num_ind_sub (3 bits)
	byte0 := uint8(codec.DataRate >> 5)
	byte1 := uint8((codec.DataRate&0x1F)<<3) | (codec.NumIndSub & 0x07)

	// Byte 2: fscod (2) + bsid (5) + reserved (1)
	byte2 := (codec.Fscod << 6) | ((codec.Bsid & 0x1F) << 1)

	// Byte 3: asvc (1) + bsmod (3) + acmod (3) + lfeon (1)
	byte3 := (asvc << 7) | ((codec.Bsmod & 0x07) << 4) | ((codec.Acmod & 0x07) << 1) | lfeon

	if codec.NumDepSub > 0 {
		// 6 bytes: with chan_loc
		// Byte 4: reserved (3) + num_dep_sub (4) + chan_loc high bit
		byte4 := ((codec.NumDepSub & 0x0F) << 1) | uint8((codec.ChanLoc>>8)&0x01)
		// Byte 5: chan_loc low 8 bits
		byte5 := uint8(codec.ChanLoc & 0xFF)
		return &Dec3{
			Payload: []byte{byte0, byte1, byte2, byte3, byte4, byte5},
		}
	}

	// 5 bytes: no chan_loc, just 1 reserved bit
	// Byte 4: reserved (3) + num_dep_sub (4) + reserved (1)
	byte4 := (codec.NumDepSub & 0x0F) << 1
	return &Dec3{
		Payload: []byte{byte0, byte1, byte2, byte3, byte4},
	}
}
