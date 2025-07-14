package argocdhelper

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"html"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/golang-jwt/jwt/v5"
	log "github.com/sirupsen/logrus"
	"github.com/skratchdot/open-golang/open"
	"golang.org/x/oauth2"

	argocdclient "github.com/argoproj/argo-cd/v2/pkg/apiclient"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient/application"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient/cluster"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient/project"
	sessionpkg "github.com/argoproj/argo-cd/v2/pkg/apiclient/session"
	settingspkg "github.com/argoproj/argo-cd/v2/pkg/apiclient/settings"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/argoproj/argo-cd/v2/util/io"
	"github.com/argoproj/argo-cd/v2/util/localconfig"
	oidcutil "github.com/argoproj/argo-cd/v2/util/oidc"
	"github.com/argoproj/argo-cd/v2/util/rand"
	argoerr "github.com/argoproj/pkg/errors"
	"github.com/santi1s/yak/cli"
	"github.com/santi1s/yak/internal/constant"
	"github.com/santi1s/yak/internal/helper"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

var (
	nonNamespacedAPIResources = []string{
		"PriorityClass",
		"StorageClass",
		"ClusterRole",
		"ClusterRoleBinding",
		"CustomResourceDefinition",
		"IngressClass",
		"CSIDriver",
		"Namespace",
		"ConstraintTemplate",
	}
)

type LoginParams struct {
	ArgocdServer   string
	ArgocdUsername string
	ArgocdPassword string
}

type ArgoCDClient struct {
	ProjectClient project.ProjectServiceClient
	ClusterClient cluster.ClusterServiceClient
	AppClient     application.ApplicationServiceClient
}

type AppResource struct {
	Kind      string
	Name      string
	Group     string
	Namespace string
}

type IgnoredOrphanResource struct {
	Kind  string
	Name  string
	Group string
}

type SyncWindowConfig struct {
	Kind     string
	Schedule string
	Duration string
	TimeZone string
}

var DefaultSyncWindowConfig = SyncWindowConfig{
	Kind:     "deny",
	Schedule: "* * * * *",
	Duration: "1m",
	TimeZone: "UTC",
}

func ArgoCDAddr(flag string) (string, error) {
	if flag != "" {
		return flag, nil
	}
	awsProfile := os.Getenv("AWS_PROFILE")
	if awsProfile == "" {
		return "", fmt.Errorf("could not determine ArgoCD address: AWS_PROFILE environment variable is not set")
	}
	return "argocd-" + os.Getenv("AWS_PROFILE") + ".doctolib.net", nil
}

func readLocalConfig() (*localconfig.LocalConfig, error) {
	cfgPath, err := localconfig.DefaultLocalConfigPath()
	if err != nil {
		return nil, fmt.Errorf("could not determine config path: %s", err)
	}

	localCfg, err := localconfig.ReadLocalConfig(cfgPath)
	if err != nil {
		return nil, fmt.Errorf("error while reading config file: %s", err)
	}
	if localCfg == nil {
		localCfg = &localconfig.LocalConfig{}
	}
	return localCfg, nil
}

func ArgocdLogin(p *LoginParams) (*ArgoCDClient, error) {
	var (
		err      error
		localCfg *localconfig.LocalConfig
		token    *jwt.Token
	)

	p.ArgocdServer, err = ArgoCDAddr(p.ArgocdServer)
	if err != nil {
		return nil, err
	}

	localCfg, err = readLocalConfig()
	if err != nil {
		return nil, err
	}

	// If no context, we need to login for the first time
	_, err = localCfg.ResolveContext(p.ArgocdServer)
	if err != nil {
		localCfg, err = login(p)
		if err != nil {
			return nil, err
		}
	}

	// we retrieve the user token
	token, err = getUserToken(localCfg, p.ArgocdServer)
	if err != nil {
		return nil, err
	}

	// Check if token is expired
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if exp, ok := claims["exp"].(float64); ok && time.Now().Unix() > int64(exp) {
			localCfg, err = login(p)
			if err != nil {
				return nil, err
			}
			token, err = getUserToken(localCfg, p.ArgocdServer)
			if err != nil {
				return nil, err
			}
		}
	}

	return getArgocdClient(p.ArgocdServer, token.Raw)
}

// Try to log in with one of the supported methods ; server address is used for context AND user
// Return config read after connection
// login is highly inspired by ArgoCD's command.NewLoginCommand.Run function
func login(p *LoginParams) (*localconfig.LocalConfig, error) {
	var ctxName string

	clientOpts := argocdclient.ClientOptions{
		ConfigPath: "",
		ServerAddr: p.ArgocdServer,
		Insecure:   true,
		GRPCWeb:    true,
	}

	if ctxName == "" {
		ctxName = p.ArgocdServer
	}

	// Perform the login
	var tokenString string
	var refreshToken string
	acdClient, err := argocdclient.NewClient(&clientOpts)
	if err != nil {
		return nil, err
	}
	setConn, setIf, err := acdClient.NewSettingsClient()
	if err != nil {
		return nil, err
	}
	defer io.Close(setConn)
	if p.ArgocdUsername != "" && p.ArgocdPassword != "" {
		sessConn, sessionIf, err := acdClient.NewSessionClient()
		if err != nil {
			return nil, err
		}
		defer io.Close(sessConn)
		sessionRequest := sessionpkg.SessionCreateRequest{
			Username: p.ArgocdUsername,
			Password: p.ArgocdPassword,
		}
		createdSession, err := sessionIf.Create(context.TODO(), &sessionRequest)
		if err != nil {
			return nil, err
		}
		tokenString = createdSession.Token
	} else {
		httpClient, err := acdClient.HTTPClient()
		if err != nil {
			return nil, err
		}
		ctx := oidc.ClientContext(context.TODO(), httpClient)
		acdSet, err := setIf.Get(ctx, &settingspkg.SettingsQuery{})
		if err != nil {
			return nil, err
		}
		oauth2conf, provider, err := acdClient.OIDCConfig(ctx, acdSet)
		if err != nil {
			return nil, err
		}
		tokenString, refreshToken = oauth2Login(ctx, 8085, acdSet.GetOIDCConfig(), oauth2conf, provider)
	}
	parser := jwt.NewParser(jwt.WithoutClaimsValidation())
	claims := jwt.MapClaims{}
	_, _, err = parser.ParseUnverified(tokenString, &claims)
	if err != nil {
		return nil, err
	}

	// login successful. Persist the config
	localCfg, err := readLocalConfig()
	if err != nil {
		return nil, err
	}

	if localCfg == nil {
		localCfg = &localconfig.LocalConfig{}
	}

	localCfg.UpsertServer(localconfig.Server{
		Server: p.ArgocdServer,
	})
	localCfg.UpsertUser(localconfig.User{
		Name:         ctxName,
		AuthToken:    tokenString,
		RefreshToken: refreshToken,
	})
	if ctxName == "" {
		ctxName = p.ArgocdServer
	}
	localCfg.CurrentContext = ctxName
	localCfg.UpsertContext(localconfig.ContextRef{
		Name:   ctxName,
		User:   ctxName,
		Server: p.ArgocdServer,
	})

	cfgPath, err := localconfig.DefaultLocalConfigPath()
	if err != nil {
		return nil, fmt.Errorf("could not determine config path: %s", err)
	}

	err = localconfig.WriteLocalConfig(*localCfg, cfgPath)
	if err != nil {
		return nil, err
	}

	// make sure we can read config from local file
	localCfg, err = readLocalConfig()
	if err != nil {
		return nil, err
	}

	return localCfg, nil
}

// ArgoCD oauth2Login function wasn't exposed so I had to copy/paste here...
// oauth2Login opens a browser, runs a temporary HTTP server to delegate OAuth2 login flow and
// returns the JWT token and a refresh token (if supported)
func oauth2Login(
	ctx context.Context,
	port int,
	oidcSettings *settingspkg.OIDCConfig,
	oauth2conf *oauth2.Config,
	provider *oidc.Provider,
) (string, string) {
	oauth2conf.RedirectURL = fmt.Sprintf("http://localhost:%d/auth/callback", port)
	oidcConf, err := oidcutil.ParseConfig(provider)
	argoerr.CheckError(err)
	log.Debug("OIDC Configuration:")
	log.Debugf("  supported_scopes: %v", oidcConf.ScopesSupported)
	log.Debugf("  response_types_supported: %v", oidcConf.ResponseTypesSupported)

	// handledRequests ensures we do not handle more requests than necessary
	handledRequests := 0
	// completionChan is to signal flow completed. Non-empty string indicates error
	completionChan := make(chan string)
	// stateNonce is an OAuth2 state nonce
	// According to the spec (https://www.rfc-editor.org/rfc/rfc6749#section-10.10), this must be guessable with
	// probability <= 2^(-128). The following call generates one of 52^24 random strings, ~= 2^136 possibilities.
	stateNonce, err := rand.String(24)
	argoerr.CheckError(err)
	var tokenString string
	var refreshToken string

	handleErr := func(w http.ResponseWriter, errMsg string) {
		http.Error(w, html.EscapeString(errMsg), http.StatusBadRequest)
		completionChan <- errMsg
	}

	// PKCE implementation of https://tools.ietf.org/html/rfc7636
	codeVerifier, err := rand.StringFromCharset(
		43,
		"ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-._~",
	)
	argoerr.CheckError(err)
	codeChallengeHash := sha256.Sum256([]byte(codeVerifier))
	codeChallenge := base64.RawURLEncoding.EncodeToString(codeChallengeHash[:])

	// Authorization redirect callback from OAuth2 auth flow.
	// Handles both implicit and authorization code flow
	callbackHandler := func(w http.ResponseWriter, r *http.Request) {
		log.Debugf("Callback: %s", r.URL)

		if formErr := r.FormValue("error"); formErr != "" {
			handleErr(w, fmt.Sprintf("%s: %s", formErr, r.FormValue("error_description")))
			return
		}

		handledRequests++
		if handledRequests > 2 {
			// Since implicit flow will redirect back to ourselves, this counter ensures we do not
			// fallinto a redirect loop (e.g. user visits the page by hand)
			handleErr(w, "Unable to complete login flow: too many redirects")
			return
		}

		if len(r.Form) == 0 {
			// If we get here, no form data was set. We presume to be performing an implicit login
			// flow where the id_token is contained in a URL fragment, making it inaccessible to be
			// read from the request. This javascript will redirect the browser to send the
			// fragments as query parameters so our callback handler can read and return token.
			fmt.Fprintf(w, `<script>window.location.search = window.location.hash.substring(1)</script>`)
			return
		}

		if state := r.FormValue("state"); state != stateNonce {
			handleErr(w, "Unknown state nonce")
			return
		}

		tokenString = r.FormValue("id_token")
		if tokenString == "" {
			code := r.FormValue("code")
			if code == "" {
				handleErr(w, fmt.Sprintf("no code in request: %q", r.Form))
				return
			}
			opts := []oauth2.AuthCodeOption{oauth2.SetAuthURLParam("code_verifier", codeVerifier)}
			tok, err := oauth2conf.Exchange(ctx, code, opts...)
			if err != nil {
				handleErr(w, err.Error())
				return
			}
			var ok bool
			tokenString, ok = tok.Extra("id_token").(string)
			if !ok {
				handleErr(w, "no id_token in token response")
				return
			}
			refreshToken, _ = tok.Extra("refresh_token").(string)
		}
		successPage := `
		<div style="height:100px; width:100%!; display:flex; flex-direction: column; justify-content: center; align-items:center; background-color:#2ecc71; color:white; font-size:22"><div>Authentication successful!</div></div>
		<p style="margin-top:20px; font-size:18; text-align:center">Authentication was successful, you can now return to CLI. This page will close automatically</p>
		<script>window.onload=function(){setTimeout(this.close, 4000)}</script>
		`
		fmt.Fprint(w, successPage)
		completionChan <- ""
	}
	srv := &http.Server{Addr: "localhost:" + strconv.Itoa(port)} // #nosec G112
	http.HandleFunc("/auth/callback", callbackHandler)

	// Redirect user to login & consent page to ask for permission for the scopes specified above.
	fmt.Printf("Opening browser for authentication\n")

	var url string
	grantType := oidcutil.InferGrantType(oidcConf)
	opts := []oauth2.AuthCodeOption{oauth2.AccessTypeOffline}
	if claimsRequested := oidcSettings.GetIDTokenClaims(); claimsRequested != nil {
		opts = oidcutil.AppendClaimsAuthenticationRequestParameter(opts, claimsRequested)
	}

	switch grantType {
	case oidcutil.GrantTypeAuthorizationCode:
		opts = append(opts, oauth2.SetAuthURLParam("code_challenge", codeChallenge))
		opts = append(opts, oauth2.SetAuthURLParam("code_challenge_method", "S256"))
		url = oauth2conf.AuthCodeURL(stateNonce, opts...)
	case oidcutil.GrantTypeImplicit:
		url, err = oidcutil.ImplicitFlowURL(oauth2conf, stateNonce, opts...)
		argoerr.CheckError(err)
	default:
		log.Fatalf("Unsupported grant type: %v", grantType)
	}
	fmt.Printf("Performing %s flow login: %s\n", grantType, url)
	time.Sleep(1 * time.Second)
	err = open.Start(url)
	argoerr.CheckError(err)
	go func() {
		log.Debugf("Listen: %s", srv.Addr)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("Temporary HTTP server failed: %s", err)
		}
	}()
	errMsg := <-completionChan
	if errMsg != "" {
		log.Fatal(errMsg)
	}
	fmt.Printf("Authentication successful\n")
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
	log.Debugf("Token: %s", tokenString)
	log.Debugf("Refresh Token: %s", refreshToken)
	return tokenString, refreshToken
}

func getUserToken(cfg *localconfig.LocalConfig, contextName string) (*jwt.Token, error) {
	user, err := cfg.GetUser(contextName)
	if err != nil {
		return nil, fmt.Errorf("unable to get user '%s' in config: %s", contextName, err)
	}

	parser := jwt.NewParser(jwt.WithoutClaimsValidation())
	token, _, err := parser.ParseUnverified(user.AuthToken, jwt.MapClaims{})
	if err != nil {
		return nil, fmt.Errorf("error while parsing JWT: %s", err)
	}
	return token, nil
}

func getArgocdClient(serverAddr string, authToken string) (*ArgoCDClient, error) {
	clientOpts := argocdclient.ClientOptions{
		ServerAddr: serverAddr,
		AuthToken:  authToken,
		Insecure:   true,
		GRPCWeb:    true,
	}

	apiClient, err := argocdclient.NewClient(&clientOpts)
	if err != nil {
		return nil, err
	}

	_, projectClient, err := apiClient.NewProjectClient()
	if err != nil {
		return nil, err
	}

	_, clusterClient, err := apiClient.NewClusterClient()
	if err != nil {
		return nil, err
	}

	_, appClient, err := apiClient.NewApplicationClient()
	if err != nil {
		return nil, err
	}

	return &ArgoCDClient{
		ProjectClient: projectClient,
		ClusterClient: clusterClient,
		AppClient:     appClient,
	}, nil
}

// what a - b means when a and b are arrays of appResource
func SubstractAppResources(a []AppResource, b []AppResource) []AppResource {
	var appResourceDiff []AppResource
	var foundEquality bool
	for _, appResourceInA := range a {
		foundEquality = false
		for _, appResourceInB := range b {
			if reflect.DeepEqual(appResourceInA, appResourceInB) {
				foundEquality = true
				break
			}
		}
		if !foundEquality {
			appResourceDiff = append(appResourceDiff, appResourceInA)
		}
	}
	return appResourceDiff
}

func GetArgoCDProject(projectClient project.ProjectServiceClient, projectName string) (*v1alpha1.AppProject, error) {
	myProject, err := projectClient.Get(context.Background(), &project.ProjectQuery{Name: projectName})
	if err != nil {
		return nil, err
	}
	return myProject, nil
}

func UpdateArgoCDProject(projectClient project.ProjectServiceClient, argoProject *v1alpha1.AppProject) error {
	_, err := projectClient.Update(context.Background(), &project.ProjectUpdateRequest{Project: argoProject})
	return err
}

func GetApplication(appClient application.ApplicationServiceClient, appName string, projectName string) (*v1alpha1.Application, error) {
	myApps, err := appClient.List(context.Background(), &application.ApplicationQuery{Project: []string{projectName}, Name: &appName})

	if err != nil {
		return nil, err
	}
	if (len(myApps.Items)) == 0 {
		return nil, fmt.Errorf("application %s not found in project %s", appName, projectName)
	}
	myApp := &myApps.Items[0]
	return myApp, nil
}

// Get all applications from a project
func GetAllApplications(appClient application.ApplicationServiceClient, projectName string) (*v1alpha1.ApplicationList, error) {
	myApps, err := appClient.List(context.Background(), &application.ApplicationQuery{Project: []string{projectName}})

	return myApps, err
}

func GetSyncWindow(project *v1alpha1.AppProject) int {
	for i, window := range project.Spec.SyncWindows {
		if window.Kind == DefaultSyncWindowConfig.Kind &&
			window.Schedule == DefaultSyncWindowConfig.Schedule &&
			window.Duration == DefaultSyncWindowConfig.Duration &&
			window.TimeZone == DefaultSyncWindowConfig.TimeZone {
			return i
		}
	}
	return -1
}

func UpdateWindow(project *v1alpha1.AppProject, applicationList []string) error {
	// Find and delete the window, if it exists
	windowIndex := GetSyncWindow(project)
	if windowIndex != -1 {
		err := project.Spec.DeleteWindow(windowIndex)
		if err != nil {
			return err
		}
	}

	return project.Spec.AddWindow(DefaultSyncWindowConfig.Kind, DefaultSyncWindowConfig.Schedule, DefaultSyncWindowConfig.Duration, applicationList, []string{""}, []string{""}, true, DefaultSyncWindowConfig.TimeZone)
}

func GetOrphanedNonNamespacedResources(appClient application.ApplicationServiceClient, projectName string) ([]AppResource, error) {
	var orphanedNonNamespacedResources []AppResource // serves for output

	// Connect to the kubernetes API
	config, _, err := helper.InitKubeClusterConfig()
	if err != nil {
		return []AppResource{}, err
	}
	log.Debug("Successfully connected to the Kubernetes API")

	// Get all non-namespaced resources running in the cluster
	allRunningNonNamespacedResources, err := getNonNamespacedResourcesRunningInKube(config)

	if err != nil {
		return []AppResource{}, err
	}
	log.Debug("Successfully grabbed all non-namespaced resources running inside the current Kubernetes cluster")

	// We will remove elements for this initial list until we only keep the non-namespaced orphan resources
	orphanedNonNamespacedResources = allRunningNonNamespacedResources

	// Get all non-namespaced resources managed by ArgoCD inside each app
	nonNamespacedResourcesManagedByArgoCD, err := getNonNamespacedResourcesArgoCD(appClient, projectName)

	if err != nil {
		return []AppResource{}, err
	}
	log.Debug("Successfully grabbed all non-namespaced resources attached to ArgoCD apps")

	// Remove the non-namespaced resources that are managed by argocd inside its apps.
	for _, AppResourcesArgoCD := range nonNamespacedResourcesManagedByArgoCD {
		orphanedNonNamespacedResources = SubstractAppResources(orphanedNonNamespacedResources, AppResourcesArgoCD)
	}

	// Retrieve the list of ignored orphan resources from the argocd project yml file
	ignored, err := getIgnoredOrphanResources(config, "argocd", projectName)
	if err != nil {
		return []AppResource{}, err
	}
	log.Debug("Successfully grabbed all orphan resources ignored by ArgoCD")

	// Based on this list of ignored let's clean the final result of orphaned resources
	for i := 0; i < len(orphanedNonNamespacedResources); i++ {
		if IsResourceOrphanIgnored(orphanedNonNamespacedResources[i], ignored) {
			orphanedNonNamespacedResources = append(orphanedNonNamespacedResources[:i], orphanedNonNamespacedResources[i+1:]...)
			i--
		}
	}

	return orphanedNonNamespacedResources, err
}

func IsResourceOrphanIgnored(resource AppResource, ignoredOrphanResource []IgnoredOrphanResource) bool {
	match := false
	for _, ignored := range ignoredOrphanResource {
		if ignored.Group != resource.Group {
			continue
		}
		if ignored.Kind == resource.Kind {
			if ignored.Name == resource.Name {
				return true
			}
			if ignored.Name == "*" {
				return true
			}
		}
	}
	return match
}

func getOrphanedResourcesInApp(appClient application.ApplicationServiceClient, appName string, appNamespace string) ([]AppResource, error) {
	orphanedResourcesInApp := []AppResource{}
	response, err := appClient.ResourceTree(context.TODO(), &application.ResourcesQuery{
		ApplicationName: &appName, AppNamespace: &appNamespace,
	})
	if err != nil {
		return orphanedResourcesInApp, err
	}

	for _, orphan := range response.OrphanedNodes {
		orphanedResourcesInApp = append(orphanedResourcesInApp, AppResource{Kind: orphan.Kind, Group: orphan.Group, Name: orphan.Name, Namespace: orphan.Namespace})
	}
	return orphanedResourcesInApp, err
}

func OrphanedResourcesArgoCD(appClient application.ApplicationServiceClient, projectName string) (map[string][]AppResource, error) {
	orphanedResources := make(map[string][]AppResource)
	applications, err := appClient.List(context.TODO(), &application.ApplicationQuery{Project: []string{projectName}})
	if err != nil {
		return orphanedResources, err
	}
	for _, app := range applications.Items {
		// Get the resources of each argocd application
		orphanedResourcesInApp, err := getOrphanedResourcesInApp(appClient, app.Name, app.Namespace)
		if err != nil {
			return orphanedResources, err
		} else {
			orphanedResources[app.Name] = orphanedResourcesInApp
		}
	}
	return orphanedResources, err
}

func getNonNamespacedResourcesInApp(appClient application.ApplicationServiceClient, appName string, appNamespace string, projectName string) ([]AppResource, error) {
	resourcesInApp := []AppResource{}
	response, err := appClient.ResourceTree(context.TODO(), &application.ResourcesQuery{Project: &projectName,
		ApplicationName: &appName, AppNamespace: &appNamespace,
	})
	if err != nil {
		return resourcesInApp, err
	}
	for _, node := range response.Nodes {
		if helper.ListContainsString(nonNamespacedAPIResources, node.Kind) {
			resourcesInApp = append(resourcesInApp, AppResource{Kind: node.Kind, Group: node.Group, Name: node.Name, Namespace: ""})
		}
	}
	return resourcesInApp, err
}

func getNonNamespacedResourcesArgoCD(appClient application.ApplicationServiceClient, projectName string) (map[string][]AppResource, error) {
	nonNamespacedResources := make(map[string][]AppResource)
	applications, err := appClient.List(context.TODO(), &application.ApplicationQuery{Project: []string{projectName}})
	for _, app := range applications.Items {
		// Get the resources of each argocd application
		orphanedResourcesInApp, err := getNonNamespacedResourcesInApp(appClient, app.Name, app.Namespace, projectName)

		if err != nil {
			return nonNamespacedResources, err
		} else {
			nonNamespacedResources[app.Name] = orphanedResourcesInApp
		}
	}
	return nonNamespacedResources, err
}

func getNonNamespacedResourcesRunningInKube(config *rest.Config) ([]AppResource, error) {
	var resourceKind string
	//var resourceVersion string
	var resourceGroup string
	var nonNamespacedResources []AppResource
	var nonNamespacedResource AppResource
	clientDynamic, err := dynamic.NewForConfig(config)

	if err != nil {
		_, _ = cli.PrintlnErr("[ERR] Unable to create a dynamic client to Kubernetes API", err)
		return nonNamespacedResources, err
	}
	for _, nonNamespacedKindResources := range nonNamespacedAPIResources {
		resourceKind = nonNamespacedKindResources
		resourceGroup = constant.KubeResourceGroupFromKind[nonNamespacedKindResources]
		list, err := fetchResourcesOfKind(clientDynamic, nonNamespacedKindResources, constant.KubeResourceGroupFromKind[nonNamespacedKindResources], "v1")

		if err != nil {
			_, _ = cli.PrintfErr("[ERR] Unable to fetch resources kind/group %s/%s, %v\n", resourceKind, resourceGroup, err)
			continue
		}
		if list != nil {
			for _, item := range list.Items {
				nonNamespacedResource = AppResource{Kind: resourceKind, Name: item.GetName(), Namespace: "", Group: resourceGroup}
				nonNamespacedResources = append(nonNamespacedResources, nonNamespacedResource)
			}
		}
	}
	return nonNamespacedResources, nil
}

func fetchResourcesOfKind(client dynamic.Interface, resourceKind, resourceGroup, resourceVersion string) (*unstructured.UnstructuredList, error) {
	resourceInterface := client.Resource(
		schema.GroupVersionResource{
			Group:    resourceGroup,
			Version:  resourceVersion,
			Resource: strings.ToLower(resourceKind) + "s", // plural form
		},
	)
	list, err := resourceInterface.List(context.TODO(), v1.ListOptions{})
	if (err != nil) || (list == nil) {
		resourceInterface = client.Resource(
			schema.GroupVersionResource{
				Group:    resourceGroup,
				Version:  resourceVersion,
				Resource: strings.ToLower(resourceKind) + "es", // plural form
			},
		)
		list, err = resourceInterface.List(context.TODO(), v1.ListOptions{})
		return list, err
	}
	return list, nil
}

// get all orphan ignored resources in an argocd project template
func getIgnoredOrphanResources(config *rest.Config, namespace, projectName string) ([]IgnoredOrphanResource, error) {
	// Get from Kubernetes the body of ArgoCD project resource
	var ignoredMap []IgnoredOrphanResource
	var kind, name, group string
	clientDynamic, _ := dynamic.NewForConfig(config)
	projectResource := schema.GroupVersionResource{
		Group:    "argoproj.io",
		Version:  "v1alpha1",
		Resource: "appprojects",
	}

	projectInterface := clientDynamic.Resource(projectResource).Namespace(namespace)
	project, err := projectInterface.Get(context.TODO(), projectName, v1.GetOptions{})
	if err != nil {
		return ignoredMap, err
	}
	orphanedResourcesIgnore, _, err := unstructured.NestedSlice(project.Object, "spec", "orphanedResources", "ignore")
	if err != nil {
		return ignoredMap, err
	}
	for _, orphanedResourceIgnore := range orphanedResourcesIgnore {
		kind = "*"
		name = "*"
		group = ""
		for k, v := range orphanedResourceIgnore.(map[string]interface{}) {
			switch k {
			case "kind":
				kind = v.(string)
			case "name":
				name = v.(string)
			case "group":
				group = v.(string)
			default:
				cli.Println("Unexpected field")
			}
		}
		ignoredMap = append(ignoredMap, IgnoredOrphanResource{Kind: kind, Name: name, Group: group})
	}
	return ignoredMap, nil
}

// PaginationDisplayer is an interface for displaying paginated content
type PaginationDisplayer interface {
	DisplayPage(start, end int)
	GetTotal() int
	DisplaySummary()
}

// PaginateOutput provides paginated output for any content that implements PaginationDisplayer
func PaginateOutput(displayer PaginationDisplayer, pageSize int) {
	totalItems := displayer.GetTotal()
	if pageSize <= 0 {
		pageSize = 20 // default page size
	}

	totalPages := (totalItems + pageSize - 1) / pageSize
	currentPage := 1

	reader := bufio.NewReader(os.Stdin)

	for {
		start := (currentPage - 1) * pageSize
		end := currentPage * pageSize

		cli.Printf("\n--- Page %d of %d (showing %d-%d of %d items) ---\n",
			currentPage, totalPages, start+1, minInt(end, totalItems), totalItems)

		displayer.DisplayPage(start, end)

		if currentPage >= totalPages {
			displayer.DisplaySummary()
			cli.Printf("\nEnd of results. Press Enter to continue...")
			_, _, _ = reader.ReadLine()
			break
		}

		cli.Printf("\nPress Enter for next page, 'q' to quit, or 'p' for previous page: ")
		input, _, _ := reader.ReadLine()
		command := strings.TrimSpace(string(input))

		switch strings.ToLower(command) {
		case "q", "quit", "exit":
			displayer.DisplaySummary()
			return
		case "p", "prev", "previous":
			if currentPage > 1 {
				currentPage--
			}
		case "":
			currentPage++
		default:
			cli.Printf("Invalid command. Use Enter (next), 'p' (previous), or 'q' (quit)\n")
		}
	}
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
