package passport

type Sex string

const Male Sex = "male"
const Female Sex = "female"

type UserInfo struct {
	FirstName       string `json:"first_name"`
	LastName        string `json:"last_name"`
	DisplayName     string `json:"display_name"`
	RealName        string `json:"real_name"`
	IsAvatarEmpty   bool   `json:"is_avatar_empty"`
	DefaultAvatarId string `json:"default_avatar_id"`
	Login           string `json:"login"`
	Sex             Sex    `json:"sex"`
	Id              string `json:"id,string"`
}

type Passport interface {
	ResolveUserInfo(token string) (UserInfo, error)
}

type UnauthorizedError struct {
	msg string
}

func (this UnauthorizedError) Error() string {
	if this.msg == "" {
		return "Unauthorized"
	}
	return "Unauthorized: " + this.msg
}
