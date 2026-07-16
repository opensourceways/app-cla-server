package watch

import "time"

type Config struct {
	// unit second
	Interval        int               `json:"interval"`
	CLAUpdateConfig CLAUpdateConfig   `json:"cla_update_config"`
	SendEmailConfig NotifyAdminConfig `json:"send_email_config"`
}

func (cfg *Config) SetDefault() {
	if cfg.Interval <= 0 {
		cfg.Interval = 3600
	}
}

func (cfg *Config) ConfigItems() []interface{} {
	return []interface{}{
		&cfg.CLAUpdateConfig,
		&cfg.SendEmailConfig,
	}
}

func (cfg *Config) intervalDuration() time.Duration {
	return time.Second * time.Duration(cfg.Interval)
}

type CLAUpdateConfig struct {
	MsgChannelSize     int `json:"msg_channel_size"`
	PythonRetryTimes   int `json:"python_retry_times"`
	GenAllDiffInterval int `json:"gen_all_diff_interval"`
}

func (cfg *CLAUpdateConfig) SetDefault() {
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

func (cfg *CLAUpdateConfig) genAllDiffInterval() time.Duration {
	return time.Second * time.Duration(cfg.GenAllDiffInterval)
}

type NotifyAdminConfig struct {
	SendEmailInterval          int      `json:"send_email_interval"`
	NotifyCorpAdminInterval    int      `json:"notify_corp_admin_interval"`
	NotifyIndividualInterval   int      `json:"notify_individual_interval"`
	NotifyIndividualRemindDays *int     `json:"notify_individual_remind_days"`
	NotifyCorpAdminRemindDays  *int     `json:"notify_corp_admin_remind_days"`
	NotifyBatchSize            int      `json:"notify_batch_size"`
	EnabledCommunityOrgs       []string `json:"enabled_community_orgs"`
	NotifyEmailTo              string   `json:"notify_email_to"`
}

func (cfg *NotifyAdminConfig) SetDefault() {
	if cfg.SendEmailInterval <= 0 {
		cfg.SendEmailInterval = 10
	}

	if cfg.NotifyCorpAdminInterval <= 0 {
		cfg.NotifyCorpAdminInterval = 86400
	}

	if cfg.NotifyIndividualInterval <= 0 {
		cfg.NotifyIndividualInterval = 86400
	}

	if cfg.NotifyIndividualRemindDays == nil {
		v := 90
		cfg.NotifyIndividualRemindDays = &v
	}

	if cfg.NotifyCorpAdminRemindDays == nil {
		v := 90
		cfg.NotifyCorpAdminRemindDays = &v
	}

	if cfg.NotifyBatchSize <= 0 {
		cfg.NotifyBatchSize = 500
	}
}

func (cfg *NotifyAdminConfig) genSendEmailInterval() time.Duration {
	return time.Second * time.Duration(cfg.SendEmailInterval)
}

func (cfg *NotifyAdminConfig) genNotifyCorpAdminInterval() time.Duration {
	return time.Second * time.Duration(cfg.NotifyCorpAdminInterval)
}

func (cfg *NotifyAdminConfig) genNotifyIndividualInterval() time.Duration {
	return time.Second * time.Duration(cfg.NotifyIndividualInterval)
}

func (cfg *NotifyAdminConfig) genNotifyIndividualRemindDays() int {
	return *cfg.NotifyIndividualRemindDays
}

func (cfg *NotifyAdminConfig) genNotifyCorpAdminRemindDays() int {
	return *cfg.NotifyCorpAdminRemindDays
}

func (cfg *NotifyAdminConfig) genNotifyBatchSize() int {
	return cfg.NotifyBatchSize
}

func (cfg *NotifyAdminConfig) genNotifyEmailTo() string {
	return cfg.NotifyEmailTo
}

func (cfg *NotifyAdminConfig) isCommunityEnabled(alias string) bool {
	if len(cfg.EnabledCommunityOrgs) == 0 {
		return true
	}

	for _, org := range cfg.EnabledCommunityOrgs {
		if org == alias {
			return true
		}
	}

	return false
}
