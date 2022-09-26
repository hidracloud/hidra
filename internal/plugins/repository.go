package plugins

var (
	plugins = make(map[string]PluginInterface)
)

// AddPlugin adds a plugin.
func AddPlugin(name string, plugin PluginInterface) {
	plugins[name] = plugin
}

// GetPlugin returns a plugin.
func GetPlugin(name string) PluginInterface {
	return plugins[name]
}
