package memeory

import (
	"container/list"
	"github.com/wojh217/learn_go_web_session/session"
	"sync"
	"time"
)

func init() {
	memoryProvider.sessions = make(map[string]*list.Element, 0)
	session.Register("memory", memoryProvider)
}

type MemoryProvider struct {
	lock     sync.Mutex
	sessions map[string]*list.Element
	list     *list.List
}

// SessionInit 根据sessionId生成session
func (m *MemoryProvider) SessionInit(sid string) (session.Session, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	v := make(map[interface{}]interface{}, 0)
	newsess := &SessionStore{sid: sid, timeAccessed: time.Now(), value: v}
	element := m.list.PushFront(newsess)
	m.sessions[sid] = element
	return newsess, nil
}

func (m *MemoryProvider) SessionRead(sid string) (session.Session, error) {
	if element, ok := m.sessions[sid]; ok {
		return element.Value.(*SessionStore), nil
	} else {
		sess, err := m.SessionInit(sid)
		return sess, err
	}
}

func (m *MemoryProvider) SessionDestroy(sid string) error {
	if element, ok := m.sessions[sid]; ok {
		delete(m.sessions, sid) // 不加锁吗？
		m.list.Remove(element)
		return nil
	}
	return nil
}

func (m *MemoryProvider) SessionGC(maxLifeTime int64) {
	m.lock.Lock()
	defer m.lock.Unlock()

	for {
		element := m.list.Back()
		if element == nil {
			break
		}
		if element.Value.(*SessionStore).timeAccessed.Unix()+maxLifeTime < time.Now().Unix() {
			m.list.Remove(element)
			delete(m.sessions, element.Value.(*SessionStore).sid)
		} else {
			break
		}
	}
}
func (m *MemoryProvider) SessionUpdate(sid string) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	if element, ok := m.sessions[sid]; ok {
		element.Value.(*SessionStore).timeAccessed = time.Now()
		m.list.MoveToFront(element)
		return nil
	}
	return nil
}

var memoryProvider = &MemoryProvider{list: list.New()}

type SessionStore struct {
	sid          string
	timeAccessed time.Time
	value        map[interface{}]interface{}
}

func (st *SessionStore) Set(key, value interface{}) error {
	st.value[key] = value
	memoryProvider.SessionUpdate(st.sid)
	return nil
}

func (st *SessionStore) Get(key interface{}) interface{} {
	memoryProvider.SessionUpdate(st.sid)
	if v, ok := st.value[key]; ok {
		return v
	} else {
		return nil
	}
}

func (st *SessionStore) Delete(key interface{}) error {
	delete(st.value, key)
	memoryProvider.SessionUpdate(st.sid)
	return nil
}

func (st *SessionStore) SessionID() string {
	return st.sid
}
