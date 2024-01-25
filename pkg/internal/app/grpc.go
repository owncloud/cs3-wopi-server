package app

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net"
	"net/url"
	"path"
	"strconv"

	appproviderv1beta1 "github.com/cs3org/go-cs3apis/cs3/app/provider/v1beta1"
	gatewayv1beta1 "github.com/cs3org/go-cs3apis/cs3/gateway/v1beta1"
	userv1beta1 "github.com/cs3org/go-cs3apis/cs3/identity/user/v1beta1"
	rpcv1beta1 "github.com/cs3org/go-cs3apis/cs3/rpc/v1beta1"
	providerv1beta1 "github.com/cs3org/go-cs3apis/cs3/storage/provider/v1beta1"
	"github.com/golang-jwt/jwt"
	"google.golang.org/grpc"
)

func (app *demoApp) GRPCServer(ctx context.Context) error {
	opts := []grpc.ServerOption{}
	app.grpcServer = grpc.NewServer(opts...)

	// register the app provider interface / OpenInApp call
	appproviderv1beta1.RegisterProviderAPIServer(app.grpcServer, app)

	l, err := net.Listen("tcp", app.Config.GRPC.BindAddr)
	if err != nil {
		return err
	}
	go app.grpcServer.Serve(l)

	return nil
}

func (app *demoApp) OpenInApp(
	ctx context.Context,
	req *appproviderv1beta1.OpenInAppRequest,
) (*appproviderv1beta1.OpenInAppResponse, error) {

	// get the current user
	var user *userv1beta1.User = nil
	meReq := &gatewayv1beta1.WhoAmIRequest{
		Token: req.AccessToken,
	}
	meResp, err := app.gwc.WhoAmI(ctx, meReq)
	if err == nil {
		if meResp.Status.Code == rpcv1beta1.Code_CODE_OK {
			user = meResp.User
		}
	}

	// required for the response, it will be used also for logs
	providerFileRef := providerv1beta1.Reference{
		ResourceId: req.GetResourceInfo().GetId(),
		Path:       ".",
	}

	// build a urlsafe and stable file reference that can be used for proxy routing,
	// so that all sessions on one file end on the same office server

	c := sha256.New()
	c.Write([]byte(req.ResourceInfo.Id.StorageId + "$" + req.ResourceInfo.Id.SpaceId + "!" + req.ResourceInfo.Id.OpaqueId))
	fileRef := hex.EncodeToString(c.Sum(nil))

	// get the file extension to use the right wopi app url
	fileExt := path.Ext(req.GetResourceInfo().Path)

	var viewAppURL string
	var editAppURL string
	if viewAppURLs, ok := app.appURLs["view"]; ok {
		if url := viewAppURLs[fileExt]; ok {
			viewAppURL = url
		}
	}
	if editAppURLs, ok := app.appURLs["edit"]; ok {
		if url, ok := editAppURLs[fileExt]; ok {
			editAppURL = url
		}
	}

	if editAppURL == "" {
		// assuming that an view action is always available in the /hosting/discovery manifest
		// eg. Collabora does support viewing jpgs but no editing
		// eg. OnlyOffice does support viewing pdfs but no editing
		// there is no known case of supporting edit only without view
		editAppURL = viewAppURL
	}

	wopiSrcURL := url.URL{
		Scheme: app.Config.HTTP.Scheme,
		Host:   app.Config.HTTP.Addr,
		Path:   path.Join("wopi", "files", fileRef),
	}

	addWopiSrcQueryParam := func(baseURL string) (string, error) {
		u, err := url.Parse(baseURL)
		if err != nil {
			return "", err
		}

		q := u.Query()
		q.Add("WOPISrc", wopiSrcURL.String())
		qs := q.Encode()
		u.RawQuery = qs

		return u.String(), nil
	}

	viewAppURL, err = addWopiSrcQueryParam(viewAppURL)
	if err != nil {
		app.Logger.Error().
			Err(err).
			Str("FileReference", providerFileRef.String()).
			Str("ViewMode", req.ViewMode.String()).
			Str("Requester", user.GetId().String()).
			Msg("OpenInApp: error parsing viewAppUrl")
		return nil, err
	}
	editAppURL, err = addWopiSrcQueryParam(editAppURL)
	if err != nil {
		app.Logger.Error().
			Err(err).
			Str("FileReference", providerFileRef.String()).
			Str("ViewMode", req.ViewMode.String()).
			Str("Requester", user.GetId().String()).
			Msg("OpenInApp: error parsing editAppUrl")
		return nil, err
	}

	appURL := viewAppURL
	if req.ViewMode == appproviderv1beta1.OpenInAppRequest_VIEW_MODE_READ_WRITE {
		appURL = editAppURL
	}

	cryptedReqAccessToken, err := EncryptAES([]byte(app.Config.WopiSecret), req.AccessToken)
	if err != nil {
		app.Logger.Error().
			Err(err).
			Str("FileReference", providerFileRef.String()).
			Str("ViewMode", req.ViewMode.String()).
			Str("Requester", user.GetId().String()).
			Msg("OpenInApp: error encrypting access token")
		return &appproviderv1beta1.OpenInAppResponse{
			Status: &rpcv1beta1.Status{Code: rpcv1beta1.Code_CODE_INTERNAL},
		}, err
	}

	wopiContext := WopiContext{
		AccessToken:   cryptedReqAccessToken,
		FileReference: providerFileRef,
		User:          user,
		ViewMode:      req.ViewMode,
		EditAppUrl:    editAppURL,
		ViewAppUrl:    viewAppURL,
	}

	cs3Claims := &jwt.StandardClaims{}
	cs3JWTparser := jwt.Parser{}
	_, _, err = cs3JWTparser.ParseUnverified(req.AccessToken, cs3Claims)
	if err != nil {
		app.Logger.Error().
			Err(err).
			Str("FileReference", providerFileRef.String()).
			Str("ViewMode", req.ViewMode.String()).
			Str("Requester", user.GetId().String()).
			Msg("OpenInApp: error parsing JWT token")
		return nil, err
	}

	claims := &Claims{
		WopiContext: wopiContext,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: cs3Claims.ExpiresAt,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err := token.SignedString([]byte(app.Config.WopiSecret))

	if err != nil {
		app.Logger.Error().
			Err(err).
			Str("FileReference", providerFileRef.String()).
			Str("ViewMode", req.ViewMode.String()).
			Str("Requester", user.GetId().String()).
			Msg("OpenInApp: error signing access token")
		return &appproviderv1beta1.OpenInAppResponse{
			Status: &rpcv1beta1.Status{Code: rpcv1beta1.Code_CODE_INTERNAL},
		}, err
	}

	app.Logger.Debug().
		Str("FileReference", providerFileRef.String()).
		Str("ViewMode", req.ViewMode.String()).
		Str("Requester", user.GetId().String()).
		Msg("OpenInApp: success")

	return &appproviderv1beta1.OpenInAppResponse{
		Status: &rpcv1beta1.Status{Code: rpcv1beta1.Code_CODE_OK},
		AppUrl: &appproviderv1beta1.OpenInAppURL{
			AppUrl: appURL,
			Method: "POST",
			FormParameters: map[string]string{
				// these parameters will be passed to the web server by the app provider application
				"access_token": accessToken,
				// milliseconds since Jan 1, 1970 UTC as required in https://docs.microsoft.com/en-us/microsoft-365/cloud-storage-partner-program/rest/concepts#access_token_ttl
				"access_token_ttl": strconv.FormatInt(claims.ExpiresAt*1000, 10),
			},
		},
	}, nil
}
