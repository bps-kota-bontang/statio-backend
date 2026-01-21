package dto

type ConfigurationResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Key   string `json:"key"`
	Value string `json:"value"`
}

type UpdateConfigurationRequest struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}
