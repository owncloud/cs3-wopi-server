package cs3wopiserver

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/owncloud/cs3-wopi-server/pkg/internal/app"
)

func Start() error {
	ctx := context.Background()

	app, err := app.New()
	if err != nil {
		return err
	}

	if err := app.RegisterOcisService(ctx); err != nil {
		return err
	}

	if err := app.WopiDiscovery(ctx); err != nil {
		return err
	}

	if err := app.GetCS3apiClient(); err != nil {
		return err
	}

	if err := app.RegisterDemoApp(ctx); err != nil {
		return err
	}

	if err := app.GRPCServer(ctx); err != nil {
		return err
	}

	// Setting up the HTTP server should be the last step because
	// there is a healthcheck set inside.
	if err := app.HTTPServer(ctx); err != nil {
		return err
	}

	// Wait here until CTRL-C or other term signal is received.
	app.Logger.Info().Msg("WOPI is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	return nil
}
