package main

import "go.uber.org/zap"

func createLogger() (*zap.Logger, error) {
	cfg := zap.NewDevelopmentConfig()
	cfg.OutputPaths = []string{
		"stdout", "trolljitrs.log",
	}
	return cfg.Build()
}
