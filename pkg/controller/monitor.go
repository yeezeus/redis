package controller

import (
	"fmt"

	tapi "github.com/k8sdb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/k8sdb/apimachinery/pkg/monitor"
)

func (c *Controller) newMonitorController(redis *tapi.Xdb) (monitor.Monitor, error) {
	monitorSpec := redis.Spec.Monitor

	if monitorSpec == nil {
		return nil, fmt.Errorf("MonitorSpec not found in %v", redis.Spec)
	}

	if monitorSpec.Prometheus != nil {
		return monitor.NewPrometheusController(c.Client, c.ApiExtKubeClient, c.promClient, c.opt.OperatorNamespace), nil
	}

	return nil, fmt.Errorf("Monitoring controller not found for %v", monitorSpec)
}

func (c *Controller) addMonitor(redis *tapi.Xdb) error {
	ctrl, err := c.newMonitorController(redis)
	if err != nil {
		return err
	}
	return ctrl.AddMonitor(redis.ObjectMeta, redis.Spec.Monitor)
}

func (c *Controller) deleteMonitor(redis *tapi.Xdb) error {
	ctrl, err := c.newMonitorController(redis)
	if err != nil {
		return err
	}
	return ctrl.DeleteMonitor(redis.ObjectMeta, redis.Spec.Monitor)
}

func (c *Controller) updateMonitor(oldXdb, updatedXdb *tapi.Xdb) error {
	var err error
	var ctrl monitor.Monitor
	if updatedXdb.Spec.Monitor == nil {
		ctrl, err = c.newMonitorController(oldXdb)
	} else {
		ctrl, err = c.newMonitorController(updatedXdb)
	}
	if err != nil {
		return err
	}
	return ctrl.UpdateMonitor(updatedXdb.ObjectMeta, oldXdb.Spec.Monitor, updatedXdb.Spec.Monitor)
}
