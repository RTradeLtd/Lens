package lens

// ConfigOpts are options used to configure the lens service
type ConfigOpts struct {
	UseChainAlgorithm bool
	DataStorePath     string
}

// MetaData is a piece of meta data from a given object after being lensed
type MetaData struct {
	Summary []string `json:"summary"`
}
