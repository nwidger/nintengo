package nintengo

type NROM struct {
	*ROMFile
}

func NewNROM(romf *ROMFile) *NROM {
	return &NROM{ROMFile: romf}
}

func (nrom *NROM) Mappings() (fetch, store []uint16) {
	fetch = []uint16{}
	store = []uint16{}

	// PRG bank 1
	for i := uint32(0x8000); i <= 0xbfff; i++ {
		fetch = append(fetch, uint16(i))
	}

	// PRG bank 2
	for i := uint32(0xc000); i <= 0xffff; i++ {
		fetch = append(fetch, uint16(i))
	}

	return
}

func (nrom *NROM) Reset() {

}

func (nrom *NROM) Fetch(address uint16) (value uint8) {
	index := address & 0x3fff

	switch {
	// PRG bank 1
	case address >= 0x8000 && address <= 0xbfff:
		if nrom.ROMFile.prgBanks > 0 {
			value = nrom.ROMFile.romBanks[0][index]
		}
	// PRG bank 2
	case address >= 0xc000 && address <= 0xffff:
		if nrom.ROMFile.prgBanks > 0 {
			value = nrom.ROMFile.romBanks[nrom.ROMFile.prgBanks-1][index]
		}
	}

	return
}

func (nrom *NROM) Store(address uint16, value uint8) (oldValue uint8) {
	return
}
