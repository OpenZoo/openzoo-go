package format

import "io"

const (
	HIGH_SCORE_COUNT = 30
)

func NewHighScoreList(count int) []THighScoreEntry {
	list := make([]THighScoreEntry, count)
	for i := 0; i < count; i++ {
		list[i].Name = ""
		list[i].Score = -1
	}
	return list
}

func HighScoresDeserialize(r io.Reader) ([]THighScoreEntry, error) {
	list := NewHighScoreList(HIGH_SCORE_COUNT)

	for i := 0; i < len(list); i++ {
		err := ReadHighScoreEntry(r, &list[i])
		if err != nil {
			return NewHighScoreList(HIGH_SCORE_COUNT), err
		}
	}

	return list, nil
}

func HighScoresSerialize(w io.Writer, list []THighScoreEntry) error {
	for i := 0; i < len(list); i++ {
		err := WriteHighScoreEntry(w, list[i])
		if err != nil {
			return err
		}
	}
	return nil
}
