package route

import (
	"alb-manager/conf"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
)

type Route struct {
	ClusterID              string            `json:"clusterid"`
	RouteID                string            `json:"id"`
	ServiceID              string            `json:"service_id"`
	Desc                   string            `json:"desc"`
	Hosts                  []string          `json:"hosts"`
	Uris                   []string          `json:"uris"`
	Methods                []string          `json:"methods"`
	EnableWebsocket        int8              `json:"websocket"`
	Status                 int               `json:"status"`
	Vars                   []Varsstruct      `json:"vars"`
	ProxyRewriteUriStatus  int               `json:"proxy-rewrite-uri_status"`
	ProxyRewriteUri        ProxyRewriteUri   `json:"proxy-rewrite-uri"`
	ProxyRewriteHostStatus int               `json:"proxy-rewrite-host_status"`
	ProxyRewriteHost       map[string]string `json:"proxy-rewrite-host"`
	ProxyRewriteHeaders    map[string]string `json:"proxy-rewrite-headers"`
	ResponseRewriteHeaders map[string]string `json:"response-rewrite-headers"`
	RedirectType           int               `json:"redirect_type"`
	Redirect               Redirect          `json:"redirect"`
}

type Varsstruct struct {
	Type  string   `json:"type"`
	Var   string   `json:"var"`
	Op    string   `json:"op"`
	Value []string `json:"value"`
	Not   bool     `json:"not"`
}

type Redirect struct {
	HttpToHttps bool   `json:"http_to_https"`
	RetCode     int    `json:"ret_code"`
	Uri         string `json:"uri"`
}

type ProxyRewrite struct {
	Headers   int8              `json:"headers"`
	Host      map[string]string `json:"host"`
	RegexUri  map[string]string `json:"regex_uri"`
	StaticUri string            `json:"static_uri"`
}

type ProxyRewriteUri struct {
	StaticUri string `json:"static_uri"`
	RegexUri  struct {
		Src string `json:"src"`
		Dst string `json:"dst"`
	} `json:"regex_uri"`
}
type ResIdobj struct {
	Node struct {
		Value struct {
			Id string
		}
	}
}

type ResValueobj struct {
	Node struct {
		Value RouteResObj
	}
}

type GetRouteData struct {
	Url       string
	Token     string
	ClusterID string   `json:"clusterid"`
	RouteIds  []string `json:"ids"`
}

type RouteResObj struct {
	RouteID          string          `json:"id"`
	Uri              string          `json:"uri"`
	Uris             []string        `json:"uris"`
	Name             string          `json:"name"`
	Desc             string          `json:"desc"`
	Methods          []string        `json:"methods"`
	Host             string          `json:"host"`
	Hosts            []string        `json:"hosts"`
	Enable_websocket bool            `json:"enable_websocket"`
	Vars             [][]interface{} `json:"vars"`
	Plugins          struct {
		ProxyRewrite struct {
			Headers   map[string]string `json:"headers"`
			Host      string            `json:"host"`
			RegexUri  []string          `json:"regex_uri"`
			StaticUri string            `json:"uri"`
		} `json:"proxy-rewrite"`
		ResponseRewrite struct {
			Headers map[string]string `json:"headers"`
		} `json:"response-rewrite"`
		Redirect struct {
			HttpToHttps bool   `json:"http_to_https"`
			RetCode     int    `json:"ret_code"`
			Uri         string `json:"uri"`
		} `json:"redirect"`
	} `json:"plugins"`
	ServiceId string `json:"service_id"`
	Status    int    `json:"status"`
}

type ResValuePluginsobj struct {
	Node struct {
		Value RoutePluginsResObj
	}
}

type RoutePluginsResObj struct {
	Plugins map[string]interface{} `json:"plugins"`
}

var (
	resCode int
	resMsg  string
	reSult  map[string]string
)

func Test(c *gin.Context) {
	var route Route

	if err := c.ShouldBindJSON(&route); err != nil {
		c.AbortWithStatusJSON(
			http.StatusInternalServerError,
			gin.H{"error": err.Error()})
		return
	}

	//data := make(map[string]interface{})
	//data["name"] = "name"
	//data["upstream_id"] = "xxxxxx"
	//bytesData, _ := json.Marshal(data)

	//viper, err := utils.InitConfig("../../config/config.yaml")
	//if err != nil {
	//  resCode = 1
	//  resMsg = err.Error()
	//  log.Println("读取配置文件错误：", resMsg)
	//
	//} else {
	//  url := viper.GetString(env + "." + clusterid + ".url")
	//  token := viper.GetString(env + "." + clusterid + ".token")
	//}
	//创建一个请求对象
	//req, _ := http.NewRequest("POST", Url+"route/", bytes.NewReader(bytesData))
	//req.Header.Add("X-API-KEY", Token)
	//res, msg := httpRequest(req)

	routedata := make(map[string]interface{})
	routedata["service_id"] = route.ServiceID
	routedata["methods"] = route.Methods
	routedata["status"] = route.Status
	routedata["desc"] = route.Desc

	if route.EnableWebsocket == 1 {
		routedata["enable_websocket"] = true
	} else {
		routedata["enable_websocket"] = false
	}

	if len(route.Hosts) == 1 {
		routedata["host"] = route.Hosts[0]
	} else {
		routedata["hosts"] = route.Hosts
	}

	if len(route.Uris) == 1 {
		routedata["uri"] = route.Uris[0]
	} else {
		routedata["uris"] = route.Uris
	}

	//vars处理逻辑
	//s3 := make([]int,3,5)
	//s3 = append(s3,"xxx")
	//{"type":"http","var": "user","op": "IN","value": ["1","2"],"not":true}

	vars := make([]([]interface{}), 0, 5)
	//v := route.Vars[0]

	routedata["vars"] = vars

	//plugins逻辑begin
	plugins := make(map[string]interface{})

	redirect := make(map[string]interface{})
	plugins_redirect_flag := 0
	if route.RedirectType == 1 {
		redirect["http_to_https"] = true
		plugins_redirect_flag = 1
	} else if route.RedirectType == 2 {
		redirect["ret_code"] = route.Redirect.RetCode
		redirect["uri"] = route.Redirect.Uri
		plugins_redirect_flag = 1
	}
	if plugins_redirect_flag == 1 {
		plugins["redirect"] = redirect
	}

	//pliguin: proxy_rewrite逻辑
	proxy_rewrite := make(map[string]interface{})
	plugins_proxy_rewrite_flag := 0

	if route.ProxyRewriteHostStatus == 1 {
		proxy_rewrite["host"] = route.ProxyRewriteHost["host"]
		plugins_proxy_rewrite_flag = 1
	}

	if len(route.ProxyRewriteHeaders) != 0 {
		proxy_rewrite["headers"] = route.ProxyRewriteHeaders
		plugins_proxy_rewrite_flag = 1
	}

	if route.ProxyRewriteUriStatus == 1 {
		proxy_rewrite["uri"] = route.ProxyRewriteUri.StaticUri
		plugins_proxy_rewrite_flag = 1
	} else if route.ProxyRewriteUriStatus == 2 {
		proxy_rewrite["regex_uri"] = []string{route.ProxyRewriteUri.RegexUri.Src, route.ProxyRewriteUri.RegexUri.Dst}
		plugins_proxy_rewrite_flag = 1
	}

	if plugins_proxy_rewrite_flag == 1 {
		plugins["proxy-rewrite"] = proxy_rewrite
	}

	//pliguin: response_rewrite逻辑
	response_rewrite := make(map[string]interface{})
	plugins_response_rewrite_flag := 0

	if len(route.ResponseRewriteHeaders) != 0 {
		response_rewrite["headers"] = route.ResponseRewriteHeaders
		plugins_response_rewrite_flag = 1
	}

	if plugins_response_rewrite_flag == 1 {
		plugins["response-rewrite"] = response_rewrite
	}

	routedata["plugins"] = plugins
	//plugins逻辑end

	bytesData, _ := json.Marshal(routedata)
	req, _ := http.NewRequest("POST", "http://192.168.26.206:80/apisix/admin/routes/", bytes.NewReader(bytesData))
	req.Header.Add("X-API-KEY", "q5kbJbYiuzyeL8a7bKVeoabce8n9RZJ6")
	//client := http.Client{}
	//发送http请求
	//res, _ := client.Do(req)

	//fmt.Println(routedata)

	res, r_id := httpRequestResponse(req)
	reSult = make(map[string]string)

	if res == false {
		resCode = 1
		resMsg = "failed"
	} else {
		resCode = 0
		resMsg = "success"
		reSult["id"] = r_id
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    resCode,
		"message": resMsg,
		"result":  reSult,
	})
}

func CreateRoutes(c *gin.Context) {
	var route Route

	if err := c.ShouldBindJSON(&route); err != nil {
		c.AbortWithStatusJSON(
			http.StatusInternalServerError,
			gin.H{"error": err.Error()})
		return
	}

	clusterid := route.ClusterID
	url := conf.ViperConfig.GetString(clusterid + ".url")
	token := conf.ViperConfig.GetString(clusterid + ".token")

	routedata := make(map[string]interface{})
	routedata["service_id"] = route.ServiceID
	routedata["methods"] = route.Methods
	routedata["status"] = route.Status
	routedata["desc"] = route.Desc
	routedata["name"] = route.Hosts[0]

	if route.EnableWebsocket == 1 {
		routedata["enable_websocket"] = true
	} else {
		routedata["enable_websocket"] = false
	}

	if len(route.Hosts) == 1 {
		routedata["host"] = route.Hosts[0]
	} else {
		routedata["hosts"] = route.Hosts
	}

	if len(route.Uris) == 1 {
		routedata["uri"] = route.Uris[0]
	} else {
		routedata["uris"] = route.Uris
	}

	if len(route.Vars) > 0 {
		vars := make([]([]interface{}), 0, 5)
		for _, v := range route.Vars {
			va := make([]interface{}, 0, 5)
			v_type := v.Type
			v_var := v.Var

			if v_type == "ngx" {
				va = append(va, v_var)
			} else {
				va = append(va, v_type+"_"+v_var)
			}

			v_not := v.Not
			if v_not {
				va = append(va, "!")
			}

			v_op := v.Op
			va = append(va, v_op)

			if v_op == "IN" {
				v_value := v.Value
				va = append(va, v_value)
			} else {
				v_value := v.Value[0]
				if v_value == "''" {
					va = append(va, nil)
				} else {
					va = append(va, v_value)
				}
			}

			vars = append(vars, va)
		}
		routedata["vars"] = vars
		routedata["priority"] = 10
	}
	//plugins逻辑begin
	plugins := make(map[string]interface{})

	//plugin: redirect
	redirect := make(map[string]interface{})
	plugins_redirect_flag := 0
	if route.RedirectType == 1 {
		redirect["http_to_https"] = true
		plugins_redirect_flag = 1
	} else if route.RedirectType == 2 {
		redirect["ret_code"] = route.Redirect.RetCode
		redirect["uri"] = route.Redirect.Uri
		plugins_redirect_flag = 1
	}
	if plugins_redirect_flag == 1 {
		plugins["redirect"] = redirect
	}

	//plugin:proxy-rewrite
	proxy_rewrite := make(map[string]interface{})
	plugins_proxy_rewrite_flag := 0

	if route.ProxyRewriteHostStatus == 1 {
		proxy_rewrite["host"] = route.ProxyRewriteHost["host"]
		plugins_proxy_rewrite_flag = 1
	}

	if len(route.ProxyRewriteHeaders) != 0 {
		proxy_rewrite["headers"] = route.ProxyRewriteHeaders
		plugins_proxy_rewrite_flag = 1
	}

	if route.ProxyRewriteUriStatus == 1 {
		proxy_rewrite["uri"] = route.ProxyRewriteUri.StaticUri
		plugins_proxy_rewrite_flag = 1
	} else if route.ProxyRewriteUriStatus == 2 {
		proxy_rewrite["regex_uri"] = []string{route.ProxyRewriteUri.RegexUri.Src, route.ProxyRewriteUri.RegexUri.Dst}
		plugins_proxy_rewrite_flag = 1
	}

	if plugins_proxy_rewrite_flag == 1 {
		plugins["proxy-rewrite"] = proxy_rewrite
	}

	//plugins: response-rewrite
	response_rewrite := make(map[string]interface{})
	plugins_response_rewrite_flag := 0

	if len(route.ResponseRewriteHeaders) != 0 {
		response_rewrite["headers"] = route.ResponseRewriteHeaders
		plugins_response_rewrite_flag = 1
	}

	if plugins_response_rewrite_flag == 1 {
		plugins["response-rewrite"] = response_rewrite
	}

	routedata["plugins"] = plugins
	//plugins逻辑end

	bytesData, _ := json.Marshal(routedata)
	req, _ := http.NewRequest("POST", url+"routes/", bytes.NewReader(bytesData))
	req.Header.Add("X-API-KEY", token)

	res, r_id := httpRequestResponse(req)
	reSult = make(map[string]string)

	if res == false {
		resCode = 1
		resMsg = "failed"
	} else {
		resCode = 0
		resMsg = "success"
		reSult["id"] = r_id
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    resCode,
		"message": resMsg,
		"result":  reSult,
	})
}

func UpdateRoutes(c *gin.Context) {
	var route Route

	if err := c.ShouldBindJSON(&route); err != nil {
		c.AbortWithStatusJSON(
			http.StatusInternalServerError,
			gin.H{"error": err.Error()})
		return
	}

	clusterid := route.ClusterID
	url := conf.ViperConfig.GetString(clusterid + ".url")
	token := conf.ViperConfig.GetString(clusterid + ".token")

	route_id := route.RouteID
	routedata := make(map[string]interface{})
	routedata["service_id"] = route.ServiceID
	routedata["methods"] = route.Methods
	routedata["status"] = route.Status
	routedata["desc"] = route.Desc
	routedata["name"] = route.Hosts[0]

	if route.EnableWebsocket == 1 {
		routedata["enable_websocket"] = true
	} else {
		routedata["enable_websocket"] = false
	}

	if len(route.Hosts) == 1 {
		routedata["host"] = route.Hosts[0]
	} else {
		routedata["hosts"] = route.Hosts
	}

	if len(route.Uris) == 1 {
		routedata["uri"] = route.Uris[0]
	} else {
		routedata["uris"] = route.Uris
	}

	if len(route.Vars) > 0 {
		vars := make([]([]interface{}), 0, 5)
		for _, v := range route.Vars {
			va := make([]interface{}, 0, 5)

			v_type := v.Type
			v_var := v.Var

			if v_type == "ngx" {
				va = append(va, v_var)
			} else {
				va = append(va, v_type+"_"+v_var)
			}

			v_not := v.Not
			if v_not {
				va = append(va, "!")
			}

			v_op := v.Op
			va = append(va, v_op)

			if v_op == "IN" {
				v_value := v.Value
				va = append(va, v_value)
			} else {
				v_value := v.Value[0]
				if v_value == "''" {
					va = append(va, nil)
				} else {
					va = append(va, v_value)
				}
			}

			vars = append(vars, va)
		}
		routedata["vars"] = vars
		routedata["priority"] = 10
	}

	//plugins逻辑begin
	plugins := make(map[string]interface{})

	//plugins:redirect
	redirect := make(map[string]interface{})
	plugins_redirect_flag := 0
	if route.RedirectType == 1 {
		redirect["http_to_https"] = true
		plugins_redirect_flag = 1
	} else if route.RedirectType == 2 {
		redirect["ret_code"] = route.Redirect.RetCode
		redirect["uri"] = route.Redirect.Uri
		plugins_redirect_flag = 1
	}
	if plugins_redirect_flag == 1 {
		plugins["redirect"] = redirect
	}

	//plugins: proxy-rewrite
	proxy_rewrite := make(map[string]interface{})
	plugins_proxy_rewrite_flag := 0

	if route.ProxyRewriteHostStatus == 1 {
		proxy_rewrite["host"] = route.ProxyRewriteHost["host"]
		plugins_proxy_rewrite_flag = 1
	}

	if len(route.ProxyRewriteHeaders) != 0 {
		proxy_rewrite["headers"] = route.ProxyRewriteHeaders
		plugins_proxy_rewrite_flag = 1
	}

	if route.ProxyRewriteUriStatus == 1 {
		proxy_rewrite["uri"] = route.ProxyRewriteUri.StaticUri
		plugins_proxy_rewrite_flag = 1
	} else if route.ProxyRewriteUriStatus == 2 {
		proxy_rewrite["regex_uri"] = []string{route.ProxyRewriteUri.RegexUri.Src, route.ProxyRewriteUri.RegexUri.Dst}
		plugins_proxy_rewrite_flag = 1
	}

	if plugins_proxy_rewrite_flag == 1 {
		plugins["proxy-rewrite"] = proxy_rewrite
	}

	//plugins: response-rewrite
	response_rewrite := make(map[string]interface{})
	plugins_response_rewrite_flag := 0

	if len(route.ResponseRewriteHeaders) != 0 {
		response_rewrite["headers"] = route.ResponseRewriteHeaders
		plugins_response_rewrite_flag = 1
	}

	if plugins_response_rewrite_flag == 1 {
		plugins["response-rewrite"] = response_rewrite
	}

	//当前已有其他插件保持原样
	req_get, _ := http.NewRequest("GET", url+"routes/"+route_id, nil)
	req_get.Header.Add("X-API-KEY", token)
	res, route_data := httpGetPluginsValue(req_get)
	//p := route_data.Plugins
	//fmt.Println(p)
	for k, v := range route_data.Plugins {
		if k != "proxy-rewrite" && k != "redirect" && k != "response-rewrite" {
			plugins[k] = v
		}
	}

	routedata["plugins"] = plugins
	//plugins逻辑end

	bytesData, _ := json.Marshal(routedata)
	req, _ := http.NewRequest("PUT", url+"routes/"+route_id, bytes.NewReader(bytesData))
	req.Header.Add("X-API-KEY", token)

	res, r_id := httpRequestResponse(req)
	reSult = make(map[string]string)

	if res == false {
		resCode = 1
		resMsg = "failed"
	} else {
		resCode = 0
		resMsg = "success"
		reSult["id"] = r_id
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    resCode,
		"message": resMsg,
		"result":  reSult,
	})
}

func DeleteRoutes(c *gin.Context) {
	var route Route

	if err := c.ShouldBindJSON(&route); err != nil {
		c.AbortWithStatusJSON(
			http.StatusInternalServerError,
			gin.H{"error": err.Error()})
		return
	}

	route_id := route.RouteID

	clusterid := route.ClusterID
	url := conf.ViperConfig.GetString(clusterid + ".url")
	token := conf.ViperConfig.GetString(clusterid + ".token")
	req, _ := http.NewRequest("DELETE", url+"routes/"+route_id, nil)
	req.Header.Add("X-API-KEY", token)

	res, r_id := httpRequestResponse(req)
	reSult = make(map[string]string)

	if res == false {
		resCode = 1
		resMsg = "failed"
	} else {
		resCode = 0
		resMsg = "success"
		reSult["id"] = r_id
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    resCode,
		"message": resMsg,
		"result":  reSult,
	})
}

func GetRoutes(c *gin.Context) {
	var getroutedata GetRouteData

	if err := c.ShouldBindJSON(&getroutedata); err != nil {
		c.AbortWithStatusJSON(
			http.StatusInternalServerError,
			gin.H{"error": err.Error()})
		return
	}

	clusterid := getroutedata.ClusterID
	url := conf.ViperConfig.GetString(clusterid + ".url")
	token := conf.ViperConfig.GetString(clusterid + ".token")

	//reSult := make(map[string]interface{})
	var reSult sync.Map
	var wg sync.WaitGroup
	for _, route_id := range getroutedata.RouteIds {
		wg.Add(1)
		go func(route_id string) {
			defer wg.Done()

			req, _ := http.NewRequest("GET", url+"routes/"+route_id, nil)
			req.Header.Add("X-API-KEY", token)

			res, route_data := httpRequestResponseValue(req)
			//fmt.Println(route_data.Plugins.ProxyRewrite)

			result_data := make(map[string]interface{})
			result_data["name"] = route_data.Name
			result_data["desc"] = route_data.Desc
			result_data["service_id"] = route_data.ServiceId
			result_data["methods"] = route_data.Methods
			result_data["status"] = route_data.Status

			if len(route_data.Uris) > 0 {
				result_data["uris"] = route_data.Uris
			} else {
				uris := make([]string, 0, 1)
				uris = append(uris, route_data.Uri)
				result_data["uris"] = uris
			}

			if len(route_data.Hosts) > 0 {
				result_data["hosts"] = route_data.Hosts
			} else {
				hosts := make([]string, 0, 1)
				hosts = append(hosts, route_data.Host)
				result_data["hosts"] = hosts
			}

			if route_data.Enable_websocket {
				result_data["websocket"] = 1
			} else {
				result_data["websocket"] = 0
			}

			redirect := make(map[string]interface{})
			if route_data.Plugins.Redirect.HttpToHttps {
				result_data["redirect_type"] = 1
			} else {
				if route_data.Plugins.Redirect.Uri != "" {
					result_data["redirect_type"] = 2
					redirect["ret_code"] = route_data.Plugins.Redirect.RetCode
					redirect["uri"] = route_data.Plugins.Redirect.Uri
					result_data["redirect"] = redirect
				} else {
					result_data["redirect_type"] = 0
					result_data["redirect"] = redirect
				}
			}

			//proxy_rewrite
			proxy_rewrite_host := make(map[string]interface{})
			if len(route_data.Plugins.ProxyRewrite.Host) > 0 {
				proxy_rewrite_host["host"] = route_data.Plugins.ProxyRewrite.Host
				result_data["proxy-rewrite-host_status"] = 1
				result_data["proxy-rewrite-host"] = proxy_rewrite_host
			} else {
				result_data["proxy-rewrite-host_status"] = 0
				result_data["proxy-rewrite-host"] = proxy_rewrite_host
			}

			if len(route_data.Plugins.ProxyRewrite.Headers) > 0 {
				result_data["proxy-rewrite-headers"] = route_data.Plugins.ProxyRewrite.Headers
			} else {
				result_data["proxy-rewrite-headers"] = make(map[string]interface{})
			}

			proxy_rewrite_uri := make(map[string]interface{})
			if route_data.Plugins.ProxyRewrite.StaticUri != "" {
				result_data["proxy-rewrite-uri_status"] = 1
				proxy_rewrite_uri["static_uri"] = route_data.Plugins.ProxyRewrite.StaticUri
				result_data["proxy-rewrite-uri"] = proxy_rewrite_uri
			} else if len(route_data.Plugins.ProxyRewrite.RegexUri) > 0 {
				result_data["proxy-rewrite-uri_status"] = 2
				regex_uri := make(map[string]interface{})
				regex_uri["src"] = route_data.Plugins.ProxyRewrite.RegexUri[0]
				regex_uri["dst"] = route_data.Plugins.ProxyRewrite.RegexUri[1]
				proxy_rewrite_uri["regex_uri"] = regex_uri
				result_data["proxy-rewrite-uri"] = proxy_rewrite_uri
			} else {
				result_data["proxy-rewrite-uri_status"] = 0
				result_data["proxy-rewrite-uri"] = proxy_rewrite_uri
			}

			//response_rewrite
			if len(route_data.Plugins.ResponseRewrite.Headers) > 0 {
				result_data["response-rewrite-headers"] = route_data.Plugins.ResponseRewrite.Headers
			} else {
				result_data["response-rewrite-headers"] = make(map[string]interface{})
			}

			vars := make([]map[string]interface{}, 0, 5)
			if len(route_data.Vars) > 0 {
				for _, v := range route_data.Vars {
					va := make(map[string]interface{})
					var value [1]string
					type_var := v[0].(string)

					type_var_1 := strings.Split(type_var, "_")[0]
					if type_var_1 != "http" && type_var_1 != "arg" && type_var_1 != "post" && type_var_1 != "cookie" {
						va["type"] = "ngx"
						va["var"] = type_var_1
					} else {
						va["type"] = strings.Split(type_var, "_")[0]
						va["var"] = strings.Split(type_var, "_")[1]
					}

					if v[1] == "!" {
						va["not"] = true
					} else {
						va["not"] = false
					}
					va["op"] = v[len(v)-2]
					if va["op"] != "IN" {
						if v[len(v)-1] == nil {
							v[len(v)-1] = "''"
						}
						value[0] = v[len(v)-1].(string)
						va["value"] = value
					} else {
						va["value"] = v[len(v)-1]
					}
					vars = append(vars, va)
				}
			}
			result_data["vars"] = vars

			if res == true {
				//reSult[route_id] = result_data
				reSult.Store(route_id, result_data)
			}

			resCode = 0
			resMsg = "success"

		}(route_id)
	}
	wg.Wait()
	r := map[string]interface{}{}
	reSult.Range(func(key, value interface{}) bool {
		r[fmt.Sprint(key)] = value
		return true
	})
	c.JSON(http.StatusOK, gin.H{
		"code":    resCode,
		"message": resMsg,
		"result":  r,
	})
}

func UpdownRoutes(c *gin.Context) {
	var route Route

	if err := c.ShouldBindJSON(&route); err != nil {
		c.AbortWithStatusJSON(
			http.StatusInternalServerError,
			gin.H{"error": err.Error()})
		return
	}

	route_id := route.RouteID
	clusterid := route.ClusterID
	routedata := make(map[string]interface{})
	routedata["status"] = route.Status

	bytesData, _ := json.Marshal(routedata)

	url := conf.ViperConfig.GetString(clusterid + ".url")
	token := conf.ViperConfig.GetString(clusterid + ".token")

	req, _ := http.NewRequest("PATCH", url+"routes/"+route_id, bytes.NewReader(bytesData))
	req.Header.Add("X-API-KEY", token)

	res, r_id := httpRequestResponse(req)
	reSult = make(map[string]string)

	if res == false {
		resCode = 1
		resMsg = "failed"
	} else {
		resCode = 0
		resMsg = "success"
		reSult["id"] = r_id
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    resCode,
		"message": resMsg,
		"result":  reSult,
	})
}

//发送http请求并返回响应内容
func httpRequestResponse(req *http.Request) (bool, string) {
	//创建一个http客户端
	client := http.Client{}
	//发送http请求
	response, err := client.Do(req)
	if err != nil {
		return false, err.Error()
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return false, err.Error()
	}

	if response.StatusCode == 200 || response.StatusCode == 201 {
		var resIdobj ResIdobj
		if err = json.Unmarshal(body, &resIdobj); err == nil {
			return true, resIdobj.Node.Value.Id
		} else {
			return false, err.Error()
		}
	}

	return false, string(body)
}

func httpRequestResponseValue(req *http.Request) (bool, RouteResObj) {
	//创建一个http客户端
	var v RouteResObj
	client := http.Client{}
	//发送http请求
	response, err := client.Do(req)
	if err != nil {
		return false, v
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return false, v
	}

	if response.StatusCode == 200 || response.StatusCode == 201 {
		var ResValueobj ResValueobj
		if err = json.Unmarshal(body, &ResValueobj); err == nil {
			return true, ResValueobj.Node.Value
		} else {
			return false, v
		}
	}

	return false, v
}

func httpGetPluginsValue(req *http.Request) (bool, RoutePluginsResObj) {
	//创建一个http客户端
	var v RoutePluginsResObj
	client := http.Client{}
	//发送http请求
	response, err := client.Do(req)
	if err != nil {
		return false, v
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return false, v
	}

	if response.StatusCode == 200 || response.StatusCode == 201 {
		var ResValuePluginsobj ResValuePluginsobj
		if err = json.Unmarshal(body, &ResValuePluginsobj); err == nil {
			return true, ResValuePluginsobj.Node.Value
		} else {
			return false, v
		}
	}

	return false, v
}
