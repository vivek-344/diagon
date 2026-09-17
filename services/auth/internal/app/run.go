package app

func (a *App) Run() {
	a.logger.Info(
		"starting auth service",
		"grpc_port", a.cfg.App.GRPCPort,
	)
}
