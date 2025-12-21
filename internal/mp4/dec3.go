package mp4

import amp4 "github.com/abema/go-mp4"

// Dec3 is the E-AC-3 decoder configuration box.
// This is a simplified implementation since go-mp4 doesn't have native Dec3 support.
type Dec3 struct {
	amp4.Box
	// Raw dec3 payload containing E-AC-3 specific configuration
	// Structure (per ETSI TS 102 366):
	// - data_rate (13 bits) + num_ind_sub (3 bits)
	// - For each independent substream:
	//   - fscod (2) + bsid (5) + reserved (1) + asvc (1) + bsmod (3) + acmod (3) + lfeon (1)
	//   - reserved (3) + num_dep_sub (4) + chan_loc (9) if num_dep_sub > 0
	Payload []byte `mp4:"0,size=8"`
}

// GetType returns the box type for dec3.
func (*Dec3) GetType() amp4.BoxType {
	return amp4.StrToBoxType("dec3")
}
