package twilio

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/twilio/twilio-go"
)

type Verify struct {
	twilio *twilio.RestClient

	phone string
	sid   string

	log *logrus.Logger
}

func NewVerify(twilio *twilio.RestClient, phone string, sid string, log *logrus.Logger) *Verify {
	return &Verify{
		twilio: twilio,

		phone: phone,
		sid:   sid,

		log: log,
	}
}

func (t *Verify) SendCode(ctx context.Context, code string, phone string) error {
	// validParams := &api.CreateValidationRequestParams{}
	// validParams.SetFriendlyName("Social Network")
	// validParams.SetPhoneNumber(phone)

	// _, err := t.twilio.Api.CreateValidationRequest(validParams)
	// if !errors.Is(err, core.ErrTwilioVerifyPhone) {
	// 	params := &api.CreateMessageParams{}
	// 	params.SetTo(phone)
	// 	params.SetFrom(t.phone)
	// 	params.SetBody("Your verification code is: " + code)

	// 	_, err = t.twilio.Api.CreateMessage(params)

	// 	return err
	// } else if err != nil {
	// 	return err
	// }

	// params := &api.CreateMessageParams{}
	// params.SetTo(phone)
	// params.SetFrom(t.phone)
	// params.SetBody("Your verification code is: " + code)

	// logrus.Info(params)

	// _, err := t.twilio.Api.CreateMessage(params)

	t.log.WithFields(logrus.Fields{"phone": phone, "code": code}).Info("sent code")

	return nil
}
