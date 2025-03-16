package main

type Config struct {
	Qbitorrent     QbitorrentConfig `json:"qbitorrent"`
	KinopoiskToken string           `json:"kinopoisk_token"`
}

type QbitorrentConfig struct {
	BaseUrl      string `json:"base_url"`
	Category     string `json:"category"`
	RealSavePath string `json:"real_save_path"`
	Username     string `json:"username"`
	Password     string `json:"password"`
}
