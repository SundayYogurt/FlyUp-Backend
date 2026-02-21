package dto

type UniversityCreateRequest struct {
	NameTH string `json:"name_th"` // ชื่อมหาวิทยาลัย
	NameEN string `json:"name_en"`
	Domain string `json:"domain"` // email domain เช่น kmutnb.ac.th
}

type UniversityUpdateRequest struct {
	NameTH string  `json:"name_th"` // ชื่อมหาวิทยาลัย
	NameEN string  `json:"name_en"`
	Domain *string `json:"domain,omitempty"`
}

type UniversityResponse struct {
	ID     uint   `json:"id"`
	NameTH string `json:"name_th"` // ชื่อมหาวิทยาลัย
	NameEN string `json:"name_en"`
	Domain string `json:"domain"`
}
