package format

import (
	"errors"
	"io"
)

type (
	WorldDeserializeFlag uint32
)

const (
	WorldDeserializeTitleOnly WorldDeserializeFlag = 0x00000001

	BOARD_NAME_LENGTH      = 50
	WORLD_FILE_HEADER_SIZE = 512
)

var ErrWrongZZTVersion = errors.New("You need a newer version of ZZT!")

func BoardSerialize(b *TBoard, w io.Writer) error {
	var (
		ix, iy int16
		rle    TRleTile
	)
	err := WritePString(w, []byte(b.Name), BOARD_NAME_LENGTH)
	if err != nil {
		return err
	}

	ix = 1
	iy = 1
	rle.Count = 1
	rle.Tile = b.Tiles.Get(ix, iy)
	for {
		ix++
		if ix > int16(b.Tiles.Width) {
			ix = 1
			iy++
		}
		if b.Tiles.Get(ix, iy).Color == rle.Tile.Color && b.Tiles.Get(ix, iy).Element == rle.Tile.Element && rle.Count < 255 && iy <= int16(b.Tiles.Height) {
			rle.Count++
		} else {
			err = WritePByte(w, rle.Count)
			if err != nil {
				return err
			}
			err = WritePByte(w, rle.Tile.Element)
			if err != nil {
				return err
			}
			err = WritePByte(w, rle.Tile.Color)
			if err != nil {
				return err
			}

			rle.Tile = b.Tiles.Get(ix, iy)
			rle.Count = 1
		}
		if iy > int16(b.Tiles.Height) {
			break
		}
	}
	err = WriteBoardInfo(w, b.Info)
	if err != nil {
		return err
	}
	err = WritePShort(w, b.Stats.Count)
	if err != nil {
		return err
	}
	for ix = 0; ix <= b.Stats.Count; ix++ {
		stat := b.Stats.At(ix)
		if stat.DataLen > 0 {
			for iy = 1; iy <= ix-1; iy++ {
				if b.Stats.At(iy).Data == stat.Data {
					stat.DataLen = -iy
				}
			}
		}
		err = WriteStat(w, *b.Stats.At(ix))
		if err != nil {
			return err
		}
		if stat.DataLen > 0 {
			err = WritePBytes(w, *b.Stats.At(ix).Data, int(b.Stats.At(ix).DataLen))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func BoardDeserialize(b *TBoard, r io.Reader) error {
	var (
		ix, iy int16
		rle    TRleTile
	)
	err := ReadPString(r, &b.Name, BOARD_NAME_LENGTH)
	if err != nil {
		return err
	}
	ix = 1
	iy = 1
	rle.Count = 0
	for {
		if rle.Count <= 0 {
			err = ReadPByte(r, &rle.Count)
			if err != nil {
				return err
			}
			err = ReadPByte(r, &rle.Tile.Element)
			if err != nil {
				return err
			}
			err = ReadPByte(r, &rle.Tile.Color)
			if err != nil {
				return err
			}
		}
		b.Tiles.Set(ix, iy, rle.Tile)
		ix++
		if ix > b.Tiles.Width {
			ix = 1
			iy++
		}
		rle.Count--
		if iy > b.Tiles.Height {
			break
		}
	}
	err = ReadBoardInfo(r, &b.Info)
	if err != nil {
		return err
	}
	err = ReadPShort(r, &b.Stats.Count)
	if err != nil {
		return err
	}
	for ix = 0; ix <= b.Stats.Count; ix++ {
		stat := b.Stats.At(ix)
		err = ReadStat(r, stat)
		if err != nil {
			return err
		}
		if stat.DataLen > 0 {
			data := make([]byte, stat.DataLen)
			_, err = r.Read(data)
			if err != nil {
				return err
			}
			stat.Data = &data
		} else if stat.DataLen < 0 {
			stat.Data = b.Stats.At(-stat.DataLen).Data
			stat.DataLen = b.Stats.At(-stat.DataLen).DataLen
		}
	}
	return nil
}

func WorldDeserialize(f io.ReadSeeker, w *TWorld, flags WorldDeserializeFlag, onLoad func(step, stepMax int)) error {
	var (
		boardId  int16
		boardLen uint16
	)

	var boardCount int16
	if err := ReadPShort(f, &boardCount); err != nil {
		return err
	}
	if boardCount < 0 {
		if boardCount != -1 {
			return ErrWrongZZTVersion
		} else {
			if err := ReadPShort(f, &boardCount); err != nil {
				return err
			}
		}
	}
	if err := ReadWorldInfo(f, &w.Info); err != nil {
		return err
	}
	if (flags & WorldDeserializeTitleOnly) != 0 {
		boardCount = 0
		w.Info.CurrentBoard = 0
		w.Info.IsSave = true
	}
	_, err := f.Seek(WORLD_FILE_HEADER_SIZE, io.SeekStart)
	if err != nil {
		return err
	}
	w.BoardData = make([][]byte, boardCount+1)
	for boardId = 0; boardId <= boardCount; boardId++ {
		onLoad(int(boardId), int(boardCount))
		err = ReadPUShort(f, &boardLen)
		if err != nil {
			return err
		}

		data := make([]byte, boardLen)
		_, err := f.Read(data)
		if err != nil {
			return err
		}
		w.BoardData[boardId] = data
	}

	return nil
}

func WorldSerialize(f io.WriteSeeker, w *TWorld) error {
	// version
	if err := WritePShort(f, -1); err != nil {
		return err
	}
	if err := WritePShort(f, int16(len(w.BoardData)-1)); err != nil {
		return err
	}
	if err := WriteWorldInfo(f, w.Info); err != nil {
		return err
	}
	if _, err := f.Seek(WORLD_FILE_HEADER_SIZE, io.SeekStart); err != nil {
		return err
	}

	for i := 0; i < len(w.BoardData); i++ {
		if err := WritePUShort(f, uint16(len(w.BoardData[i]))); err != nil {
			return err
		}
		if _, err := f.Write(w.BoardData[i]); err != nil {
			return err
		}
	}
	return nil
}
