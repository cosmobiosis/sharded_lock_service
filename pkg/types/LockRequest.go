package types

type RWFlag string
const (
	READ RWFlag = "read"
	WRITE     = "write"
)

type LockRequest struct {
	Key    string
	Rwflag   RWFlag
	ClientId string
}
