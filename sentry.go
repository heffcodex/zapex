package zapex

import "github.com/getsentry/sentry-go"

func newSentryHub() (*sentry.Hub, error) {
	c, err := sentry.NewClient(sentry.ClientOptions{})
	if err != nil {
		return nil, err
	}

	return sentry.NewHub(c, sentry.NewScope()), nil
}
