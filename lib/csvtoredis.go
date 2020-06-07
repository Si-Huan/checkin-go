package cks

import (
	"encoding/csv"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

type RedisDB struct {
	parent *Cks
	conn   *redis.Client
}

func NewRedisDB(parent *Cks) *RedisDB {

	r := new(RedisDB)

	r.parent = parent

	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	_, err := client.Ping(ctx).Result()

	if err != nil {
		r.parent.logger.WithFields(logrus.Fields{
			"scope": "redis/NewRedisDB",
		}).Fatal(err)
	}

	r.conn = client

	r.parent.logger.WithFields(logrus.Fields{
		"scope": "redis/NewRedisDB",
	}).Info("RedisDB Connected!")

	return r

}

func (r *RedisDB) LoadData() {

	r.conn.FlushAll(ctx)
	r.parent.logger.WithFields(logrus.Fields{
		"scope": "csv/LoadData",
	}).Warn("FlushALl Done!")

	fileName := "./data.csv"

	fs, err := os.Open(fileName)
	if err != nil {
		r.parent.logger.WithFields(logrus.Fields{
			"scope": "csv/OpenFile",
		}).Fatal(err)
	}
	defer fs.Close()

	csvr := csv.NewReader(fs)
	content, err := csvr.ReadAll()

	if err != nil {
		r.parent.logger.WithFields(logrus.Fields{
			"scope": "csv/ReadAll",
		}).Fatal(err)
	}

	for _, row := range content {
		r.AddStudent(row)
	}

	r.parent.logger.WithFields(logrus.Fields{
		"scope": "csv/LoadData",
	}).Info("Loaded!")

}

func (r *RedisDB) AddStudent(row []string) {

	s := NewStudent(row[0], row[1], row[2])
	r.Save(s)

	r.conn.SAdd(ctx, "AllStudent", s.sid)
	r.conn.SAdd(ctx, s.Class, s.sid)

}

func (r *RedisDB) GetStudent(sid string) *Resp {
	resp := new(Resp)

	s, err := r.QueryStudent(sid)
	if err != nil {
		resp.Code = -1
		resp.Msg = "查无此人"
		return resp
	}

	if s.docard {
		resp.Code = 1
		resp.Msg = "已打卡"
		return resp
	}

	resp.Code = 0
	resp.Msg = "OK"
	resp.Data = s

	return resp
}

func (r *RedisDB) CheckIn(req *http.Request) *Resp {

	sid := req.FormValue("sid")

	resp := new(Resp)

	s, err := r.QueryStudent(sid)
	if err != nil {
		resp.Code = -1
		resp.Msg = "查无此人"
		return resp
	}

	if s.docard {
		resp.Code = 1
		resp.Msg = "已打卡"
		return resp
	}

	infos, abn, err := GetInfosFromForm(req)
	if err != nil {
		resp.Code = -2
		resp.Msg = "Bad Request"
		return resp
	}
	s.infos = infos
	s.abnormal = abn
	s.docard = true

	r.Save(*s)
	r.conn.SAdd(ctx, "OKStudent", sid)

	r.parent.logger.WithFields(logrus.Fields{
		"scope": "Server/CheckIn",
	}).Info(sid)

	resp.Code = 0
	resp.Msg = "OK"
	return resp
}

func (r *RedisDB) Check(class string) *Resp {
	resp := new(Resp)
	ci := NewCheckInfo()

	var all []string
	if class == "All" {
		all, _ = r.conn.SMembers(ctx, "AllStudent").Result()
	} else {
		all, _ = r.conn.SMembers(ctx, class).Result()
		if len(all) == 0 {
			resp.Code = -1
			resp.Msg = "查无此班"
			return resp
		}
	}

	for _, sid := range all {
		s, _ := r.QueryStudent(sid)

		if s.docard == false {
			ci.Nocheck = append(ci.Nocheck, *s)
		} else {
			if len(s.abnormal) > 0 {
				abn := NewAbnormal(*s)
				for _, abinfo := range s.abnormal {
					abn.Infos[abinfo] = s.infos[abinfo]
				}
				ci.Abnormal = append(ci.Abnormal, abn)
			} else {
				ci.Checked = append(ci.Checked, *s)
			}
		}
	}

	r.parent.logger.WithFields(logrus.Fields{
		"scope": "Server/Check",
	}).Info(class)

	resp.Code = 0
	resp.Msg = "OK"
	resp.Data = ci
	return resp
}

func (r *RedisDB) QueryStudent(sid string) (*Student, error) {
	d, _ := r.conn.HGetAll(ctx, sid).Result()
	if len(d) == 0 {
		return nil, errors.New("no sid")
	}

	s := new(Student)
	s.abnormal = make([]string, 0)
	s.infos = make(map[string]string)

	s.sid = d["学号"]
	s.Name = d["姓名"]
	s.Class = d["班级"]

	if d["打卡"] == "1" {
		s.docard = true
	} else {
		s.docard = false
	}

	s.abnormal = strings.Split(d["异常"], "|")
	s.abnormal = s.abnormal[:len(s.abnormal)-1]

	for k, _ := range NormalInfos {
		s.infos[k] = d[k]
	}

	return s, nil
}

func (r *RedisDB) Save(s Student) {
	var data = make(map[string]interface{})

	for k, v := range s.infos {
		data[k] = v
	}
	data["学号"] = s.sid
	data["姓名"] = s.Name
	data["班级"] = s.Class
	data["打卡"] = s.docard

	abn := ""
	for _, info := range s.abnormal {
		abn += info + "|"
	}
	data["异常"] = abn
	_, err := r.conn.HMSet(ctx, s.sid, data).Result()
	if err != nil {
		r.parent.logger.WithFields(logrus.Fields{
			"scope": "StudentInfo/Save",
		}).Warn(err)
	}
}

func GetInfosFromForm(req *http.Request) (map[string]string, []string, error) {
	infos := make(map[string]string)
	abn := make([]string, 0)

	for k, nv := range NormalInfos {
		v := req.FormValue(k)
		if v == "" {
			return nil, nil, errors.New("Bad Request")
		} else {
			infos[k] = v
			if v != nv && nv != "" {
				abn = append(abn, k)
			}
		}
	}
	return infos, abn, nil

}

func (r *RedisDB) Reload() {
	nTime := time.Now()
	yesTime := nTime.AddDate(0, 0, -1)
	Day := yesTime.Format("2006-01-02")
	filename := "./backup/" + Day + ".csv"

	newFile, err := os.Create(filename)
	if err != nil {
		r.parent.logger.WithFields(logrus.Fields{
			"scope": "DumpData/CreateFile",
		}).Fatal(err)
	}
	defer newFile.Close()

	newFile.WriteString("\xEF\xBB\xBF") // 写入UTF-8 BOM，防止中文乱码

	w := csv.NewWriter(newFile)
	header := []string{"学号", "姓名", "班级", "打卡", "异常", "体温", "现居地", "居住地变更", "十五天内是否有过感冒"} //标题
	data := [][]string{
		header,
	}

	all, _ := r.conn.SMembers(ctx, "AllStudent").Result()
	for _, sid := range all {
		d, _ := r.conn.HGetAll(ctx, sid).Result()
		row := make([]string, 0)
		for _, k := range header {
			row = append(row, d[k])
		}
		data = append(data, row)
	}
	w.WriteAll(data)
	w.Flush()

	err = w.Error()
	if err != nil {
		r.parent.logger.WithFields(logrus.Fields{
			"scope": "DumpData/WriteFile",
		}).Fatal(err)
	}
	r.parent.logger.WithFields(logrus.Fields{
		"scope": "csv/DumpData",
	}).Info("Dumped!")
	r.LoadData()
}
