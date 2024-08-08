package manage

type PluginNameList []string

type ListerInterface interface {
	GetResourceNamespace() string
	Discover(chan PluginNameList)
	NewPlugin(string) PluginInterface
}
