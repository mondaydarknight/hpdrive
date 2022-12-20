package app

type ColOrdering int

const (
	FileName ColOrdering = iota
	Size
	LastModified
)

func (c ColOrdering) String() string {
	return [...]string{"fileName", "size", "lastModified"}[c]
}

type OrderDirection int

const (
	Ascending OrderDirection = iota
	Descending
)

func (o OrderDirection) String() string {
	return [...]string{"ASC", "DESC"}[o]
}

var OrderDirectionMap = map[string]OrderDirection{
	"Ascending":  Ascending,
	"Descending": Descending,
}

type Dir struct {
	IsDirectory bool     `json:"isDirectory"`
	Files       []string `json:"files"`
}
