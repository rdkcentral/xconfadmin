package taggingapi_config

import "github.com/go-akka/configuration"

// TagSyncConfig holds the knobs for the tag sync job (detecting and
// repairing members missing from XDAS). All keys live under the
// xconfwebconfig.xconf root, matching the group service connector
// configuration.
type TagSyncConfig struct {
	RateLimit               int
	WorkerCount             int
	ChunkSize               int
	CheckpointIntervalSecs  int
	BreakerWindow           int
	BreakerMinSample        int
	BreakerErrorRatePercent int
	BreakerMaxConsecErrors  int
}

func NewTagSyncConfig(conf *configuration.Config) *TagSyncConfig {
	return &TagSyncConfig{
		RateLimit:               int(conf.GetInt32("xconfwebconfig.xconf.tag_sync_rate_limit", 100)),
		WorkerCount:             int(conf.GetInt32("xconfwebconfig.xconf.tag_sync_worker_count", 20)),
		ChunkSize:               int(conf.GetInt32("xconfwebconfig.xconf.tag_sync_chunk_size", 5000)),
		CheckpointIntervalSecs:  int(conf.GetInt32("xconfwebconfig.xconf.tag_sync_checkpoint_interval_secs", 30)),
		BreakerWindow:           int(conf.GetInt32("xconfwebconfig.xconf.tag_sync_breaker_window", 200)),
		BreakerMinSample:        int(conf.GetInt32("xconfwebconfig.xconf.tag_sync_breaker_min_sample", 500)),
		BreakerErrorRatePercent: int(conf.GetInt32("xconfwebconfig.xconf.tag_sync_breaker_error_rate_percent", 25)),
		BreakerMaxConsecErrors:  int(conf.GetInt32("xconfwebconfig.xconf.tag_sync_breaker_max_consec_errors", 10)),
	}
}
