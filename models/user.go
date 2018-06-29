package models

import (
	"errors"
	"fmt"
	"strconv"
	"sync"
	"user/models/mymongo"
	"user/models/myredis"

	"github.com/astaxie/beego"
	"github.com/gomodule/redigo/redis"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type User struct {
	Id            int    `bson:"_id" json:"Id,omitempty" redis:"Id"`
	Phone         string `bson:"Phone" json:"Phone,omitempty" redis:"Phone"`
	Email         string `bson:"Email" json:"Email,omitempty" redis:"Email"`
	Nickname      string `bson:"Nickname" json:"Nickname,omitempty" redis:"Nickname"`
	Password      string `bson:"Password" json:"Password,omitempty" redis:"Password"`
	Authority     uint64 `bson:"Authority" json:"Authority,omitempty "redis:"Authority"`
	HeadSculpture string `bson:"HeadSculpture" json:"HeadSculpture, omitempty" redis:"HeadSculpture"`
	CreateTime    uint64 `bson:"CreateTime" json:"CreateTime,omitempty" redis:"CreateTime"`
	UpdateTime    uint64 `bson:"UpdateTime" json:"UpdateTime,omitempty" redis:"UpdateTime"`
}

const (
	db   = "user"
	user = "user"
)

var (
	userFieldCount = 9
	Id             = "Id"
	Phone          = "Phone"
	Email          = "Email"
	Nickname       = "Nickname"
	Password       = "Password"
	Authority      = "Authority"
	HeadSculpture  = "HeadSculpture"
	CreateTime     = "CreateTime"
	UpdateTime     = "UpdateTime"
)

type Account struct {
	Id int `bson:"_id" json:"Id,omitempty" redis:"Id`
}

const (
	SaleSystem      = 0x00000001
	IM              = 0x00000002
	Product         = 0x00000004
	Community       = 0x00000008
	CustomerService = 0x00000010
	Info            = 0x00000020
	RMS             = 0x00000040
)

const (
	UserId    = 0
	UserPhone = 1
	UserEmail = 2
)

var mapAccountType map[int]int
var mapAppType map[uint64]int
var mapUserInfo map[string]int

func Init() {
	mapAppType = make(map[uint64]int)
	mapAppType[SaleSystem] = 0
	mapAppType[IM] = 0
	mapAppType[Product] = 0
	mapAppType[Community] = 0
	mapAppType[CustomerService] = 0
	mapAppType[Info] = 0
	mapAppType[RMS] = 0

	mapAccountType = make(map[int]int)
	mapAccountType[UserId] = 0
	mapAccountType[UserPhone] = 0
	mapAccountType[UserEmail] = 0

	mapUserInfo = make(map[string]int)
	mapUserInfo["Id"] = 0
	mapUserInfo["Phone"] = 1
	mapUserInfo["Email"] = 2
	mapUserInfo["Nickname"] = 3
	mapUserInfo["Password"] = 4
	mapUserInfo["Authority"] = 5
	mapUserInfo["HeadSculpture"] = 6
	mapUserInfo["CreateTime"] = 7
	mapUserInfo["UpdateTime"] = 8
}

func FindAppId(Type uint64) (code int, ok bool) {
	code, ok = mapAppType[Type]
	return
}

func FindAccountType(Type int) (code int, ok bool) {
	code, ok = mapAccountType[Type]
	return
}

func (info *User) Insert() (code int, err error) {
	mConn := mymongo.Conn()
	defer mConn.Close()

	c := mConn.DB(db).C(user)
	err = c.Insert(info)

	if err != nil {
		if mgo.IsDup(err) {
			code = ErrDupRows
		} else {
			code = ErrDatabase
		}
		beego.Error("fail to insert.err=%s,uid=%d", err.Error(), info.Id)
		return
	} else {
		code = 0
	}

	conn := myredis.GetConn()
	if conn.Err() != nil {
		err = conn.Err()
		code = ErrDatabase
		beego.Error("fail to insert.err=%s,uid=%d", err.Error(), info.Id)
		return
	}
	defer conn.Close()

	strKey := fmt.Sprintf("%d:%d:%d", myredis.REDIS_T_HASH, myredis.USER_INFO, info.Id)
	_, err = conn.Do("HMSET", redis.Args{}.Add(strKey).AddFlat(info)...)

	if err != nil {
		beego.Error("fail to insert.err=%s,uid=%d", err.Error(), info.Id)
		code = ErrDatabase
		return
	}

	strKey = fmt.Sprintf("%d:%d", myredis.REDIS_T_SORTSET, myredis.USER_ID_LIST)
	_, err = conn.Do("ZADD", strKey, info.CreateTime, info.Id)

	if err != nil {
		beego.Error("fail to insert.err=%s,uid=%d", err.Error(), info.Id)
		code = ErrDatabase
		return
	}

	if "" != info.Email {
		strKey = fmt.Sprintf("%d:%d", myredis.REDIS_T_HASH, myredis.USER_EMAIL_LIST)
		_, err = conn.Do("HMSET", strKey, info.Email, info.Id)

		if err != nil {
			beego.Error("fail to insert.err=%s,uid=%d", err.Error(), info.Id)
			code = ErrDatabase
			return
		}
	}

	if "" != info.Phone {
		strKey = fmt.Sprintf("%d:%d", myredis.REDIS_T_HASH, myredis.USER_PHONE_LIST)
		_, err = conn.Do("HMSET", strKey, info.Phone, info.Id)

		if err != nil {
			beego.Error("fail to insert.err=%s,uid=%d", err.Error(), info.Id)
			code = ErrDatabase
			return
		}
	}
	return
}

func (info *User) ChangePasswd(oldPasswd string, newPasswd string) (code int, err error) {
	mConn := mymongo.Conn()
	defer mConn.Close()

	c := mConn.DB(db).C(user)
	err = c.Update(bson.M{"_id": info.Id, "Password": oldPasswd}, bson.M{"$set": bson.M{"Password": newPasswd}})

	if err != nil {
		beego.Error("fail to ChangePasswd.err=%s,uid=%d", err.Error(), info.Id)
		if err == mgo.ErrNotFound {
			return ErrNotFound, err
		}

		return ErrDatabase, err
	}

	conn := myredis.GetConn()
	if conn.Err() != nil {
		err = conn.Err()
		code = ErrDatabase
		beego.Error("fail to ChangeAuthority.err=%s,uid=%d", err.Error(), info.Id)
		return
	}
	defer conn.Close()

	strKey := fmt.Sprintf("%d:%d:%d", myredis.REDIS_T_HASH, myredis.USER_INFO, info.Id)
	_, err = conn.Do("HMSET", strKey, Password, newPasswd)

	if err != nil {
		beego.Error("fail to ChangePasswd.err=%s,uid=%d", err.Error(), info.Id)
		code = ErrDatabase
	}

	return
}

func (info *User) ChangeAuthority(oldAuthority uint64, newAuthority uint64) (code int, err error) {
	mConn := mymongo.Conn()
	defer mConn.Close()

	c := mConn.DB(db).C(user)
	err = c.Update(bson.M{"_id": info.Id, "Authority": oldAuthority}, bson.M{"$set": bson.M{"Authority": newAuthority}})

	if err != nil {
		beego.Error("fail to ChangeAuthority.err=%s,uid=%d", err.Error(), info.Id)
		if err == mgo.ErrNotFound {
			return ErrNotFound, err
		}

		return ErrDatabase, err
	}

	conn := myredis.GetConn()
	if conn.Err() != nil {
		err = conn.Err()
		code = ErrDatabase
		beego.Error("fail to ChangeAuthority.err=%s,uid=%d", err.Error(), info.Id)
		return
	}
	defer conn.Close()

	strKey := fmt.Sprintf("%d:%d:%d", myredis.REDIS_T_HASH, myredis.USER_INFO, info.Id)
	_, err = conn.Do("HMSET", strKey, Authority, newAuthority)

	if err != nil {
		beego.Error("fail to ChangeAuthority.err=%s,uid=%d", err.Error(), info.Id)
		code = ErrDatabase
	}
	return
}

func (info *User) ChangeHeadSculpture(oldHeadSculpture string, newHeadSculpture string) (code int, err error) {
	mConn := mymongo.Conn()
	defer mConn.Close()

	c := mConn.DB(db).C(user)
	err = c.Update(bson.M{"_id": info.Id, "HeadSculpture": oldHeadSculpture}, bson.M{"$set": bson.M{"HeadSculpture": newHeadSculpture}})

	if err != nil {
		beego.Error("fail to ChangeHeadSculpture.err=%s,uid=%d", err.Error(), info.Id)
		if err == mgo.ErrNotFound {
			return ErrNotFound, err
		}

		return ErrDatabase, err
	}

	conn := myredis.GetConn()
	if conn.Err() != nil {
		err = conn.Err()
		code = ErrDatabase
		beego.Error("fail to ChangeHeadSculpture.err=%s,uid=%d", err.Error(), info.Id)
		return
	}
	defer conn.Close()

	strKey := fmt.Sprintf("%d:%d:%d", myredis.REDIS_T_HASH, myredis.USER_INFO, info.Id)
	_, err = conn.Do("HMSET", strKey, HeadSculpture, newHeadSculpture)

	if err != nil {
		beego.Error("fail to ChangeHeadSculpture.err=%s,uid=%d", err.Error(), info.Id)
		code = ErrDatabase
	}
	return
}

func (info *User) FindByField(field int, filld_value string, dst []string) (code int, err error) {
	if code, err = info.RedisFindByField(field, filld_value, dst); nil != err {
		beego.Error("fail to FindByField.err=%s,uid=%d", err.Error(), info.Id)
		if code, err = info.MongoFindByField(field, filld_value); nil != err {
			beego.Error("fail to FindByField.err=%s,uid=%d", err.Error(), info.Id)
		}
	}

	return
}

/*
func (info *User) FindUserIsExist(field int, filld_value string) (code int, err error) {
	conn := myredis.GetConn()
	err = nil
	code = 0
	if conn.Err() != nil {
		beego.Error(conn.Err().Error())
	} else {
		defer conn.Close()
		var strKey string
		if UserId != field {
			strKey = fmt.Sprintf("%d:%d", myredis.REDIS_T_SORTSET, myredis.USER_ID_LIST)
		} else if UserPhone == field {
			strKey = fmt.Sprintf("%d:%d", myredis.REDIS_T_SORTSET, myredis.USER_PHONE_LIST)
		} else if UserEmail == field {
			strKey = fmt.Sprintf("%d:%d", myredis.REDIS_T_SORTSET, myredis.USER_EMAIL_LIST)
		}

		keys, errTemp := redis.Values(conn.Do("ZRANGEBYSCORE", strKey, filld_value, filld_value))
		if errTemp != nil {
			beego.Error(errTemp.Error())
		} else {
			if 1 != len(keys) {
				beego.Error(errTemp.Error())
			}
			return
		}

	}

	if code, err = info.MongoFindByField(field, filld_value); nil != err {
		beego.Error("fail to insert.err=%s,uid=%d", err.Error(), info.Id)
	}

	return
}
*/
func (info *User) RedisFindByField(field int, filld_value string, dst []string) (code int, err error) {
	conn := myredis.GetConn()
	err = nil
	code = 0
	if conn.Err() != nil {
		beego.Error(conn.Err().Error())
		return ErrDatabase, conn.Err()
	} else {
		defer conn.Close()
		var strKey string
		if UserId != field {
			if UserPhone == field {
				strKey = fmt.Sprintf("%d:%d", myredis.REDIS_T_HASH, myredis.USER_PHONE_LIST)
			} else if UserEmail == field {
				strKey = fmt.Sprintf("%d:%d", myredis.REDIS_T_HASH, myredis.USER_EMAIL_LIST)
			}

			id, errTemp := redis.Int(conn.Do("HGET", strKey, filld_value))
			if errTemp != nil {
				beego.Error("fail to RedisFindByField.err=%s,uid=%d", errTemp.Error(), info.Id)
				return ErrDatabase, errTemp
			} else {
				strKey = fmt.Sprintf("%d:%d:%d", myredis.REDIS_T_HASH, myredis.USER_INFO, id)
				info.Id = id
			}
		} else {
			strKey = fmt.Sprintf("%d:%d:%s", myredis.REDIS_T_HASH, myredis.USER_INFO, filld_value)
			if info.Id, err = strconv.Atoi(filld_value); nil != err {
				beego.Error("fail to RedisFindByField.err=%s,uid=%d", err.Error(), info.Id)
				return ErrDatabase, err
			}
		}

		iResult, errTemp := redis.Values(conn.Do("HMGET", redis.Args{}.Add(strKey).AddFlat(dst)...))

		if errTemp != nil {
			beego.Error(errTemp.Error())
			return ErrDatabase, errTemp
		} else {
			if len(iResult) != len(dst) {
				beego.Error("not find, key=%s, value=%s", strKey, filld_value)
				return ErrNotFound, errors.New("not find")
			}

			for index, value := range dst {
				switch mapUserInfo[value] {
				case 0:
					if info.Id, err = strconv.Atoi(string(iResult[index].([]byte))); nil != err {
						beego.Error("fail to RedisFindByField.err=%s,uid=%d", err.Error(), info.Id)
						return
					}
				case 1:
					info.Phone = string(iResult[index].([]byte))
				case 2:
					info.Email = string(iResult[index].([]byte))
				case 3:
					info.Nickname = string(iResult[index].([]byte))
				case 4:
					info.Password = string(iResult[index].([]byte))
				case 5:
					var temp int64
					if temp, err = strconv.ParseInt(string(iResult[index].([]byte)), 10, 64); nil != err {
						beego.Error("fail to RedisFindByField.err=%s,uid=%d", err.Error(), info.Id)
						code = ErrDatabase
						return
					}
					info.Authority = uint64(temp)
				case 6:
					info.HeadSculpture = string(iResult[index].([]byte))
				case 7:
					var temp int64
					if temp, err = strconv.ParseInt(string(iResult[index].([]byte)), 10, 64); nil != err {
						beego.Error("fail to RedisFindByField.err=%s,uid=%d", err.Error(), info.Id)
						code = ErrDatabase
						return
					}

					info.CreateTime = uint64(temp)
				case 8:
					var temp int64
					if temp, err = strconv.ParseInt(string(iResult[index].([]byte)), 10, 64); nil != err {
						beego.Error("fail to RedisFindByField.err=%s,uid=%d", err.Error(), info.Id)
						code = ErrDatabase
						return
					}

					info.UpdateTime = uint64(temp)
				default:
					beego.Error("not find, key=%s, value=%s", strKey, value)
					return ErrNotFound, errors.New("not find")
				}
			}
		}
	}

	return
}

/*
func (info *User) FindUserHeadSculpture(field int, filld_value string) (code int, err error) {
	conn := myredis.GetConn()
	err = nil
	code = 0
	if conn.Err() != nil {
		beego.Error(conn.Err().Error())
	} else {
		defer conn.Close()

		if UserId != field {
			var strKey string
			if UserPhone == field {
				strKey = fmt.Sprintf("%d:%d", myredis.REDIS_T_SORTSET, myredis.USER_PHONE_LIST)
			} else if UserEmail == field {
				strKey = fmt.Sprintf("%d:%d", myredis.REDIS_T_SORTSET, myredis.USER_EMAIL_LIST)
			}

			keys, errTemp := redis.Values(conn.Do("ZRANGEBYSCORE", strKey, filld_value, filld_value))
			if errTemp != nil {
				beego.Error(errTemp.Error())
				return ErrDatabase, errTemp
			} else {
				if 1 != len(keys) {
					beego.Error(errTemp.Error())
					return ErrDatabase, errTemp
				}
				filld_value = string(keys[0].([]byte))
			}
		}

		var temp UserHeadSculpture
		strKey := fmt.Sprintf("%d:%d:%s", myredis.REDIS_T_HASH, myredis.USER_INFO, filld_value)
		iResult, errTemp := redis.Values(conn.Do("HMGET", redis.Args{}.Add(strKey).AddFlat(&temp)...))

		if errTemp != nil {
			beego.Error(errTemp.Error())
		} else {
			if len(iResult) != userFieldCount {
				beego.Error(errTemp.Error())
				return ErrNotFound, errTemp
			}
			var temp int64
			if temp, err = strconv.ParseInt(string(iResult[0].([]byte)), 10, 64); nil != err {
				beego.Error("fail to insert.err=%s,uid=%d", err.Error(), info.Id)
				return
			}
			info.Authority = uint64(temp)
			info.HeadSculpture = string(iResult[1].([]byte))
		}
	}

	if code, err = info.MongoFindByField(field, filld_value); nil != err {
		beego.Error("fail to insert.err=%s,uid=%d", err.Error(), info.Id)
	}

	return
}
*/
//从mongo获取某个用户的所有信息，并存入redis
func (info *User) MongoFindByField(field int, filld_value string) (code int, err error) {
	mConn := mymongo.Conn()
	defer mConn.Close()

	c := mConn.DB(db).C(user)
	if UserId == field {
		var b int
		b, err = strconv.Atoi(filld_value)
		if err == nil {
			err = c.Find(bson.M{"_id": b}).One(info)
		}
	} else if UserPhone == field {
		err = c.Find(bson.M{"Phone": filld_value}).One(info)
	} else if UserEmail == field {
		err = c.Find(bson.M{"Email": filld_value}).One(info)
	} else {
		err = errors.New("err input")
		beego.Error("fail to MongoFindByField.err=%s,uid=%s", err.Error(), filld_value)
		code = ErrInput
		return
	}

	if err != nil {
		if err == mgo.ErrNotFound {
			code = ErrNotFound
		} else {
			code = ErrDatabase
		}
		beego.Error("fail to MongoFindByField.err=%s,uid=%d", err.Error(), info.Id)
		return
	} else {
		code = 0
	}

	conn := myredis.GetConn()
	if conn.Err() != nil {
		beego.Error(conn.Err().Error())
	} else {
		defer conn.Close()

		strKey := fmt.Sprintf("%d:%d:%d", myredis.REDIS_T_HASH, myredis.USER_INFO, info.Id)
		_, err = conn.Do("HMSET", redis.Args{}.Add(strKey).AddFlat(info)...)

		if err != nil {
			beego.Error("fail to MongoFindByField.err=%s,uid=%d", err.Error(), info.Id)
			code = ErrDatabase
			return
		}

		strKey = fmt.Sprintf("%d:%d", myredis.REDIS_T_SORTSET, myredis.USER_ID_LIST)
		_, err = conn.Do("ZADD", strKey, info.CreateTime, info.Id)

		if err != nil {
			beego.Error("fail to MongoFindByField.err=%s,uid=%d", err.Error(), info.Id)
			code = ErrDatabase
			return
		}

		if "" != info.Email {
			strKey = fmt.Sprintf("%d:%d", myredis.REDIS_T_SORTSET, myredis.USER_EMAIL_LIST)
			_, err = conn.Do("ZADD", strKey, info.Email, info.Id)

			if err != nil {
				beego.Error("fail to MongoFindByField.err=%s,uid=%d", err.Error(), info.Id)
				code = ErrDatabase
				return
			}
		}

		if "" != info.Phone {
			strKey = fmt.Sprintf("%d:%d", myredis.REDIS_T_SORTSET, myredis.USER_PHONE_LIST)
			_, err = conn.Do("ZADD", strKey, info.Phone, info.Id)

			if err != nil {
				beego.Error("fail to MongoFindByField.err=%s,uid=%d", err.Error(), info.Id)
				code = ErrDatabase
				return
			}
		}
	}
	return
}

/*
//从redis获取某个用户的所有信息
func (info *User) RedisFindByField(field int, filld_value string) (code int, err error) {
	conn := myredis.GetConn()
	err = nil
	code = 0
	if conn.Err() != nil {
		beego.Error(conn.Err().Error())
	} else {
		defer conn.Close()
		if UserId != field {
			var strKey string
			if UserPhone == field {
				strKey = fmt.Sprintf("%d:%d", myredis.REDIS_T_SORTSET, myredis.USER_PHONE_LIST)
			} else if UserEmail == field {
				strKey = fmt.Sprintf("%d:%d", myredis.REDIS_T_SORTSET, myredis.USER_EMAIL_LIST)
			}

			keys, errTemp := redis.Values(conn.Do("ZRANGEBYSCORE", strKey, filld_value, filld_value))
			if errTemp != nil {
				beego.Error(errTemp.Error())
				return ErrDatabase, errTemp
			} else {
				if 1 != len(keys) {
					beego.Error("not find key=%s, value=%s", strKey, filld_value)
					return ErrDatabase, errors.New("not find")
				}
				filld_value = string(keys[0].([]byte))
			}
		}

		strKey1 := fmt.Sprintf("%d:%d:%s", myredis.REDIS_T_HASH, myredis.USER_INFO, filld_value)
		iResult, errTemp := redis.Values(conn.Do("HMGET", redis.Args{}.Add(strKey1).AddFlat(info)...))

		if errTemp != nil {
			beego.Error(errTemp.Error())
		} else {
			if len(iResult) != userFieldCount {
				beego.Error("not find key=%s, value=%s", strKey1, filld_value)
				return ErrNotFound, errors.New("not find")
			}
			if info.Id, err = strconv.Atoi(string(iResult[0].([]byte))); nil != err {
				beego.Error("fail to insert.err=%s,uid=%d", err.Error(), info.Id)
				return
			}
			info.Phone = string(iResult[1].([]byte))
			info.Email = string(iResult[2].([]byte))
			info.Nickname = string(iResult[3].([]byte))
			info.Password = string(iResult[4].([]byte))
			var temp int64
			if temp, err = strconv.ParseInt(string(iResult[5].([]byte)), 10, 64); nil != err {
				beego.Error("fail to insert.err=%s,uid=%d", err.Error(), info.Id)
				return
			}
			info.Authority = uint64(temp)
			info.HeadSculpture = string(iResult[6].([]byte))

			if temp, err = strconv.ParseInt(string(iResult[7].([]byte)), 10, 64); nil != err {
				beego.Error("fail to insert.err=%s,uid=%d", err.Error(), info.Id)
				return
			}

			info.CreateTime = uint64(temp)

			if temp, err = strconv.ParseInt(string(iResult[8].([]byte)), 10, 64); nil != err {
				beego.Error("fail to insert.err=%s,uid=%d", err.Error(), info.Id)
				return
			}

			info.UpdateTime = uint64(temp)
		}
	}

	return
}
*/
func GetUserListInfo(PageSzie int, PageIndex int, dst []string) (Recodes []StUserInfo, code int, err error) {
	conn := myredis.GetConn()
	err = nil
	code = 0
	if conn.Err() != nil {
		beego.Error(conn.Err().Error())
		err = conn.Err()
		code = ErrDatabase
	} else {
		defer conn.Close()

		strKey := fmt.Sprintf("%d:%d", myredis.REDIS_T_SORTSET, myredis.USER_ID_LIST)
		keys, errTemp := redis.Values(conn.Do("ZRANGE", strKey, PageSzie*(PageIndex-1), PageSzie*PageIndex))

		if errTemp != nil {
			beego.Error("fail to GetUserListInfo.err=%s,pageindex=%d,pagesize", errTemp.Error(), PageSzie, PageIndex)
			mConn := mymongo.Conn()
			defer mConn.Close()
			c := mConn.DB(db).C(user)
			err = c.Find(nil).Sort("-CreateTime").Skip(int(PageSzie * (PageIndex - 1))).Limit(int(PageSzie)).All(&Recodes)

			if err != nil {
				if err == mgo.ErrNotFound {
					code = ErrNotFound
				} else {
					code = ErrDatabase
				}
				beego.Error("fail to GetUserListInfo.err=%s,pageindex=%d,pagesize", err.Error(), PageSzie, PageIndex)
			}
		} else {
			var temp User

			for _, value := range keys {
				if code, err = temp.FindByField(UserId, string(value.([]byte)), dst); nil != err {
					beego.Error("fail to GetUserListInfo.err=%s,pageindex=%d,pagesize", err.Error(), PageSzie, PageIndex)
					code = ErrDatabase
					return
				}
				Recodes = append(Recodes, StUserInfo{temp.Id, temp.Nickname, temp.Authority, temp.CreateTime})
			}
		}
	}

	return
}

func GetUserListSize() (count int, code int, err error) {
	conn := myredis.GetConn()
	err = nil
	code = 0
	if conn.Err() != nil {
		beego.Error(conn.Err().Error())
	} else {
		defer conn.Close()
		strKey := fmt.Sprintf("%d:%d", myredis.REDIS_T_SORTSET, myredis.USER_ID_LIST)
		Count, errTemp := redis.Int(conn.Do("ZCARD", strKey))

		if errTemp != nil {
			beego.Error("fail to GetUserListSize.err=%s", errTemp.Error())
		} else {
			count = Count
			return
		}
	}

	mConn := mymongo.Conn()
	defer mConn.Close()

	c := mConn.DB(db).C(user)
	count, err = c.Find(nil).Count()

	if err != nil {
		if err == mgo.ErrNotFound {
			code = ErrNotFound
		} else {
			code = ErrDatabase
		}
		beego.Error("fail to GetUserListSize.err=%s", err.Error())
	} else {
		code = 0
	}
	return
}

func (info *User) InsertToken(token string) (code int, err error) {
	conn := myredis.GetConn()
	if conn.Err() != nil {
		err = conn.Err()
		code = ErrDatabase
		beego.Error("fail to InsertToken.err=%s,uid=%d", err.Error(), info.Id)
		return
	}
	defer conn.Close()

	strKey := fmt.Sprintf("%d:%d", myredis.REDIS_T_HASH, myredis.TOKEN_LIST)
	_, err = conn.Do("HSET", strKey, info.Id, token)

	if err != nil {
		beego.Error("fail to InsertToken.err=%s,uid=%d", err.Error(), info.Id)
		code = ErrDatabase
	}
	return
}

func (info *User) FindToken() (code int, err error) {
	conn := myredis.GetConn()
	if conn.Err() != nil {
		err = conn.Err()
		code = ErrDatabase
		beego.Error("fail to FindToken.err=%s,uid=%d", err.Error(), info.Id)
		return
	}
	defer conn.Close()

	strKey := fmt.Sprintf("%d:%d", myredis.REDIS_T_HASH, myredis.TOKEN_LIST)
	_, err = redis.String(conn.Do("HGET", strKey, info.Id))

	if err != nil {
		beego.Error("fail to FindToken.err=%s,uid=%d", err.Error(), info.Id)
		code = ErrNotFound
	}
	return
}

func (info *User) RemoveToken() (code int, err error) {
	conn := myredis.GetConn()
	if conn.Err() != nil {
		err = conn.Err()
		code = ErrDatabase
		beego.Error("fail to RemoveToken.err=%s,uid=%d", err.Error(), info.Id)
		return
	}
	defer conn.Close()

	strKey := fmt.Sprintf("%d:%d", myredis.REDIS_T_HASH, myredis.TOKEN_LIST)
	_, err = conn.Do("HDEL", strKey, info.Id)

	if err != nil {
		beego.Error("fail to RemoveToken.err=%s,uid=%d", err.Error(), info.Id)
		code = ErrDatabase
	}
	return
}

var lock sync.Mutex

func GetUserId(UserId *int) (code int, err error) {
	lock.Lock()
	mConn := mymongo.Conn()
	defer mConn.Close()

	c := mConn.DB("").C("account")
	temp := Account{}
	err = c.Find(nil).One(&temp)
	if err != nil {
		lock.Unlock()
		if err == mgo.ErrNotFound {
			return ErrNotFound, err
		}

		return ErrDatabase, err
	}

	*UserId = temp.Id + 1
	New := Account{Id: *UserId}
	err = c.Insert(New)

	if err != nil {
		lock.Unlock()
		if err == mgo.ErrNotFound {
			return ErrNotFound, err
		}

		return ErrDatabase, err
	}

	err = c.Remove(temp)

	if err != nil {
		lock.Unlock()
		if err == mgo.ErrNotFound {
			return ErrNotFound, err
		}

		return ErrDatabase, err
	}

	lock.Unlock()
	return 0, nil
}
