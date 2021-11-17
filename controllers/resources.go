package controllers

import (
	"github.com/dataworkbench/hdfs-operator/api/v1"
	com "github.com/dataworkbench/hdfs-operator/common"
	dn "github.com/dataworkbench/hdfs-operator/controllers/datanode"
	jn "github.com/dataworkbench/hdfs-operator/controllers/journalnode"
	nn "github.com/dataworkbench/hdfs-operator/controllers/namenode"
	"github.com/dataworkbench/hdfs-operator/controllers/yarn"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"reflect"
)

type HdfsResources struct {
	StatefulSets  []appsv1.StatefulSet
	Datanode      appsv1.StatefulSet
	ConfigMaps    []corev1.ConfigMap
	Services      []corev1.Service
}

func BuildExpectedResources(hdfs v1.HDFS) (HdfsResources, error) {

	VersionHandler(hdfs.Spec.Version)

	configs, err := BuildConfigMaps(hdfs)
	if err != nil {
		return HdfsResources{}, err
	}

	services, err := BuildServices(hdfs)
	if err != nil {
		return HdfsResources{}, err
	}

	statefulSets, err := BuildStatefulSets(hdfs)
	if err != nil {
		return HdfsResources{}, err
	}

	dnSet, err := dn.BuildStatefulSet(hdfs)
	if err != nil {
		return HdfsResources{}, err
	}

	return HdfsResources{
		StatefulSets: statefulSets,
		Datanode:     dnSet,
		ConfigMaps:   configs ,
		Services:     services,
	}, nil
}

func VersionHandler(version string) {
	if version[0:1] == "3" {
		com.DatanodeRpcPort = 9864
		com.NamenodeHttpPort = 9870
		com.NamenodeRpcPort = 9820
	}
}

func BuildConfigMaps(hdfs v1.HDFS) (c []corev1.ConfigMap,err error) {

	config, err := com.BuildHdfsConfig(hdfs, com.GetName(hdfs.Name, com.CommonConfigName))
	if err != nil {
		return c, err
	}
	nnScripts := nn.BuildConfigMap(hdfs)
	dnScripts := dn.BuildConfigMap(hdfs)

	if !reflect.DeepEqual(hdfs.Spec.Yarn, v1.Yarn{}) {
		yarnConfig ,err:= yarn.BuildConfigMap(hdfs)
		if err != nil {
			return c, err
		}
		c = append( c,yarnConfig)
	}

	return append(c, config, nnScripts, dnScripts), nil
}

func BuildServices(hdfs v1.HDFS) (svc []corev1.Service,err error) {

	nnSvc := com.HeadlessService(hdfs,
		com.GetName(hdfs.Name, hdfs.Spec.Namenode.Name),
		nn.GetDefaultServicePorts())
	jnSvc := com.HeadlessService(hdfs, com.GetName(hdfs.Name,
		hdfs.Spec.Journalnode.Name),
		jn.GetDefaultServicePorts())

	if !reflect.DeepEqual(hdfs.Spec.Yarn, v1.Yarn{}) {
		rmSvc := com.HeadlessService(hdfs,
			com.GetName(hdfs.Name, hdfs.Spec.Yarn.Name)+"-rm",
			yarn.GetRMServicePorts())
		nmSvc := com.HeadlessService(hdfs,
			com.GetName(hdfs.Name, hdfs.Spec.Yarn.Name)+"-nm",
			yarn.GetNMServicePorts())
		svc = append(svc, rmSvc, nmSvc )
	}
	return append(svc, nnSvc, jnSvc ), nil
}

func BuildStatefulSets(hdfs v1.HDFS) (s []appsv1.StatefulSet,err error) {

	nnStatefulSet, err := nn.BuildStatefulSet(hdfs)
	if err != nil {
		return s, err
	}
	jnStatefulSet, err := jn.BuildStatefulSet(hdfs)
	if err != nil {
		return s, err
	}

	if !reflect.DeepEqual(hdfs.Spec.Yarn, v1.Yarn{}) {
		rmStatefulSet, err := yarn.BuildRMStatefulSet(hdfs)
		if err != nil {
			return s, err
		}
		nmStatefulSet, err := yarn.BuildNMStatefulSet(hdfs)
		if err != nil {
			return s, err
		}
		s = append(s,rmStatefulSet,nmStatefulSet)

	}

	return append(s, nnStatefulSet, jnStatefulSet ), nil
}