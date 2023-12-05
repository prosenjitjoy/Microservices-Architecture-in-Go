package memory

import (
	"context"
	"errors"
	"main/discovery"
	"sync"
	"time"
)

type ServiceName string
type InstanceID string

// Registry defines an in-momory service registry.
type Registry struct {
	sync.RWMutex
	serviceAddrs map[ServiceName]map[InstanceID]*serviceInstance
}

type serviceInstance struct {
	hostPort   string
	lastActive time.Time
}

// NewRegistry creates a new in-memory service registry instance.
func NewRegistry() *Registry {
	return &Registry{
		serviceAddrs: map[ServiceName]map[InstanceID]*serviceInstance{},
	}
}

// Register creates a service record in the registry.
func (r *Registry) Register(ctx context.Context, instanceId string, serviceName string, hostPort string) error {
	r.Lock()
	defer r.Unlock()

	if _, ok := r.serviceAddrs[ServiceName(serviceName)]; !ok {
		r.serviceAddrs[ServiceName(serviceName)] = map[InstanceID]*serviceInstance{}
	}
	r.serviceAddrs[ServiceName(serviceName)][InstanceID(instanceId)] = &serviceInstance{
		hostPort:   hostPort,
		lastActive: time.Now(),
	}
	return nil
}

// Deregister removes a service record from the registry.
func (r *Registry) Deregister(ctx context.Context, instanceId string, serviceName string) error {
	r.Lock()
	defer r.Unlock()

	if _, ok := r.serviceAddrs[ServiceName(serviceName)]; !ok {
		return nil
	}
	delete(r.serviceAddrs[ServiceName(serviceName)], InstanceID(instanceId))
	return nil
}

// ReportHealthyState is a push mechanism for reporting healthy state to the registry.
func (r *Registry) ReportHealthyState(instanceId string, serviceName string) error {
	r.Lock()
	defer r.Unlock()

	if _, ok := r.serviceAddrs[ServiceName(serviceName)]; !ok {
		return errors.New("service is not registered yet")
	}
	if _, ok := r.serviceAddrs[ServiceName(serviceName)][InstanceID(instanceId)]; !ok {
		return errors.New("service instance is not registered yet")
	}
	r.serviceAddrs[ServiceName(serviceName)][InstanceID(instanceId)].lastActive = time.Now()
	return nil
}

// ServiceAddresses returns the list of addresses of active instances of the given service.
func (r *Registry) ServiceAddresses(ctx context.Context, serviceName string) ([]string, error) {
	r.RLock()
	defer r.RUnlock()

	if len(r.serviceAddrs[ServiceName(serviceName)]) == 0 {
		return nil, discovery.ErrNotFound
	}
	var res []string
	for _, i := range r.serviceAddrs[ServiceName(serviceName)] {
		if !i.lastActive.Before(time.Now().Add(-5 * time.Second)) {
			res = append(res, i.hostPort)
		}
	}
	return res, nil
}
