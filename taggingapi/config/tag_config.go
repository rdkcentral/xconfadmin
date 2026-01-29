package taggingapi_config

import "github.com/go-akka/configuration"

type TaggingApiConfig struct {
	BatchLimit  int
	WorkerCount int
}

func NewTaggingApiConfig(conf *configuration.Config) *TaggingApiConfig {
	return &TaggingApiConfig{
		BatchLimit:  int(conf.GetInt32("webconfig.xconf.tag_members_batch_limit", 2000)),
		WorkerCount: int(conf.GetInt32("webconfig.xconf.tag_update_worker_count", 20)),
	}
}
