package model

type BaseModel struct {
}

func (b *BaseModel) MakeOrderSort() {
}

type BaseParams struct {
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	Sort     string `json:"sort"`
	Order    string `json:"order"`
}
