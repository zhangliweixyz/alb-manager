package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
	"io/ioutil"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"
)

//发送http请求
func httpRequest(req *http.Request) (bool, string) {
	//创建一个http客户端
	client := http.Client{}
	//发送http请求
	_, err := client.Do(req)
	if err != nil {
		return false, err.Error()
	}

	return true, "success"
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
		var bodyIdObj BodyId
		if err = json.Unmarshal(body, &bodyIdObj); err == nil {
			return true, bodyIdObj.Node.Value.Id
		} else {
			return false, err.Error()
		}
	}

	return false, string(body)
}

//根据查询body提取节点信息
func byteToGetMap(data []byte) (bool, *map[string]int, string) {
	nodesMap := map[string]int{}
	var bodyNodesObj BodyNodes
	if err := json.Unmarshal(data, &bodyNodesObj); err == nil {
		nodes := bodyNodesObj.Node.Value.Upstream.Nodes
		bytesRes, _ := json.Marshal(nodes)

		typeOfnodes := reflect.TypeOf(nodes)
		//存储类型是Slice
		if typeOfnodes.Kind() == reflect.Slice {
			var nodesSlice []map[string]interface{}
			if err = json.Unmarshal(bytesRes, &nodesSlice); err == nil {
				for _, node := range nodesSlice {
					host := node["host"].(string)
					port := strconv.FormatFloat(node["port"].(float64), 'f', -1, 64)
					weight := int(node["weight"].(float64))
					nodesMap[host+":"+port] = weight
				}
			} else {
				return false, nil, err.Error()
			}
			//存储类型是Map
		} else if typeOfnodes.Kind() == reflect.Map {
			if err = json.Unmarshal(bytesRes, &nodesMap); err != nil {
				return false, nil, err.Error()
			}
			//既非Slice也非Map
		} else {
			return false, nil, "nodes类型错误"
		}
		//bodyNodesObj解析错误
	} else {
		return false, nil, err.Error()
	}

	return true, &nodesMap, ""
}

//更新service节点前检查
func (svc *UpDownId) ServiceNodesUpDown() (bool, string) {
	//添加etcd锁，避免读写apisix数据冲突
	//初始化etcd客户端
	config := clientv3.Config{
		Endpoints:   svc.Endpoints,
		DialTimeout: 5 * time.Second,
		Username:    svc.Username,
		Password:    svc.Password,
	}
	client, err := clientv3.New(config)
	if err != nil {
		return false, err.Error()
	}
	defer client.Close()

	//创建一个session，并根据业务情况设置锁的ttl
	s, _ := concurrency.NewSession(client, concurrency.WithTTL(3))
	defer s.Close()
	//初始化一个锁的实例，并进行加锁
	mu := concurrency.NewMutex(s, "/mutex-alb/"+svc.Id)
	if err = mu.Lock(context.TODO()); err != nil {
		return false, err.Error()
	}


	res, nodesPtr, msg := svc.GetServiceNodes()
	if res == false {
		return res, msg
	}
	nodes := make(map[string]interface{})
	for k, v := range *nodesPtr {
		nodes[k] = v
	}

	if svc.Action == "up" {
		nodes[svc.Server] = svc.Weight
	}
	if svc.Action == "down" {
		if nodes[svc.Server] == nil || len(nodes) == 1 {
			return true, "success"
		}
		var upnodes = 0
		for node := range nodes {
			if nodes[node] != 0 {
				upnodes = upnodes + 1
			}
		}
		if upnodes == 1 && nodes[svc.Server] != 0 {
			msg = svc.Server + " 是最后一个有效节点，不允许摘除"
			return false, msg
		}
		nodes[svc.Server] = 0
	}
	if svc.Action == "del" {
		nodes[svc.Server] = nil
	}

	svc.Nodes = nodes
	res, msg = svc.UpdateServiceNodes()


	//操作完成，执行解锁
	if err = mu.Unlock(context.TODO()); err != nil {
		return false, err.Error()
	}

	return res, msg
}

//上下线获取service节点信息
func (svc *UpDownId) GetServiceNodes() (bool, *map[string]int, string) {
	if svc.Id == "" {
		svc.Id = " "
	}
	req, _ := http.NewRequest("GET", svc.Url+"services/"+svc.Id, nil)
	req.Header.Add("X-API-KEY", svc.Token)

	client := http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return false, nil, err.Error()
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return false, nil, err.Error()
	}

	if response.StatusCode == 200 || response.StatusCode == 201 {
		res, nodesPtr, msg := byteToGetMap(body)
		if res == false {
			return false, nil, msg
		}
		return true, nodesPtr, ""
	}

	//GET请求失败
	return false, nil, string(body)
}

//更新service节点信息
func (svc *UpDownId) UpdateServiceNodes() (bool, string) {
	data := make(map[string]interface{})
	data["upstream"] = map[string]interface{}{"nodes": svc.Nodes}
	bytesData, _ := json.Marshal(data)

	//创建一个请求对象
	req, _ := http.NewRequest("PATCH", svc.Url+"services/"+svc.Id, bytes.NewReader(bytesData))
	req.Header.Add("X-API-KEY", svc.Token)
	res, msg := httpRequest(req)

	return res, msg
}

//创建or更新service，1创建，2更新
func (svc *Service) CreateOrUpdateService(op int) (bool, string) {
	data := make(map[string]interface{})
	upstream := make(map[string]interface{})
	//有服务名和环境则表示选择服务，否则为手动填写
	//名称都改为手动填写，临时注释
	//if svc.Servicename != "" && svc.Env != "" {
	//	data["name"] = svc.Servicename + "-" + svc.Env
	//} else {
	//	data["name"] = svc.Name
	//}
	data["name"] = svc.Name
	data["desc"] = svc.Desc

	upstream["type"] = svc.Loadbalance.Type
	if svc.Loadbalance.Type == "chash" {
		upstream["hash_on"] = svc.Loadbalance.Hash_on
		upstream["key"] = svc.Loadbalance.Key
	}
	upstream["pass_host"] = svc.Pass_host
	upstream["scheme"] = svc.Scheme
	upstream["timeout"] = map[string]int{"send": svc.Timeout.Send, "read": svc.Timeout.Read, "connect": svc.Timeout.Connect}
	//开启健康检查
	if svc.Enable_check == 1 {
		if svc.Checks.Type == "http" {
			HttpStatusBak := make(map[string][]int)
			for k, v := range HttpStatus {
				HttpStatusBak[k] = v
			}
			var statusYes []int
			var statusNo []int
			for _, v := range svc.Checks.Status {
				statusYes = append(statusYes, HttpStatusBak[v]...)
				delete(HttpStatusBak, v)
			}
			for k := range HttpStatusBak {
				statusNo = append(statusNo, HttpStatusBak[k]...)
			}
			checksActive := map[string]interface{}{
				"type":      svc.Checks.Type,
				"timeout":   svc.Checks.Timeout,
				"http_path": svc.Checks.Path,
				"healthy": map[string]interface{}{
					"interval":      svc.Checks.Interval,
					"successes":     svc.Checks.Successes,
					"http_statuses": statusYes,
				},
				"unhealthy": map[string]interface{}{
					"interval":      svc.Checks.Interval,
					"http_failures": svc.Checks.Failures,
					"http_statuses": statusNo,
				},
				"req_headers": svc.Checks.Headers,
			}
			upstream["checks"] = map[string]interface{}{"active": checksActive}
		} else if svc.Checks.Type == "tcp" {
			checksActive := map[string]interface{}{
				"type":    svc.Checks.Type,
				"timeout": svc.Checks.Timeout,
				"healthy": map[string]interface{}{
					"interval":  svc.Checks.Interval,
					"successes": svc.Checks.Successes,
				},
				"unhealthy": map[string]interface{}{
					"interval":     svc.Checks.Interval,
					"tcp_failures": svc.Checks.Failures,
				},
			}
			upstream["checks"] = map[string]interface{}{"active": checksActive}
		}
	}

	res, msg := true, ""
	if op == 1 { //创建
		upstream["nodes"] = svc.Nodes
		data["upstream"] = upstream
		bytesData, _ := json.Marshal(data)
		res, msg = svc.CreateService(bytesData)

	} else { //编辑
		// 关联服务也允许添加或删除节点
		//if svc.Servicename != "" && svc.Env != "" {
		//	var nodesPtr *map[string]int
		//	res, nodesPtr, msg = svc.GetServiceNodes()
		//	if res == false {
		//		return false, msg
		//	}
		//	nodes := *nodesPtr
		//	for k := range nodes {
		//		w, ok := svc.Nodes[k]
		//		if ok {
		//			nodes[k] = w
		//		}
		//	}
		//	upstream["nodes"] = nodes
		//} else {
		//	upstream["nodes"] = svc.Nodes
		//}
		upstream["nodes"] = svc.Nodes
		data["upstream"] = upstream
		bytesData, _ := json.Marshal(data)
		res, msg = svc.UpdateService(bytesData)
	}

	return res, msg
}

//创建service
func (svc *Service) CreateService(data []byte) (bool, string) {
	//创建一个请求对象
	req, _ := http.NewRequest("POST", svc.Url+"services/", bytes.NewReader(data))
	req.Header.Add("X-API-KEY", svc.Token)
	res, msg := httpRequestResponse(req)
	if res == false {
		return res, msg
	}
	svc.Id = msg

	return true, "success"
}

//更新service
func (svc *Service) UpdateService(data []byte) (bool, string) {
	//创建一个请求对象
	req, _ := http.NewRequest("PUT", svc.Url+"services/"+svc.Id, bytes.NewReader(data))
	req.Header.Add("X-API-KEY", svc.Token)

	res, msg := httpRequest(req)
	if res == false {
		return res, msg
	}

	return true, "success"
}

//编辑时获取service节点信息
func (svc *Service) GetServiceNodes() (bool, *map[string]int, string) {
	req, _ := http.NewRequest("GET", svc.Url+"services/"+svc.Id, nil)
	req.Header.Add("X-API-KEY", svc.Token)

	client := http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return false, nil, err.Error()
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return false, nil, err.Error()
	}

	if response.StatusCode == 200 || response.StatusCode == 201 {
		res, nodesPtr, msg := byteToGetMap(body)
		if res == false {
			return false, nil, msg
		}
		return true, nodesPtr, ""
	}

	//GET请求失败
	return false, nil, string(body)
}

//删除service
func (svc *DeleteSvc) DeleteService() (bool, string) {
	//创建一个请求对象
	req, _ := http.NewRequest("DELETE", svc.Url+"services/"+svc.Id, nil)
	req.Header.Add("X-API-KEY", svc.Token)

	res, msg := httpRequest(req)
	if res == false {
		return res, msg
	}

	return true, "success"
}

//查询service
func (svc *GetSvcId) GetService() (bool, *GetSvcRes, string) {
	var getSvcBodyValueObj GetSvcBodyValue
	var getSvcResObj GetSvcRes
	req, _ := http.NewRequest("GET", svc.Url+"services/"+svc.Id, nil)
	req.Header.Add("X-API-KEY", svc.Token)

	client := http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return false, nil, err.Error()
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return false, nil, err.Error()
	}

	if response.StatusCode == 200 || response.StatusCode == 201 {
		if err = json.Unmarshal(body, &getSvcBodyValueObj); err == nil {
			getSvcResObj = getSvcBodyValueObj.Node.Value
			res, nodesPtr, msg := byteToGetMap(body)
			if res == false {
				return false, nil, msg
			}
			getSvcResObj.Upstream.Nodes_int = *nodesPtr
			return true, &getSvcResObj, "success"
		} else {
			return false, nil, err.Error()
		}
	}

	//GET请求失败
	return false, nil, string(body)
}

//根据IP反查service
func (svc *GetSvcByIp) GetServiceByIp() (bool, []string, string) {
	var idNodesObj IdNodes
	config := clientv3.Config{
		Endpoints:   svc.Endpoints,
		DialTimeout: 5 * time.Second,
		Username:    svc.Username,
		Password:    svc.Password,
	}
	client, err := clientv3.New(config)
	if err != nil {
		return false, nil, err.Error()
	}
	defer client.Close()

	res, err := client.Get(context.TODO(), "/apisix/services", clientv3.WithPrefix())
	if err != nil {
		return false, nil, err.Error()
	}
	var idList []string
	for _, v := range res.Kvs {
		if string(v.Key) == "/apisix/services/" {
			continue
		}
		if err = json.Unmarshal(v.Value, &idNodesObj); err != nil {
			return false, nil, err.Error()
		}
		nodes := fmt.Sprintf("%v", idNodesObj.Upstream.Nodes)
		for _, ip := range svc.Ips {
			if strings.Contains(nodes, ip) {
				idList = append(idList, idNodesObj.Id)
				break
			}
		}
	}

	return true, idList, ""
}

//查询serivce节点数量
func (svc *GetSvcNodes) GetServiceNods() (bool, *map[string]int, string) {
	var idNodesObj IdNodes
	config := clientv3.Config{
		Endpoints:   svc.Endpoints,
		DialTimeout: 5 * time.Second,
		Username:    svc.Username,
		Password:    svc.Password,
	}
	client, err := clientv3.New(config)
	if err != nil {
		return false, nil, err.Error()
	}
	defer client.Close()

	res, err := client.Get(context.TODO(), "/apisix/services", clientv3.WithPrefix())
	if err != nil {
		return false, nil, err.Error()
	}
	idMap := make(map[string]int)
	for _, v := range svc.Ids {
		idMap[v] = 0
	}
	for _, v := range res.Kvs {
		if string(v.Key) == "/apisix/services/" {
			continue
		}
		if err = json.Unmarshal(v.Value, &idNodesObj); err != nil {
			return false, nil, err.Error()
		}
		nodes := reflect.ValueOf(idNodesObj.Upstream.Nodes)
		if _, ok := idMap[idNodesObj.Id]; ok {
			idMap[idNodesObj.Id] = nodes.Len()
		}
	}

	return true, &idMap, ""
}
