package dto

type UserInfoResponse struct {
	Provinsi     string `json:"provinsi"`
	Sub          string `json:"sub"`
	Organisasi   string `json:"organisasi"`
	Jabatan      string `json:"jabatan"`
	Eselon       string `json:"eselon"`
	Kabupaten    string `json:"kabupaten"`
	FirstName    string `json:"first-name"`
	NIP          string `json:"nip"`
	Foto         string `json:"foto"`
	AlamatKantor string `json:"alamat-kantor"`
	Golongan     string `json:"golongan"`
	Name         string `json:"name"`
	NipLama      string `json:"nip-lama"`
	Email        string `json:"email"`
	LastName     string `json:"last-name"`
	Username     string `json:"username"`
}
