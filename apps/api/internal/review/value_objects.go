package review

import "errors"

type Rating int32

func NewRating(value int32) (Rating, error) {
	if value < 1 || value > 5 {
		return 0, errors.New("rating must be between 1 and 5")
	}
	return Rating(value), nil
}

func (r Rating) Value() int32 { return int32(r) }
