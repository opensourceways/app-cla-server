package watch

import "time"

type Config struct {
	// unit second
	Interval           int `json:"interval"`
	MsgChannelSize     int `json:"msg_channel_size"`
	PythonRetryTimes   int `json:"python_retry_times"`
	GenAllDiffInterval int `json:"gen_all_diff_interval"`
}

func (cfg *Config) SetDefault() {
	if cfg.Interval <= 0 {
		cfg.Interval = 3600
	}

	if cfg.MsgChannelSize <= 0 {
		cfg.MsgChannelSize = 100
	}

	if cfg.PythonRetryTimes <= 0 {
		cfg.PythonRetryTimes = 3
	}

	if cfg.GenAllDiffInterval <= 0 {
		cfg.GenAllDiffInterval = 600
	}
}

func (cfg *Config) intervalDuration() time.Duration {
	return time.Second * time.Duration(cfg.Interval)
}

func (cfg *Config) genAllDiffInterval() time.Duration {
	return time.Second * time.Duration(cfg.GenAllDiffInterval)
}
