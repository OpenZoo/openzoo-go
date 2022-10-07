package main

func HighScoresClear() {
	for i := 0; i < HIGH_SCORE_COUNT; i++ {
		HighScoreList[i].Name = ""
		HighScoreList[i].Score = -1
	}
}

func HighScoresLoad() {
	f, err := VfsOpen(World.Info.Name + ".HI")
	if err != nil {
		HighScoresClear()
		return
	}
	defer f.Close()

	for i := 0; i < HIGH_SCORE_COUNT; i++ {
		err = ReadHighScoreEntry(f, &HighScoreList[i])
		if err != nil {
			HighScoresClear()
			return
		}
	}
}

func HighScoresSave() {
	f, err := VfsCreate(World.Info.Name + ".HI")
	if err != nil {
		return
	}
	defer f.Close()

	for i := 0; i < HIGH_SCORE_COUNT; i++ {
		err = WriteHighScoreEntry(f, HighScoreList[i])
		if err != nil {
			return
		}
	}
}

func HighScoresInitTextWindow(state *TTextWindowState) bool {
	var (
		i        int16
		scoreStr string
	)
	TextWindowInitState(state)
	TextWindowAppend(state, "Score  Name")
	TextWindowAppend(state, "-----  ----------------------------------")
	for i = 1; i <= HIGH_SCORE_COUNT; i++ {
		if Length(HighScoreList[i-1].Name) != 0 {
			scoreStr = StrWidth(HighScoreList[i-1].Score, 5)
			TextWindowAppend(state, scoreStr+"  "+HighScoreList[i-1].Name)
		}
	}
	return len(state.Lines) > 2
}

func HighScoresDisplay(linePos int) {
	var state TTextWindowState
	state.LinePos = linePos
	if HighScoresInitTextWindow(&state) {
		state.Title = "High scores for " + World.Info.Name
		TextWindowDrawOpen(&state)
		TextWindowSelect(&state, false, true)
		TextWindowDrawClose(&state)
	}
	TextWindowFree(&state)
}

func HighScoresAdd(score int16) {
	var (
		textWindow TTextWindowState
		name       string
	)
	listPos := 1
	for listPos <= HIGH_SCORE_COUNT && score < HighScoreList[listPos-1].Score {
		listPos++
	}
	if listPos <= HIGH_SCORE_COUNT && score > 0 {
		for i := HIGH_SCORE_COUNT - 1; i >= listPos; i-- {
			HighScoreList[i+1-1] = HighScoreList[i-1]
		}
		HighScoreList[listPos-1].Score = score
		HighScoreList[listPos-1].Name = "-- You! --"
		HighScoresInitTextWindow(&textWindow)
		textWindow.LinePos = listPos
		textWindow.Title = "New high score for " + World.Info.Name
		TextWindowDrawOpen(&textWindow)
		TextWindowDraw(&textWindow, false, false)
		name = ""
		PopupPromptString("Congratulations!  Enter your name:", &name)
		HighScoreList[listPos-1].Name = name
		HighScoresSave()
		TextWindowDrawClose(&textWindow)
		TransitionDrawToBoard()
		TextWindowFree(&textWindow)
	}
}
