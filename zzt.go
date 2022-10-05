package main

import (
	"bufio"
	"os"
)

// uses: Crt, Dos, Video, Keys, Sounds, Input, TxtWind, GameVars, Elements, Editor, Oop, Game

func ParseArguments() {
	var (
		i    int
		pArg string
	)
	for i = 1; i < len(os.Args); i++ {
		pArg = os.Args[i]
		if pArg[0] == '/' {
			switch UpCase(pArg[1]) {
			case 'R':
				ResetConfig = true
			}
		} else {
			StartupWorldFileName = pArg
			if Length(StartupWorldFileName) > 4 && StartupWorldFileName[Length(StartupWorldFileName)-3-1] == '.' {
				StartupWorldFileName = Copy(StartupWorldFileName, 1, Length(StartupWorldFileName)-4)
			}
		}
	}
}

func GameConfigure() {
	ParsingConfigFile = true
	EditorEnabled = true
	ConfigRegistration = ""
	ConfigWorldFile = ""
	GameVersion = "3.2"
	{
		f, err := os.Open(PathFindCaseInsensitiveFile("zzt.cfg"))
		if err == nil {
			bf := bufio.NewReader(f)
			line, _, _ := bf.ReadLine()
			ConfigWorldFile = string(line)
			line, _, _ = bf.ReadLine()
			ConfigRegistration = string(line)
		}
		f.Close()
	}
	if Length(ConfigWorldFile) > 0 {
		if ConfigWorldFile[0] == '*' {
			EditorEnabled = false
			ConfigWorldFile = ConfigWorldFile[1:]
		}
		if Length(ConfigWorldFile) > 0 {
			StartupWorldFileName = ConfigWorldFile
		}
	}
	InputInitDevices()
	ParsingConfigFile = false
	Window(1, 1, 80, 25)
	TextBackground(Black)
	ClrScr()
	TextColor(White)
	TextColor(White)
	WriteLn("")
	WriteLn("                                 <=-  ZZT  -=>")
	TextColor(Yellow)
	if Length(ConfigRegistration) == 0 {
		WriteLn("                             Shareware version 3.2")
	} else {
		WriteLn("                                  Version  3.2")
	}
	WriteLn("                            Created by Tim Sweeney")
	GotoXY(1, 7)
	TextColor(Blue)
	Write("================================================================================")
	GotoXY(1, 24)
	Write("================================================================================")
	TextColor(White)
	GotoXY(30, 7)
	Write(" Game Configuration ")
	GotoXY(1, 25)
	Write(" Copyright (c) 1991 Epic MegaGames                         Press ... to abort")
	TextColor(Black)
	TextBackground(LightGray)
	GotoXY(66, 25)
	Write("ESC")
	Window(1, 8, 80, 23)
	TextColor(Yellow)
	TextBackground(Black)
	ClrScr()
	TextColor(Yellow)
	if !InputConfigure() {
		GameTitleExitRequested = true
	} else {
		TextColor(LightGreen)
		if !VideoConfigure() {
			GameTitleExitRequested = true
		}
	}
	Window(1, 1, 80, 25)
}

func ZZTMain() {
	WorldFileDescCount = 7
	WorldFileDescKeys[0] = "TOWN"
	WorldFileDescValues[0] = "TOWN       The Town of ZZT"
	WorldFileDescKeys[1] = "DEMO"
	WorldFileDescValues[1] = "DEMO       Demo of the ZZT World Editor"
	WorldFileDescKeys[2] = "CAVES"
	WorldFileDescValues[2] = "CAVES      The Caves of ZZT"
	WorldFileDescKeys[3] = "DUNGEONS"
	WorldFileDescValues[3] = "DUNGEONS   The Dungeons of ZZT"
	WorldFileDescKeys[4] = "CITY"
	WorldFileDescValues[4] = "CITY       Underground City of ZZT"
	WorldFileDescKeys[5] = "BEST"
	WorldFileDescValues[5] = "BEST       The Best of ZZT"
	WorldFileDescKeys[6] = "TOUR"
	WorldFileDescValues[6] = "TOUR       Guided Tour ZZT's Other Worlds"
	Randomize()
	SetCBreak(false)
	InitialTextAttr = TextAttr
	StartupWorldFileName = "TOWN"
	ResourceDataFileName = "ZZT.DAT"
	ResetConfig = false
	GameTitleExitRequested = false
	GameConfigure()
	ParseArguments()
	if !GameTitleExitRequested {
		VideoInstall(80, Blue)
		OrderPrintId = &GameVersion
		TextWindowInit(5, 3, 50, 18)
		VideoHideCursor()
		ClrScr()
		TickSpeed = 4
		DebugEnabled = false
		SavedGameFileName = "SAVED"
		SavedBoardFileName = "TEMP"
		GenerateTransitionTable()
		WorldCreate()
		GameTitleLoop()
	}
	SoundClearQueue()
	TextAttr = InitialTextAttr
	ClrScr()
	if Length(ConfigRegistration) == 0 {
		GamePrintRegisterMessage()
	} else {
		WriteLn("")
		WriteLn("  Registered version -- Thank you for playing ZZT.")
		WriteLn("")
	}
	VideoShowCursor()
}
