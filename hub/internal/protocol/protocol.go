package protocol

type NestJson struct {
	Name         string
	Private      bool
	Version      string
	Author       string
	Description  string
	License      string
	Dist         string
	Dependencies map[string]string
}
