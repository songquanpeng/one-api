package model

import (
	"errors"
	"fmt"
	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/blacklist"
	"github.com/songquanpeng/one-api/common/config"
	"github.com/songquanpeng/one-api/common/helper"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/common/random"
	"gorm.io/gorm"
	"strings"
)

const (
	RoleGuestUser  = 0
	RoleCommonUser = 1
	RoleAdminUser  = 10
	RoleRootUser   = 100
)

const (
	UserStatusEnabled  = 1 // don't use 0, 0 is the default value!
	UserStatusDisabled = 2 // also don't use 0
	UserStatusDeleted  = 3
)

// User if you add sensitive fields, don't forget to clean them in setupLogin function.
// Otherwise, the sensitive information will be saved on local storage in plain text!
type User struct {
	Id               int    `json:"id"`
	Username         string `json:"username" gorm:"unique;index" validate:"max=12"`
	Password         string `json:"password" gorm:"not null;" validate:"min=8,max=20"`
	DisplayName      string `json:"display_name" gorm:"index" validate:"max=20"`
	Role             int    `json:"role" gorm:"type:int;default:1"`   // admin, util
	Status           int    `json:"status" gorm:"type:int;default:1"` // enabled, disabled
	Email            string `json:"email" gorm:"index" validate:"max=50"`
	GitHubId         string `json:"github_id" gorm:"column:github_id;index"`
	WeChatId         string `json:"wechat_id" gorm:"column:wechat_id;index"`
	LarkId           string `json:"lark_id" gorm:"column:lark_id;index"`
	VerificationCode string `json:"verification_code" gorm:"-:all"`                                    // this field is only for Email verification, don't save it to database!
	AccessToken      string `json:"access_token" gorm:"type:char(32);column:access_token;uniqueIndex"` // this token is for system management
	Quota            int64  `json:"quota" gorm:"bigint;default:0"`
	UsedQuota        int64  `json:"used_quota" gorm:"bigint;default:0;column:used_quota"` // used quota
	RequestCount     int    `json:"request_count" gorm:"type:int;default:0;"`             // request number
	Group            string `json:"group" gorm:"type:varchar(32);default:'default'"`
	AffCode          string `json:"aff_code" gorm:"type:varchar(32);column:aff_code;uniqueIndex"`
	InviterId        int    `json:"inviter_id" gorm:"type:int;column:inviter_id;index"`
}

func GetMaxUserId() int {
	var user User
	DB.Last(&user)
	return user.Id
}

func GetAllUsers(startIdx int, num int, order string) (users []*User, err error) {
	query := DB.Limit(num).Offset(startIdx).Omit("password").Where("status != ?", UserStatusDeleted)

	switch order {
	case "quota":
		query = query.Order("quota desc")
	case "used_quota":
		query = query.Order("used_quota desc")
	case "request_count":
		query = query.Order("request_count desc")
	default:
		query = query.Order("id desc")
	}

	err = query.Find(&users).Error
	return users, err
}

func SearchUsers(keyword string) (users []*User, err error) {
	if !common.UsingPostgreSQL {
		err = DB.Omit("password").Where("id = ? or username LIKE ? or email LIKE ? or display_name LIKE ?", keyword, keyword+"%", keyword+"%", keyword+"%").Find(&users).Error
	} else {
		err = DB.Omit("password").Where("username LIKE ? or email LIKE ? or display_name LIKE ?", keyword+"%", keyword+"%", keyword+"%").Find(&users).Error
	}
	return users, err
}

func GetUserById(id int, selectAll bool) (*User, error) {
	if id == 0 {
		return nil, errors.New("id 为空！")
	}
	user := User{Id: id}
	var err error = nil
	if selectAll {
		err = DB.First(&user, "id = ?", id).Error
	} else {
		err = DB.Omit("password").First(&user, "id = ?", id).Error
	}
	return &user, err
}

func GetUserIdByAffCode(affCode string) (int, error) {
	if affCode == "" {
		return 0, errors.New("affCode 为空！")
	}
	var user User
	err := DB.Select("id").First(&user, "aff_code = ?", affCode).Error
	return user.Id, err
}

func DeleteUserById(id int) (err error) {
	if id == 0 {
		return errors.New("id 为空！")
	}
	user := User{Id: id}
	return user.Delete()
}

func (user *User) Insert(inviterId int) error {
	var err error
	if user.Password != "" {
		user.Password, err = common.Password2Hash(user.Password)
		if err != nil {
			return err
		}
	}
	user.Quota = config.QuotaForNewUser
	user.AccessToken = random.GetUUID()
	user.AffCode = random.GetRandomString(4)
	result := DB.Create(user)
	if result.Error != nil {
		return result.Error
	}
	if config.QuotaForNewUser > 0 {
		RecordLog(user.Id, LogTypeSystem, fmt.Sprintf("新用户注册赠送 %s", common.LogQuota(config.QuotaForNewUser)))
	}
	if inviterId != 0 {
		if config.QuotaForInvitee > 0 {
			_ = IncreaseUserQuota(user.Id, config.QuotaForInvitee)
			RecordLog(user.Id, LogTypeSystem, fmt.Sprintf("使用邀请码赠送 %s", common.LogQuota(config.QuotaForInvitee)))
		}
		if config.QuotaForInviter > 0 {
			_ = IncreaseUserQuota(inviterId, config.QuotaForInviter)
			RecordLog(inviterId, LogTypeSystem, fmt.Sprintf("邀请用户赠送 %s", common.LogQuota(config.QuotaForInviter)))
		}
	}
	// create default token
	cleanToken := Token{
		UserId:         user.Id,
		Name:           "default",
		Key:            random.GenerateKey(),
		CreatedTime:    helper.GetTimestamp(),
		AccessedTime:   helper.GetTimestamp(),
		ExpiredTime:    -1,
		RemainQuota:    -1,
		UnlimitedQuota: true,
	}
	result.Error = cleanToken.Insert()
	if result.Error != nil {
		// do not block
		logger.SysError(fmt.Sprintf("create default token for user %d failed: %s", user.Id, result.Error.Error()))
	}
	return nil
}

func (user *User) Update(updatePassword bool) error {
	var err error
	if updatePassword {
		user.Password, err = common.Password2Hash(user.Password)
		if err != nil {
			return err
		}
	}
	if user.Status == UserStatusDisabled {
		blacklist.BanUser(user.Id)
	} else if user.Status == UserStatusEnabled {
		blacklist.UnbanUser(user.Id)
	}
	err = DB.Model(user).Updates(user).Error
	return err
}

func (user *User) Delete() error {
	if user.Id == 0 {
		return errors.New("id 为空！")
	}
	blacklist.BanUser(user.Id)
	user.Username = fmt.Sprintf("deleted_%s", random.GetUUID())
	user.Status = UserStatusDeleted
	err := DB.Model(user).Updates(user).Error
	return err
}

// ValidateAndFill check password & user status
func (user *User) ValidateAndFill() (err error) {
	// When querying with struct, GORM will only query with non-zero fields,
	// that means if your field’s value is 0, '', false or other zero values,
	// it won’t be used to build query conditions
	password := user.Password
	if user.Username == "" || password == "" {
		return errors.New("用户名或密码为空")
	}
	err = DB.Where("username = ?", user.Username).First(user).Error
	if err != nil {
		// we must make sure check username firstly
		// consider this case: a malicious user set his username as other's email
		err := DB.Where("email = ?", user.Username).First(user).Error
		if err != nil {
			return errors.New("用户名或密码错误，或用户已被封禁")
		}
	}
	okay := common.ValidatePasswordAndHash(password, user.Password)
	if !okay || user.Status != UserStatusEnabled {
		return errors.New("用户名或密码错误，或用户已被封禁")
	}
	return nil
}

func (user *User) FillUserById() error {
	if user.Id == 0 {
		return errors.New("id 为空！")
	}
	DB.Where(User{Id: user.Id}).First(user)
	return nil
}

func (user *User) FillUserByEmail() error {
	if user.Email == "" {
		return errors.New("email 为空！")
	}
	DB.Where(User{Email: user.Email}).First(user)
	return nil
}

func (user *User) FillUserByGitHubId() error {
	if user.GitHubId == "" {
		return errors.New("GitHub id 为空！")
	}
	DB.Where(User{GitHubId: user.GitHubId}).First(user)
	return nil
}

func (user *User) FillUserByLarkId() error {
	if user.LarkId == "" {
		return errors.New("lark id 为空！")
	}
	DB.Where(User{LarkId: user.LarkId}).First(user)
	return nil
}

func (user *User) FillUserByWeChatId() error {
	if user.WeChatId == "" {
		return errors.New("WeChat id 为空！")
	}
	DB.Where(User{WeChatId: user.WeChatId}).First(user)
	return nil
}

func (user *User) FillUserByUsername() error {
	if user.Username == "" {
		return errors.New("username 为空！")
	}
	DB.Where(User{Username: user.Username}).First(user)
	return nil
}

func IsEmailAlreadyTaken(email string) bool {
	return DB.Where("email = ?", email).Find(&User{}).RowsAffected == 1
}

func IsWeChatIdAlreadyTaken(wechatId string) bool {
	return DB.Where("wechat_id = ?", wechatId).Find(&User{}).RowsAffected == 1
}

func IsGitHubIdAlreadyTaken(githubId string) bool {
	return DB.Where("github_id = ?", githubId).Find(&User{}).RowsAffected == 1
}

func IsLarkIdAlreadyTaken(githubId string) bool {
	return DB.Where("lark_id = ?", githubId).Find(&User{}).RowsAffected == 1
}

func IsUsernameAlreadyTaken(username string) bool {
	return DB.Where("username = ?", username).Find(&User{}).RowsAffected == 1
}

func ResetUserPasswordByEmail(email string, password string) error {
	if email == "" || password == "" {
		return errors.New("邮箱地址或密码为空！")
	}
	hashedPassword, err := common.Password2Hash(password)
	if err != nil {
		return err
	}
	err = DB.Model(&User{}).Where("email = ?", email).Update("password", hashedPassword).Error
	return err
}

func IsAdmin(userId int) bool {
	if userId == 0 {
		return false
	}
	var user User
	err := DB.Where("id = ?", userId).Select("role").Find(&user).Error
	if err != nil {
		logger.SysError("no such user " + err.Error())
		return false
	}
	return user.Role >= RoleAdminUser
}

func IsUserEnabled(userId int) (bool, error) {
	if userId == 0 {
		return false, errors.New("user id is empty")
	}
	var user User
	err := DB.Where("id = ?", userId).Select("status").Find(&user).Error
	if err != nil {
		return false, err
	}
	return user.Status == UserStatusEnabled, nil
}

func ValidateAccessToken(token string) (user *User) {
	if token == "" {
		return nil
	}
	token = strings.Replace(token, "Bearer ", "", 1)
	user = &User{}
	if DB.Where("access_token = ?", token).First(user).RowsAffected == 1 {
		return user
	}
	return nil
}

func GetUserQuota(id int) (quota int64, err error) {
	err = DB.Model(&User{}).Where("id = ?", id).Select("quota").Find(&quota).Error
	return quota, err
}

func GetUserUsedQuota(id int) (quota int64, err error) {
	err = DB.Model(&User{}).Where("id = ?", id).Select("used_quota").Find(&quota).Error
	return quota, err
}

func GetUserEmail(id int) (email string, err error) {
	err = DB.Model(&User{}).Where("id = ?", id).Select("email").Find(&email).Error
	return email, err
}

func GetUserGroup(id int) (group string, err error) {
	groupCol := "`group`"
	if common.UsingPostgreSQL {
		groupCol = `"group"`
	}

	err = DB.Model(&User{}).Where("id = ?", id).Select(groupCol).Find(&group).Error
	return group, err
}

func IncreaseUserQuota(id int, quota int64) (err error) {
	if quota < 0 {
		return errors.New("quota 不能为负数！")
	}
	if config.BatchUpdateEnabled {
		addNewRecord(BatchUpdateTypeUserQuota, id, quota)
		return nil
	}
	return increaseUserQuota(id, quota)
}

func increaseUserQuota(id int, quota int64) (err error) {
	err = DB.Model(&User{}).Where("id = ?", id).Update("quota", gorm.Expr("quota + ?", quota)).Error
	return err
}

func DecreaseUserQuota(id int, quota int64) (err error) {
	if quota < 0 {
		return errors.New("quota 不能为负数！")
	}
	if config.BatchUpdateEnabled {
		addNewRecord(BatchUpdateTypeUserQuota, id, -quota)
		return nil
	}
	return decreaseUserQuota(id, quota)
}

func decreaseUserQuota(id int, quota int64) (err error) {
	err = DB.Model(&User{}).Where("id = ?", id).Update("quota", gorm.Expr("quota - ?", quota)).Error
	return err
}

func GetRootUserEmail() (email string) {
	DB.Model(&User{}).Where("role = ?", RoleRootUser).Select("email").Find(&email)
	return email
}

func UpdateUserUsedQuotaAndRequestCount(id int, quota int64) {
	if config.BatchUpdateEnabled {
		addNewRecord(BatchUpdateTypeUsedQuota, id, quota)
		addNewRecord(BatchUpdateTypeRequestCount, id, 1)
		return
	}
	updateUserUsedQuotaAndRequestCount(id, quota, 1)
}

func updateUserUsedQuotaAndRequestCount(id int, quota int64, count int) {
	err := DB.Model(&User{}).Where("id = ?", id).Updates(
		map[string]interface{}{
			"used_quota":    gorm.Expr("used_quota + ?", quota),
			"request_count": gorm.Expr("request_count + ?", count),
		},
	).Error
	if err != nil {
		logger.SysError("failed to update user used quota and request count: " + err.Error())
	}
}

func updateUserUsedQuota(id int, quota int64) {
	err := DB.Model(&User{}).Where("id = ?", id).Updates(
		map[string]interface{}{
			"used_quota": gorm.Expr("used_quota + ?", quota),
		},
	).Error
	if err != nil {
		logger.SysError("failed to update user used quota: " + err.Error())
	}
}

func updateUserRequestCount(id int, count int) {
	err := DB.Model(&User{}).Where("id = ?", id).Update("request_count", gorm.Expr("request_count + ?", count)).Error
	if err != nil {
		logger.SysError("failed to update user request count: " + err.Error())
	}
}

func GetUsernameById(id int) (username string) {
	DB.Model(&User{}).Where("id = ?", id).Select("username").Find(&username)
	return username
}
