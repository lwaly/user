package controllers

import (
	"encoding/json"
	"strconv"
	"time"
	"user/common"
	"user/models"
	"user/models/myredis"

	"github.com/astaxie/beego"
)

var user_id = 1

// Operations about Users
type UserController struct {
	beego.Controller
}

const (
	USER_ADD      = 5001
	USER_DEL      = 5003
	TOKEN_INVALID = 5005
)

type StQueueUserBody struct {
	UId int `json:"UId" valid:"Required"`
}

type StQueueTokenBody struct {
	Token string `json:"Token" valid:"Required"`
}

type StQueue struct {
	Version int16           `json:"ver" valid:"Required"`
	Cmd     int32           `json:"cmd" valid:"Required"`
	Seq     int             `json:"seq" valid:"Required"`
	Body    json.RawMessage `json:"body"`
}

var g_seq = 0

func (info *UserController) Register() {
	form := models.StRegister{}
	json.Unmarshal(info.Ctx.Input.RequestBody, &form)

	defer info.ServeJSON()

	if "" == form.Name || "" == form.Password || "" == form.Account {
		info.Data["json"] = models.NewErrorInfo(ErrCodeInputData, ErrInputData)
		return
	}

	if _, ok := models.FindAccountType(form.AccountType); ok != true {
		info.Data["json"] = models.NewErrorInfo(ErrCodeInputData, ErrInputData)
		return
	}

	if _, ok := models.FindAppId(form.AppId); ok != true {
		info.Data["json"] = models.NewErrorInfo(ErrCodeInputData, ErrInputData)
		return
	}

	regDate := uint64(time.Now().UnixNano())
	user := models.User{}

	if code, err := user.FindByField(form.AccountType, form.Account, []string{models.Id}); err != nil {
		beego.Error("find:", err)
		if code == models.ErrNotFound {
			if _, err := models.GetUserId(&user.Id); err != nil {
				beego.Error("find:", err)
				info.Data["json"] = models.NewErrorInfo(err.Error())
				return
			}
			if models.UserPhone == form.AccountType {
				user.Phone = form.Account
				user.Email = ""
			} else if models.UserEmail == form.AccountType {
				user.Email = form.Account
				user.Phone = ""
			}

			user.Authority |= form.AppId
			user.Nickname = form.Name
			user.Password = form.Password
			user.CreateTime = regDate
			user.UpdateTime = regDate
			if _, err := user.Insert(); err != nil {
				beego.Error("find:", err)
				info.Data["json"] = models.NewErrorInfo(err.Error())
				return
			}
		} else {
			info.Data["json"] = models.NewErrorInfo(err.Error())
			return
		}
	} else {
		beego.Error("find:", err)
		info.Data["json"] = models.NewErrorInfo(ErrDupUser)
		return
	}

	msg := models.StRegisterRsp{Id: user.Id}
	if b, err := json.Marshal(msg); nil == err {
		info.Data["json"] = models.NewNormalInfo(b)
	} else {
		beego.Error("ErrUnknown:", err)
		info.Data["json"] = models.NewErrorInfo(ErrUnknown)
		return
	}

	body := StQueueUserBody{user.Id}
	if b, err := json.Marshal(body); nil == err {
		temp := StQueue{
			Version: 1,
			Cmd:     USER_ADD,
			Body:    b,
		}
		SendRedisQueue(&temp)
	}
	return
}

func (info *UserController) Login() {
	form := models.StLogin{}
	json.Unmarshal(info.Ctx.Input.RequestBody, &form)

	defer info.ServeJSON()

	if "" == form.Password || "" == form.Account {
		info.Data["json"] = models.NewErrorInfo(ErrCodeInputData, ErrInputData)
		return
	}

	if _, ok := models.FindAccountType(form.AccountType); ok != true {
		info.Data["json"] = models.NewErrorInfo(ErrCodeInputData, ErrInputData)
		return
	}

	if _, ok := models.FindAppId(form.AppId); ok != true {
		info.Data["json"] = models.NewErrorInfo(ErrCodeInputData, ErrInputData)
		return
	}

	user := models.User{}

	if _, err := user.FindByField(form.AccountType, form.Account, []string{models.Password, models.Authority}); err != nil {
		beego.Error("find:", err)
		info.Data["json"] = models.NewErrorInfo(err.Error())
		return
	}

	if code, err := user.FindToken(); err != nil {
		beego.Error("find:", err)
		info.Data["json"] = models.NewErrorInfo(code, err.Error())
		return
	}

	if form.Password != user.Password {
		info.Data["json"] = models.NewErrorInfo(ErrCodeInputData, ErrInputData)
		return
	}

	if 0 == user.Authority&form.AppId {
		info.Data["json"] = models.NewErrorInfo(ErrCodeInputData, ErrInputData)
		return
	}

	retInfo := models.StLoginInfo{
		Id:     user.Id,
		RoleId: user.Authority}

	retInfo.Token = common.Create_token(user.Id)

	if code, err := user.InsertToken(retInfo.Token); nil != err {
		info.Data["json"] = models.NewErrorInfo(code, err.Error())
		return
	}
	if b, err := json.Marshal(retInfo); nil == err {
		info.Data["json"] = models.NewNormalInfo(b)
	} else {
		beego.Error("ErrUnknown:", err)
		info.Data["json"] = models.NewErrorInfo(ErrUnknown)
		return
	}
	return
}

func (info *UserController) Logout() {
	form := models.StLogout{}
	json.Unmarshal(info.Ctx.Input.RequestBody, &form)

	defer info.ServeJSON()

	if _, ok := models.FindAppId(form.AppId); ok != true {
		info.Data["json"] = models.NewErrorInfo(ErrCodeInputData, ErrInputData)
		return
	}

	_, err := common.Token_auth(form.Token)
	if err != nil {
		info.Data["json"] = common.ErrExpired
		return
	}

	user := models.User{}
	if _, err := user.FindByField(models.UserId, strconv.Itoa(form.Id), []string{models.Authority}); err != nil {
		beego.Error("find:", err)
		info.Data["json"] = models.NewErrorInfo(err.Error())
		return
	}

	if 0 == user.Authority&form.AppId {
		info.Data["json"] = models.NewErrorInfo(ErrCodeInputData, ErrInputData)
		return
	}

	if code, err := user.RemoveToken(); nil != err {
		info.Data["json"] = models.NewErrorInfo(code, err.Error())
		return
	}

	var msg []byte
	info.Data["json"] = models.NewNormalInfo(msg)

	body := StQueueTokenBody{form.Token}
	if b, err := json.Marshal(body); nil == err {
		temp := StQueue{
			Version: 1,
			Cmd:     TOKEN_INVALID,
			Body:    b,
		}
		SendRedisQueue(&temp)
	}

	return
}

func (info *UserController) ChangePasswd() {
	form := models.StChangePasswd{}
	json.Unmarshal(info.Ctx.Input.RequestBody, &form)

	defer info.ServeJSON()

	if "" == form.OldPassword || "" == form.NewPassword {
		info.Data["json"] = models.NewErrorInfo(ErrCodeInputData, ErrInputData)
		return
	}

	if _, ok := models.FindAppId(form.AppId); ok != true {
		info.Data["json"] = models.NewErrorInfo(ErrCodeInputData, ErrInputData)
		return
	}

	_, err := common.Token_auth(form.Token)
	if err != nil {
		info.Data["json"] = common.ErrExpired
		return
	}

	user := models.User{}
	if _, err := user.FindByField(models.UserId, strconv.Itoa(form.Id), []string{models.Password, models.Authority}); err != nil {
		beego.Error("find:", err)
		info.Data["json"] = models.NewErrorInfo(err.Error())
		return
	}

	if form.OldPassword != user.Password {
		info.Data["json"] = models.NewErrorInfo(ErrCodeInputData, ErrInputData)
		return
	}

	if 0 == user.Authority&form.AppId {
		info.Data["json"] = models.NewErrorInfo(ErrCodeInputData, ErrInputData)
		return
	}

	if _, err := user.ChangePasswd(form.OldPassword, form.NewPassword); err != nil {
		beego.Error("find:", err)
		info.Data["json"] = models.NewErrorInfo(err.Error())
		return
	}

	if code, err := user.RemoveToken(); nil != err {
		info.Data["json"] = models.NewErrorInfo(code, err.Error())
		return
	}

	var msg []byte
	info.Data["json"] = models.NewNormalInfo(msg)

	body := StQueueTokenBody{form.Token}
	if b, err := json.Marshal(body); nil == err {
		temp := StQueue{
			Version: 1,
			Cmd:     TOKEN_INVALID,
			Body:    b,
		}
		SendRedisQueue(&temp)
	}
	return
}

func (info *UserController) ChangeHeadSculpture() {
	form := models.StChangeHeadSculpture{}
	json.Unmarshal(info.Ctx.Input.RequestBody, &form)

	defer info.ServeJSON()

	if _, ok := models.FindAppId(form.AppId); ok != true {
		info.Data["json"] = models.NewErrorInfo(ErrCodeInputData, ErrInputData)
		return
	}

	_, err := common.Token_auth(form.Token)
	if err != nil {
		info.Data["json"] = common.ErrExpired
		return
	}

	user := models.User{}
	if _, err := user.FindByField(models.UserId, strconv.Itoa(form.Id), []string{models.Authority, models.HeadSculpture}); err != nil {
		beego.Error("find:", err)
		info.Data["json"] = models.NewErrorInfo(err.Error())
		return
	}

	if 0 == user.Authority&form.AppId {
		info.Data["json"] = models.NewErrorInfo(ErrCodeInputData, ErrInputData)
		return
	}

	if _, err := user.ChangeHeadSculpture(user.HeadSculpture, form.HeadSculpture); err != nil {
		beego.Error("find:", err)
		info.Data["json"] = models.NewErrorInfo(err.Error())
		return
	}

	var msg []byte
	info.Data["json"] = models.NewNormalInfo(msg)
	return
}

func (info *UserController) GetHeadSculpture() {
	form := models.StGetHeadSculpture{}
	json.Unmarshal(info.Ctx.Input.RequestBody, &form)

	defer info.ServeJSON()

	if _, ok := models.FindAppId(form.AppId); ok != true {
		info.Data["json"] = models.NewErrorInfo(ErrCodeInputData, ErrInputData)
		return
	}

	_, err := common.Token_auth(form.Token)
	if err != nil {
		info.Data["json"] = common.ErrExpired
		return
	}

	user := models.User{}
	if _, err := user.FindByField(models.UserId, strconv.Itoa(form.Id), []string{models.Authority, models.HeadSculpture}); err != nil {
		beego.Error("find:", err)
		info.Data["json"] = models.NewErrorInfo(err.Error())
		return
	}

	if 0 == user.Authority&form.AppId {
		info.Data["json"] = models.NewErrorInfo(ErrCodeInputData, ErrInputData)
		return
	}

	retInfo := models.StHeadSculptureInfo{
		Id:            form.Id,
		HeadSculpture: user.HeadSculpture}

	if b, err := json.Marshal(retInfo); nil == err {
		info.Data["json"] = models.NewNormalInfo(b)
	} else {
		beego.Error("ErrUnknown:", err)
		info.Data["json"] = models.NewErrorInfo(ErrUnknown)
		return
	}
	return
}

func (info *UserController) DelRole() {
	form := models.StDelRole{}
	json.Unmarshal(info.Ctx.Input.RequestBody, &form)

	defer info.ServeJSON()

	if _, ok := models.FindAppId(form.AppId); ok != true {
		info.Data["json"] = models.NewErrorInfo(ErrCodeInputData, ErrInputData)
		return
	}

	_, err := common.Token_auth(form.Token)
	if err != nil {
		info.Data["json"] = common.ErrExpired
		return
	}

	user := models.User{}
	if _, err := user.FindByField(models.UserId, strconv.Itoa(form.Id), []string{models.Authority}); err != nil {
		beego.Error("find:", err)
		info.Data["json"] = models.NewErrorInfo(err.Error())
		return
	}

	if 0 == user.Authority&form.AppId {
		info.Data["json"] = models.NewErrorInfo(ErrCodeInputData, ErrInputData)
		return
	}
	Authority := 0x7fffffffffffffff ^ form.AppId
	if _, err := user.ChangeAuthority(user.Authority, user.Authority&Authority); err != nil {
		beego.Error("ChangeAuthority:", err)
		info.Data["json"] = models.NewErrorInfo(err.Error())
		return
	}

	var msg []byte
	info.Data["json"] = models.NewNormalInfo(msg)
	return
}

func (info *UserController) AddRole() {
	form := models.StAddRole{}
	json.Unmarshal(info.Ctx.Input.RequestBody, &form)

	defer info.ServeJSON()

	if _, ok := models.FindAppId(form.AppId); ok != true {
		info.Data["json"] = models.NewErrorInfo(ErrCodeInputData, ErrInputData)
		return
	}

	_, err := common.Token_auth(form.Token)
	if err != nil {
		info.Data["json"] = common.ErrExpired
		return
	}

	user := models.User{}
	if _, err := user.FindByField(models.UserId, strconv.Itoa(form.Id), []string{models.Authority}); err != nil {
		beego.Error("find:", err)
		info.Data["json"] = models.NewErrorInfo(err.Error())
		return
	}

	if 0 == user.Authority&models.RMS {
		info.Data["json"] = models.NewErrorInfo(ErrCodeInputData, ErrInputData)
		return
	}

	if 1 == user.Authority&form.AppId {
		info.Data["json"] = models.NewErrorInfo(ErrCodeInputData, ErrInputData)
		return
	}

	if _, err := user.ChangeAuthority(user.Authority, user.Authority|form.AppId); err != nil {
		beego.Error("ChangeAuthority:", err)
		info.Data["json"] = models.NewErrorInfo(err.Error())
		return
	}

	var msg []byte
	info.Data["json"] = models.NewNormalInfo(msg)
	return
}

func (info *UserController) UpdateRole() {
	form := models.StAddRole{}
	json.Unmarshal(info.Ctx.Input.RequestBody, &form)

	defer info.ServeJSON()

	_, err := common.Token_auth(form.Token)
	if err != nil {
		info.Data["json"] = common.ErrExpired
		return
	}

	user := models.User{}
	if _, err := user.FindByField(models.UserId, strconv.Itoa(form.Id), []string{models.Authority}); err != nil {
		beego.Error("find:", err)
		info.Data["json"] = models.NewErrorInfo(err.Error())
		return
	}

	if 0 == user.Authority&models.RMS {
		info.Data["json"] = models.NewErrorInfo(ErrCodeInputData, ErrInputData)
		return
	}

	if _, err := user.ChangeAuthority(user.Authority, form.AppId); err != nil {
		beego.Error("ChangeAuthority:", err)
		info.Data["json"] = models.NewErrorInfo(err.Error())
		return
	}

	var msg []byte
	info.Data["json"] = models.NewNormalInfo(msg)
	return
}

func (info *UserController) RetrievePassword() {
	form := models.StRetrievePassword{}
	json.Unmarshal(info.Ctx.Input.RequestBody, &form)

	defer info.ServeJSON()

	if _, ok := models.FindAccountType(form.AccountType); ok != true {
		info.Data["json"] = models.NewErrorInfo(ErrCodeInputData, ErrInputData)
		return
	}

	if _, ok := models.FindAppId(form.AppId); ok != true {
		info.Data["json"] = models.NewErrorInfo(ErrCodeInputData, ErrInputData)
		return
	}

	user := models.User{}
	if _, err := user.FindByField(form.AccountType, form.Account, []string{models.Authority}); err != nil {
		beego.Error("find:", err)
		info.Data["json"] = models.NewErrorInfo(err.Error())
		return
	}

	if 0 == user.Authority&form.AppId {
		info.Data["json"] = models.NewErrorInfo(ErrCodeInputData, ErrInputData)
		return
	}

	var msg []byte
	info.Data["json"] = models.NewNormalInfo(msg)
	return
}

func (info *UserController) GetUserInfo() {
	form := models.StLogin{}
	json.Unmarshal(info.Ctx.Input.RequestBody, &form)

	defer info.ServeJSON()

	if _, ok := models.FindAccountType(form.AccountType); ok != true {
		info.Data["json"] = models.NewErrorInfo(ErrCodeInputData, ErrInputData)
		return
	}

	if _, ok := models.FindAppId(form.AppId); ok != true {
		info.Data["json"] = models.NewErrorInfo(ErrCodeInputData, ErrInputData)
		return
	}

	user := models.User{}

	if _, err := user.FindByField(form.AccountType, form.Account, []string{models.Id, models.Nickname, models.Authority, models.CreateTime}); err != nil {
		beego.Error("find:", err)
		info.Data["json"] = models.NewErrorInfo(err.Error())
		return
	}

	retInfo := models.StUserInfo{
		Id:         user.Id,
		Name:       user.Nickname,
		RoleId:     user.Authority,
		CreateTime: user.CreateTime}

	if b, err := json.Marshal(retInfo); nil == err {
		info.Data["json"] = models.NewNormalInfo(b)
	} else {
		beego.Error("ErrUnknown:", err)
		info.Data["json"] = models.NewErrorInfo(ErrUnknown)
		return
	}
	return
}

func (info *UserController) GetUserListInfo() {
	form := models.StGetUserList{}
	json.Unmarshal(info.Ctx.Input.RequestBody, &form)

	defer info.ServeJSON()

	_, err := common.Token_auth(form.Token)
	if err != nil {
		info.Data["json"] = common.ErrExpired
		return
	}

	if count, _, err := models.GetUserListSize(); err != nil {
		beego.Error("find:", err)
		info.Data["json"] = models.NewErrorInfo(err.Error())
	} else {
		if Records, _, err := models.GetUserListInfo(form.PageSize, form.PageIndex, []string{models.Id, models.Nickname, models.Authority, models.CreateTime}); err != nil {
			beego.Error("find:", err)
			info.Data["json"] = models.NewErrorInfo(err.Error())
		} else {
			retInfo := models.StUserListInfo{
				PageTotal: count,
				PageIndex: form.PageIndex,
				Info:      Records,
			}
			if b, err := json.Marshal(retInfo); nil == err {
				info.Data["json"] = models.NewNormalInfo(b)
			} else {
				beego.Error("ErrUnknown:", err)
				info.Data["json"] = models.NewErrorInfo(ErrUnknown)
				return
			}
		}
	}
	return
}

func SendRedisQueue(info *StQueue) (err error) {
	g_seq++
	info.Seq = g_seq
	szBytes, errTemp := json.Marshal(info)
	if nil != errTemp {
		beego.Error(errTemp)
		return errTemp
	}
	conn := myredis.GetConn()
	err = nil
	if conn.Err() != nil {
		beego.Error(conn.Err().Error())
		err = conn.Err()
	} else {
		defer conn.Close()

		for _, value := range myredis.RedisQueueKeys {
			if _, err = conn.Do("lpush", value, szBytes); err != nil {
				beego.Error(err)
				return
			}
		}
	}
	return
}
