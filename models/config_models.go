package models

// Config ...
type Config struct {
	HTTP struct {
		Port       int    `yaml:"port"`
		IsTLS      bool   `yaml:"is_tls"`
		ServerCert string `yaml:"server_cert"`
		ServerKey  string `yaml:"server_key"`
	} `yaml:"http"`

	Simulator struct {
		EnableDelay bool   `yaml:"enableDelay"`
		DelayType   string `yaml:"delayType"`
		DelaySec    int    `yaml:"delaySec"`
	} `yaml:"simulator"`

	System struct {
		CookieName string `yaml:"cookie_name"`
		CookieKey  string `yaml:"cookie_key"`
	} `yaml:"system"`
}
