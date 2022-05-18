package main

type Update struct {
	UpdateId int     `json:"update_id"`
	Message  Message `json:"message"`
}

type Message struct {
	Chat Chat   `json:"chat"`
	Text string `json:"text"`
}

type Chat struct {
	ChatId int `json:"id"`
}

type RestResponse struct {
	Result []Update `json:"result"`
}

type BotMessage struct {
	ChatId int    `json:"chat_id"`
	Text   string `json:"text"`
}

type BaseMsg struct {
	id       int
	login    string
	name     string
	secname  string
	whom     string
	who      string
	nproject int
	message  string
	naprav   string
	mainid   string
	chat     int
}

type BaseMsg2 struct {
	id       int
	who      string
	nproject int
	message  string
}

type userExists struct {
	userexisis int
}

type Config struct {
	Token string `yaml:"token"`
	Db    struct {
		DBURL    string `yaml:"dburl"`
		Server   string `yaml:"server"`
		Port     int    `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		Database string `yaml:"database"`
	}
}
