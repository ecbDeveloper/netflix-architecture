package episode

import "errors"

type Season int32

func NewSeason(value int32) (Season, error) {
	if value <= 0 {
		return 0, errors.New("season must be greater than zero")
	}
	return Season(value), nil
}

func (s Season) Value() int32 { return int32(s) }

type EpisodeNumber int32

func NewEpisodeNumber(value int32) (EpisodeNumber, error) {
	if value <= 0 {
		return 0, errors.New("episode number must be greater than zero")
	}
	return EpisodeNumber(value), nil
}

func (e EpisodeNumber) Value() int32 { return int32(e) }

type Duration int32

func NewDuration(value int32) (Duration, error) {
	if value <= 0 {
		return 0, errors.New("duration must be greater than zero")
	}
	return Duration(value), nil
}

func (d Duration) Value() int32 { return int32(d) }
