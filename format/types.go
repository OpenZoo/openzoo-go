package format

type (
	TTile struct {
		Element byte
		Color   byte
	}
	TStat struct {
		X, Y         byte
		StepX, StepY int16
		Cycle        int16
		P1, P2, P3   byte
		Follower     int16
		Leader       int16
		Under        TTile
		Data         *[]byte
		DataPos      int16
		DataLen      int16
		Padding1     [4]byte
		Padding2     [8]byte
	}
	TRleTile struct {
		Count byte
		Tile  TTile
	}
	TBoardInfo struct {
		MaxShots          byte
		IsDark            bool
		NeighborBoards    [4]byte
		ReenterWhenZapped bool
		Message           string
		StartPlayerX      byte
		StartPlayerY      byte
		TimeLimitSec      int16
		Padding           [16]byte
	}
	TWorldInfo struct {
		Ammo           int16
		Gems           int16
		Keys           [7]bool
		Health         int16
		CurrentBoard   int16
		Torches        int16
		TorchTicks     int16
		EnergizerTicks int16
		Padding1       int16
		Score          int16
		Name           string
		Flags          []string
		BoardTimeSec   int16
		BoardTimeHsec  int16
		IsSave         bool
		Padding2       [14]byte
	}
	TTileStorage struct {
		Width  int16
		Height int16
		pitch  int
		tiles  []TTile
	}
	TStatStorage struct {
		Count int16
		Max   int16
		stats []TStat
	}
	TBoard struct {
		Name  string
		Tiles TTileStorage
		Stats TStatStorage
		Info  TBoardInfo
	}
	TWorld struct {
		BoardData [][]byte
		Info      TWorldInfo
	}
	THighScoreEntry struct {
		Name  string
		Score int16
	}
)

func NewTileStorage(width, height int16) TTileStorage {
	return TTileStorage{
		Width:  width,
		Height: height,
		pitch:  int(width) + 2,
		tiles:  make([]TTile, (width+2)*(height+2)),
	}
}

func NewStatStorage(max int16) TStatStorage {
	t := TStatStorage{
		Count: 0,
		Max:   max,
		stats: make([]TStat, max+3),
	}
	t.stats[0] = TStat{
		X:        0,
		Y:        1,
		StepX:    256,
		StepY:    256,
		Cycle:    256,
		P1:       0,
		P2:       1,
		P3:       0,
		Follower: 1,
		Leader:   1,
		Under: TTile{
			Element: 0x01,
			Color:   0x00,
		},
		DataPos: 1,
		DataLen: 1,
	}
	return t
}

func (t *TTileStorage) Get(x, y int16) TTile {
	if x >= 0 && y >= 0 && x <= t.Width+1 && y <= t.Height+1 {
		return t.tiles[int(y)*t.pitch+int(x)]
	} else {
		// TODO: This hardcodes "board edge" as element 1.
		return TTile{Element: 0x01, Color: 0x00}
	}
}

func (t *TTileStorage) Pointer(x, y int16) *TTile {
	if x >= 0 && y >= 0 && x <= t.Width+1 && y <= t.Height+1 {
		return &t.tiles[int(y)*t.pitch+int(x)]
	} else {
		return nil
	}
}

func (t *TTileStorage) With(x, y int16, w func(*TTile)) bool {
	if x >= 0 && y >= 0 && x <= t.Width+1 && y <= t.Height+1 {
		w(&t.tiles[int(y)*t.pitch+int(x)])
		return true
	} else {
		return false
	}
}

func (t *TTileStorage) Set(x, y int16, tile TTile) {
	if x >= 0 && y >= 0 && x <= t.Width+1 && y <= t.Height+1 {
		t.tiles[int(y)*t.pitch+int(x)] = tile
	}
}

func (t *TTileStorage) SetElement(x, y int16, element byte) {
	if x >= 0 && y >= 0 && x <= t.Width+1 && y <= t.Height+1 {
		t.tiles[int(y)*t.pitch+int(x)].Element = element
	}
}

func (t *TTileStorage) SetColor(x, y int16, color byte) {
	if x >= 0 && y >= 0 && x <= t.Width+1 && y <= t.Height+1 {
		t.tiles[int(y)*t.pitch+int(x)].Color = color
	}
}

func (b *TStatStorage) At(i int16) *TStat {
	if i < -1 || i > (b.Max+1) {
		return &b.stats[0]
	} else {
		return &b.stats[i+1]
	}
}
