package models

type SecretType int

const (
	LoginPasswordType SecretType = iota
	TextDataType
	BinaryDataType
	BankCardType
)

func (st SecretType) String() string {
	switch st {
	case LoginPasswordType:
		return "login"
	case TextDataType:
		return "text"
	case BinaryDataType:
		return "binary"
	case BankCardType:
		return "bankcard"
	default:
		return "unknown"
	}
}

type Secret struct {
	ID       int        `json:"id"`
	UserID   int        `json:"user_id"`
	Type     SecretType `json:"type"`
	Data     []byte     `json:"data"`
	Metadata string     `json:"metadata"`
}
