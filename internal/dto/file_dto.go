package dto

type FileResponse struct {
	Name string `json:"name"`
	File []byte `json:"file"`
}
