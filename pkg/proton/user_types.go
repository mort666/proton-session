package proton

type User struct {
	ID          string
	Name        string
	DisplayName string
	Email       string
	Keys        Keys

	UsedSpace uint64
	MaxSpace  uint64
	MaxUpload uint64

	Credit   int
	Currency string

	ProductUsedSpace ProductUsedSpace
}

type DeleteUserReq struct {
	Reason   string
	Feedback string
	Email    string
}

type ProductUsedSpace struct {
	Calendar uint64
	Contact  uint64
	Drive    uint64
	Mail     uint64
	Pass     uint64
}

type Keys []Key

type Key struct {
	ID         string
	PrivateKey []byte
	Token      string
	Signature  string
	Primary    Bool
	Active     Bool
	Flags      KeyState
}

type KeyState int

const (
	KeyStateTrusted KeyState = 1 << iota // 2^0 = 1 means the key is not compromised (i.e. if we can trust signatures coming from it)
	KeyStateActive                       // 2^1 = 2 means the key is still in use (i.e. not obsolete, we can encrypt messages to it)
)