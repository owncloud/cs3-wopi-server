package app

import (
	"github.com/dchest/uniuri"
)

type Service struct {
	Namespace string
	Name      string `env:"WOPI_SERVICE_NAME"`
}

func (s Service) GetServiceFQDN() string {
	return s.Namespace + "." + s.Name
}

type GRPC struct {
	BindAddr string `env:"WOPI_GRPC_BIND_ADDR"`
}

type HTTP struct {
	Addr     string `env:"WOPI_HTTP_ADDR"`
	BindAddr string `env:"WOPI_HTTP_BIND_ADDR"`
	Scheme   string `env:"WOPI_HTTP_SCHEME"`
}

type WopiApp struct {
	Addr     string `env:"WOPI_APP_ADDR"`
	Insecure bool   `env:"WOPI_APP_INSECURE"`
}

type CS3api struct {
	GatewayServiceName     string `env:"WOPI_CS3API_GATEWAY_SERVICENAME"`
	CS3DataGatewayInsecure bool   `env:"WOPI_CS3API_DATA_GATEWAY_INSECURE"`
}

type Log struct {
	Level  string `env:"WOPI_LOG_LEVEL"`
	Pretty bool   `env:"WOPI_LOG_PRETTY"`
	Color  bool   `env:"WOPI_LOG_COLOR"`
	File   string `env:"WOPI_LOG_FILE"`
}

type Option func(o *Options)

type Options struct {
	AppName        string
	AppDescription string
	AppIcon        string
	AppLockName    string
	WopiSecret     string
	CS3api         CS3api
	Service        Service
	GRPC           GRPC
	HTTP           HTTP
	WopiApp        WopiApp
	Log            Log
}

func WithAppName(name string) Option {
	return func(o *Options) {
		o.AppName = name
	}
}

func WithAppDescription(description string) Option {
	return func(o *Options) {
		o.AppDescription = description
	}
}

func WithAppIcon(icon string) Option {
	return func(o *Options) {
		o.AppIcon = icon
	}
}

func WithAppLockName(lockName string) Option {
	return func(o *Options) {
		o.AppLockName = lockName
	}
}

func WithWopiSecret(wopiSecret string) Option {
	return func(o *Options) {
		o.WopiSecret = wopiSecret
	}
}

func WithCS3api(cs3api CS3api) Option {
	return func(o *Options) {
		o.CS3api = cs3api
	}
}

func WithService(service Service) Option {
	return func(o *Options) {
		o.Service = service
	}
}

func WithGRPC(grpc GRPC) Option {
	return func(o *Options) {
		o.GRPC = grpc
	}
}

func WithHTTP(http HTTP) Option {
	return func(o *Options) {
		o.HTTP = http
	}
}

func WithWopiApp(wopiApp WopiApp) Option {
	return func(o *Options) {
		o.WopiApp = wopiApp
	}
}

func WithLog(log Log) Option {
	return func(o *Options) {
		o.Log = log
	}
}

// Get a copy of the default options. You can use this method as a way
// to get a new `Options` instance, and then overwrite the values with
// any of the `With...` methods
func GetDefaultOptions() Options {
	return Options{
		AppName:        "WOPI app",
		AppDescription: "Open office documents with a WOPI app",
		AppIcon:        "image-edit",
		AppLockName:    "com.github.owncloud.cs3-wopi-server",
		WopiSecret:     uniuri.NewLen(32),
		CS3api: CS3api{
			GatewayServiceName:     "com.owncloud.api.gateway",
			CS3DataGatewayInsecure: true, // TODO: this should have a secure default
		},
		Service: Service{
			Namespace: "com.github.owncloud.cs3-wopi-server",
		},
		GRPC: GRPC{
			BindAddr: "127.0.0.1:5678",
		},
		HTTP: HTTP{
			Addr:     "127.0.0.1:6789",
			BindAddr: "127.0.0.1:6789",
			Scheme:   "http",
		},
		WopiApp: WopiApp{
			Addr:     "https://127.0.0.1:8080",
			Insecure: true, // TODO: this should have a secure default
		},
		Log: Log{
			Level:  "error",
			Pretty: true,
			Color:  true,
			File:   "",
		},
	}
}
