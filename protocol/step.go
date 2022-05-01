package protocol

type Execute struct {
	Comment string
	Use     string
	Run     string
}

type Deploy struct {
	Servers  []*Server
	Mappers  []*Mapper
	Executes []*Execute
}

type Mapper struct {
	Source string
	Target string
}
