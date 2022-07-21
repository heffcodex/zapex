package sentry

import "github.com/getsentry/sentry-go"

func makeExceptions(key string, errs ...error) []sentry.Exception {
	exceptions := make([]sentry.Exception, 0, len(errs))

	for _, err := range errs {
		exceptions = append(exceptions, sentry.Exception{
			Type:       key,
			Value:      err.Error(),
			Stacktrace: prepareStacktrace(err),
		})
	}

	return exceptions
}

func prepareStacktrace(err error) *sentry.Stacktrace {
	stack := sentry.ExtractStacktrace(err)
	if stack != nil {
		return stack
	}

	stack = sentry.NewStacktrace()
	stack.Frames = stack.Frames[:len(stack.Frames)-omitHeadStackFrames]

	return stack
}
