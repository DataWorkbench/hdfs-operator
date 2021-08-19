package common

import (
	"encoding/xml"
	hdfsv1 "github.com/dataworkbench/hdfs-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	CoreSiteFileName        = "core-site.xml"
	HdfsSiteFileName        = "hdfs-site.xml"
	CommonConfigName        = "common-config"
	VolumesConfigMapName    = "hdfs-config"
	HdfsConfigMountPath     = "/etc/hadoop-custom-conf"
)

func BuildHdfsConfig(hdfs hdfsv1.HDFS,name string) (corev1.ConfigMap, error) {
	coreSiteData, err := RenderCoreSiteCfg()
	if err != nil {
		return corev1.ConfigMap{}, err
	}
	hdfsSiteData, err := RenderHdfsSiteCfg()
	if err != nil {
		return corev1.ConfigMap{}, err
	}
	return corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: hdfs.Namespace,
			Name:      name,
			//Labels:    label.NewConfigLabels(),
		},
		Data: map[string]string{
			CoreSiteFileName: string(coreSiteData),
			HdfsSiteFileName: string(hdfsSiteData),
		},
	},nil
}

type configuration struct {
	XMLName       xml.Name     `xml:"configuration"`
	Version       string       `xml:"version,attr"`
	Configuration []property   `xml:"configuration"`
}

type property struct {
	XMLName      xml.Name     `xml:"property"`
	Name         string       `xml:"name"`
	Value        string       `xml:"value"`
}

func RenderCoreSiteCfg()([]byte, error){
	var c = configuration{}
	c.Configuration =append(c.Configuration,property{
		Name: "fs.defaultFS",
		Value: "hdfs://hdfs-k8s",
	},property{
		Name: "ha.zookeeper.quorum",
		Value: "zk-0.zk-hs.default.svc.cluster.local:2181,zk-1.zk-hs.default.svc.cluster.local:2181,zk-2.zk-hs.default.svc.cluster.local:2181",
	})
	return xml.MarshalIndent(c, " ", " ")
}

func RenderHdfsSiteCfg()([]byte, error){

	var c = configuration{}
	c.Configuration =append(c.Configuration,property{
		Name: "dfs.nameservices",
		Value: "hdfs-k8s",
	},property{
		Name: "dfs.ha.namenodes.hdfs-k8s",
		Value: "nn0,nn1",
	},property{
		Name: "dfs.namenode.rpc-address.hdfs-k8s"+"nn0",
		Value: "my-hdfs-namenode-0."+"my-hdfs-namenode.default.svc.cluster.local"+":8020",
	},property{
		Name: "dfs.namenode.rpc-address.hdfs-k8s.nn1",
		Value: "my-hdfs-namenode-1."+"my-hdfs-namenode.default.svc.cluster.local"+":8020",
	},property{
		Name: "dfs.namenode.http-address.hdfs-k8s.nn0",
		Value: "my-hdfs-namenode-0.my-hdfs-namenode.default.svc.cluster.local:50070",
	},property{
		Name: "dfs.namenode.http-address.hdfs-k8s.nn1",
		Value: "my-hdfs-namenode-1.my-hdfs-namenode.default.svc.cluster.local:50070",
	},property{
		Name: "dfs.namenode.shared.edits.dir",
		Value: "qjournal://my-hdfs-journalnode-1.my-hdfs-journalnode.default.svc.cluster.local:8485;my-hdfs-journalnode-2.my-hdfs-journalnode.default.svc.cluster.local:8485;my-hdfs-journalnode-0.my-hdfs-journalnode.default.svc.cluster.local:8485/hdfs-k8s",
	},property{
		Name: "dfs.ha.automatic-failover.enabled",
		Value: "true",
	},property{
		Name: "dfs.ha.fencing.methods",
		Value: "shell(/bin/true)",     //
	},property{
		Name: "dfs.journalnode.edits.dir",
		Value: "/hadoop/dfs/journal",
	},property{
		Name: "dfs.client.failover.proxy.provider.hdfs-k8s",
		Value: "org.apache.hadoop.hdfs.server.namenode.ha.ConfiguredFailoverProxyProvider",
	},property{
		Name: "dfs.namenode.name.dir",
		Value: "file:///hadoop/dfs/name",
	},property{
		Name: "dfs.namenode.datanode.registration.ip-hostname-check",
		Value: "false",
	},property{
		Name: "dfs.datanode.data.dir",
		Value: "/mnt/hdfs/dn-data",
	})
	return xml.MarshalIndent(c, " ", " ")
}
