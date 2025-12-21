package mp4

// CodecEAC3 is the E-AC-3 (Enhanced AC-3 / Dolby Digital Plus) codec.
type CodecEAC3 struct {
	SampleRate   int
	ChannelCount int
}

// IsVideo implements Codec.
func (CodecEAC3) IsVideo() bool {
	return false
}

func (*CodecEAC3) isCodec() {}
