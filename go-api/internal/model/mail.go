package model

type Mail struct {
	ID       int    `json:"id"`
	FromUID  int    `json:"from_uid"`
	FromName string `json:"from_name"`
	ToUID    int    `json:"to_uid"`
	ToName   string `json:"to_name"`
	Title    string `json:"title"`
	Content  string `json:"content"`
	FileURL  string `json:"file_url"`
	FileName string `json:"file_name"`
	Status   int    `json:"status"` // 0=未读 1=已读
	AddTime  string `json:"addtime"`
}

type MailSendRequest struct {
	ToUID    int    `json:"to_uid" binding:"required"`
	Title    string `json:"title" binding:"required"`
	Content  string `json:"content"`
	FileURL  string `json:"file_url"`
	FileName string `json:"file_name"`
}

type MailListRequest struct {
	Page  int    `json:"page" form:"page"`
	Limit int    `json:"limit" form:"limit"`
	Type  string `json:"type" form:"type"` // inbox / sent
}
