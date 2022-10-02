package main

import (
	"encoding/binary"
	"io"
)

// Pascal-style serializer

func ReadPBool(r io.Reader, data *bool) error {
	var b byte
	if err := binary.Read(r, binary.LittleEndian, &b); err != nil {
		return err
	}
	*data = b != 0
	return nil
}

func ReadPByte(r io.Reader, data *byte) error {
	return binary.Read(r, binary.LittleEndian, data)
}

func ReadPShort(r io.Reader, data *int16) error {
	return binary.Read(r, binary.LittleEndian, data)
}

func ReadPUShort(r io.Reader, data *uint16) error {
	return binary.Read(r, binary.LittleEndian, data)
}

func ReadPLongint(r io.Reader, data *int32) error {
	return binary.Read(r, binary.LittleEndian, data)
}

func ReadPString(r io.Reader, data *string, length int) error {
	var realLength byte
	if err := ReadPByte(r, &realLength); err != nil {
		return err
	}
	if realLength > byte(length) {
		realLength = byte(length)
	}
	dataB := make([]byte, length)
	if _, err := r.Read(dataB); err != nil {
		return err
	}
	*data = string(dataB[0:realLength])
	return nil
}

func WritePByte(w io.Writer, data byte) error {
	return binary.Write(w, binary.LittleEndian, data)
}

func WritePBool(w io.Writer, data bool) error {
	if data {
		return WritePByte(w, 1)
	} else {
		return WritePByte(w, 0)
	}
}

func WritePShort(w io.Writer, data int16) error {
	return binary.Write(w, binary.LittleEndian, data)
}

func WritePUShort(w io.Writer, data uint16) error {
	return binary.Write(w, binary.LittleEndian, data)
}

func WritePLongint(w io.Writer, data int32) error {
	return binary.Write(w, binary.LittleEndian, data)
}

func WritePString(w io.Writer, data []byte, length int) error {
	if err := WritePByte(w, byte(len(data))); err != nil {
		return err
	}
	if len(data) < length {
		if _, err := w.Write(data); err != nil {
			return err
		}
		if _, err := w.Write(make([]byte, length-len(data))); err != nil {
			return err
		}
	} else {
		if _, err := w.Write(data[0:length]); err != nil {
			return err
		}
	}
	return nil
}

//go:generate go run serialize_gen.go BoardInfo TBoardInfo MaxShots:u8 IsDark:bool NeighborBoards:array ReenterWhenZapped:bool Message:string:58 StartPlayerX:u8 StartPlayerY:u8 TimeLimitSec:i16 Padding:array

func Read7BoolArray(r io.Reader, data *[7]bool) error {
	for i := 0; i < 7; i++ {
		if err := ReadPBool(r, &((*data)[i])); err != nil {
			return err
		}
	}
	return nil
}

func Write7BoolArray(w io.Writer, data [7]bool) error {
	for i := 0; i < 7; i++ {
		if err := WritePBool(w, data[i]); err != nil {
			return err
		}
	}
	return nil
}

//go:generate go run serialize_gen.go WorldInfo TWorldInfo Ammo:i16 Gems:i16 Keys:!7BoolArray Health:i16 CurrentBoard:i16 Torches:i16 TorchTicks:i16 EnergizerTicks:i16 Padding1:i16 Score:i16 Name:string:20 Flags[0]:string:20 Flags[1]:string:20 Flags[2]:string:20 Flags[3]:string:20 Flags[4]:string:20 Flags[5]:string:20 Flags[6]:string:20 Flags[7]:string:20 Flags[8]:string:20 Flags[9]:string:20 BoardTimeSec:i16 BoardTimeHsec:i16 IsSave:bool Padding2:array

//go:generate go run serialize_gen.go Stat TStat X:u8 Y:u8 StepX:i16 StepY:i16 Cycle:i16 P1:u8 P2:u8 P3:u8 Follower:i16 Leader:i16 Under.Element:u8 Under.Color:u8 nil:4 DataPos:i16 DataLen:i16 Padding:array
