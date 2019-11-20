package mail

import "github.com/jeckbjy/gsk/apm/alert"

// TODO: implement me
// https://github.com/golang/go/wiki/SendingMail
type mailAlert struct {
}

func (a *mailAlert) Name() string {
	return "mail"
}

func (a *mailAlert) Send(event *alert.Event) error {
	return nil
}
