package main

type AppConf struct {
	SeriaDelay float64 `json:"seria_delay"`
	SeriaMax   float64 `json:"seria_max"`
}

type DbConf struct {
	User string `json:"user"`
	Host string `json:"host"`
	Port string `json:"port"`
}

type Conf struct {
	AppConf AppConf `json:"app"`
	DbConf  DbConf  `json:"db"`
}
