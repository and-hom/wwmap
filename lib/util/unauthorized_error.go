package util

type UnauthorizedError struct {
	msg string
}

func (this UnauthorizedError) Error() string {
	if this.msg == "" {
		return "Unauthorized"
	}
	return "Unauthorized: " + this.msg
}
