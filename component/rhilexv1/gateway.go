// Copyright (C) 2025 wwhai
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package rhilex

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

// Gateway
type Gateway struct {
	ctx         context.Context
	cancel      context.CancelFunc
	config      RhilexConfig
	inits       map[string]func() error
	northerns   *GenericResourceManager
	southerns   *GenericResourceManager
	plugins     *GenericResourceManager
	natives     *GenericResourceManager
	queue       *GenericMessageQueue
	broker      *Broker
	cronManager *CronManager
	cache       *GatewayInternalCache
	logger      *logrus.Logger
}

// NewGateway
func NewGateway(logger *logrus.Logger) *Gateway {
	gateway := new(Gateway)
	gateway.logger = logger
	gateway.inits = make(map[string]func() error)
	gateway.config = RhilexConfig{
		AppId:         "rhilex",
		MaxQueueSize:  1024,
		GomaxProcs:    4,
		EnablePProf:   false,
		DebugMode:     false,
		LogLevel:      "info",
		LogMaxSize:    10,
		LogMaxBackups: 5,
		LogMaxAge:     30,
	}
	gateway.northerns = NewGenericResourceManager(gateway)
	gateway.southerns = NewGenericResourceManager(gateway)
	gateway.plugins = NewGenericResourceManager(gateway)
	gateway.natives = NewGenericResourceManager(gateway)
	gateway.cache = NewGatewayInternalCache(5 * time.Second)
	gateway.queue = NewGenericMessageQueue(1024)
	gateway.cronManager = NewCronManager("./")
	gateway.broker = NewBroker(1024)
	return gateway
}

// CallInitFuncByName calls a specific initialization function by name
func (g *Gateway) Start(ctx context.Context, cancel context.CancelFunc, config RhilexConfig) error {
	g.ctx = ctx
	g.cancel = cancel
	g.config = config
	g.CallInitFunc()
	g.northerns.StartMonitoring()
	g.southerns.StartMonitoring()
	g.plugins.StartMonitoring()
	g.natives.StartMonitoring()
	return nil
}

// Stop stops the gateway and all resource managers
func (g *Gateway) Stop() error {
	g.northerns.StopMonitoring()
	g.southerns.StopMonitoring()
	g.plugins.StopMonitoring()
	g.natives.StopMonitoring()
	g.queue.Destroy()
	g.cache.StopCleanup()
	g.broker.Close()
	g.cronManager.Stop()
	g.cancel()
	return nil
}

// RegisterInitFunc registers an initialization function
func (g *Gateway) RegisterInitFunc(name string, fn func() error) {
	if _, exists := g.inits[name]; exists {
		g.logger.Errorf("Init function %s already registered", name)
		return
	}
	g.logger.Infof("Registering init function %s", name)
	// Check if the function is nil
	if fn == nil {
		g.logger.Errorf("Init function %s is nil", name)
		return
	}
	g.inits[name] = fn
}

// CallInitFunc calls all registered initialization functions
func (g *Gateway) CallInitFunc() {
	for name, fn := range g.inits {
		if err := fn(); err != nil {
			g.logger.Errorf("Failed to execute init function %s: %v", name, err)
		} else {
			g.logger.Infof("Successfully executed init function %s", name)
		}
	}
}

// GetManager retrieves a specific resource manager by name
func (g *Gateway) GetManager(managerName string) (*GenericResourceManager, error) {
	switch managerName {
	case "northerns":
		return g.northerns, nil
	case "southerns":
		return g.southerns, nil
	case "plugins":
		return g.plugins, nil
	case "natives":
		return g.natives, nil
	default:
		return nil, fmt.Errorf("resource manager %s not found", managerName)
	}
}

// LogResourceStatus logs the status of all resources in all managers
func (g *Gateway) LogResourceStatus() {
	g.logger.Info("Logging resource statuses for all managers...")

	logManagerStatus := func(managerName string, manager *GenericResourceManager) {
		resources := manager.GetResourceList()
		g.logger.Infof("Manager: %s, Total Resources: %d", managerName, len(resources))
		for _, resource := range resources {
			g.logger.Infof("Resource UUID: %s, Status: %v", resource.Worker().UUID, resource.Status())
		}
	}

	logManagerStatus("northerns", g.northerns)
	logManagerStatus("southerns", g.southerns)
	logManagerStatus("plugins", g.plugins)
	logManagerStatus("natives", g.natives)

	g.logger.Info("Finished logging resource statuses.")
}

// ReloadAllManagers reloads all resources in all managers
func (g *Gateway) ReloadAllManagers() {
	g.logger.Info("Reloading all resources in all managers...")

	reloadManager := func(managerName string, manager *GenericResourceManager) {
		resources := manager.GetResourceList()
		for _, resource := range resources {
			if err := manager.ReloadResource(resource.Worker().UUID); err != nil {
				g.logger.Errorf("Failed to reload resource in %s: %v", managerName, err)
			} else {
				g.logger.Infof("Successfully reloaded resource in %s: %s", managerName, resource.Worker().UUID)
			}
		}
	}

	reloadManager("northerns", g.northerns)
	reloadManager("southerns", g.southerns)
	reloadManager("plugins", g.plugins)
	reloadManager("natives", g.natives)

	g.logger.Info("Finished reloading all resources.")
}

// LoadSouthernResource loads a resource into the Southern resource manager
func (g *Gateway) LoadSouthernResource(resourceType, uuid string, config map[string]any) error {
	g.logger.Infof("Loading resource %s into the Southern resource manager...", uuid)
	if err := g.southerns.CreateResource(resourceType, uuid, config); err != nil {
		g.logger.Errorf("Failed to load resource %s into the Southern resource manager: %v", uuid, err)
		return err
	}
	g.logger.Infof("Successfully loaded resource %s into the Southern resource manager.", uuid)
	return nil
}

// LoadNorthernResource loads a resource into the Northern resource manager
func (g *Gateway) LoadNorthernResource(resourceType, uuid string, config map[string]any) error {
	g.logger.Infof("Loading resource %s into the Northern resource manager...", uuid)
	if err := g.northerns.CreateResource(resourceType, uuid, config); err != nil {
		g.logger.Errorf("Failed to load resource %s into the Northern resource manager: %v", uuid, err)
		return err
	}
	g.logger.Infof("Successfully loaded resource %s into the Northern resource manager.", uuid)
	return nil
}

// LoadPluginResource loads a resource into the Plugin resource manager
func (g *Gateway) LoadPluginResource(resourceType, uuid string, config map[string]any) error {
	g.logger.Infof("Loading resource %s into the Plugin resource manager...", uuid)
	if err := g.plugins.CreateResource(resourceType, uuid, config); err != nil {
		g.logger.Errorf("Failed to load resource %s into the Plugin resource manager: %v", uuid, err)
		return err
	}
	g.logger.Infof("Successfully loaded resource %s into the Plugin resource manager.", uuid)
	return nil
}

// LoadNativeResource loads a resource into the Native resource manager
func (g *Gateway) LoadNativeResource(resourceType, uuid string, config map[string]any) error {
	g.logger.Infof("Loading resource %s into the Native resource manager...", uuid)
	if err := g.natives.CreateResource(resourceType, uuid, config); err != nil {
		g.logger.Errorf("Failed to load resource %s into the Native resource manager: %v", uuid, err)
		return err
	}
	g.logger.Infof("Successfully loaded resource %s into the Native resource manager.", uuid)
	return nil
}

// RemoveSouthernResource removes a resource from the Southern resource manager
func (g *Gateway) RemoveSouthernResource(uuid string) error {
	g.logger.Infof("Removing resource %s from the Southern resource manager...", uuid)
	if err := g.southerns.StopResource(uuid); err != nil {
		g.logger.Errorf("Failed to remove resource %s from the Southern resource manager: %v", uuid, err)
		return err
	}
	g.logger.Infof("Successfully removed resource %s from the Southern resource manager.", uuid)
	return nil
}

// RemoveNorthernResource removes a resource from the Northern resource manager
func (g *Gateway) RemoveNorthernResource(uuid string) error {
	g.logger.Infof("Removing resource %s from the Northern resource manager...", uuid)
	if err := g.northerns.StopResource(uuid); err != nil {
		g.logger.Errorf("Failed to remove resource %s from the Northern resource manager: %v", uuid, err)
		return err
	}
	g.logger.Infof("Successfully removed resource %s from the Northern resource manager.", uuid)
	return nil
}

// RemovePluginResource removes a resource from the Plugin resource manager
func (g *Gateway) RemovePluginResource(uuid string) error {
	g.logger.Infof("Removing resource %s from the Plugin resource manager...", uuid)
	if err := g.plugins.StopResource(uuid); err != nil {
		g.logger.Errorf("Failed to remove resource %s from the Plugin resource manager: %v", uuid, err)
		return err
	}
	g.logger.Infof("Successfully removed resource %s from the Plugin resource manager.", uuid)
	return nil
}

// RemoveNativeResource removes a resource from the Native resource manager
func (g *Gateway) RemoveNativeResource(uuid string) error {
	g.logger.Infof("Removing resource %s from the Native resource manager...", uuid)
	if err := g.natives.StopResource(uuid); err != nil {
		g.logger.Errorf("Failed to remove resource %s from the Native resource manager: %v", uuid, err)
		return err
	}
	g.logger.Infof("Successfully removed resource %s from the Native resource manager.", uuid)
	return nil
}

// GetSouthernResource retrieves a resource from the Southern resource manager by UUID
func (g *Gateway) GetSouthernResource(uuid string) (GenericResource, error) {
	g.logger.Infof("Retrieving resource %s from the Southern resource manager...", uuid)
	resource, err := g.southerns.GetResource(uuid)
	if err != nil {
		g.logger.Errorf("Failed to retrieve resource %s from the Southern resource manager: %v", uuid, err)
		return nil, err
	}
	g.logger.Infof("Successfully retrieved resource %s from the Southern resource manager.", uuid)
	return resource, nil
}

// GetNorthernResource retrieves a resource from the Northern resource manager by UUID
func (g *Gateway) GetNorthernResource(uuid string) (GenericResource, error) {
	g.logger.Infof("Retrieving resource %s from the Northern resource manager...", uuid)
	resource, err := g.northerns.GetResource(uuid)
	if err != nil {
		g.logger.Errorf("Failed to retrieve resource %s from the Northern resource manager: %v", uuid, err)
		return nil, err
	}
	g.logger.Infof("Successfully retrieved resource %s from the Northern resource manager.", uuid)
	return resource, nil
}

// GetPluginResource retrieves a resource from the Plugin resource manager by UUID
func (g *Gateway) GetPluginResource(uuid string) (GenericResource, error) {
	g.logger.Infof("Retrieving resource %s from the Plugin resource manager...", uuid)
	resource, err := g.plugins.GetResource(uuid)
	if err != nil {
		g.logger.Errorf("Failed to retrieve resource %s from the Plugin resource manager: %v", uuid, err)
		return nil, err
	}
	g.logger.Infof("Successfully retrieved resource %s from the Plugin resource manager.", uuid)
	return resource, nil
}

// GetNativeResource retrieves a resource from the Native resource manager by UUID
func (g *Gateway) GetNativeResource(uuid string) (GenericResource, error) {
	g.logger.Infof("Retrieving resource %s from the Native resource manager...", uuid)
	resource, err := g.natives.GetResource(uuid)
	if err != nil {
		g.logger.Errorf("Failed to retrieve resource %s from the Native resource manager: %v", uuid, err)
		return nil, err
	}
	g.logger.Infof("Successfully retrieved resource %s from the Native resource manager.", uuid)
	return resource, nil
}

// GetSouthernServices retrieves the services of a resource from the Southern resource manager by UUID
func (g *Gateway) GetSouthernServices(uuid string) ([]ResourceService, error) {
	g.logger.Infof("Retrieving services for resource %s from the Southern resource manager...", uuid)
	resource, err := g.southerns.GetResource(uuid)
	if err != nil {
		g.logger.Errorf("Failed to retrieve resource %s from the Southern resource manager: %v", uuid, err)
		return nil, err
	}
	services := resource.Services()
	g.logger.Infof("Successfully retrieved services for resource %s from the Southern resource manager.", uuid)
	return services, nil
}

// GetNorthernServices retrieves the services of a resource from the Northern resource manager by UUID
func (g *Gateway) GetNorthernServices(uuid string) ([]ResourceService, error) {
	g.logger.Infof("Retrieving services for resource %s from the Northern resource manager...", uuid)
	resource, err := g.northerns.GetResource(uuid)
	if err != nil {
		g.logger.Errorf("Failed to retrieve resource %s from the Northern resource manager: %v", uuid, err)
		return nil, err
	}
	services := resource.Services()
	g.logger.Infof("Successfully retrieved services for resource %s from the Northern resource manager.", uuid)
	return services, nil
}

// GetPluginServices retrieves the services of a resource from the Plugin resource manager by UUID
func (g *Gateway) GetPluginServices(uuid string) ([]ResourceService, error) {
	g.logger.Infof("Retrieving services for resource %s from the Plugin resource manager...", uuid)
	resource, err := g.plugins.GetResource(uuid)
	if err != nil {
		g.logger.Errorf("Failed to retrieve resource %s from the Plugin resource manager: %v", uuid, err)
		return nil, err
	}
	services := resource.Services()
	g.logger.Infof("Successfully retrieved services for resource %s from the Plugin resource manager.", uuid)
	return services, nil
}

// GetNativeServices retrieves the services of a resource from the Native resource manager by UUID
func (g *Gateway) GetNativeServices(uuid string) ([]ResourceService, error) {
	g.logger.Infof("Retrieving services for resource %s from the Native resource manager...", uuid)
	resource, err := g.natives.GetResource(uuid)
	if err != nil {
		g.logger.Errorf("Failed to retrieve resource %s from the Native resource manager: %v", uuid, err)
		return nil, err
	}
	services := resource.Services()
	g.logger.Infof("Successfully retrieved services for resource %s from the Native resource manager.", uuid)
	return services, nil
}

// CallSouthernOnService invokes a service on a resource in the Southern resource manager by UUID
func (g *Gateway) CallSouthernOnService(uuid string, request ResourceServiceRequest) (ResourceServiceResponse, error) {
	g.logger.Infof("Invoking service %s on resource %s in the Southern resource manager...", request.Name, uuid)
	resource, err := g.southerns.GetResource(uuid)
	if err != nil {
		g.logger.Errorf("Failed to retrieve resource %s from the Southern resource manager: %v", uuid, err)
		return ResourceServiceResponse{}, err
	}
	response, err := resource.OnService(request)
	if err != nil {
		g.logger.Errorf("Failed to invoke service %s on resource %s in the Southern resource manager: %v", request.Name, uuid, err)
		return ResourceServiceResponse{}, err
	}
	g.logger.Infof("Successfully invoked service %s on resource %s in the Southern resource manager.", request.Name, uuid)
	return response, nil
}

// CallNorthernOnService invokes a service on a resource in the Northern resource manager by UUID
func (g *Gateway) CallNorthernOnService(uuid string, request ResourceServiceRequest) (ResourceServiceResponse, error) {
	g.logger.Infof("Invoking service %s on resource %s in the Northern resource manager...", request.Name, uuid)
	resource, err := g.northerns.GetResource(uuid)
	if err != nil {
		g.logger.Errorf("Failed to retrieve resource %s from the Northern resource manager: %v", uuid, err)
		return ResourceServiceResponse{}, err
	}
	response, err := resource.OnService(request)
	if err != nil {
		g.logger.Errorf("Failed to invoke service %s on resource %s in the Northern resource manager: %v", request.Name, uuid, err)
		return ResourceServiceResponse{}, err
	}
	g.logger.Infof("Successfully invoked service %s on resource %s in the Northern resource manager.", request.Name, uuid)
	return response, nil
}

// CallPluginOnService invokes a service on a resource in the Plugin resource manager by UUID
func (g *Gateway) CallPluginOnService(uuid string, request ResourceServiceRequest) (ResourceServiceResponse, error) {
	g.logger.Infof("Invoking service %s on resource %s in the Plugin resource manager...", request.Name, uuid)
	resource, err := g.plugins.GetResource(uuid)
	if err != nil {
		g.logger.Errorf("Failed to retrieve resource %s from the Plugin resource manager: %v", uuid, err)
		return ResourceServiceResponse{}, err
	}
	response, err := resource.OnService(request)
	if err != nil {
		g.logger.Errorf("Failed to invoke service %s on resource %s in the Plugin resource manager: %v", request.Name, uuid, err)
		return ResourceServiceResponse{}, err
	}
	g.logger.Infof("Successfully invoked service %s on resource %s in the Plugin resource manager.", request.Name, uuid)
	return response, nil
}

// CallNativeOnService invokes a service on a resource in the Native resource manager by UUID
func (g *Gateway) CallNativeOnService(uuid string, request ResourceServiceRequest) (ResourceServiceResponse, error) {
	g.logger.Infof("Invoking service %s on resource %s in the Native resource manager...", request.Name, uuid)
	resource, err := g.natives.GetResource(uuid)
	if err != nil {
		g.logger.Errorf("Failed to retrieve resource %s from the Native resource manager: %v", uuid, err)
		return ResourceServiceResponse{}, err
	}
	response, err := resource.OnService(request)
	if err != nil {
		g.logger.Errorf("Failed to invoke service %s on resource %s in the Native resource manager: %v", request.Name, uuid, err)
		return ResourceServiceResponse{}, err
	}
	g.logger.Infof("Successfully invoked service %s on resource %s in the Native resource manager.", request.Name, uuid)
	return response, nil
}

// GatewaySnapshot takes a snapshot of the current state of all resources in all resource managers
func (g *Gateway) GatewaySnapshot() map[string]any {
	g.logger.Info("Taking a snapshot of the current state of all resources in all managers...")

	snapshot := make(map[string]any)

	// Helper function to take a snapshot of a specific manager
	takeManagerSnapshot := func(manager *GenericResourceManager) map[string]any {
		managerSnapshot := make(map[string]any)
		resources := manager.GetResourceList()
		for _, resource := range resources {
			managerSnapshot[resource.Worker().UUID] = map[string]any{
				"Status":   resource.Status(),
				"Services": resource.Services(),
				"Worker": map[string]any{
					"Config": resource.Worker().GetConfig(),
					"Type":   resource.Worker().Type,
				},
			}
		}
		return managerSnapshot
	}

	// Take snapshots of all managers
	snapshot["northerns"] = takeManagerSnapshot(g.northerns)
	snapshot["southerns"] = takeManagerSnapshot(g.southerns)
	snapshot["plugins"] = takeManagerSnapshot(g.plugins)
	snapshot["natives"] = takeManagerSnapshot(g.natives)

	// Add a timestamp to the snapshot
	snapshot["Timestamp"] = time.Now().Format(time.RFC3339)

	g.logger.Info("Successfully took a snapshot of all resources.")
	return snapshot
}

// GetLogger retrieves the logger for the gateway
func (g *Gateway) GetLogger() *logrus.Logger {
	return g.logger
}

// SetLogger sets the logger for the gateway
func (g *Gateway) SetLogger(logger *logrus.Logger) {
	g.logger = logger
}

// GetNorthernResourceManager retrieves the Northern resource manager
func (g *Gateway) GetNorthernResourceManager() *GenericResourceManager {
	return g.northerns
}

// GetSouthernResourceManager retrieves the Southern resource manager
func (g *Gateway) GetSouthernResourceManager() *GenericResourceManager {
	return g.southerns
}

// GetPluginResourceManager retrieves the Plugin resource manager
func (g *Gateway) GetPluginResourceManager() *GenericResourceManager {
	return g.plugins
}

// GetNativeResourceManager retrieves the Native resource manager
func (g *Gateway) GetNativeResourceManager() *GenericResourceManager {
	return g.natives
}

// GetQueue retrieves the message queue for the gateway
func (g *Gateway) GetQueue() *GenericMessageQueue {
	return g.queue
}

// GetCache retrieves the internal cache for the gateway
func (g *Gateway) GetCache() *GatewayInternalCache {
	return g.cache
}

// GetGateway retrieves the gateway instance
func (g *Gateway) GetGateway() *Gateway {
	return g
}

// GetBroker retrieves the broker for the gateway
func (g *Gateway) GetBroker() *Broker {
	return g.broker
}

// GetCronManager retrieves the cron manager for the gateway
func (g *Gateway) GetCronManager() *CronManager {
	return g.cronManager
}

// GetInits retrieves the initialization functions for the gateway
func (g *Gateway) GetInits() map[string]func() error {
	return g.inits
}

// GetInitFunc retrieves a specific initialization function by name
func (g *Gateway) GetInitFunc(name string) (func() error, error) {
	if fn, exists := g.inits[name]; exists {
		return fn, nil
	}
	return nil, fmt.Errorf("init function %s not found", name)
}

// SetInitFunc sets a specific initialization function by name
func (g *Gateway) SetInitFunc(name string, fn func() error) {
	if _, exists := g.inits[name]; exists {
		g.logger.Errorf("Init function %s already registered", name)
		return
	}
	g.logger.Infof("Setting init function %s", name)
	// Check if the function is nil
	if fn == nil {
		g.logger.Errorf("Init function %s is nil", name)
		return
	}
	g.inits[name] = fn
}

// GetAllInitFunc retrieves all initialization functions
func (g *Gateway) GetAllInitFunc() map[string]func() error {
	return g.inits
}
