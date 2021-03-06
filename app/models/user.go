package models

import (
	"context"
	"errors"
	"fmt"
	"gvue-scaffold/pkg/helper"
	"gvue-scaffold/pkg/log"
	"net/url"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
)

// User is user model
type User struct {
	Model
	EmailVerifiedAt *time.Time `json:"email_verified_at"`
	Name            string     `json:"name" gorm:"size:50"`
	Email           string     `json:"email" gorm:"size:50"`
	Password        string     `json:"-" gorm:"size:50"`
	Avatar          string     `json:"avatar"`
	Token           string     `json:"token"`
}

// FindUser find user by key
func FindUser(key, value string) (*User, bool) {
	var user User
	if mysqlCli.Where(key+" = ?", value).First(&user); user.ID > 0 {
		return &user, true
	}
	return nil, false
}

// Save a user
func (u *User) Save() error {
	return mysqlCli.Save(u).Error
}

// New a user
func (u *User) New() (uint, error) {
	if err := mysqlCli.Create(u).Error; err != nil {
		return 0, err
	}
	return u.ID, nil
}

// SendWelcomeEmail send welcome email to user
func (u *User) SendWelcomeEmail() {
	link, err := u.getSignedURL(TypeVerify)
	if err != nil {
		log.Error("get url sign error: ", err)
		return
	}
	body := fmt.Sprintf("<h3>%s您好:</h3><p>欢迎注册%s，请点击链接: <a href='%s'>%s</a> 进行确认邮箱</p><p>或者直接复制链接 %s 到浏览器打开</p><p>有效期30分钟</p>", u.Name, viper.GetString("app.name"), link, link, link)
	if err = helper.SendEmail(u.Email, "欢迎注册", body); err != nil {
		log.Error("send email: ", err)
	}
}

// SendVerifyEmail send verify email
func (u *User) SendVerifyEmail() error {
	link, err := u.getSignedURL(TypeVerify)
	if err != nil {
		log.Error("get url sign error: ", err)
		return err
	}
	body := fmt.Sprintf("<h3>%s您好:</h3><p>您申请验证邮箱，请点击链接: <a href='%s'>%s</a> 进行确认邮箱</p><p>或者直接复制链接 %s 到浏览器打开</p><p>有效期30分钟</p>", viper.GetString("app.name"), link, link, link)
	return helper.SendEmail(u.Email, "验证邮箱", body)
}

// SendResetEmail send reset password email to user
func (u *User) SendResetEmail() error {
	link, err := u.getSignedURL(TypeReset)
	if err != nil {
		return err
	}
	body := fmt.Sprintf("<h3>%s您好:</h3><p>您申请了重置密码，请点击链接: <a href='%s'>%s</a> 进行重置</p><p>或者直接复制链接 %s 到浏览器打开</p><p>有效期30分钟</p>", u.Name, link, link, link)
	return helper.SendEmail(u.Email, "重置密码", body)
}

var (
	keySignPrefix = "user:sign:%s:%s"
	// TypeReset 重置类型的url
	TypeReset = "reset"
	// TypeVerify 验证类型的url
	TypeVerify = "verify"
)

func (u *User) getSignedURL(t string) (string, error) {
	host := viper.GetString("app.url")
	//签名
	signstr := helper.Md5(u.Email)
	values := url.Values{}
	//redis key
	key := fmt.Sprintf(keySignPrefix, t, signstr)
	switch t {
	case TypeReset:
		host += "/password/reset"
		values.Set("email", u.Email)
	case TypeVerify:
		host += "/verification"
	default:
		return "", errors.New("签名失败")
	}
	values.Set("sign", signstr)
	query := values.Encode()
	//构造url
	myurlstr := host + "?" + query
	myurl, err := url.Parse(myurlstr)
	if err != nil {
		return "", err
	}
	//存到redis
	if _, err := redisCli.Set(context.Background(), key, u.Email, time.Minute*30).Result(); err != nil {
		return "", err
	}
	//解析url
	return myurl.String(), nil
}

// DecodeSignURL decode sign in url
func DecodeSignURL(t, sign string) (*User, error) {
	if t != TypeReset && t != TypeVerify {
		return nil, errors.New("链接错误")
	}
	key := fmt.Sprintf(keySignPrefix, t, sign)
	result, err := redisCli.Get(context.Background(), key).Result()
	if err == redis.Nil {
		return nil, errors.New("签名不存在或过期~链接错误")
	} else if err != nil {
		return nil, err
	}
	user, exists := FindUser("email", result)
	if !exists {
		return nil, errors.New("签名不存在或过期~链接错误")
	}
	return user, nil
}
