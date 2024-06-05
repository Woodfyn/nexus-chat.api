package core

import "github.com/go-playground/validator/v10"

var validate *validator.Validate

func init() {
	validate = validator.New()
}

func (a *AuthRegister) Validate() error {
	return validate.Struct(a)
}

func (a *AuthLogin) Validate() error {
	return validate.Struct(a)
}

func (u *UpdateUserReq) Validate() error {
	return validate.Struct(u)
}

func (s *SendMessageReq) Validate() error {
	return validate.Struct(s)
}

func (c *CreateChatGroupReq) Validate() error {
	return validate.Struct(c)
}

func (c *CreateDefaultChatReq) Validate() error {
	return validate.Struct(c)
}

func (c *UpdateGroupChatAdminReq) Validate() error {
	return validate.Struct(c)
}

func (c *UpdateGroupChatNameReq) Validate() error {
	return validate.Struct(c)
}

func (c *JoinChatGroupReq) Validate() error {
	return validate.Struct(c)
}
