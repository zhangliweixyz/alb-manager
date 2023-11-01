package service

import (
	"alb-manager/conf"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strconv"
)

//节点上下线
func NodeUpDown(c *gin.Context) {
	var upDownObj UpDown
	if err := c.ShouldBindJSON(&upDownObj); err != nil {
		log.Println("NodeUpDown json解析错误：", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log.Println("节点上下线信息：", upDownObj)

	resCode := 0
	resMsg := "success"
	reSult := make(map[string]string)

	for _, where := range upDownObj.Where {
		url := conf.ViperConfig.GetString(where.Clusterid + ".url")
		token := conf.ViperConfig.GetString(where.Clusterid + ".token")
		endpoints := conf.ViperConfig.GetStringSlice(where.Clusterid + ".etcd")
		username := conf.ViperConfig.GetString(where.Clusterid + ".etcduser")
		password := conf.ViperConfig.GetString(where.Clusterid + ".etcdpass")
		svc := &UpDownId{
			Url:       url,
			Token:     token,
			Nodes:     make(map[string]interface{}),
			Clusterid: where.Clusterid,
			Id:        where.Id,
			Server:    upDownObj.Server,
			Weight:    upDownObj.Weight,
			Action:    upDownObj.Action,
			Endpoints: endpoints,
			Username:  username,
			Password:  password,
		}
		res, msg := svc.ServiceNodesUpDown()
		if res == false {
			resCode = 1
			resMsg = "服务器组ID " + where.Id + "：" + msg
			log.Println("节点上下线 "+where.Id+" 执行结果：", false, msg)
			break
		}
	}
	if resCode == 0 {
		log.Println("节点上下线执行结果：", true, "success")
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    resCode,
		"message": resMsg,
		"result":  reSult,
	})
}

//创建服务器组
func CreateServerGroup(c *gin.Context) {
	var serviceObj Service
	if err := c.ShouldBindJSON(&serviceObj); err != nil {
		log.Println("CreateServerGroup json解析错误：", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log.Println("创建服务器组信息：", serviceObj)

	resCode := 0
	resMsg := "success"
	reSult := make(map[string]string)

	serviceObj.Url = conf.ViperConfig.GetString(serviceObj.Clusterid + ".url")
	serviceObj.Token = conf.ViperConfig.GetString(serviceObj.Clusterid + ".token")
	res, msg := serviceObj.CreateOrUpdateService(1)
	if res == false {
		resCode = 1
		resMsg = msg
	} else {
		reSult["id"] = serviceObj.Id
	}
	log.Println("创建服务器组执行结果：", res, msg)

	c.JSON(http.StatusOK, gin.H{
		"code":    resCode,
		"message": resMsg,
		"result":  reSult,
	})
}

//编辑服务器组
func UpdateServerGroup(c *gin.Context) {
	var serviceObj Service
	if err := c.ShouldBindJSON(&serviceObj); err != nil {
		log.Println("UpdateServerGroup json解析错误：", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log.Println("编辑服务器组信息：", serviceObj)

	resCode := 0
	resMsg := "success"
	reSult := make(map[string]string)

	serviceObj.Url = conf.ViperConfig.GetString(serviceObj.Clusterid + ".url")
	serviceObj.Token = conf.ViperConfig.GetString(serviceObj.Clusterid + ".token")
	res, msg := serviceObj.CreateOrUpdateService(2)
	if res == false {
		resCode = 1
		resMsg = msg
	}
	log.Println("编辑服务器组执行结果：", res, msg)

	c.JSON(http.StatusOK, gin.H{
		"code":    resCode,
		"message": resMsg,
		"result":  reSult,
	})
}

//删除服务器组
func DeleteServerGroup(c *gin.Context) {
	var deleteSvcObj DeleteSvc
	if err := c.ShouldBindJSON(&deleteSvcObj); err != nil {
		log.Println("DeleteServerGroup json解析错误：", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log.Println("删除服务器组信息：", deleteSvcObj)

	resCode := 0
	resMsg := "success"
	reSult := make(map[string]string)

	deleteSvcObj.Url = conf.ViperConfig.GetString(deleteSvcObj.Clusterid + ".url")
	deleteSvcObj.Token = conf.ViperConfig.GetString(deleteSvcObj.Clusterid + ".token")

	res, msg := deleteSvcObj.DeleteService()
	if res == false {
		resCode = 1
		resMsg = msg
	}
	log.Println("删除服务器组执行结果：", res, msg)

	c.JSON(http.StatusOK, gin.H{
		"code":    resCode,
		"message": resMsg,
		"result":  reSult,
	})
}

//查询服务器组
func GetServerGroup(c *gin.Context) {
	var getSvcObj GetSvc
	if err := c.ShouldBindJSON(&getSvcObj); err != nil {
		log.Println("GetServerGroup json解析错误：", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log.Println("查询服务器组信息：", getSvcObj)

	resCode := 0
	resMsg := "success"
	reSult2 := make(map[string]interface{})

	url := conf.ViperConfig.GetString(getSvcObj.Clusterid + ".url")
	token := conf.ViperConfig.GetString(getSvcObj.Clusterid + ".token")
	for _, id := range getSvcObj.Ids {
		svc := &GetSvcId{
			Url:       url,
			Token:     token,
			Id:        id,
			Clusterid: getSvcObj.Clusterid,
		}
		res, svcPtr, msg := svc.GetService()
		if res == false {
			resCode = 1
			resMsg = "服务器组ID " + id + "：" + msg
			log.Println("查询服务器组ID "+id+" 执行结果：", false, msg)
			break
		}
		svcData := *svcPtr
		usData := svcData.Upstream

		reSult2Data := make(map[string]interface{})
		reSult2Data["name"] = svcData.Name
		reSult2Data["desc"] = svcData.Desc
		if usData.Type == "chash" {
			reSult2Data["loadbalance"] = map[string]string{
				"type":    usData.Type,
				"hash_on": usData.Hash_on,
				"key":     usData.Key,
			}
		} else {
			reSult2Data["loadbalance"] = map[string]string{
				"type": usData.Type,
			}
		}
		reSult2Data["nodes"] = usData.Nodes_int
		reSult2Data["pass_host"] = usData.Pass_host
		reSult2Data["scheme"] = usData.Scheme
		reSult2Data["timeout"] = map[string]int{"connect": usData.Timeout.Connect, "send": usData.Timeout.Send, "read": usData.Timeout.Read}
		reSult2Data["enable_check"] = 0
		if usData.Checks.Active.Type == "http" {
			var status []string
			statusMap := make(map[string]string)
			for _, v := range usData.Checks.Active.Healthy.Http_statuses {
				s := strconv.Itoa(v)[0:1] + "xx"
				statusMap[s] = s
			}
			for k := range statusMap {
				status = append(status, k)
			}
			reSult2Data["enable_check"] = 1
			reSult2Data["checks"] = map[string]interface{}{
				"type":      usData.Checks.Active.Type,
				"path":      usData.Checks.Active.Http_path,
				"headers":   usData.Checks.Active.Req_headers,
				"interval":  usData.Checks.Active.Healthy.Interval,
				"timeout":   usData.Checks.Active.Timeout,
				"successes": usData.Checks.Active.Healthy.Successes,
				"failures":  usData.Checks.Active.Unhealthy.Http_failures,
				"status":    status,
			}
		} else if usData.Checks.Active.Type == "tcp" {
			reSult2Data["enable_check"] = 1
			reSult2Data["checks"] = map[string]interface{}{
				"type":      usData.Checks.Active.Type,
				"interval":  usData.Checks.Active.Healthy.Interval,
				"timeout":   usData.Checks.Active.Timeout,
				"successes": usData.Checks.Active.Healthy.Successes,
				"failures":  usData.Checks.Active.Unhealthy.Tcp_failures,
			}

		}
		reSult2[id] = reSult2Data
	}
	if resCode == 0 {
		log.Println("查询服务器组执行结果：", true, "success")
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    resCode,
		"message": resMsg,
		"result":  reSult2,
	})
}

//反查服务器组
func GetServerGroupByIp(c *gin.Context) {
	var getSvcObj GetSvcByIp
	if err := c.ShouldBindJSON(&getSvcObj); err != nil {
		log.Println("GetServerGroupByIp json解析错误：", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log.Println("反查服务器组信息：", getSvcObj)

	resCode := 0
	resMsg := "success"
	reSult := make(map[string][]string)

	getSvcObj.Endpoints = conf.ViperConfig.GetStringSlice(getSvcObj.Clusterid + ".etcd")
	getSvcObj.Username = conf.ViperConfig.GetString(getSvcObj.Clusterid + ".etcduser")
	getSvcObj.Password = conf.ViperConfig.GetString(getSvcObj.Clusterid + ".etcdpass")

	res, svcList, msg := getSvcObj.GetServiceByIp()
	if res == false {
		resCode = 1
		resMsg = msg
	} else {
		reSult["ids"] = svcList
	}
	log.Println("反查服务器组执行结果：", res, msg)

	c.JSON(http.StatusOK, gin.H{
		"code":    resCode,
		"message": resMsg,
		"result":  reSult,
	})
}

//查询节点数量
func GetServerGroupNodes(c *gin.Context) {
	var getSvcObj GetSvcNodes
	if err := c.ShouldBindJSON(&getSvcObj); err != nil {
		log.Println("GetServerGroupNodes json解析错误：", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log.Println("查询节点数量信息：", getSvcObj)

	resCode := 0
	resMsg := "success"
	reSult := make(map[string]int)

	getSvcObj.Endpoints = conf.ViperConfig.GetStringSlice(getSvcObj.Clusterid + ".etcd")
	getSvcObj.Username = conf.ViperConfig.GetString(getSvcObj.Clusterid + ".etcduser")
	getSvcObj.Password = conf.ViperConfig.GetString(getSvcObj.Clusterid + ".etcdpass")

	res, svcPtr, msg := getSvcObj.GetServiceNods()
	if res == false {
		resCode = 1
		resMsg = msg
	} else {
		reSult = *svcPtr
	}
	log.Println("查询节点数量执行结果：", res, msg)

	c.JSON(http.StatusOK, gin.H{
		"code":    resCode,
		"message": resMsg,
		"result":  reSult,
	})
}
