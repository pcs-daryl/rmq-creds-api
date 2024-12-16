package main

import (
	"aaaas/rmq-permissions-api/pkg/api/config"
	"aaaas/rmq-permissions-api/pkg/api/handlers"

	commonCfg "github.com/pcs-aa-aas/commons/pkg/api/config"
	"github.com/pcs-aa-aas/commons/pkg/api/server"
)

func main() {
	configPath := "conf/api.conf"
	configSections := []string{"server"}
	serverCfgImpl := config.NewServerConfigImpl()
	middlewareConf := commonCfg.NewMiddlewareConfig(commonCfg.DisableKubeconfigMiddleware())
	routes := []server.APIHandlerGroup{handlers.HandlerGroup{}}
	// server.Run(configPath, configSections, routes, serverCfgImpl, "")
	server.RunWithMiddlewareConfigs(configPath, configSections, routes, serverCfgImpl, "conf/supervisorconf", middlewareConf)
}
