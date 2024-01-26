package app

import (
	"context"
	"errors"

	"github.com/owncloud/ocis/v2/ocis-pkg/log"

	registryv1beta1 "github.com/cs3org/go-cs3apis/cs3/app/registry/v1beta1"
	gatewayv1beta1 "github.com/cs3org/go-cs3apis/cs3/gateway/v1beta1"
	rpcv1beta1 "github.com/cs3org/go-cs3apis/cs3/rpc/v1beta1"
	"github.com/cs3org/reva/v2/pkg/mime"
	"github.com/cs3org/reva/v2/pkg/rgrpc/todo/pool"
	"github.com/gofrs/uuid"
	"github.com/owncloud/ocis/v2/ocis-pkg/config/envdecode"
	"github.com/owncloud/ocis/v2/ocis-pkg/registry"
	"google.golang.org/grpc"
)

type Config struct {
	Service
	GRPC
	HTTP
	WopiApp
	CS3api

	WopiSecret     string `env:"WOPI_SECRET"` // used as jwt secret and to encrypt access tokens
	AppName        string `env:"WOPI_APP_NAME"`
	AppDescription string `env:"WOPI_APP_DESCRIPTION"`
	AppIcon        string `env:"WOPI_APP_ICON"`
	AppLockName    string `env:"WOPI_APP_LOCK_NAME"`
}

type demoApp struct {
	gwc        gatewayv1beta1.GatewayAPIClient
	grpcServer *grpc.Server

	appURLs map[string]map[string]string

	Config Config

	Logger log.Logger
}

func New(opts ...Option) (*demoApp, error) {
	// get the default options
	options := GetDefaultOptions()

	// overwrite values with the env variables
	err := envdecode.Decode(&options)
	if err != nil {
		if !errors.Is(err, envdecode.ErrNoTargetFieldsAreSet) {
			return nil, err
		}
	}

	// a second overwrite with the options passed as parameters
	for _, o := range opts {
		o(&options)
	}

	// create instance using the final options
	app := &demoApp{
		Config: Config{
			AppName:        options.AppName,
			AppDescription: options.AppDescription,
			AppIcon:        options.AppIcon,
			AppLockName:    options.AppLockName,
			WopiSecret:     options.WopiSecret,
			CS3api:         options.CS3api,
			Service:        options.Service,
			GRPC:           options.GRPC,
			HTTP:           options.HTTP,
			WopiApp:        options.WopiApp,
		},
	}

	// configure the app logger with the options
	app.Logger = log.NewLogger(
		log.Name("wopiserver"), // currently hardcoded
		log.Level(options.Log.Level),
		log.Pretty(options.Log.Pretty),
		log.Color(options.Log.Color),
		log.File(options.Log.File),
	)
	return app, nil
}

func (app *demoApp) GetCS3apiClient() error {
	// establish a connection to the cs3 api endpoint
	// in this case a REVA gateway, started by oCIS
	gwc, err := pool.GetGatewayServiceClient(app.Config.CS3api.GatewayServiceName)
	if err != nil {
		return err
	}
	app.gwc = gwc

	return nil
}

func (app *demoApp) RegisterOcisService(ctx context.Context) error {
	svc := registry.BuildGRPCService(app.Config.Service.GetServiceFQDN(), uuid.Must(uuid.NewV4()).String(), app.Config.GRPC.BindAddr, "0.0.0")
	return registry.RegisterService(ctx, svc, app.Logger)
}

func (app *demoApp) RegisterDemoApp(ctx context.Context) error {
	mimeTypesMap := make(map[string]bool)
	for _, extensions := range app.appURLs {
		for ext := range extensions {
			m := mime.Detect(false, ext)
			mimeTypesMap[m] = true
		}
	}

	mimeTypes := make([]string, 0, len(mimeTypesMap))
	for m := range mimeTypesMap {
		mimeTypes = append(mimeTypes, m)
	}

	// TODO: REVA has way to filter supported mimetypes (do we need to implement it here or is it in the registry?)

	// TODO: an added app provider shouldn't last forever. Instead the registry should use a TTL
	// and delete providers that didn't register again. If an app provider dies or get's disconnected,
	// the users will be no longer available to choose to open a file with it (currently, opening a file just fails)
	req := &registryv1beta1.AddAppProviderRequest{
		Provider: &registryv1beta1.ProviderInfo{
			Name:        app.Config.AppName,
			Description: app.Config.AppDescription,
			Icon:        app.Config.AppIcon,
			Address:     app.Config.Service.GetServiceFQDN(),
			MimeTypes:   mimeTypes,
		},
	}

	resp, err := app.gwc.AddAppProvider(ctx, req)
	if err != nil {
		app.Logger.Error().Err(err).Msg("AddAppProvider failed")
		return err
	}

	if resp.Status.Code != rpcv1beta1.Code_CODE_OK {
		app.Logger.Error().Str("status_code", resp.Status.Code.String()).Msg("AddAppProvider failed")
		return errors.New("status code != CODE_OK")
	}

	return nil
}
