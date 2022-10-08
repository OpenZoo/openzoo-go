package main

import "github.com/OpenZoo/openzoo-go/format"

func HighScoresLoad() {
	f, err := VfsOpen(World.Info.Name + ".HI")
	if err != nil {
		HighScoreList = format.NewHighScoreList(HIGH_SCORE_COUNT)
		return
	}
	defer f.Close()

	HighScoreList, _ = format.HighScoresDeserialize(f)
}

func HighScoresSave() {
	f, err := VfsCreate(World.Info.Name + ".HI")
	if err != nil {
		return
	}
	defer f.Close()

	format.HighScoresSerialize(f, HighScoreList)
}

func HighScoresInitTextWindow(state *TTextWindowState) bool {
	var (
		i        int16
		scoreStr string
	)
	state.Init()
	state.Append("Score  Name")
	state.Append("-----  ----------------------------------")
	for i = 1; i <= HIGH_SCORE_COUNT; i++ {
		if Length(HighScoreList[i-1].Name) != 0 {
			scoreStr = StrWidth(HighScoreList[i-1].Score, 5)
			state.Append(scoreStr + "  " + HighScoreList[i-1].Name)
		}
	}
	return len(state.Lines) > 2
}

func HighScoresDisplay(linePos int) {
	var state TTextWindowState
	state.LinePos = linePos
	if HighScoresInitTextWindow(&state) {
		state.Title = "High scores for " + World.Info.Name
		state.DrawOpen()
		state.Select(false, true)
		state.DrawClose()
	}
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
		textWindow.DrawOpen()
		textWindow.Draw(false, false)
		name = ""
		PopupPromptString("Congratulations!  Enter your name:", &name)
		HighScoreList[listPos-1].Name = name
		HighScoresSave()
		textWindow.DrawClose()
		TransitionDrawToBoard()
	}
}
