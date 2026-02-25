package model

type ChatList struct {
	ListID   int    `json:"list_id"`
	User1    int    `json:"user1"`
	User2    int    `json:"user2"`
	LastMsg  string `json:"last_msg"`
	LastTime string `json:"last_time"`
}

type ChatSession struct {
	ListID      int    `json:"list_id"`
	TargetUID   int    `json:"uid"`
	TargetName  string `json:"name"`
	Avatar      string `json:"avatar"`
	LastMsg     string `json:"last_msg"`
	LastTime    string `json:"last_time"`
	UnreadCount int    `json:"unread_count"`
	Online      bool   `json:"online"`
}

type ChatMsg struct {
	MsgID   int    `json:"msg_id"`
	ListID  int    `json:"list_id"`
	FromUID int    `json:"from_uid"`
	ToUID   int    `json:"to_uid"`
	Content string `json:"content"`
	Img     string `json:"img"`
	Status  string `json:"status"`
	AddTime string `json:"addtime"`
}

type AdminChatSession struct {
	ListID      int    `json:"list_id"`
	User1       int    `json:"user1"`
	User1Name   string `json:"user1_name"`
	User1Avatar string `json:"user1_avatar"`
	User2       int    `json:"user2"`
	User2Name   string `json:"user2_name"`
	User2Avatar string `json:"user2_avatar"`
	LastMsg     string `json:"last_msg"`
	LastTime    string `json:"last_time"`
	UnreadCount int    `json:"unread_count"`
	User1Online bool   `json:"user1_online"`
	User2Online bool   `json:"user2_online"`
	LastFromUID int    `json:"last_from_uid"`
}

type ChatSendRequest struct {
	ListID  int    `json:"list_id" binding:"required"`
	ToUID   int    `json:"to_uid" binding:"required"`
	Content string `json:"content"`
}

type ChatSendImageRequest struct {
	ListID int `form:"list_id" binding:"required"`
	ToUID  int `form:"to_uid" binding:"required"`
}

type ChatMessagesRequest struct {
	Limit int `form:"limit,default=50"`
}

type ChatNewRequest struct {
	AfterID int `form:"after_id,default=0"`
}

type ChatHistoryRequest struct {
	BeforeID int `form:"before_id,default=0"`
	Limit    int `form:"limit,default=20"`
}

type ChatCreateRequest struct {
	TargetUID int `json:"target_uid" binding:"required"`
}
