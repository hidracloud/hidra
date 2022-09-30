package plugins

var (
	plugins            = make(map[string]PluginInterface)
	pluginDescriptions = make(map[string]string)
)

// AddPlugin adds a plugin.
func AddPlugin(name, description string, plugin PluginInterface) {
	plugins[name] = plugin
	pluginDescriptions[name] = description
}

// GetPlugin returns a plugin.
func GetPlugin(name string) PluginInterface {
	return plugins[name]
}

// GetPlugins returns all plugins.
func GetPlugins() map[string]PluginInterface {
	return plugins
}

// GetPluginDescription returns a plugin description.
func GetPluginDescription(name string) string {
	return pluginDescriptions[name]
}
