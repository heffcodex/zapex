package zfield

import (
	"github.com/getsentry/sentry-go"
	"github.com/heffcodex/zapex/consts"
	"github.com/pkg/errors"
	"github.com/valyala/bytebufferpool"
	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"net/http"
	"net/url"
	"strings"
)

const (
	HeaderCookie = fasthttp.HeaderCookie
)

var (
	HTTPRequestFieldBodyDumpLimitBytes int64 = 512 * 1024 * 1024
	httpRequestBodyPool                bytebufferpool.Pool
	sensitiveHeaders                   = map[string]bool{
		"Authorization":   true,
		"Cookie":          true,
		"X-Forwarded-For": true,
		"X-Real-Ip":       true,
	}

	_ zapcore.ObjectMarshaler = (*HTTPRequest)(nil)
	_ ZField                  = (*HTTPRequest)(nil)
)

type HTTPRequest struct {
	Method        string
	URI           *url.URL
	Proto         string
	ContentLength int64
	Headers       map[string][]string
	Body          *bytebufferpool.ByteBuffer
	DumpErrors    []error
}

func (f *HTTPRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("method", f.Method)
	enc.AddString("uri", f.URI.String())
	enc.AddString("proto", f.Proto)
	enc.AddInt64("content_length", f.ContentLength)

	enc.AddByteString("body", f.Body.B)
	httpRequestBodyPool.Put(f.Body)

	enc.AddString("dump_errors", strings.Join(errsToStrings(f.DumpErrors...), "; "))

	return nil
}

func (f *HTTPRequest) ToField() zap.Field {
	return zap.Field{Key: consts.KeyHTTPRequest, Type: zapcore.ObjectMarshalerType, Interface: f}
}

func (f *HTTPRequest) ToSentry(c *sentry.Client) *sentry.Request {
	r := &sentry.Request{
		URL:         f.URI.String(),
		Method:      f.Method,
		Data:        f.Body.String(),
		QueryString: f.URI.RawQuery,
		Cookies:     "",
		Headers:     make(map[string]string, len(f.Headers)),
		Env:         make(map[string]string),
	}

	if c.Options().SendDefaultPII {
		for _, cookie := range f.Headers[HeaderCookie] {
			r.Cookies += cookie + "; "
		}

		for k, v := range f.Headers {
			r.Headers[k] = strings.Join(v, ",")
		}
	} else {
		for k, v := range f.Headers {
			if _, ok := sensitiveHeaders[k]; !ok {
				r.Headers[k] = strings.Join(v, ",")
			}
		}
	}

	if len(f.DumpErrors) > 0 {
		r.Env["DUMP_ERRORS"] = strings.Join(errsToStrings(f.DumpErrors...), "; ")
	}

	return r
}

func NewHTTPRequest(r *http.Request) *HTTPRequest {
	if r == nil {
		return nil
	}

	f := &HTTPRequest{
		Method:        r.Method,
		URI:           r.URL,
		Proto:         r.Proto,
		ContentLength: r.ContentLength,
		Headers:       r.Header.Clone(),
		Body:          httpRequestBodyPool.Get(),
		DumpErrors:    nil,
	}

	read, err := f.Body.ReadFrom(r.Body)
	if err != nil {
		f.DumpErrors = append(f.DumpErrors, errors.Wrap(err, "cannot dump http.Request body to buffer"))
	} else if read > HTTPRequestFieldBodyDumpLimitBytes {
		f.Body.B = f.Body.B[:HTTPRequestFieldBodyDumpLimitBytes]
	}

	return f
}

func NewFastHTTPRequest(r *fasthttp.Request) *HTTPRequest {
	if r == nil {
		return nil
	}

	f := &HTTPRequest{
		Method:        string(r.Header.Method()),
		URI:           new(url.URL),
		Proto:         string(r.Header.Protocol()),
		ContentLength: int64(r.Header.ContentLength()),
		Headers:       make(map[string][]string, len(r.Header.Header())),
		Body:          httpRequestBodyPool.Get(),
		DumpErrors:    nil,
	}

	uri, err := url.ParseRequestURI(r.URI().String())
	if err != nil {
		f.DumpErrors = append(f.DumpErrors, errors.Wrap(err, "cannot parse fasthttp.Request URI"))
	} else {
		f.URI = uri
	}

	r.Header.VisitAll(func(key, value []byte) {
		f.Headers[string(key)] = append(f.Headers[string(key)], string(value))
	})

	body, err := r.BodyUncompressed()
	if err != nil {
		f.DumpErrors = append(f.DumpErrors, errors.Wrap(err, "cannot get uncompressed fasthttp.Request body"))
	} else if len(body) > 0 {
		_, err = f.Body.Write(body[:min(int64(len(body)), HTTPRequestFieldBodyDumpLimitBytes)])
		if err != nil {
			f.DumpErrors = append(f.DumpErrors, errors.Wrap(err, "cannot dump fasthttp.Request body to buffer"))
		}
	}

	return f
}
