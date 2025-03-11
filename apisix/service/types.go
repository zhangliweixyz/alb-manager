package service

//节点上下线相关类型
type UpDown struct {
	Server string `json:"server"`
	Weight int    `json:"weight"`
	Action string `json:"action"`
	Where  []struct {
		Clusterid string `json:"clusterid"`
		Id        string `json:"id"`
	}
}
type UpDownId struct {
	Url       string
	Token     string
	Nodes     map[string]interface{}
	Clusterid string
	Id        string
	Server    string
	Weight    int
	Action    string
	Endpoints []string
	Username  string
	Password  string
}
type BodyNodes struct {
	Node struct {
		Value struct {
			Upstream struct {
				Nodes interface{}
			}
		}
	}
}

//创建和编辑服务器组相关类型
type Loadbalance struct {
	Type    string
	Hash_on string
	Key     string
}
type Timeout struct {
	Connect int
	Send    int
	Read    int
}
type Checks struct {
	Type      string
	Path      string
	Headers   []string
	Interval  int
	Timeout   int
	Failures  int
	Successes int
	Status    []string
}
type Service struct {
	Url          string
	Token        string
	Clusterid    string `json:"clusterid"`
	Servicename  string `json:"servicename"`
	Env          string `json:"env"`
	Id           string `json:"id"`
	Name         string `json:"name"`
	Desc         string `json:"desc"`
	Loadbalance  `json:"loadbalance"`
	Nodes        map[string]int `json:"nodes"`
	Pass_host    string         `json:"pass_host"`
	Scheme       string         `json:"scheme"`
	Timeout      `json:"timeout"`
	Enable_check int `json:"enable_check"`
	Checks       `json:"checks"`
}
type BodyId struct {
	Node struct {
		Value struct {
			Id string
		}
	}
}

var HttpStatus = map[string][]int{
	"2xx": {200, 201, 202, 203, 204, 205, 206},
	"3xx": {300, 301, 302, 303, 304, 305, 306, 307},
	"4xx": {400, 401, 402, 403, 404, 405, 406, 407, 408, 409, 410, 411, 412, 413, 414, 415, 416, 417},
	"5xx": {500, 501, 502, 503, 504, 505},
}

//删除服务器组相关类型
type DeleteSvc struct {
	Url       string
	Token     string
	Clusterid string `json:"clusterid"`
	Id        string `json:"id"`
}

//查询服务器组相关类型
type GetSvc struct {
	Clusterid string   `json:"clusterid"`
	Ids       []string `json:"ids"`
}
type GetSvcId struct {
	Url       string
	Token     string
	Id        string
	Clusterid string
}
type GetSvcRes struct {
	Name     string
	Desc     string
	Upstream struct {
		Type      string
		Hash_on   string
		Key       string
		Pass_host string
		Scheme    string
		Timeout   Timeout
		Nodes     interface{}
		Nodes_int map[string]int
		Checks    struct {
			Active struct {
				Type        string
				Http_path   string
				Req_headers []string
				Timeout     int
				Healthy     struct {
					Http_statuses []int
					Interval      int
					Successes     int
				}
				Unhealthy struct {
					Http_statuses []int
					Interval      int
					Http_failures int
					Tcp_failures  int
				}
			}
		}
	}
}
type GetSvcBodyValue struct {
	Node struct {
		Value GetSvcRes
	}
}

//反查服务器组相关类型
type GetSvcByIp struct {
	Endpoints []string
	Username  string
	Password  string
	Clusterid string   `json:"clusterid"`
	Ips       []string `json:"ips"`
}
type IdNodes struct {
	Upstream struct {
		Nodes interface{}
	}
	Id string
}

//查询节点数量相关类型
type GetSvcNodes struct {
	Endpoints []string
	Username  string
	Password  string
	Clusterid string   `json:"clusterid"`
	Ids       []string `json:"ids"`
}
