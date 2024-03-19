package dto

import "owlllovo/ginessential/model"

type UserDto struct {
	Name      string `json:"name"`
	Telephone string `json:"telephone"`
	Role      string `json:"role"`
}

func ToUserDto(user model.User) UserDto {
	return UserDto{
		Name:      user.Name,
		Telephone: user.Telephone,
		Role:      user.Role,
	}
}
