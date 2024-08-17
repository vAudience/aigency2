package models

// define TokenDirection as enum
type TokenDirection string //@name TokenDirection

const (
	InputToken  TokenDirection = "input"
	OutputToken TokenDirection = "output"
)

func (td TokenDirection) String() string {
	return string(td)
}

func (td TokenDirection) Set(value string) {
	switch value {
	case "input":
		td = InputToken
	case "output":
		td = OutputToken
	}
}

func (td TokenDirection) Equals(compareDirectionString string) bool {
	return td.String() == compareDirectionString
}
