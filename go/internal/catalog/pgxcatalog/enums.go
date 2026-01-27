package pgxcatalog

type enum interface {
	IsValid() bool
	String() string
}

type workKind string

const (
	workKindMovie        workKind = "movie"
	workKindMovieEdition workKind = "movie_edition"
	workKindExtra        workKind = "extra"
)

func (wk workKind) IsValid() bool {
	switch wk {
	case workKindMovie, workKindMovieEdition, workKindExtra:
		return true
	default:
		return false
	}
}

func (wk workKind) String() string {
	return string(wk)
}

type sourceKind string

const (
	sourceKindFile sourceKind = "file"
	sourceKindDisc sourceKind = "disc"
)

func (sk sourceKind) IsValid() bool {
	switch sk {
	case sourceKindFile, sourceKindDisc:
		return true
	default:
		return false
	}
}

func (sk sourceKind) String() string {
	return string(sk)
}

type planKind string

const (
	planKindDirect       planKind = "direct"
	planKindChapterRange planKind = "chapter_range"
)

func (pk planKind) IsValid() bool {
	switch pk {
	case planKindDirect, planKindChapterRange:
		return true
	default:
		return false
	}
}
func (pk planKind) String() string {
	return string(pk)
}
