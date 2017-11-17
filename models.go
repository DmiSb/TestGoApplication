package main

type App struct {
	Bundle string `json:"bundle"`
}

type Geo struct {
	Country string `json:"country"`
}

type Device struct {
	Ifa string `json:"ifa"`
	Geo Geo    `json:"geo"`
	Os  string `json:"os"`
}

type Data struct {
	App    App    `json:"app"`
	Device Device `json:"device"`
}
