package cks

type Student struct {
	sid      string   `json:"学号"`
	Name     string   `json:"姓名"`
	Class    string   `json:"班级"`
	docard   bool     `json:"打卡"`
	abnormal []string `json:"异常"`
	infos    map[string]string
}

var NormalInfos = map[string]string{
	"现居地":        "",
	"体温":         "正常",
	"十五天内是否有过感冒": "无",
	"居住地变更":      "无",
	// "北京旅居史":      "无",
	// "家属北京旅居史":    "无",
	"应检尽检四类人员": "不属于",
}

type CheckInfo struct {
	Checked  []Student
	Nocheck  []Student
	Abnormal []Abnormal
}

type Abnormal struct {
	Student Student
	Infos   map[string]string
}

func NewCheckInfo() CheckInfo {
	ci := new(CheckInfo)
	ci.Checked = make([]Student, 0)
	ci.Nocheck = make([]Student, 0)
	ci.Abnormal = make([]Abnormal, 0)

	return *ci
}

func NewAbnormal(s Student) Abnormal {
	abn := new(Abnormal)
	abn.Student = s
	abn.Infos = make(map[string]string, 0)
	return *abn
}

func NewStudent(sid string, name string, class string) Student {

	s := Student{sid, name, class, false, []string{}, NormalInfos}

	return s
}
