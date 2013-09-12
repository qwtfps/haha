// Copyright 2011 Google Inc. All rights reserved.

// Use of this source code is governed by the Apache 2.0

// license that can be found in the LICENSE file.

package myoohoohoo2

import (
	"appengine"
	//"appengine/blobstore"
	"appengine/datastore"
	"encoding/json"
	"fmt"
	//"io"
	"net/http"
	//"os"
	. "github.com/qiniu/api/conf"
	//"github.com/qiniu/api/io"
	"github.com/qiniu/api/rs"
	"strconv"
	"time"
)

type UserStruct struct {
	Uid       int
	UserName  string
	Password  string `json:"-"`
	Sex       string
	Date      time.Time
	TokenTime time.Time `json:"-"`
}

type AudioStruct struct {
	Aid        int
	Uid        int
	AudioKey   string
	AudioTitle string
	IsValid    bool `json:"-"`
	Favorite   int
	Date       time.Time
	Size       int
}

type AudioSlice struct {
	Audios []AudioStruct
}

type CollectionAudio struct {
	Uid int
	Aid int
}

//func main() {
//	mux := http.NewServeMux()
//	r := mux.NewRouter()
//	r.HandleFunc("/", root)
//	r.HandleFunc("/register", register)
//	r.HandleFunc("/login", login)
//	r.HandleFunc("/generate", generate) //get toekn
//	r.HandleFunc("/checkvalid", checkvalid)
//	r.HandleFunc("/query", query)
//	r.HandleFunc("/delaudio", delaudio) //delete audio=set IsValid = false
//	r.HandleFunc("/recqiniu", recqiniu) //get sth from qiniu
//}

func init() {
	ACCESS_KEY = "iN7NgwM31j4-BZacMjPrOQBs34UG1maYCAQmhdCV"
	SECRET_KEY = "6QTOr2Jg1gcZEWDQXKOGZh5PziC2MCV5KsntT70j"
	http.HandleFunc("/", root)
	http.HandleFunc("/register", register)
	http.HandleFunc("/login", login)
	http.HandleFunc("/generate", generate) //get toekn
	http.HandleFunc("/checkvalid", checkvalid)
	http.HandleFunc("/query", query)
	http.HandleFunc("/delaudio", delaudio) //delete audio=set IsValid = false
	http.HandleFunc("/recqiniu", recqiniu) //get sth from qiniu

	http.HandleFunc("/addcollect", addcollect)
	http.HandleFunc("/delcollect", delcollect)
	//http.HandleFunc("/getinfo", getinfo)//for myself to check users audios count,need xxx=yyy?
}

func linkJson(status string, subKey string, val string) string {
	return "{\"status\":" + status + "," + "\"" + subKey + "\"" + ":" + "\"" + string(val) + "\"" + "}"
}

func root(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "root!")
	//var s AudioSlice
	//s.Audios = append(s.Audios, AudioStruct{Aid: 10001, Uid: 111001, AudioKey: "aehjaewb", Favorite: 1, Date: time.Now(), Size: 123456})
	//s.Audios = append(s.Audios, AudioStruct{Aid: 121312, Uid: 21312, AudioKey: "feewdwED", Favorite: 1, Date: time.Now(), Size: 123456})

	//b, err := json.Marshal(s)
	//if err != nil {
	//	fmt.Fprintln(w, linkJson("0", "msg", "no audio"))
	//}
	//fmt.Fprint(w, "{\"status\":\"1\""+","+string(b)+"}")
	//fmt.Fprintln(w, linkJson("1", nil, string(b)))
}

func uptoken(bucketName string, uid string) string {
	//body := "x:uid=" + uid + "&key=$(etag)&size=$(fsize)" // + "&gentime=" + string(time.Now().Unix())
	body := "uid=$(x:uid)&key=$(etag)&size=$(fsize)" + "&gentime=" + string(time.Now().Unix())
	putPolicy := rs.PutPolicy{
		Scope:        bucketName,
		CallbackUrl:  "http://riji001.com/api/202/travel_diary.php?method_name=profile", //http://<your domain>/recqiniu
		CallbackBody: body,                                                              //gae body   eg:test=$(x:test)&key=$(etag)&size=$(fsize)&uid=$(endUser)
		//ReturnUrl:   returnUrl,
		//ReturnBody:  returnBody,
		//AsyncOps:    asyncOps,
		EndUser: uid,           //uid
		Expires: 3600 * 24 * 7, // 1week?
	}
	return putPolicy.Token(nil)
}

func generate(w http.ResponseWriter, r *http.Request) {
	//get uid
	if "POST" == r.Method {
		//uid := r.FormValue("uid")
		uid, _ := strconv.Atoi(r.FormValue("uid"))
		c := appengine.NewContext(r)
		q1 := datastore.NewQuery("UserStruct").Filter("Uid =", uid)
		existUser := make([]UserStruct, 0, 1)
		if _, err := q1.GetAll(c, &existUser); err != nil {
			fmt.Fprint(w, "{\"status\":\"0\"}")
			return
		}
		if len(existUser) > 0 {
			token := uptoken("qtestbucket", r.FormValue("uid"))
			//update user tokentime
			thisUser := existUser[0]

			thisUser = UserStruct{
				Uid:       thisUser.Uid,
				UserName:  thisUser.UserName,
				Password:  thisUser.Password,
				Sex:       thisUser.Sex,
				Date:      thisUser.Date,
				TokenTime: time.Now(),
			}
			//kkk, err1 := datastore.Put(c, datastore.NewIncompleteKey(c, "UserStruct", nil), &thisUser)
			fmt.Fprintln(w, thisUser)
			key_str := "UserStruct" + r.FormValue("uid")
			key := datastore.NewKey(c, "UserStruct", key_str, 0, nil)
			_, err1 := datastore.Put(c, key, &thisUser)
			if err1 != nil {
				fmt.Fprintln(w, linkJson("0", "uploadToken", ""))
			} else {
				fmt.Fprintln(w, linkJson("1", "uploadToken", token))
				//fmt.Fprintln(w, "success")
			}

		} else {
			fmt.Fprintln(w, linkJson("0", "uploadToken", "")) //not find this uid
		}

	}

}

func checkvalid(w http.ResponseWriter, r *http.Request) {
	if "POST" == r.Method {
		uid, _ := strconv.Atoi(r.FormValue("uid"))
		c := appengine.NewContext(r)
		q1 := datastore.NewQuery("UserStruct").Filter("Uid =", uid)
		existUser := make([]UserStruct, 0, 1)
		if _, err := q1.GetAll(c, &existUser); err != nil {
			fmt.Fprint(w, "{\"status\":\"0\"}")
			return
		}
		if len(existUser) > 0 {
			thisUser := existUser[0]
			result := isvalid(thisUser.TokenTime)
			if result {
				fmt.Fprintln(w, linkJson("1", "msg", "success"))
			} else {
				fmt.Fprintln(w, linkJson("0", "msg", "fail")) //need re generate token
			}
		} else {
			fmt.Fprintln(w, linkJson("0", "uploadToken", "")) //not find this uid
		}
	}
}

func isvalid(gentime time.Time) bool {
	//now is before valid time is right
	if time.Now().Before(gentime.Add(1000 * 1000 * 1000 * 3600 * 24 * 7)) {
		return true
	} else {
		return false
	}
	return false
}

//http://requestb.in/   测试七牛返回数据
func recqiniu(w http.ResponseWriter, r *http.Request) {
	if "POST" == r.Method {
		uid := r.FormValue("uid")
		audiokey := r.FormValue("key")
		audiotitle := r.FormValue("title")
		size := r.FormValue("size")

		c := appengine.NewContext(r)

		q := datastore.NewQuery("AudioStruct") //query count
		audios := make([]AudioStruct, 0, 10)   //10need max or auto add
		if _, err := q.GetAll(c, &audios); err != nil {

			return
		}
		count := len(audios)
		//uid_str := strconv.Itoa(count + 1)
		uid_int, _ := strconv.Atoi(uid)
		aid_int := count + 1
		size_int, _ := strconv.Atoi(size)
		audio := AudioStruct{
			Aid:        aid_int,
			Uid:        uid_int,
			AudioKey:   audiokey,
			AudioTitle: audiotitle,
			IsValid:    true,
			Favorite:   0,
			Date:       time.Now(),
			Size:       size_int,
		}
		//_, err1 := datastore.Put(c, datastore.NewIncompleteKey(c, "AudioStruct", nil), &audio)
		aid_str := strconv.Itoa(aid_int)
		key_str := "AudioStruct" + aid_str //应该从七牛返回的来存，这个key，或者从手机传出去的时候就规则好key
		key := datastore.NewKey(c, "AudioStruct", key_str, 0, nil)
		fmt.Fprint(w, "reg:")
		fmt.Fprint(w, key)
		_, err1 := datastore.Put(c, key, &audio)
		if err1 != nil {
			fmt.Fprintln(w, linkJson("0", "msg", "fail"))
			return
		} else {
			fmt.Fprintln(w, linkJson("1", "msg", "success"))
		}

	}
}

//http://localhost:8080/register?username=aaa&password=123456&sex=1
func register(w http.ResponseWriter, r *http.Request) {

	if "POST" == r.Method {

		c := appengine.NewContext(r)
		q := datastore.NewQuery("UserStruct") //query count
		users := make([]UserStruct, 0, 10)    //need max or auto add
		if _, err := q.GetAll(c, &users); err != nil {
			fmt.Fprint(w, "{\"status\":\"0\"}")
			return
		}

		q1 := datastore.NewQuery("UserStruct").Filter("UserName =", r.FormValue("username"))
		existUser := make([]UserStruct, 0, 1)
		if _, err := q1.GetAll(c, &existUser); err != nil {
			fmt.Fprint(w, "{\"status\":\"0\"}")
			return
		}
		if len(existUser) > 0 {
			//msg，用户名已存在
			fmt.Fprint(w, "{\"status\":\"0\",\"msg\":\"1\"}")
			return
		}

		count := len(users)
		//uid_str := strconv.Itoa(count + 1)
		//uid_int, _ := strconv.Atoi(uid_str)
		uid_int := 100000 + count + 1
		//uid_str := strconv.Itoa(uid_int)
		u := UserStruct{
			Uid:      uid_int,
			UserName: r.FormValue("username"),
			Password: r.FormValue("password"),
			Sex:      r.FormValue("sex"),
			Date:     time.Now(),
			//TokenTime: nil,
		}
		uid_str := strconv.Itoa(uid_int)
		key_str := "UserStruct" + uid_str
		key := datastore.NewKey(c, "UserStruct", key_str, 0, nil)
		_, err := datastore.Put(c, key, &u)
		//_, err := datastore.Put(c, datastore.NewIncompleteKey(c, "UserStruct", nil), &u)
		if err != nil {
			fmt.Fprint(w, "{\"status\":\"0\"}")
			return
		} else {
			var s = make(map[string]interface{})
			s["Uid"] = u.Uid
			s["UserName"] = u.UserName
			s["Sex"] = u.Sex
			s["Date"] = u.Date.Format("2006-01-02 15:04:05")
			b, err := json.Marshal(s)
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Fprintln(w, linkJson("1", "userinfo", string(b)))
			//注册成功后,记录原来输入的,客户端执行登录
		}
	} else {
		fmt.Fprintln(w, "get")
	}

}

func login(w http.ResponseWriter, r *http.Request) {
	if "POST" == r.Method {
		c := appengine.NewContext(r)
		q := datastore.NewQuery("UserStruct").Filter("UserName =", r.FormValue("username"))

		users := make([]UserStruct, 0, 1)
		if _, err := q.GetAll(c, &users); err != nil {
			return
		}

		if len(users) > 0 {
			realPwd := users[0].Password
			if r.FormValue("password") == realPwd {
				//返回相应信息
				//response := fmt.Sprintf("%v", users[0].Password)
				var s = make(map[string]interface{})
				s["Uid"] = users[0].Uid
				s["UserName"] = users[0].UserName
				s["Sex"] = users[0].Sex
				s["Date"] = users[0].Date.Format("2006-01-02 15:04:05")
				//response := "userinfo:" + s
				b, err := json.Marshal(s)
				if err != nil {
					fmt.Println(err)
					return
				}
				fmt.Fprintln(w, linkJson("1", "userinfo", string(b)))
				//fmt.Fprint(w, "{\"status\":\"1\"}")
			}
		} else {
			fmt.Fprint(w, "{\"status\":\"0\"}")
		}
	} else {
		fmt.Fprintln(w, "login get")
	}

}

type Sizer interface {
	Size() int64
}

func query(w http.ResponseWriter, r *http.Request) {

	if "POST" == r.Method {
		c := appengine.NewContext(r)
		//q := datastore.NewQuery("AudioStruct").Filter("UserName =", r.FormValue("username"))
		//type order audio
		page, _ := strconv.Atoi(r.FormValue("page"))
		countbegin := (page - 1) * 20 //page * 20
		q := datastore.NewQuery("AudioStruct").Filter("IsValid =", true).Order("-Date").Offset(countbegin).Limit(20)

		audios := make([]AudioStruct, 0, 10) //need max or auto add
		if _, err := q.GetAll(c, &audios); err != nil {

			return
		}

		if len(audios) > 0 {
			fmt.Fprint(w, "has audio")
			getAudio := audios[0].AudioKey
			fmt.Fprint(w, getAudio)

			var s AudioSlice
			for _, value := range audios {
				s.Audios = append(s.Audios, AudioStruct{Aid: value.Aid, Uid: value.Uid, AudioKey: value.AudioKey, AudioTitle: value.AudioTitle, Favorite: value.Favorite, Date: value.Date, Size: value.Size})
			}
			b, err := json.Marshal(s)
			if err != nil {
				fmt.Fprintln(w, linkJson("0", "msg", "no audio"))
			}
			fmt.Fprint(w, "{\"status\":\"1\""+","+string(b)+"}")
			//fmt.Fprintln(w, linkJson("1", "msg", string(b)))

		} else {
			fmt.Fprintln(w, linkJson("0", "msg", "no audio"))
		}

	}
}

func delaudio(w http.ResponseWriter, r *http.Request) {
	if "POST" == r.Method {
	}
}

func addcollect(w http.ResponseWriter, r *http.Request) {
	if "POST" == r.Method {
	}
}
func delcollect(w http.ResponseWriter, r *http.Request) {
	if "POST" == r.Method {
	}
}

//c := appengine.NewContext(r)
//uid, _ := strconv.Atoi(r.FormValue("uid"))
//q := datastore.NewQuery("UserStruct").Filter("Uid =", uid)
//q := datastore.NewQuery("UserStruct").Order("-Date")
//q = q.Filter("Sex =", "1")
//users := make([]UserStruct, 0, 10)
//if _, err := q.GetAll(c, &users); err != nil {

//	return
//}

//if len(users) > 0 {
//	fmt.Fprint(w, "has user")
//	//getUser := users[0].TokenTime
//	name := users[0].UserName
//	fmt.Fprint(w, name)

//	fmt.Fprint(w, strconv.Itoa(len(users)))
//} else {
//	fmt.Fprint(w, "no user")
//}
//return

//type Server struct {
//	ServerName string
//	ServerIP   string
//}

//type Serverslice struct {
//	Servers []Server
//}

//func main() {
//	var s Serverslice
//	s.Servers = append(s.Servers, Server{ServerName: "Shanghai_VPN", ServerIP: "127.0.0.1"})
//	s.Servers = append(s.Servers, Server{ServerName: "Beijing_VPN", ServerIP: "127.0.0.2"})
//	b, err := json.Marshal(s)
//	if err != nil {
//		fmt.Println("json err:", err)
//	}
//	fmt.Println(string(b))
//}
//http://hi.baidu.com/liuhelishuang/item/035bc33f23c389c21b9696a7
//http://golang.usr.cc/thread-52517-1-1.html//可能有用，上传file的defer后正确的
//https://github.com/jimmykuu/gopher/blob/master/src/gopher/account.go
