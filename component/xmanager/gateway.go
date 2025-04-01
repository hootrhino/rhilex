package xmanager

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

// Gateway
type Gateway struct {
	Northerns *GenericResourceManager
	Southerns *GenericResourceManager
	Plugins   *GenericResourceManager
	Natives   *GenericResourceManager
	Logger    *logrus.Logger
}

// NewGateway
func NewGateway(logger *logrus.Logger) *Gateway {
	gateway := new(Gateway)
	gateway.Logger = logger
	gateway.Logger.SetFormatter(&logrus.TextFormatter{})
	gateway.Northerns = NewGenericResourceManager(gateway)
	gateway.Southerns = NewGenericResourceManager(gateway)
	gateway.Plugins = NewGenericResourceManager(gateway)
	gateway.Natives = NewGenericResourceManager(gateway)
	return gateway
}

// StartAllManagers starts monitoring for all resource managers
func (g *Gateway) StartAllManagers() {
	g.Logger.Info("Starting all resource managers...")
	g.Northerns.StartMonitoring()
	g.Southerns.StartMonitoring()
	g.Plugins.StartMonitoring()
	g.Natives.StartMonitoring()
	g.Logger.Info("All resource managers started.")
}

// StopAllManagers stops monitoring for all resource managers
func (g *Gateway) StopAllManagers() {
	g.Logger.Info("Stopping all resource managers...")
	g.Northerns.StopMonitoring()
	g.Southerns.StopMonitoring()
	g.Plugins.StopMonitoring()
	g.Natives.StopMonitoring()
	g.Logger.Info("All resource managers stopped.")
}

// GetManager retrieves a specific resource manager by name
func (g *Gateway) GetManager(managerName string) (*GenericResourceManager, error) {
	switch managerName {
	case "Northerns":
		return g.Northerns, nil
	case "Southerns":
		return g.Southerns, nil
	case "Plugins":
		return g.Plugins, nil
	case "Natives":
		return g.Natives, nil
	default:
		return nil, fmt.Errorf("resource manager %s not found", managerName)
	}
}

// LogResourceStatus logs the status of all resources in all managers
func (g *Gateway) LogResourceStatus() {
	g.Logger.Info("Logging resource statuses for all managers...")

	logManagerStatus := func(managerName string, manager *GenericResourceManager) {
		resources := manager.GetResourceList()
		g.Logger.Infof("Manager: %s, Total Resources: %d", managerName, len(resources))
		for _, resource := range resources {
			g.Logger.Infof("Resource UUID: %s, Status: %v", resource.Details().UUID, resource.Status())
		}
	}

	logManagerStatus("Northerns", g.Northerns)
	logManagerStatus("Southerns", g.Southerns)
	logManagerStatus("Plugins", g.Plugins)
	logManagerStatus("Natives", g.Natives)

	g.Logger.Info("Finished logging resource statuses.")
}

// ReloadAllManagers reloads all resources in all managers
func (g *Gateway) ReloadAllManagers() {
	g.Logger.Info("Reloading all resources in all managers...")

	reloadManager := func(managerName string, manager *GenericResourceManager) {
		resources := manager.GetResourceList()
		for _, resource := range resources {
			if err := manager.ReloadResource(resource.Details().UUID); err != nil {
				g.Logger.Errorf("Failed to reload resource in %s: %v", managerName, err)
			} else {
				g.Logger.Infof("Successfully reloaded resource in %s: %s", managerName, resource.Details().UUID)
			}
		}
	}

	reloadManager("Northerns", g.Northerns)
	reloadManager("Southerns", g.Southerns)
	reloadManager("Plugins", g.Plugins)
	reloadManager("Natives", g.Natives)

	g.Logger.Info("Finished reloading all resources.")
}

// LoadSouthernResource loads a resource into the Southern resource manager
func (g *Gateway) LoadSouthernResource(resourceType, uuid string, config map[string]any) error {
	g.Logger.Infof("Loading resource %s into the Southern resource manager...", uuid)
	if err := g.Southerns.CreateResource(resourceType, uuid, config); err != nil {
		g.Logger.Errorf("Failed to load resource %s into the Southern resource manager: %v", uuid, err)
		return err
	}
	g.Logger.Infof("Successfully loaded resource %s into the Southern resource manager.", uuid)
	return nil
}

// LoadNorthernResource loads a resource into the Northern resource manager
func (g *Gateway) LoadNorthernResource(resourceType, uuid string, config map[string]any) error {
	g.Logger.Infof("Loading resource %s into the Northern resource manager...", uuid)
	if err := g.Northerns.CreateResource(resourceType, uuid, config); err != nil {
		g.Logger.Errorf("Failed to load resource %s into the Northern resource manager: %v", uuid, err)
		return err
	}
	g.Logger.Infof("Successfully loaded resource %s into the Northern resource manager.", uuid)
	return nil
}

// LoadPluginResource loads a resource into the Plugin resource manager
func (g *Gateway) LoadPluginResource(resourceType, uuid string, config map[string]any) error {
	g.Logger.Infof("Loading resource %s into the Plugin resource manager...", uuid)
	if err := g.Plugins.CreateResource(resourceType, uuid, config); err != nil {
		g.Logger.Errorf("Failed to load resource %s into the Plugin resource manager: %v", uuid, err)
		return err
	}
	g.Logger.Infof("Successfully loaded resource %s into the Plugin resource manager.", uuid)
	return nil
}

// LoadNativeResource loads a resource into the Native resource manager
func (g *Gateway) LoadNativeResource(resourceType, uuid string, config map[string]any) error {
	g.Logger.Infof("Loading resource %s into the Native resource manager...", uuid)
	if err := g.Natives.CreateResource(resourceType, uuid, config); err != nil {
		g.Logger.Errorf("Failed to load resource %s into the Native resource manager: %v", uuid, err)
		return err
	}
	g.Logger.Infof("Successfully loaded resource %s into the Native resource manager.", uuid)
	return nil
}

// RemoveSouthernResource removes a resource from the Southern resource manager
func (g *Gateway) RemoveSouthernResource(uuid string) error {
	g.Logger.Infof("Removing resource %s from the Southern resource manager...", uuid)
	if err := g.Southerns.StopResource(uuid); err != nil {
		g.Logger.Errorf("Failed to remove resource %s from the Southern resource manager: %v", uuid, err)
		return err
	}
	g.Logger.Infof("Successfully removed resource %s from the Southern resource manager.", uuid)
	return nil
}

// RemoveNorthernResource removes a resource from the Northern resource manager
func (g *Gateway) RemoveNorthernResource(uuid string) error {
	g.Logger.Infof("Removing resource %s from the Northern resource manager...", uuid)
	if err := g.Northerns.StopResource(uuid); err != nil {
		g.Logger.Errorf("Failed to remove resource %s from the Northern resource manager: %v", uuid, err)
		return err
	}
	g.Logger.Infof("Successfully removed resource %s from the Northern resource manager.", uuid)
	return nil
}

// RemovePluginResource removes a resource from the Plugin resource manager
func (g *Gateway) RemovePluginResource(uuid string) error {
	g.Logger.Infof("Removing resource %s from the Plugin resource manager...", uuid)
	if err := g.Plugins.StopResource(uuid); err != nil {
		g.Logger.Errorf("Failed to remove resource %s from the Plugin resource manager: %v", uuid, err)
		return err
	}
	g.Logger.Infof("Successfully removed resource %s from the Plugin resource manager.", uuid)
	return nil
}

// RemoveNativeResource removes a resource from the Native resource manager
func (g *Gateway) RemoveNativeResource(uuid string) error {
	g.Logger.Infof("Removing resource %s from the Native resource manager...", uuid)
	if err := g.Natives.StopResource(uuid); err != nil {
		g.Logger.Errorf("Failed to remove resource %s from the Native resource manager: %v", uuid, err)
		return err
	}
	g.Logger.Infof("Successfully removed resource %s from the Native resource manager.", uuid)
	return nil
}

// GetSouthernResource retrieves a resource from the Southern resource manager by UUID
func (g *Gateway) GetSouthernResource(uuid string) (GenericResource, error) {
	g.Logger.Infof("Retrieving resource %s from the Southern resource manager...", uuid)
	resource, err := g.Southerns.GetResource(uuid)
	if err != nil {
		g.Logger.Errorf("Failed to retrieve resource %s from the Southern resource manager: %v", uuid, err)
		return nil, err
	}
	g.Logger.Infof("Successfully retrieved resource %s from the Southern resource manager.", uuid)
	return resource, nil
}

// GetNorthernResource retrieves a resource from the Northern resource manager by UUID
func (g *Gateway) GetNorthernResource(uuid string) (GenericResource, error) {
	g.Logger.Infof("Retrieving resource %s from the Northern resource manager...", uuid)
	resource, err := g.Northerns.GetResource(uuid)
	if err != nil {
		g.Logger.Errorf("Failed to retrieve resource %s from the Northern resource manager: %v", uuid, err)
		return nil, err
	}
	g.Logger.Infof("Successfully retrieved resource %s from the Northern resource manager.", uuid)
	return resource, nil
}

// GetPluginResource retrieves a resource from the Plugin resource manager by UUID
func (g *Gateway) GetPluginResource(uuid string) (GenericResource, error) {
	g.Logger.Infof("Retrieving resource %s from the Plugin resource manager...", uuid)
	resource, err := g.Plugins.GetResource(uuid)
	if err != nil {
		g.Logger.Errorf("Failed to retrieve resource %s from the Plugin resource manager: %v", uuid, err)
		return nil, err
	}
	g.Logger.Infof("Successfully retrieved resource %s from the Plugin resource manager.", uuid)
	return resource, nil
}

// GetNativeResource retrieves a resource from the Native resource manager by UUID
func (g *Gateway) GetNativeResource(uuid string) (GenericResource, error) {
	g.Logger.Infof("Retrieving resource %s from the Native resource manager...", uuid)
	resource, err := g.Natives.GetResource(uuid)
	if err != nil {
		g.Logger.Errorf("Failed to retrieve resource %s from the Native resource manager: %v", uuid, err)
		return nil, err
	}
	g.Logger.Infof("Successfully retrieved resource %s from the Native resource manager.", uuid)
	return resource, nil
}

// GetSouthernServices retrieves the services of a resource from the Southern resource manager by UUID
func (g *Gateway) GetSouthernServices(uuid string) ([]ResourceService, error) {
	g.Logger.Infof("Retrieving services for resource %s from the Southern resource manager...", uuid)
	resource, err := g.Southerns.GetResource(uuid)
	if err != nil {
		g.Logger.Errorf("Failed to retrieve resource %s from the Southern resource manager: %v", uuid, err)
		return nil, err
	}
	services := resource.Services()
	g.Logger.Infof("Successfully retrieved services for resource %s from the Southern resource manager.", uuid)
	return services, nil
}

// GetNorthernServices retrieves the services of a resource from the Northern resource manager by UUID
func (g *Gateway) GetNorthernServices(uuid string) ([]ResourceService, error) {
	g.Logger.Infof("Retrieving services for resource %s from the Northern resource manager...", uuid)
	resource, err := g.Northerns.GetResource(uuid)
	if err != nil {
		g.Logger.Errorf("Failed to retrieve resource %s from the Northern resource manager: %v", uuid, err)
		return nil, err
	}
	services := resource.Services()
	g.Logger.Infof("Successfully retrieved services for resource %s from the Northern resource manager.", uuid)
	return services, nil
}

// GetPluginServices retrieves the services of a resource from the Plugin resource manager by UUID
func (g *Gateway) GetPluginServices(uuid string) ([]ResourceService, error) {
	g.Logger.Infof("Retrieving services for resource %s from the Plugin resource manager...", uuid)
	resource, err := g.Plugins.GetResource(uuid)
	if err != nil {
		g.Logger.Errorf("Failed to retrieve resource %s from the Plugin resource manager: %v", uuid, err)
		return nil, err
	}
	services := resource.Services()
	g.Logger.Infof("Successfully retrieved services for resource %s from the Plugin resource manager.", uuid)
	return services, nil
}

// GetNativeServices retrieves the services of a resource from the Native resource manager by UUID
func (g *Gateway) GetNativeServices(uuid string) ([]ResourceService, error) {
	g.Logger.Infof("Retrieving services for resource %s from the Native resource manager...", uuid)
	resource, err := g.Natives.GetResource(uuid)
	if err != nil {
		g.Logger.Errorf("Failed to retrieve resource %s from the Native resource manager: %v", uuid, err)
		return nil, err
	}
	services := resource.Services()
	g.Logger.Infof("Successfully retrieved services for resource %s from the Native resource manager.", uuid)
	return services, nil
}

// CallSouthernOnService invokes a service on a resource in the Southern resource manager by UUID
func (g *Gateway) CallSouthernOnService(uuid string, request ResourceServiceRequest) (ResourceServiceResponse, error) {
	g.Logger.Infof("Invoking service %s on resource %s in the Southern resource manager...", request.Name, uuid)
	resource, err := g.Southerns.GetResource(uuid)
	if err != nil {
		g.Logger.Errorf("Failed to retrieve resource %s from the Southern resource manager: %v", uuid, err)
		return ResourceServiceResponse{}, err
	}
	response, err := resource.OnService(request)
	if err != nil {
		g.Logger.Errorf("Failed to invoke service %s on resource %s in the Southern resource manager: %v", request.Name, uuid, err)
		return ResourceServiceResponse{}, err
	}
	g.Logger.Infof("Successfully invoked service %s on resource %s in the Southern resource manager.", request.Name, uuid)
	return response, nil
}

// CallNorthernOnService invokes a service on a resource in the Northern resource manager by UUID
func (g *Gateway) CallNorthernOnService(uuid string, request ResourceServiceRequest) (ResourceServiceResponse, error) {
	g.Logger.Infof("Invoking service %s on resource %s in the Northern resource manager...", request.Name, uuid)
	resource, err := g.Northerns.GetResource(uuid)
	if err != nil {
		g.Logger.Errorf("Failed to retrieve resource %s from the Northern resource manager: %v", uuid, err)
		return ResourceServiceResponse{}, err
	}
	response, err := resource.OnService(request)
	if err != nil {
		g.Logger.Errorf("Failed to invoke service %s on resource %s in the Northern resource manager: %v", request.Name, uuid, err)
		return ResourceServiceResponse{}, err
	}
	g.Logger.Infof("Successfully invoked service %s on resource %s in the Northern resource manager.", request.Name, uuid)
	return response, nil
}

// CallPluginOnService invokes a service on a resource in the Plugin resource manager by UUID
func (g *Gateway) CallPluginOnService(uuid string, request ResourceServiceRequest) (ResourceServiceResponse, error) {
	g.Logger.Infof("Invoking service %s on resource %s in the Plugin resource manager...", request.Name, uuid)
	resource, err := g.Plugins.GetResource(uuid)
	if err != nil {
		g.Logger.Errorf("Failed to retrieve resource %s from the Plugin resource manager: %v", uuid, err)
		return ResourceServiceResponse{}, err
	}
	response, err := resource.OnService(request)
	if err != nil {
		g.Logger.Errorf("Failed to invoke service %s on resource %s in the Plugin resource manager: %v", request.Name, uuid, err)
		return ResourceServiceResponse{}, err
	}
	g.Logger.Infof("Successfully invoked service %s on resource %s in the Plugin resource manager.", request.Name, uuid)
	return response, nil
}

// CallNativeOnService invokes a service on a resource in the Native resource manager by UUID
func (g *Gateway) CallNativeOnService(uuid string, request ResourceServiceRequest) (ResourceServiceResponse, error) {
	g.Logger.Infof("Invoking service %s on resource %s in the Native resource manager...", request.Name, uuid)
	resource, err := g.Natives.GetResource(uuid)
	if err != nil {
		g.Logger.Errorf("Failed to retrieve resource %s from the Native resource manager: %v", uuid, err)
		return ResourceServiceResponse{}, err
	}
	response, err := resource.OnService(request)
	if err != nil {
		g.Logger.Errorf("Failed to invoke service %s on resource %s in the Native resource manager: %v", request.Name, uuid, err)
		return ResourceServiceResponse{}, err
	}
	g.Logger.Infof("Successfully invoked service %s on resource %s in the Native resource manager.", request.Name, uuid)
	return response, nil
}

// GatewaySnapshot takes a snapshot of the current state of all resources in all resource managers
func (g *Gateway) GatewaySnapshot() map[string]any {
	g.Logger.Info("Taking a snapshot of the current state of all resources in all managers...")

	snapshot := make(map[string]any)

	// Helper function to take a snapshot of a specific manager
	takeManagerSnapshot := func(manager *GenericResourceManager) map[string]any {
		managerSnapshot := make(map[string]any)
		resources := manager.GetResourceList()
		for _, resource := range resources {
			managerSnapshot[resource.Details().UUID] = map[string]any{
				"Status":   resource.Status(),
				"Services": resource.Services(),
				"Details": map[string]any{
					"Config": resource.Details().GetConfig(),
					"Type":   resource.Details().Type,
				},
			}
		}
		return managerSnapshot
	}

	// Take snapshots of all managers
	snapshot["Northerns"] = takeManagerSnapshot(g.Northerns)
	snapshot["Southerns"] = takeManagerSnapshot(g.Southerns)
	snapshot["Plugins"] = takeManagerSnapshot(g.Plugins)
	snapshot["Natives"] = takeManagerSnapshot(g.Natives)

	// Add a timestamp to the snapshot
	snapshot["Timestamp"] = time.Now().Format(time.RFC3339)

	g.Logger.Info("Successfully took a snapshot of all resources.")
	return snapshot
}
