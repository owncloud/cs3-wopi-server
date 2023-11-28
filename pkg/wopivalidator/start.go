package wopivalidator

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	gatewayv1beta1 "github.com/cs3org/go-cs3apis/cs3/gateway/v1beta1"
	userv1beta1 "github.com/cs3org/go-cs3apis/cs3/identity/user/v1beta1"
	rpcv1beta1 "github.com/cs3org/go-cs3apis/cs3/rpc/v1beta1"
	providerv1beta1 "github.com/cs3org/go-cs3apis/cs3/storage/provider/v1beta1"
	ctxpkg "github.com/cs3org/reva/v2/pkg/ctx"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/owncloud/ocis/v2/ocis-pkg/log"
	"github.com/owncloud/ocis/v2/ocis-pkg/registry"
	"github.com/wkloucek/cs3-wopi-server/pkg/internal/app"
	"google.golang.org/grpc/metadata"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func Run(username, password, testGroup, testName string) error {
	ctx := context.Background()

	app, err := app.New()
	if err != nil {
		return err
	}

	// 0. initialize registry
	registry.GetRegistry()
	if err := app.GetCS3apiClient(); err != nil {
		return err
	}

	// 0. Login
	ctx, user, err := Login(ctx, username, password, app.GatewayAPIClient)
	if err != nil {
		return err
	}

	ref, err := buildRef(ctx, user, app.GatewayAPIClient)
	if err != nil {
		return err
	}

	// 1. Upload test.wopitest
	err = UploadWopiTestFile(ctx, ref, app.GatewayAPIClient, app.Config.CS3DataGatewayInsecure, app.Logger)
	if err != nil {
		return err
	}

	// 2. open in app
	wopiValidatorArguments, err := OpenInApp(ctx, ref, app.GatewayAPIClient)
	if err != nil {
		return err
	}

	// 3. start wopi-validator in docker
	err = RunWopiValidator(ctx, wopiValidatorArguments, testGroup, testName)
	if err != nil {
		return err
	}
	return nil
}

func RunWopiValidator(ctx context.Context, arguments *WopiValidatorArguments, testGroup string, testName string) error {
	fmt.Println("Starting wopi-validation ....")
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return err
	}
	defer cli.Close()

	cmd := []string{"-w", arguments.WopiSrc, "-t", arguments.AccessToken, "-l", arguments.AccessTokenTtl}
	if testGroup != "" {
		cmd = append(cmd, "-g", testGroup)
	} else {
		if testName != "" {
			cmd = append(cmd, "-n", testName)
		}
	}
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image:        "owncloudci/wopi-validator",
		Cmd:          cmd,
		AttachStderr: true,
		AttachStdout: true,
	}, &container.HostConfig{}, nil, nil, "")
	if err != nil {
		return err
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return err
	}

	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNextExit)
	select {
	case err := <-errCh:
		if err != nil {
			return err
		}
	case <-statusCh:
		out, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true})
		if err != nil {
			return err
		}
		_, err = io.Copy(os.Stdout, out)
		if err != nil {
			return err
		}
	}

	return nil
}

func OpenInApp(ctx context.Context, ref *providerv1beta1.Reference, app gatewayv1beta1.GatewayAPIClient) (*WopiValidatorArguments, error) {
	var req = &gatewayv1beta1.OpenInAppRequest{
		Ref:      ref,
		ViewMode: gatewayv1beta1.OpenInAppRequest_VIEW_MODE_READ_WRITE,
		App:      "WOPI app",
	}
	res, err := app.OpenInApp(ctx, req)
	if err != nil {
		return nil, err
	}

	WopiSrc := res.AppUrl.AppUrl
	u, err := url.Parse(WopiSrc)
	if err != nil {
		return nil, err
	}
	accessToken := res.AppUrl.FormParameters["access_token"]
	accessTokenTtl := res.AppUrl.FormParameters["access_token_ttl"]
	return &WopiValidatorArguments{
		WopiSrc:        u.Query().Get("WOPISrc"),
		AccessToken:    accessToken,
		AccessTokenTtl: accessTokenTtl,
	}, err
}

type WopiValidatorArguments struct {
	WopiSrc        string
	AccessToken    string
	AccessTokenTtl string
}

func Login(ctx context.Context, user string, pass string, gwc gatewayv1beta1.GatewayAPIClient) (context.Context, *userv1beta1.User, error) {
	req := &gatewayv1beta1.AuthenticateRequest{
		Type:         "basic",
		ClientId:     user,
		ClientSecret: pass,
	}

	// login
	res, err := gwc.Authenticate(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	if res.Status.Code != rpcv1beta1.Code_CODE_OK {
		return nil, nil, statusToError(res.Status)
	}

	ctx = ctxpkg.ContextSetToken(ctx, res.Token)
	ctx = metadata.AppendToOutgoingContext(ctx, ctxpkg.TokenHeader, res.Token)
	return ctx, res.User, nil
}

func statusToError(status *rpcv1beta1.Status) error {
	return fmt.Errorf("error: code=%+v msg=%q support_trace=%q", status.Code, status.Message, status.Trace)
}

func listStorageSpacesUserFilter(id string) *providerv1beta1.ListStorageSpacesRequest_Filter {
	return &providerv1beta1.ListStorageSpacesRequest_Filter{
		Type: providerv1beta1.ListStorageSpacesRequest_Filter_TYPE_USER,
		Term: &providerv1beta1.ListStorageSpacesRequest_Filter_User{
			User: &userv1beta1.UserId{
				OpaqueId: id,
			},
		},
	}
}

func listStorageSpacesTypeFilter(spaceType string) *providerv1beta1.ListStorageSpacesRequest_Filter {
	return &providerv1beta1.ListStorageSpacesRequest_Filter{
		Type: providerv1beta1.ListStorageSpacesRequest_Filter_TYPE_SPACE_TYPE,
		Term: &providerv1beta1.ListStorageSpacesRequest_Filter_SpaceType{
			SpaceType: spaceType,
		},
	}
}

func ListUserSpace(ctx context.Context, user *userv1beta1.User, gwc gatewayv1beta1.GatewayAPIClient) (*providerv1beta1.ListStorageSpacesResponse, error) {
	filters := []*providerv1beta1.ListStorageSpacesRequest_Filter{}
	filters = append(filters, listStorageSpacesUserFilter(user.GetId().OpaqueId))
	filters = append(filters, listStorageSpacesTypeFilter("personal"))

	res, err := gwc.ListStorageSpaces(ctx, &providerv1beta1.ListStorageSpacesRequest{
		Filters: filters,
	})
	return res, err
}

func buildRef(ctx context.Context, user *userv1beta1.User, gwc gatewayv1beta1.GatewayAPIClient) (*providerv1beta1.Reference, error) {
	space, err := ListUserSpace(ctx, user, gwc)
	if err != nil {
		return nil, err
	}
	ref := providerv1beta1.Reference{
		ResourceId: &providerv1beta1.ResourceId{
			StorageId: space.StorageSpaces[0].Id.OpaqueId,
			OpaqueId:  user.Id.OpaqueId,
			SpaceId:   user.Id.OpaqueId,
		},
		Path: "/test.wopitest",
	}

	return &ref, nil
}

func UploadWopiTestFile(ctx context.Context, ref *providerv1beta1.Reference, gwc gatewayv1beta1.GatewayAPIClient, insecure bool, logger log.Logger) error {

	content := io.NopCloser(strings.NewReader(""))
	req := &providerv1beta1.InitiateFileUploadRequest{
		Ref: ref,
	}

	resp, err := gwc.InitiateFileUpload(ctx, req)
	if err != nil {
		logger.Error().Err(
			err,
		).Str(
			"FileReference", ref.String(),
		).Msg("UploadHelper: InitiateFileUpload failed")
		return err
	}

	if resp.Status.Code != rpcv1beta1.Code_CODE_OK {
		return statusToError(resp.Status)
	}

	uploadEndpoint := ""
	uploadToken := ""

	for _, proto := range resp.Protocols {
		if proto.Protocol == "simple" || proto.Protocol == "spaces" {
			uploadEndpoint = proto.UploadEndpoint
			uploadToken = proto.Token
		}
	}

	if uploadEndpoint == "" {
		return errors.New("upload endpoint or token is missing")
	}

	httpClient := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: insecure,
			},
		},
	}

	httpReq, err := http.NewRequest(http.MethodPut, uploadEndpoint, content)
	if err != nil {
		return err
	}

	if uploadToken != "" {
		// public link uploads have the token in the upload endpoint
		httpReq.Header.Add("X-Reva-Transfer", uploadToken)
	}
	// TODO: the access token shouldn't be needed
	t := ctxpkg.ContextMustGetToken(ctx)
	httpReq.Header.Add("x-access-token", t)

	httpResp, err := httpClient.Do(httpReq)
	if err != nil {
		return err
	}

	if httpResp.StatusCode != http.StatusOK {
		return errors.New("status code was not 200")
	}

	return nil
}
