package mpeg4audio

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolveChannelCount(t *testing.T) {
	t.Run("config with explicit channel count", func(t *testing.T) {
		config := &AudioSpecificConfig{
			Type:         ObjectTypeAACLC,
			SampleRate:   48000,
			ChannelCount: 2,
		}
		result := ResolveChannelCount(config, nil, 6)
		require.Equal(t, 2, result)
	})

	t.Run("nil config with CPE in AU", func(t *testing.T) {
		// AU with CPE element (stereo)
		// CPE id_syn_ele (3 bits) = 001, element_instance_tag (4 bits) = 0000
		au := []byte{0x20, 0x00, 0x00, 0x00}
		result := ResolveChannelCount(nil, au, 6)
		require.Equal(t, 2, result)
	})

	t.Run("config with channel_config=0 and CPE in AU", func(t *testing.T) {
		config := &AudioSpecificConfig{
			Type:         ObjectTypeAACLC,
			SampleRate:   48000,
			ChannelCount: 0, // PCE in bitstream
		}
		// AU with CPE element (stereo)
		au := []byte{0x20, 0x00, 0x00, 0x00}
		result := ResolveChannelCount(config, au, 6)
		require.Equal(t, 2, result)
	})

	t.Run("config with channel_config=0 and SCE in AU", func(t *testing.T) {
		config := &AudioSpecificConfig{
			Type:         ObjectTypeAACLC,
			SampleRate:   48000,
			ChannelCount: 0,
		}
		// AU with SCE element (mono)
		// SCE id_syn_ele (3 bits) = 000, element_instance_tag (4 bits) = 0000
		au := []byte{0x00, 0x00, 0x00, 0x00}
		result := ResolveChannelCount(config, au, 6)
		require.Equal(t, 1, result)
	})

	t.Run("config with channel_config=0 and PCE in AU", func(t *testing.T) {
		config := &AudioSpecificConfig{
			Type:         ObjectTypeAACLC,
			SampleRate:   48000,
			ChannelCount: 0,
		}
		// AU with PCE element (stereo - 1 front CPE)
		// PCE id_syn_ele (3 bits) = 101, then PCE data
		au := []byte{0xA0, 0xA0, 0x80, 0x00, 0x04, 0x00}
		result := ResolveChannelCount(config, au, 6)
		require.Equal(t, 2, result)
	})

	t.Run("nil config with empty AU returns fallback", func(t *testing.T) {
		result := ResolveChannelCount(nil, nil, 6)
		require.Equal(t, 6, result)
	})

	t.Run("config with channel_config=0 and empty AU returns fallback", func(t *testing.T) {
		config := &AudioSpecificConfig{
			Type:         ObjectTypeAACLC,
			SampleRate:   48000,
			ChannelCount: 0,
		}
		result := ResolveChannelCount(config, nil, 2)
		require.Equal(t, 2, result)
	})

	t.Run("config with channel_config=0 and invalid AU returns fallback", func(t *testing.T) {
		config := &AudioSpecificConfig{
			Type:         ObjectTypeAACLC,
			SampleRate:   48000,
			ChannelCount: 0,
		}
		// Invalid AU - END element without any channel elements
		au := []byte{0xE0} // id_syn_ele = 7 (END)
		result := ResolveChannelCount(config, au, 2)
		require.Equal(t, 2, result)
	})
}
