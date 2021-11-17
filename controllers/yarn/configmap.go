package yarn

import (
	"encoding/xml"
	v1 "github.com/dataworkbench/hdfs-operator/api/v1"
	com "github.com/dataworkbench/hdfs-operator/common"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	YarnConfigName       = "yarn-config"
	MapredSiteFileName   = "mapred-site.xml"
	YarnSiteFileName     = "yarn-site.xml"
)

func BuildConfigMap(hdfs v1.HDFS) (corev1.ConfigMap, error) {

	mapredSiteData, err := RenderMapredSiteCfg(hdfs.Spec.Yarn.MapredSite)
	if err != nil {
		return corev1.ConfigMap{}, err
	}
	yarnSiteData, err := RenderYarnSiteCfg(hdfs)
	if err != nil {
		return corev1.ConfigMap{}, err
	}
	return corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: hdfs.Namespace,
			Name:      com.GetName(hdfs.Name, YarnConfigName),
			OwnerReferences: com.GetOwnerReference(hdfs),
		},
		Data: map[string]string{
			MapredSiteFileName: string(mapredSiteData),
			YarnSiteFileName: string(yarnSiteData),
		},
	}, nil

}

func RenderMapredSiteCfg(cfgs []v1.ClusterConfig) ([]byte, error) {

	var c = com.Configuration{}

	c.Configuration = append(c.Configuration, com.Property{
		Name:  "mapreduce.framework.name",
		Value: "yarn",
	},
	)
	for _, cfg := range cfgs {
		c.Configuration = append(c.Configuration, com.Property{
			Name:  cfg.Property,
			Value: cfg.Value,
		})
	}
	return xml.MarshalIndent(c, " ", " ")
}

func RenderYarnSiteCfg(hdfs v1.HDFS) ([]byte, error) {

	var c = com.Configuration{}

	rmPrefix := com.GetName(hdfs.Name, hdfs.Spec.Yarn.Name)+"-rm"
	rmService := rmPrefix+"."+hdfs.Namespace+".svc.cluster.local"

	c.Configuration = append(c.Configuration, com.Property{
		Name:  "yarn.resourcemanager.hostname",
		Value: rmPrefix+"-0." + rmService,
	}, com.Property{
		Name:  "yarn.nodemanager.vmem-check-enabled",
		Value: "false",
	}, com.Property{
		Name:  "yarn.nodemanager.aux-services",
		Value: "mapreduce_shuffle",
	}, com.Property{
		Name:  "yarn.nodemanager.aux-services.mapreduce_shuffle.class",
		Value: "org.apache.hadoop.mapred.ShuffleHandler",
	},com.Property{
		Name:  "yarn.nodemanager.remote-app-log-dir",
		Value: "/var/log/hadoop-yarn/apps",
	},
	)
	for _, cfg := range hdfs.Spec.Yarn.YarnSite {
		c.Configuration = append(c.Configuration, com.Property{
			Name:  cfg.Property,
			Value: cfg.Value,
		})
	}
	return xml.MarshalIndent(c, " ", " ")
}