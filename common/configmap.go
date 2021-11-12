package common

import (
	"encoding/xml"
	hdfsv1 "github.com/dataworkbench/hdfs-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strconv"
)

const (
	CoreSiteFileName     = "core-site.xml"
	HdfsSiteFileName     = "hdfs-site.xml"
	CommonConfigName     = "common-config"
	VolumesConfigMapName = "hdfs-config"
	HdfsConfigMountPath  = "/etc/hadoop-custom-conf"
)

var (
	NamenodeRpcPort       = 8020  // 9820
	NamenodeHttpPort      = 50070  // 9870
	DatanodeRpcPort       = 50075 //  9864
	//Journalnode 8485 8480
)

func BuildHdfsConfig(hdfs hdfsv1.HDFS, name string) (corev1.ConfigMap, error) {
	coreSiteData, err := RenderCoreSiteCfg(hdfs.Spec)
	if err != nil {
		return corev1.ConfigMap{}, err
	}
	hdfsSiteData, err := RenderHdfsSiteCfg(hdfs)
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
			Name:      name,
			//Labels:    label.NewConfigLabels(),
			OwnerReferences: GetOwnerReference(hdfs),
		},
		Data: map[string]string{
			CoreSiteFileName: string(coreSiteData),
			HdfsSiteFileName: string(hdfsSiteData),
		},
	}, nil
}

type configuration struct {
	XMLName       xml.Name   `xml:"configuration"`
	Version       string     `xml:"version,attr"`
	Configuration []property `xml:"configuration"`
}

type property struct {
	XMLName xml.Name `xml:"property"`
	Name    string   `xml:"name"`
	Value   string   `xml:"value"`
}

func RenderCoreSiteCfg(spec hdfsv1.HDFSSpec) ([]byte, error) {
	var c = configuration{}

	var zkCfg = property{}
	zkCfg.Name = "ha.zookeeper.quorum"
	zkCfg.Value = spec.ZkQuorum

	c.Configuration = append(c.Configuration, property{
		Name:  "fs.defaultFS",
		Value: "hdfs://hdfs-k8s",
	}, zkCfg)
	return xml.MarshalIndent(c, " ", " ")
}

func RenderHdfsSiteCfg(hdfs hdfsv1.HDFS) ([]byte, error) {

	var c = configuration{}

	// prefixe of pod and service are the same
	nnPrefix := GetName(hdfs.Name, hdfs.Spec.Namenode.Name)
	nnService := nnPrefix+"."+hdfs.Namespace+".svc.cluster.local"

	jnPrefix := GetName(hdfs.Name, hdfs.Spec.Journalnode.Name)
	jnService := jnPrefix+"."+hdfs.Namespace+".svc.cluster.local"
	editsDir :=  "qjournal://"+ jnPrefix+"-0." + jnService+":8485;"+ jnPrefix+"-1."+ jnService+":8485;"+ jnPrefix+"-2."+ jnService+":8485" + "/hdfs-k8s"

	//get dn  MountPaths
	dataDirs := ""
	for _, dir := range hdfs.Spec.Datanode.Datadirs {
		dataDirs = dataDirs+"/hadoop/dfs/data"+dir+","  // const DNDataVolumeMountPath = /hadoop/dfs/data
	}


	c.Configuration = append(c.Configuration, property{
		Name:  "dfs.nameservices",
		Value: "hdfs-k8s",
	}, property{
		Name:  "dfs.ha.namenodes.hdfs-k8s",
		Value: "nn0,nn1",
	}, property{
		Name:  "dfs.namenode.rpc-address.hdfs-k8s." + "nn0",
		Value: nnPrefix+"-0." + nnService + ":"+ strconv.Itoa(NamenodeRpcPort),
	}, property{
		Name:  "dfs.namenode.rpc-address.hdfs-k8s.nn1",
		Value: nnPrefix+"-1." + nnService + ":"+strconv.Itoa(NamenodeRpcPort),
	}, property{
		Name:  "dfs.namenode.http-address.hdfs-k8s.nn0",
		Value: nnPrefix+"-0." + nnService +":"+strconv.Itoa(NamenodeHttpPort),
	}, property{
		Name:  "dfs.namenode.http-address.hdfs-k8s.nn1",
		Value: nnPrefix+"-1." + nnService +":"+strconv.Itoa(NamenodeHttpPort),
	}, property{
		Name:  "dfs.namenode.shared.edits.dir",
		Value: editsDir,
	}, property{
		Name:  "dfs.ha.automatic-failover.enabled",
		Value: "true",
	}, property{
		Name:  "dfs.ha.fencing.methods",
		Value: "shell(/bin/true)",
	}, property{
		Name:  "dfs.journalnode.edits.dir",
		Value: "/hadoop/dfs/journal",
	}, property{
		Name:  "dfs.client.failover.proxy.provider.hdfs-k8s",
		Value: "org.apache.hadoop.hdfs.server.namenode.ha.ConfiguredFailoverProxyProvider",
	}, property{
		Name:  "dfs.namenode.name.dir",
		Value: "file:///hadoop/dfs/name",
	}, property{
		Name:  "dfs.namenode.datanode.registration.ip-hostname-check",
		Value: "false",
	}, property{
		Name:  "dfs.datanode.data.dir",
		Value: dataDirs,
		//Value: "/hadoop/dfs/data/0",
	})
	return xml.MarshalIndent(c, " ", " ")
}
