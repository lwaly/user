package models

type StRegister struct {
	AccountType int    `json:"AccountType"    valid:"Required;Mobile"`
	Account     string `json:"Account"     valid:"Required"`
	Name        string `json:"name"     valid:"Required"`
	AppId       uint64 `json:"AppId"    valid:"Required"`
	Password    string `json:"password" valid:"Required"`
}

type StRegisterRsp struct {
	Id int `json:"Id" valid:"Required"`
}

type StLogin struct {
	AccountType int    `json:"AccountType"    valid:"Required;Mobile"`
	Account     string `json:"Account"     valid:"Required"`
	AppId       uint64 `json:"AppId"    valid:"Required"`
	Password    string `json:"password" valid:"Required"`
}

type StLogout struct {
	Id    int    `json:"Id" valid:"Required"`
	AppId uint64 `json:"AppId"    valid:"Required"`
	Token string `json:"Token" valid:"Required"`
}

type StChangePasswd struct {
	Id          int    `json:"Id"    valid:"Required"`
	Token       string `json:"Token" valid:"Required"`
	AppId       uint64 `json:"AppId"    valid:"Required"`
	OldPassword string `json:"OldPassword"     valid:"Required"`
	NewPassword string `json:"NewPassword" valid:"Required"`
}

type StRetrievePassword struct {
	AccountType int    `json:"AccountType"    valid:"Required;Mobile"`
	Account     string `json:"Account"     valid:"Required"`
	AppId       uint64 `json:"AppId"    valid:"Required"`
}

type StDelRole struct {
	Id    int    `json:"Id" valid:"Required"`
	AppId uint64 `json:"AppId"    valid:"Required"`
	Token string `json:"Token" valid:"Required"`
}

type StAddRole struct {
	Id    int    `json:"Id" valid:"Required"`
	AppId uint64 `json:"AppId"    valid:"Required"`
	Token string `json:"Token" valid:"Required"`
}

type StChangeHeadSculpture struct {
	Id            int    `json:"Id" valid:"Required"`
	AppId         uint64 `json:"AppId"    valid:"Required"`
	Token         string `json:"Token" valid:"Required"`
	HeadSculpture string `json:"HeadSculpture" valid:"Required"`
}

type StGetHeadSculpture struct {
	Id    int    `json:"Id" valid:"Required"`
	AppId uint64 `json:"AppId"    valid:"Required"`
	Token string `json:"Token" valid:"Required"`
}

type StGetUserInfo struct {
	AccountType string `json:"AccountType"    valid:"Required;Mobile"`
	Account     string `json:"Account"     valid:"Required"`
	AppId       uint64 `json:"AppId"    valid:"Required"`
	Token       string `json:"Token" valid:"Required"`
}

type StGetUserList struct {
	PageSize  int    `json:"PageSize"    valid:"Required"`
	PageIndex int    `json:"PageIndex"    valid:"Required"`
	Token     string `json:"Token" valid:"Required"`
}

// LoginInfo definiton.
type StLoginInfo struct {
	Id     int    `json:"Id"`
	RoleId uint64 `json:"RoleId"`
	Token  string `json:"Token"`
}

type StHeadSculptureInfo struct {
	Id            int    `json:"Id"`
	HeadSculpture string `json:"HeadSculpture"`
}

type StUserInfo struct {
	Id         int    `bson:"_id" json:"Id"`
	Name       string `bson:"Nickname" json:"Name"`
	RoleId     uint64 `bson:"Authority" json:"RoleId"`
	CreateTime uint64 `bson:"CreateTime" json:"CreateTime,omitempty"`
}

type StUserListInfo struct {
	PageTotal int          `json:"PageTotal"    valid:"Required"`
	PageIndex int          `json:"PageIndex"    valid:"Required"`
	Info      []StUserInfo `json:"Info"`
}
