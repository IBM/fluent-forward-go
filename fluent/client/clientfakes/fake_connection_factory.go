// Code generated by counterfeiter. DO NOT EDIT.
package clientfakes

import (
	"net"
	"sync"

	"github.com/IBM/fluent-forward-go/fluent/client"
	"github.com/tinylib/msgp/msgp"
)

type FakeConnectionFactory struct {
	ConnectStub        func() error
	connectMutex       sync.RWMutex
	connectArgsForCall []struct {
	}
	connectReturns struct {
		result1 error
	}
	connectReturnsOnCall map[int]struct {
		result1 error
	}
	DisconnectStub        func() error
	disconnectMutex       sync.RWMutex
	disconnectArgsForCall []struct {
	}
	disconnectReturns struct {
		result1 error
	}
	disconnectReturnsOnCall map[int]struct {
		result1 error
	}
	NewStub        func() (net.Conn, error)
	newMutex       sync.RWMutex
	newArgsForCall []struct {
	}
	newReturns struct {
		result1 net.Conn
		result2 error
	}
	newReturnsOnCall map[int]struct {
		result1 net.Conn
		result2 error
	}
	SendMessageStub        func(msgp.Encodable) error
	sendMessageMutex       sync.RWMutex
	sendMessageArgsForCall []struct {
		arg1 msgp.Encodable
	}
	sendMessageReturns struct {
		result1 error
	}
	sendMessageReturnsOnCall map[int]struct {
		result1 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeConnectionFactory) Connect() error {
	fake.connectMutex.Lock()
	ret, specificReturn := fake.connectReturnsOnCall[len(fake.connectArgsForCall)]
	fake.connectArgsForCall = append(fake.connectArgsForCall, struct {
	}{})
	stub := fake.ConnectStub
	fakeReturns := fake.connectReturns
	fake.recordInvocation("Connect", []interface{}{})
	fake.connectMutex.Unlock()
	if stub != nil {
		return stub()
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeConnectionFactory) ConnectCallCount() int {
	fake.connectMutex.RLock()
	defer fake.connectMutex.RUnlock()
	return len(fake.connectArgsForCall)
}

func (fake *FakeConnectionFactory) ConnectCalls(stub func() error) {
	fake.connectMutex.Lock()
	defer fake.connectMutex.Unlock()
	fake.ConnectStub = stub
}

func (fake *FakeConnectionFactory) ConnectReturns(result1 error) {
	fake.connectMutex.Lock()
	defer fake.connectMutex.Unlock()
	fake.ConnectStub = nil
	fake.connectReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeConnectionFactory) ConnectReturnsOnCall(i int, result1 error) {
	fake.connectMutex.Lock()
	defer fake.connectMutex.Unlock()
	fake.ConnectStub = nil
	if fake.connectReturnsOnCall == nil {
		fake.connectReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.connectReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeConnectionFactory) Disconnect() error {
	fake.disconnectMutex.Lock()
	ret, specificReturn := fake.disconnectReturnsOnCall[len(fake.disconnectArgsForCall)]
	fake.disconnectArgsForCall = append(fake.disconnectArgsForCall, struct {
	}{})
	stub := fake.DisconnectStub
	fakeReturns := fake.disconnectReturns
	fake.recordInvocation("Disconnect", []interface{}{})
	fake.disconnectMutex.Unlock()
	if stub != nil {
		return stub()
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeConnectionFactory) DisconnectCallCount() int {
	fake.disconnectMutex.RLock()
	defer fake.disconnectMutex.RUnlock()
	return len(fake.disconnectArgsForCall)
}

func (fake *FakeConnectionFactory) DisconnectCalls(stub func() error) {
	fake.disconnectMutex.Lock()
	defer fake.disconnectMutex.Unlock()
	fake.DisconnectStub = stub
}

func (fake *FakeConnectionFactory) DisconnectReturns(result1 error) {
	fake.disconnectMutex.Lock()
	defer fake.disconnectMutex.Unlock()
	fake.DisconnectStub = nil
	fake.disconnectReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeConnectionFactory) DisconnectReturnsOnCall(i int, result1 error) {
	fake.disconnectMutex.Lock()
	defer fake.disconnectMutex.Unlock()
	fake.DisconnectStub = nil
	if fake.disconnectReturnsOnCall == nil {
		fake.disconnectReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.disconnectReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeConnectionFactory) New() (net.Conn, error) {
	fake.newMutex.Lock()
	ret, specificReturn := fake.newReturnsOnCall[len(fake.newArgsForCall)]
	fake.newArgsForCall = append(fake.newArgsForCall, struct {
	}{})
	stub := fake.NewStub
	fakeReturns := fake.newReturns
	fake.recordInvocation("New", []interface{}{})
	fake.newMutex.Unlock()
	if stub != nil {
		return stub()
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeConnectionFactory) NewCallCount() int {
	fake.newMutex.RLock()
	defer fake.newMutex.RUnlock()
	return len(fake.newArgsForCall)
}

func (fake *FakeConnectionFactory) NewCalls(stub func() (net.Conn, error)) {
	fake.newMutex.Lock()
	defer fake.newMutex.Unlock()
	fake.NewStub = stub
}

func (fake *FakeConnectionFactory) NewReturns(result1 net.Conn, result2 error) {
	fake.newMutex.Lock()
	defer fake.newMutex.Unlock()
	fake.NewStub = nil
	fake.newReturns = struct {
		result1 net.Conn
		result2 error
	}{result1, result2}
}

func (fake *FakeConnectionFactory) NewReturnsOnCall(i int, result1 net.Conn, result2 error) {
	fake.newMutex.Lock()
	defer fake.newMutex.Unlock()
	fake.NewStub = nil
	if fake.newReturnsOnCall == nil {
		fake.newReturnsOnCall = make(map[int]struct {
			result1 net.Conn
			result2 error
		})
	}
	fake.newReturnsOnCall[i] = struct {
		result1 net.Conn
		result2 error
	}{result1, result2}
}

func (fake *FakeConnectionFactory) SendMessage(arg1 msgp.Encodable) error {
	fake.sendMessageMutex.Lock()
	ret, specificReturn := fake.sendMessageReturnsOnCall[len(fake.sendMessageArgsForCall)]
	fake.sendMessageArgsForCall = append(fake.sendMessageArgsForCall, struct {
		arg1 msgp.Encodable
	}{arg1})
	stub := fake.SendMessageStub
	fakeReturns := fake.sendMessageReturns
	fake.recordInvocation("SendMessage", []interface{}{arg1})
	fake.sendMessageMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeConnectionFactory) SendMessageCallCount() int {
	fake.sendMessageMutex.RLock()
	defer fake.sendMessageMutex.RUnlock()
	return len(fake.sendMessageArgsForCall)
}

func (fake *FakeConnectionFactory) SendMessageCalls(stub func(msgp.Encodable) error) {
	fake.sendMessageMutex.Lock()
	defer fake.sendMessageMutex.Unlock()
	fake.SendMessageStub = stub
}

func (fake *FakeConnectionFactory) SendMessageArgsForCall(i int) msgp.Encodable {
	fake.sendMessageMutex.RLock()
	defer fake.sendMessageMutex.RUnlock()
	argsForCall := fake.sendMessageArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeConnectionFactory) SendMessageReturns(result1 error) {
	fake.sendMessageMutex.Lock()
	defer fake.sendMessageMutex.Unlock()
	fake.SendMessageStub = nil
	fake.sendMessageReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeConnectionFactory) SendMessageReturnsOnCall(i int, result1 error) {
	fake.sendMessageMutex.Lock()
	defer fake.sendMessageMutex.Unlock()
	fake.SendMessageStub = nil
	if fake.sendMessageReturnsOnCall == nil {
		fake.sendMessageReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.sendMessageReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeConnectionFactory) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.connectMutex.RLock()
	defer fake.connectMutex.RUnlock()
	fake.disconnectMutex.RLock()
	defer fake.disconnectMutex.RUnlock()
	fake.newMutex.RLock()
	defer fake.newMutex.RUnlock()
	fake.sendMessageMutex.RLock()
	defer fake.sendMessageMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeConnectionFactory) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}

var _ client.ConnectionFactory = new(FakeConnectionFactory)
