// Code generated manually for testing multicluster-runtime manager. DO NOT EDIT.

package mocks

import (
	"context"
	"net/http"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/mock"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
	"sigs.k8s.io/controller-runtime/pkg/config"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	mcmanager "sigs.k8s.io/multicluster-runtime/pkg/manager"
	"sigs.k8s.io/multicluster-runtime/pkg/multicluster"
)

// MockMCManager is a mock for mcmanager.Manager
type MockMCManager struct {
	mock.Mock
}

type MockMCManager_Expecter struct {
	mock *mock.Mock
}

func (_m *MockMCManager) EXPECT() *MockMCManager_Expecter {
	return &MockMCManager_Expecter{mock: &_m.Mock}
}

func NewMockMCManager(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockMCManager {
	m := &MockMCManager{}
	m.Mock.Test(t)
	t.Cleanup(func() { m.AssertExpectations(t) })
	return m
}

func (_m *MockMCManager) Add(runnable mcmanager.Runnable) error {
	ret := _m.Called(runnable)
	return ret.Error(0)
}

func (_m *MockMCManager) Elected() <-chan struct{} {
	ret := _m.Called()
	return ret.Get(0).(<-chan struct{})
}

func (_m *MockMCManager) AddMetricsServerExtraHandler(path string, handler http.Handler) error {
	ret := _m.Called(path, handler)
	return ret.Error(0)
}

func (_m *MockMCManager) AddHealthzCheck(name string, check healthz.Checker) error {
	ret := _m.Called(name, check)
	return ret.Error(0)
}

func (_m *MockMCManager) AddReadyzCheck(name string, check healthz.Checker) error {
	ret := _m.Called(name, check)
	return ret.Error(0)
}

func (_m *MockMCManager) Start(ctx context.Context) error {
	ret := _m.Called(ctx)
	return ret.Error(0)
}

func (_m *MockMCManager) GetWebhookServer() webhook.Server {
	ret := _m.Called()
	if ret.Get(0) == nil {
		return nil
	}
	return ret.Get(0).(webhook.Server)
}

func (_m *MockMCManager) GetLogger() logr.Logger {
	ret := _m.Called()
	return ret.Get(0).(logr.Logger)
}

func (_m *MockMCManager) GetControllerOptions() config.Controller {
	ret := _m.Called()
	return ret.Get(0).(config.Controller)
}

func (_m *MockMCManager) GetCluster(ctx context.Context, clusterName string) (cluster.Cluster, error) {
	ret := _m.Called(ctx, clusterName)

	var r0 cluster.Cluster
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (cluster.Cluster, error)); ok {
		return rf(ctx, clusterName)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) cluster.Cluster); ok {
		r0 = rf(ctx, clusterName)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(cluster.Cluster)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, clusterName)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetCluster_Call is a helper struct for GetCluster expectations
type MockMCManager_GetCluster_Call struct {
	*mock.Call
}

func (_e *MockMCManager_Expecter) GetCluster(ctx interface{}, clusterName interface{}) *MockMCManager_GetCluster_Call {
	return &MockMCManager_GetCluster_Call{Call: _e.mock.On("GetCluster", ctx, clusterName)}
}

func (_c *MockMCManager_GetCluster_Call) Return(_a0 cluster.Cluster, _a1 error) *MockMCManager_GetCluster_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockMCManager_GetCluster_Call) RunAndReturn(run func(context.Context, string) (cluster.Cluster, error)) *MockMCManager_GetCluster_Call {
	_c.Call.Return(run)
	return _c
}

func (_m *MockMCManager) ClusterFromContext(ctx context.Context) (cluster.Cluster, error) {
	ret := _m.Called(ctx)
	var r0 cluster.Cluster
	var r1 error
	if ret.Get(0) != nil {
		r0 = ret.Get(0).(cluster.Cluster)
	}
	r1 = ret.Error(1)
	return r0, r1
}

func (_m *MockMCManager) GetManager(ctx context.Context, clusterName string) (manager.Manager, error) {
	ret := _m.Called(ctx, clusterName)
	var r0 manager.Manager
	var r1 error
	if ret.Get(0) != nil {
		r0 = ret.Get(0).(manager.Manager)
	}
	r1 = ret.Error(1)
	return r0, r1
}

func (_m *MockMCManager) GetLocalManager() manager.Manager {
	ret := _m.Called()
	if ret.Get(0) == nil {
		return nil
	}
	return ret.Get(0).(manager.Manager)
}

func (_m *MockMCManager) GetProvider() multicluster.Provider {
	ret := _m.Called()
	if ret.Get(0) == nil {
		return nil
	}
	return ret.Get(0).(multicluster.Provider)
}

func (_m *MockMCManager) GetFieldIndexer() client.FieldIndexer {
	ret := _m.Called()
	if ret.Get(0) == nil {
		return nil
	}
	return ret.Get(0).(client.FieldIndexer)
}

// multicluster.Aware methods
func (_m *MockMCManager) Engage(ctx context.Context, clusterName string, cl cluster.Cluster) error {
	ret := _m.Called(ctx, clusterName, cl)
	return ret.Error(0)
}

func (_m *MockMCManager) Disengage(ctx context.Context, clusterName string) error {
	ret := _m.Called(ctx, clusterName)
	return ret.Error(0)
}
