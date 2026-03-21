package application

import "errors"

var (
	ErrInvalidPluginInput = errors.New("invalid_plugin_input")
	ErrPluginAlreadyExists = errors.New("plugin_already_exists")
	ErrPluginNotFound     = errors.New("plugin_not_found")
)

