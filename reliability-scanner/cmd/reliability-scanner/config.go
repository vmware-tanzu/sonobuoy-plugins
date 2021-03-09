package main

import (
	"strconv"

	"github.com/sirupsen/logrus"
	"github.com/vmware-tanzu/sonobuoy-plugins/reliability-scanner/api/v1alpha1/namespace/labels"
	"github.com/vmware-tanzu/sonobuoy-plugins/reliability-scanner/api/v1alpha1/pod/disruption"
	"github.com/vmware-tanzu/sonobuoy-plugins/reliability-scanner/api/v1alpha1/pod/probes"
	"github.com/vmware-tanzu/sonobuoy-plugins/reliability-scanner/api/v1alpha1/pod/qos"
	"github.com/vmware-tanzu/sonobuoy-plugins/reliability-scanner/internal"
)

// initializeQueriers sets up queriers based on the runners configuration.
func initializeQueriers(runner *internal.Runner) error {
	for _, checkCfg := range runner.Config.Checks {
		switch checkCfg.Kind {
		case "v1alpha1/pod/disruption":
			querier, err := disruption.NewQuerier(&disruption.QuerierSpec{})
			if err != nil {
				runner.Logger.WithFields(logrus.Fields{
					"kind":  checkCfg.Spec["kind"],
					"name":  checkCfg.Spec["name"],
					"phase": "add",
				}).Error(err)
				return err
			}
			querier.AddtoRunner(runner)
		case "v1alpha1/pod/probes":
			querier, err := probes.NewQuerier(&probes.QuerierSpec{})
			if err != nil {
				runner.Logger.WithFields(logrus.Fields{
					"kind":  checkCfg.Spec["kind"],
					"name":  checkCfg.Spec["name"],
					"phase": "add",
				}).Error(err)
				return err
			}
			querier.AddtoRunner(runner)
		case "v1alpha1/pod/qos":
			includeDetail, err := strconv.ParseBool(checkCfg.Spec["include_detail"])
			if err != nil {
				runner.Logger.WithFields(logrus.Fields{
					"kind":  checkCfg.Spec["kind"],
					"name":  checkCfg.Spec["name"],
					"phase": "add",
				}).Error(err)
				return err
			}
			querier, err := qos.NewQuerier(&qos.QuerierSpec{
				IncludeDetail:          includeDetail,
				MinimumDesiredQOSClass: checkCfg.Spec["minimum_desired_qos_class"],
			})
			if err != nil {
				runner.Logger.WithFields(logrus.Fields{
					"kind":  checkCfg.Spec["kind"],
					"name":  checkCfg.Spec["name"],
					"phase": "add",
				}).Error(err)
				return err
			}
			querier.AddtoRunner(runner)
		case "v1alpha1/namespace/labels":
			includeDetail, err := strconv.ParseBool(checkCfg.Spec["include_detail"])
			if err != nil {
				runner.Logger.WithFields(logrus.Fields{
					"check_name": checkCfg.Spec["Name"],
					"phase":      "add",
				}).Error(err)
				return err
			}
			querier, err := labels.NewQuerier(&labels.QuerierSpec{
				Key:           checkCfg.Spec["key"],
				IncludeDetail: includeDetail,
			})
			if err != nil {
				runner.Logger.WithFields(logrus.Fields{
					"check_name": checkCfg.Spec["Name"],
					"phase":      "add",
				}).Error(err)
				return err
			}
			querier.AddtoRunner(runner)
		}
	}
	return nil
}
