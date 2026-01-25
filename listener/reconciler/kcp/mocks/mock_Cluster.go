// Code generated manually for testing cluster.Cluster interface. DO NOT EDIT.

package mocks

import (
	"context"
	"net/http"

	"github.com/stretchr/testify/mock"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// MockCluster is a mock for cluster.Cluster
type MockCluster struct {
	mock.Mock
}

type MockCluster_Expecter struct {
	mock *mock.Mock
}

func (_m *MockCluster) EXPECT() *MockCluster_Expecter {
	return &MockCluster_Expecter{mock: &_m.Mock}
}

func NewMockCluster(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockCluster {
	m := &MockCluster{}
	m.Mock.Test(t)
	t.Cleanup(func() { m.AssertExpectations(t) })
	return m
}

func (_m *MockCluster) GetHTTPClient() *http.Client {
	ret := _m.Called()
	if ret.Get(0) == nil {
		return nil
	}
	return ret.Get(0).(*http.Client)
}

func (_m *MockCluster) GetConfig() *rest.Config {
	ret := _m.Called()
	if ret.Get(0) == nil {
		return nil
	}
	return ret.Get(0).(*rest.Config)
}

func (_m *MockCluster) GetCache() cache.Cache {
	ret := _m.Called()
	if ret.Get(0) == nil {
		return nil
	}
	return ret.Get(0).(cache.Cache)
}

func (_m *MockCluster) GetScheme() *runtime.Scheme {
	ret := _m.Called()
	if ret.Get(0) == nil {
		return nil
	}
	return ret.Get(0).(*runtime.Scheme)
}

func (_m *MockCluster) GetClient() client.Client {
	ret := _m.Called()
	if ret.Get(0) == nil {
		return nil
	}
	return ret.Get(0).(client.Client)
}

// GetClient_Call is a helper struct for GetClient expectations
type MockCluster_GetClient_Call struct {
	*mock.Call
}

func (_e *MockCluster_Expecter) GetClient() *MockCluster_GetClient_Call {
	return &MockCluster_GetClient_Call{Call: _e.mock.On("GetClient")}
}

func (_c *MockCluster_GetClient_Call) Return(_a0 client.Client) *MockCluster_GetClient_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_m *MockCluster) GetFieldIndexer() client.FieldIndexer {
	ret := _m.Called()
	if ret.Get(0) == nil {
		return nil
	}
	return ret.Get(0).(client.FieldIndexer)
}

func (_m *MockCluster) GetEventRecorderFor(name string) record.EventRecorder {
	ret := _m.Called(name)
	if ret.Get(0) == nil {
		return nil
	}
	return ret.Get(0).(record.EventRecorder)
}

func (_m *MockCluster) GetRESTMapper() meta.RESTMapper {
	ret := _m.Called()
	if ret.Get(0) == nil {
		return nil
	}
	return ret.Get(0).(meta.RESTMapper)
}

func (_m *MockCluster) GetAPIReader() client.Reader {
	ret := _m.Called()
	if ret.Get(0) == nil {
		return nil
	}
	return ret.Get(0).(client.Reader)
}

func (_m *MockCluster) Start(ctx context.Context) error {
	ret := _m.Called(ctx)
	return ret.Error(0)
}
