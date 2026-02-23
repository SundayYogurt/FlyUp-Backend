package dto

type UniversityCreateRequest struct {
	NameTH   string `json:"name_th,omitempty"` // ชื่อมหาวิทยาลัย
	NameEN   string `json:"name_en,omitempty"`
	Province string `json:"province,omitempty"`
	Domain   string `json:"domain,omitempty"` // email domain เช่น kmutnb.ac.th
}

type UniversityUpdateRequest struct {
	NameTH   string `json:"name_th,omitempty"` // ชื่อมหาวิทยาลัย
	NameEN   string `json:"name_en,omitempty"`
	Province string `json:"province,omitempty"`
	Domain   string `json:"domain,omitempty"` // email domain เช่น kmutnb.ac.th
}

type UniversityResponse struct {
	ID       uint   `json:"id"`
	NameTH   string `json:"name_th,omitempty"` // ชื่อมหาวิทยาลัย
	NameEN   string `json:"name_en,omitempty"`
	Province string `json:"province,omitempty"`
	Domain   string `json:"domain,omitempty"` // email domain เช่น kmutnb.ac.th
}
