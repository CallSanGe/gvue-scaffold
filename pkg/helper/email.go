package helper

import (
	"gvue-scaffold/pkg/log"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/dm"
	"github.com/spf13/viper"
)

// SendEmail send a email
func SendEmail(to, sub, body string) error {
	accessKeyID := viper.GetString("email.id")
	accessKeySecret := viper.GetString("email.secret")
	if accessKeyID == "" || accessKeySecret == "" {
		log.Warn("email key is empty")
		return nil
	}

	client, err := dm.NewClientWithAccessKey("cn-hangzhou", accessKeyID, accessKeySecret)
	if err != nil {
		return err
	}
	request := dm.CreateSingleSendMailRequest()
	request.Scheme = "https"
	request.AccountName = viper.GetString("email.from")
	request.AddressType = requests.NewInteger(1)
	request.ToAddress = to
	request.Subject = sub
	request.FromAlias = viper.GetString("app.name")
	request.HtmlBody = body
	request.ReplyToAddress = requests.NewBoolean(true)

	response, err := client.SingleSendMail(request)
	if err != nil {
		log.Error("send email error: ", err)
		return err
	}
	log.Debug("response is: ", response.String())
	return nil
}
