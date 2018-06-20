package services

import (
	"github.com/XMatrixStudio/Coffee/App/models"
	"github.com/XMatrixStudio/Violet.SDK.Go"
)

// UserService 用户服务层
type UserService interface {
	InitViolet(c violetSdk.Config)
	GetLoginURL(backURL string) (url, state string)
	LoginByCode(code string) (userID string, err error)
	GetUserInfo(id string) (user models.Users, err error)
}

type userService struct {
	Violet  violetSdk.Violet
	Model   *models.UserModel
	Service *Service
}

func (s *userService) InitViolet(c violetSdk.Config) {
	s.Violet = violetSdk.NewViolet(c)
}

func (s *userService) GetLoginURL(backURL string) (url, state string) {
	url, state = s.Violet.GetLoginURL(backURL)
	return
}

func (s *userService) LoginByCode(code string) (userID string, err error) {
	// 获取用户Token
	res, err := s.Violet.GetToken(code)
	if err != nil {
		return
	}
	// 保存数据并获取用户信息
	user, err := s.Model.GetUserByVID(res.UserID)
	if err == nil { // 数据库已存在该用户
		userID = user.ID.Hex()
		s.Model.SetUserToken(user.ID.Hex(), res.Token)
	} else if err.Error() == "not found" { // 数据库不存在此用户
		userNew, errN := s.Violet.GetUserBaseInfo(res.UserID, res.Token)
		if errN != nil {
			err = errN
			return
		}
		userBsonID, errN := s.Model.AddUser(res.UserID, res.Token, userNew.Email, userNew.Name, userNew.Info.Avatar, userNew.Info.Bio, userNew.Info.Gender)
		err = errN
		userID = userBsonID.Hex()
	}
	return
}

func (s *userService) GetUserInfo(id string) (user models.Users, err error) {
	user, err = s.Model.GetUserByID(id)
	return
}
