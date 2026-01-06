package mpeg4audio

// ResolveChannelCount determines the actual channel count for AAC streams.
//
// When channel_config=0 in ADTS or AudioSpecificConfig, the channel layout
// is defined by a Program Config Element (PCE). This function resolves the
// actual channel count by:
//  1. Returning config.ChannelCount if non-zero
//  2. Attempting to parse PCE from the access unit
//  3. Attempting to count channels from syntactic elements in the access unit
//  4. Returning the fallback value if all else fails
//
// Parameters:
//   - config: AudioSpecificConfig (may be nil)
//   - au: Access unit (raw_data_block) that may contain PCE or channel elements
//   - fallback: Value to return if channel count cannot be determined
//
// This is useful for muxers that need the actual channel count when the
// original stream used channel_config=0.
func ResolveChannelCount(config *AudioSpecificConfig, au []byte, fallback int) int {
	// If config has explicit channel count, use it
	if config != nil && config.ChannelCount > 0 {
		return config.ChannelCount
	}

	// Try to extract from the access unit
	if len(au) > 0 {
		// Try PCE first (explicit channel layout)
		if pce, err := ParsePCEFromRawDataBlock(au); err == nil {
			return pce.ChannelCount
		}

		// Try counting syntactic elements (implicit layout)
		if count, err := CountChannelsFromRawDataBlock(au); err == nil {
			return count
		}
	}

	return fallback
}
