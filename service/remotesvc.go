package service

import (
	"github.com/davyxu/cellmesh/discovery"
	"github.com/davyxu/cellnet"
	"sync"
)

type RemoteServiceContext struct {
	Name  string
	SvcID string
}

var (
	connBySvcID        = map[string]cellnet.Session{}
	connBySvcNameGuard sync.RWMutex
)

func AddRemoteService(ses cellnet.Session, svcid, name string) {

	connBySvcNameGuard.Lock()
	ses.(cellnet.ContextSet).SetContext("ctx", &RemoteServiceContext{Name: name, SvcID: svcid})
	connBySvcID[svcid] = ses
	connBySvcNameGuard.Unlock()

	log.SetColor("green").Debugf("remote service added: '%s'", svcid)
}

func SessionToContext(ses cellnet.Session) *RemoteServiceContext {
	if raw, ok := ses.(cellnet.ContextSet).GetContext("ctx"); ok {
		return raw.(*RemoteServiceContext)
	}

	return nil
}

func RemoveRemoteService(ses cellnet.Session) {

	desc := SessionToContext(ses)
	if desc != nil {

		connBySvcNameGuard.Lock()
		delete(connBySvcID, desc.SvcID)
		connBySvcNameGuard.Unlock()

		log.SetColor("yellow").Debugf("remote service removed '%s'", desc.SvcID)
	}
}

// 根据id获取远程服务的会话
func GetRemoteService(svcid string) cellnet.Session {
	connBySvcNameGuard.RLock()
	defer connBySvcNameGuard.RUnlock()

	if ses, ok := connBySvcID[svcid]; ok {

		return ses
	}

	return nil
}

// 遍历远程服务
func VisitRemoteService(callback func(ses cellnet.Session, ctx *RemoteServiceContext) bool) {
	connBySvcNameGuard.RLock()

	for _, ses := range connBySvcID {

		if !callback(ses, SessionToContext(ses)) {
			break
		}
	}

	connBySvcNameGuard.RUnlock()
}

// 根据远程服务的session，获取服务发现信息
func GetRemoteServiceDesc(ses cellnet.Session, matchSvcGroup string) *discovery.ServiceDesc {
	ctx := SessionToContext(ses)

	if ctx == nil {
		return nil
	}

	for _, sd := range DiscoveryService(LinkRules, ctx.Name, matchSvcGroup) {

		if sd.ID == ctx.SvcID {
			return sd
		}

	}

	return nil
}

// 获得一个远程服务会话的外网地址
func GetRemoteServiceWANAddress(ses cellnet.Session, matchSvcGroup string) string {

	sd := GetRemoteServiceDesc(ses, matchSvcGroup)
	if sd != nil {
		return sd.GetMeta("WANAddress")
	}

	return ""
}
