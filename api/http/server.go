package http

import (
	"github.com/portainer/portainer"

	"net/http"
)

// Server implements the portainer.Server interface
type Server struct {
	BindAddress            string
	AssetsPath             string
	AuthDisabled           bool
	EndpointManagement     bool
	UserService            portainer.UserService
	EndpointService        portainer.EndpointService
	ResourceControlService portainer.ResourceControlService
	CryptoService          portainer.CryptoService
	JWTService             portainer.JWTService
	FileService            portainer.FileService
	Settings               *portainer.Settings
	TemplatesURL           string
	Handler                *Handler
	SSL                    bool
	SSLCert                string
	SSLKey                 string
}

// Start starts the HTTP server
func (server *Server) Start() error {
	middleWareService := &middleWareService{
		jwtService:   server.JWTService,
		authDisabled: server.AuthDisabled,
	}
	proxyService := NewProxyService(server.ResourceControlService)

	var authHandler = NewAuthHandler(middleWareService)
	authHandler.UserService = server.UserService
	authHandler.CryptoService = server.CryptoService
	authHandler.JWTService = server.JWTService
	authHandler.authDisabled = server.AuthDisabled
	var userHandler = NewUserHandler(middleWareService)
	userHandler.UserService = server.UserService
	userHandler.CryptoService = server.CryptoService
	userHandler.ResourceControlService = server.ResourceControlService
	var settingsHandler = NewSettingsHandler(middleWareService)
	settingsHandler.settings = server.Settings
	var templatesHandler = NewTemplatesHandler(middleWareService)
	templatesHandler.containerTemplatesURL = server.TemplatesURL
	var dockerHandler = NewDockerHandler(middleWareService, server.ResourceControlService)
	dockerHandler.EndpointService = server.EndpointService
	dockerHandler.ProxyService = proxyService
	var websocketHandler = NewWebSocketHandler()
	websocketHandler.EndpointService = server.EndpointService
	var endpointHandler = NewEndpointHandler(middleWareService)
	endpointHandler.authorizeEndpointManagement = server.EndpointManagement
	endpointHandler.EndpointService = server.EndpointService
	endpointHandler.FileService = server.FileService
	endpointHandler.ProxyService = proxyService
	var uploadHandler = NewUploadHandler(middleWareService)
	uploadHandler.FileService = server.FileService
	var fileHandler = newFileHandler(server.AssetsPath)

	server.Handler = &Handler{
		AuthHandler:      authHandler,
		UserHandler:      userHandler,
		EndpointHandler:  endpointHandler,
		SettingsHandler:  settingsHandler,
		TemplatesHandler: templatesHandler,
		DockerHandler:    dockerHandler,
		WebSocketHandler: websocketHandler,
		FileHandler:      fileHandler,
		UploadHandler:    uploadHandler,
	}

	if server.SSL {
		return http.ListenAndServeTLS(server.BindAddress, server.SSLCert, server.SSLKey, server.Handler)
	}
	return http.ListenAndServe(server.BindAddress, server.Handler)
}
