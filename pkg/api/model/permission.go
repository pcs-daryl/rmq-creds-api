package model

type Permission struct {
	User   string `json:"user"`
	Vhost  string `json:"vhost"`
	Access Access `json:"access"`
}

type Access struct {
	Read      string `json:"read"`
	Write     string `json:"write"`
	Configure string `json:"configure"`
}