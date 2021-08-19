package controllers

import (
	"github.com/dataworkbench/hdfs-operator/api/v1"
	com "github.com/dataworkbench/hdfs-operator/controllers/common"
	dn "github.com/dataworkbench/hdfs-operator/controllers/datanode"
	jn "github.com/dataworkbench/hdfs-operator/controllers/journalnode"
	nn "github.com/dataworkbench/hdfs-operator/controllers/namenode"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

type HdfsResources struct {
	Nodes         []NodeResources
	Datanode      DataResources
	CommonConfig  corev1.ConfigMap
}

type NodeResources struct {
	StatefulSet     appsv1.StatefulSet
	HeadlessService corev1.Service
	Config          corev1.ConfigMap
}

type DataResources struct {
	DaemonSet       appsv1.DaemonSet
	Config          corev1.ConfigMap
}

func BuildExpectedResources(hdfs v1.HDFS) (HdfsResources, error) {
	config ,err:=  com.BuildHdfsConfig(hdfs,com.GetName(hdfs.Name,com.CommonConfigName))
	if err != nil {
		return HdfsResources{},err
	}
	nnResources, err := BuildNNExpectedResources(hdfs)
	if err != nil {
		return HdfsResources{},err
	}
	jnResources, err := BuildJNExpectedResources(hdfs)
	if err != nil {
		return HdfsResources{},err
	}
	dnResources, err := BuildDNExpectedResources(hdfs)
	if err != nil {
		return HdfsResources{},err
	}

	var nodes []NodeResources
	return HdfsResources{
		Nodes:         append(nodes,nnResources,jnResources),
		Datanode:      dnResources,
		CommonConfig:  config,
	},nil
}

func BuildNNExpectedResources(hdfs v1.HDFS) (NodeResources, error) {
	cfg := nn.BuildConfigMap(hdfs)
	statefulSet, err := nn.BuildStatefulSet(hdfs)
	if err != nil {
		return NodeResources{}, err
	}
	headlessSvc := com.HeadlessService(&hdfs, statefulSet.Name,nn.GetDefaultServicePorts())
	return NodeResources{
		StatefulSet:     statefulSet,
		HeadlessService: headlessSvc,
		Config:          cfg,
	},nil
}

func BuildJNExpectedResources(hdfs v1.HDFS) (NodeResources, error) {
	// build stateful set and associated headless service
	statefulSet, err := jn.BuildStatefulSet( hdfs)
	if err != nil {
		return NodeResources{}, err
	}
	headlessSvc := com.HeadlessService(&hdfs, statefulSet.Name,jn.GetDefaultServicePorts())
	return NodeResources{
		StatefulSet:     statefulSet,
		HeadlessService: headlessSvc,
	},nil
}

func BuildDNExpectedResources(hdfs v1.HDFS) (DataResources, error) {
	cfg := dn.BuildConfigMap(hdfs)
	daemonSet, err := dn.BuildDaemonSet(hdfs)
	if err != nil {
		return DataResources{}, err
	}
	return DataResources{
		DaemonSet:     daemonSet,
		Config:          cfg,
	},nil
}

