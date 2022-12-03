package sentry

import "github.com/getsentry/sentry-go"

func NewHub() (*sentry.Hub, error) {
	c, err := sentry.NewClient(sentry.ClientOptions{SendDefaultPII: true})
	if err != nil {
		return nil, err
	}

	return sentry.NewHub(c, sentry.NewScope()), nil
}
