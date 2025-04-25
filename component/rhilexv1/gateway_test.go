package rhilex

import (
	"context"
	"fmt"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// Helper function to create a test logger
func newTestLogger() *logrus.Logger {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel) // Or any level you prefer for testing
	logger.SetFormatter(&logrus.TextFormatter{
		DisableColors: true,
	})
	return logger
}

func TestNewGateway(t *testing.T) {
	logger := newTestLogger()
	gateway := NewGateway(logger)

	assert.NotNil(t, gateway, "Gateway should not be nil")
	assert.NotNil(t, gateway.logger, "Logger should not be nil")
	assert.NotNil(t, gateway.inits, "Init functions map should not be nil")
	assert.NotNil(t, gateway.config, "Config should not be nil")
	assert.NotNil(t, gateway.northerns, "Northerns manager should not be nil")
	assert.NotNil(t, gateway.southerns, "Southerns manager should not be nil")
	assert.NotNil(t, gateway.plugins, "Plugins manager should not be nil")
	assert.NotNil(t, gateway.natives, "Natives manager should not be nil")
	assert.NotNil(t, gateway.cache, "Cache should not be nil")
	assert.NotNil(t, gateway.queue, "Queue should not be nil")
	assert.NotNil(t, gateway.broker, "Broker should not be nil")
	assert.NotNil(t, gateway.cronManager, "CronManager should not be nil")
}

func TestGateway_Start_Stop(t *testing.T) {
	logger := newTestLogger()
	gateway := NewGateway(logger)
	ctx, cancel := context.WithCancel(context.Background())
	config := RhilexConfig{AppId: "test"}

	err := gateway.Start(ctx, cancel, config)
	assert.NoError(t, err, "Start should not return an error")

	err = gateway.Stop()
	assert.NoError(t, err, "Stop should not return an error")
}

func TestGateway_Register_CallInitFunc(t *testing.T) {
	logger := newTestLogger()
	gateway := NewGateway(logger)

	initFunc := func() error {
		t.Log("Init function called")
		return nil
	}

	gateway.RegisterInitFunc("testInit", initFunc)
	assert.Contains(t, gateway.inits, "testInit", "Init function should be registered")

	gateway.CallInitFunc()
}

func TestGateway_GetManager(t *testing.T) {
	logger := newTestLogger()
	gateway := NewGateway(logger)

	managers := []string{"northerns", "southerns", "plugins", "natives"}
	for _, managerName := range managers {
		manager, err := gateway.GetManager(managerName)
		assert.NoError(t, err, fmt.Sprintf("GetManager(%s) should not return an error", managerName))
		assert.NotNil(t, manager, fmt.Sprintf("GetManager(%s) should not return nil", managerName))
	}

	_, err := gateway.GetManager("invalid")
	assert.Error(t, err, "GetManager(invalid) should return an error")
}
