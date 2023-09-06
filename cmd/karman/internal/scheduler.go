package internal

import (
	"github.com/hibiken/asynq"
)

// mergedPeriodicTaskConfigProvider implements a config provider backed by multiple distinct providers.
type mergedPeriodicTaskConfigProvider struct {
	providers []asynq.PeriodicTaskConfigProvider
}

// MergePeriodicTaskConfigProviders creates a new periodic task config provider
// backed by the specified list of providers.
func MergePeriodicTaskConfigProviders(providers ...asynq.PeriodicTaskConfigProvider) asynq.PeriodicTaskConfigProvider {
	return &mergedPeriodicTaskConfigProvider{providers}
}

// GetConfigs returns the merged configs of all backing config providers.
func (p *mergedPeriodicTaskConfigProvider) GetConfigs() ([]*asynq.PeriodicTaskConfig, error) {
	var configs []*asynq.PeriodicTaskConfig
	for _, provider := range p.providers {
		providerConfigs, err := provider.GetConfigs()
		if err != nil {
			return nil, err
		}
		configs = append(configs, providerConfigs...)
	}
	return configs, nil
}
