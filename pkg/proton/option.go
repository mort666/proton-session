package proton

import (
	"net/http"

	"github.com/ProtonMail/gluon/async"
	"github.com/go-resty/resty/v2"
)

// Option represents a type that can be used to Configure the
type Option interface {
	Config(*ManagerBuilder)
}

func WithHostURL(hostURL string) Option {
	return &withHostURL{
		hostURL: hostURL,
	}
}

type withHostURL struct {
	hostURL string
}

func (opt withHostURL) Config(b *ManagerBuilder) {
	b.HostURL = opt.hostURL
}

func WithAppVersion(appVersion string) Option {
	return &withAppVersion{
		appVersion: appVersion,
	}
}

type withUserAgent struct {
	userAgent string
}

func (opt withUserAgent) Config(b *ManagerBuilder) {
	b.UserAgent = opt.userAgent
}

func WithUserAgent(userAgent string) Option {
	return &withUserAgent{
		userAgent: userAgent,
	}
}

type withAppVersion struct {
	appVersion string
}

func (opt withAppVersion) Config(b *ManagerBuilder) {
	b.AppVersion = opt.appVersion
}

func WithTransport(transport http.RoundTripper) Option {
	return &withTransport{
		transport: transport,
	}
}

type withTransport struct {
	transport http.RoundTripper
}

func (opt withTransport) Config(b *ManagerBuilder) {
	b.Transport = opt.transport
}

type withSkipVerifyProofs struct {
	skipVerifyProofs bool
}

func (opt withSkipVerifyProofs) Config(b *ManagerBuilder) {
	b.VerifyProofs = !opt.skipVerifyProofs
}

func WithSkipVerifyProofs() Option {
	return &withSkipVerifyProofs{
		skipVerifyProofs: true,
	}
}

func WithRetryCount(retryCount int) Option {
	return &withRetryCount{
		retryCount: retryCount,
	}
}

type withRetryCount struct {
	retryCount int
}

func (opt withRetryCount) Config(b *ManagerBuilder) {
	b.RetryCount = opt.retryCount
}

func WithCookieJar(jar http.CookieJar) Option {
	return &withCookieJar{
		jar: jar,
	}
}

type withCookieJar struct {
	jar http.CookieJar
}

func (opt withCookieJar) Config(b *ManagerBuilder) {
	b.CookieJar = opt.jar
}

func WithLogger(logger resty.Logger) Option {
	return &withLogger{
		logger: logger,
	}
}

type withLogger struct {
	logger resty.Logger
}

func (opt withLogger) Config(b *ManagerBuilder) {
	b.Logger = opt.logger
}

func WithDebug(debug bool) Option {
	return &withDebug{
		debug: debug,
	}
}

type withDebug struct {
	debug bool
}

func (opt withDebug) Config(b *ManagerBuilder) {
	b.Debug = opt.debug
}

func WithPanicHandler(panicHandler async.PanicHandler) Option {
	return &withPanicHandler{
		panicHandler: panicHandler,
	}
}

type withPanicHandler struct {
	panicHandler async.PanicHandler
}

func (opt withPanicHandler) Config(b *ManagerBuilder) {
	b.PanicHandler = opt.panicHandler
}
