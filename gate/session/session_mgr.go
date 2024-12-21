package session

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/eleren/newtbig/common"
	log "github.com/eleren/newtbig/logging"
	"github.com/eleren/newtbig/msg/framepb"
	"github.com/eleren/newtbig/utils"
)

type SessionMgr struct {
	sync.Map
	count   int32
	addChan chan *Session
	addMap  map[uint64]int64
}

var sessionMgrSingleton *SessionMgr = nil
var once sync.Once

func GetSessionMgr() *SessionMgr {
	once.Do(func() {
		sessionMgrSingleton = new(SessionMgr)
		sessionMgrSingleton.addChan = make(chan *Session, 9216)
		sessionMgrSingleton.addMap = make(map[uint64]int64)
		utils.SafeGO(func() {
			sessionMgrSingleton.run()
		})
	})
	return sessionMgrSingleton
}

func (s *SessionMgr) run() {
	for ss := range s.addChan {
		s.add(ss)
	}
}

func (sm *SessionMgr) AddSession(session *Session) bool {
	if session.GetUID() == 0 {
		return false
	}
	sm.addChan <- session
	return true
}

func (sm *SessionMgr) add(session *Session) bool {
	lastTime := sm.addMap[session.GetUID()]
	if lastTime != 0 && time.Now().UnixMilli()-lastTime < 3000 {
		log.Logger.Warnf("SessionMgr AddSession too much :%d", session.GetUID())
		session.OnStop()
		return false
	}

	old := sm.GetSession(session.GetUID())
	if old != nil {
		sm.Delete(session.GetUID())
		err := old.OnStop()
		if err != nil {
			log.Logger.Error("SessionMgr AddSession err :", err.Error())
		}
	} else {
		atomic.AddInt32(&sm.count, 1)
	}
	session.SendSign(common.Msg_Verify_Rst_Suc)
	_, loaded := sm.LoadOrStore(session.GetUID(), session)
	sm.addMap[session.GetUID()] = time.Now().UnixMilli()
	return !loaded
}

func (sm *SessionMgr) StopSession(uid uint64) {
	if uid == 0 {
		return
	}
	ss := sm.GetSession(uid)
	if ss != nil {
		err := ss.OnStop()
		if err != nil {
			log.Logger.Error(err.Error())
		}
	}
}

func (sm *SessionMgr) DelSession(uid uint64) {
	if uid == 0 {
		return
	}
	ss := sm.GetSession(uid)
	if ss != nil {
		sm.Delete(uid)
		atomic.AddInt32(&sm.count, -1)
	}
}

func (sm *SessionMgr) GetSession(uid uint64) *Session {
	ss, ok := sm.Load(uid)
	if ok {
		ret, _ := ss.(*Session)
		return ret
	} else {
		return nil
	}
}
func (sm *SessionMgr) StopAllSession() {
	sm.Range(func(key, value interface{}) bool {
		ss, ok := value.(*Session)
		if ok && ss != nil {
			ss.OnStop()
			atomic.AddInt32(&sm.count, -1)
			sm.Delete(key)
			log.Logger.Infof("SessionMgr closeSession key:%d", key)
		}
		return true
	})
}

func (sm *SessionMgr) Count() int32 {
	return sm.count
}

func (sm *SessionMgr) BroatCast(msg *framepb.Msg) error {
	sm.Range(func(key, value interface{}) bool {
		if ss, ok := value.(*Session); ok {
			err := ss.Broadcast(msg)
			if err != nil {
				log.Logger.Warnf("broatcast uid :%d err :%s", key.(uint64), err.Error())
			}
		}
		return true
	})
	return nil
}

func (sm *SessionMgr) Receive(msg *framepb.Msg) error {
	ss := sm.GetSession(msg.UID)
	if ss == nil {
		log.Logger.Warnf("not found session:%d  req:%d msgid:%d", msg.UID, msg.Seq, msg.ID)
		return nil
	}
	return ss.RpcMsg(msg)

}
