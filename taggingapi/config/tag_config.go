package taggingapi_config

import "github.com/go-akka/configuration"

type TaggingApiConfig struct {
	BatchLimit  int
	WorkerCount int
	// TagTypeColumnEnabled gates every read and write of TagBucketMetadata.tag_type.
	//
	// The metadata insert shares an UnloggedBatch with the member inserts, so a
	// statement naming a column the cluster does not have fails the whole batch —
	// which would break *all* tag adds, not just account ones. Keeping this off
	// until the ALTER has been applied decouples the binary rollout from the DDL
	// and makes rollback a config flip.
	TagTypeColumnEnabled bool
}

func NewTaggingApiConfig(conf *configuration.Config) *TaggingApiConfig {
	return &TaggingApiConfig{
		BatchLimit:           int(conf.GetInt32("webconfig.xconf.tag_members_batch_limit", 2000)),
		WorkerCount:          int(conf.GetInt32("webconfig.xconf.tag_update_worker_count", 20)),
		TagTypeColumnEnabled: conf.GetBoolean("xconfwebconfig.xconf.tag_type_column_enabled", false),
	}
}
