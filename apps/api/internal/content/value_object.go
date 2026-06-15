package content

type MaturityRating string

const (
	MaturityRatingL  MaturityRating = "L"
	MaturityRating10 MaturityRating = "10"
	MaturityRating12 MaturityRating = "12"
	MaturityRating14 MaturityRating = "14"
	MaturityRating16 MaturityRating = "16"
	MaturityRating18 MaturityRating = "18"
)

func (m MaturityRating) IsKidsFriendly() bool { return m == MaturityRatingL }
func (m MaturityRating) String() string       { return string(m) }

type ContentType string

const (
	ContentTypeMovie  ContentType = "MOVIE"
	ContentTypeSeries ContentType = "SERIES"
)

func (c ContentType) IsMovie() bool { return c == ContentTypeMovie }

func (c ContentType) IsSeries() bool { return c == ContentTypeSeries }

func (c ContentType) String() string { return string(c) }
