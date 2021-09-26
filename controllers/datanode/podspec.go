package datanode

import (
	v1 "github.com/dataworkbench/hdfs-operator/api/v1"
	com "github.com/dataworkbench/hdfs-operator/common"
	corev1 "k8s.io/api/core/v1"
)

const (
	UhopperDatanodeImage   = "uhopper/hadoop-datanode:2.7.2"
	QydwdDatanodeImage     = "qydwd/hadoop-datanode:2.9.2"
)

// BuildDataNodePod builds a new PodTemplateSpec for DataNode.
func BuildDataNodePod(hdfs v1.HDFS, labels map[string]string) (corev1.PodTemplateSpec, error) {
	volumes, volumeMounts := buildVolumes(hdfs.Name)

	container := buildContainer(hdfs.Spec.Journalnode.Name, volumeMounts, GetImage(hdfs.Spec.Version))

	builder := &com.PodTemplateBuilder{}
	builder.WithContainers(container).
		WithSpecVolumes(volumes...).
		WithRestartPolicy(corev1.RestartPolicyAlways).
		WithHostNetwork(defaultOptional).
		WithHostPID(defaultOptional).
		WithDNSPolicy(corev1.DNSClusterFirstWithHostNet).
		WithTemplateMetadata(labels)

	return builder.PodTemplate, nil
}

func buildVolumes(name string) (volumes []corev1.Volume, volumeMounts []corev1.VolumeMount) {

	configVolume := com.NewConfigMapVolume(com.GetName(name, com.CommonConfigName), com.VolumesConfigMapName, com.HdfsConfigMountPath)

	scriptsVolume := com.NewConfigMapVolumeWithMode(com.GetName(name, DatanodeScripts), DNScriptsVolumeName, DNScriptsVolumeMountPath, 0744)

	hostPathVolume := NewHostPathVolume(DNDataVolumeName, DNDataHostPath, DNDataVolumeMountPath)

	//DaemonSetSpec.Template.Spec.Volume
	volumes = append(volumes,
		scriptsVolume.Volume(),
		configVolume.Volume(),
		hostPathVolume.Volume(),
	)
	//DaemonSetSpec.Template.Spec.containers.volumeMounts
	volumeMounts = append(volumeMounts,
		scriptsVolume.VolumeMount(),
		configVolume.VolumeMount(),
		hostPathVolume.VolumeMount(),
	)

	return volumes, volumeMounts
}

func buildContainer(name string, volumeMounts []corev1.VolumeMount, image string) corev1.Container {

	probe := &corev1.Probe{
		Handler: corev1.Handler{
			Exec: &corev1.ExecAction{
				Command: []string{"/dn-scripts/check-status.sh"},
			},
		},
		InitialDelaySeconds: 60,
		PeriodSeconds:       30,
	}

	return corev1.Container{
		ImagePullPolicy: corev1.PullIfNotPresent,
		Image:           image,
		Name:            name,
		Env:             envVars(),
		VolumeMounts:    volumeMounts,
		LivenessProbe:   probe,
		ReadinessProbe:  probe,
		SecurityContext: &corev1.SecurityContext{Privileged: &defaultOptional},
	}
}

func envVars() []corev1.EnvVar {
	return []corev1.EnvVar{
		{Name: "HADOOP_CUSTOM_CONF_DIR", Value: "/etc/hadoop-custom-conf"},
		{Name: "MULTIHOMED_NETWORK", Value: "0"},
	}
}

// HostPathVolume defines a volume to expose a configmap
type HostPathVolume struct {
	name      string //volume and volumeMounts associated name
	hostPath  string //volumes.hostPath.path
	mountPath string
}

// NewHostPathVolume creates a new ConfigMapVolume
func NewHostPathVolume(name, hostPath, mountPath string) HostPathVolume {
	return HostPathVolume{
		name:      name,
		hostPath:  hostPath,
		mountPath: mountPath,
	}
}

// Volume returns the k8s volume.
func (cm HostPathVolume) Volume() corev1.Volume {
	return corev1.Volume{
		Name: cm.name,
		VolumeSource: corev1.VolumeSource{
			HostPath: &corev1.HostPathVolumeSource{
				Path: cm.hostPath,
			},
		},
	}
}

// VolumeMount returns the k8s volume mount.
func (cm HostPathVolume) VolumeMount() corev1.VolumeMount {
	return corev1.VolumeMount{
		Name:      cm.name,
		MountPath: cm.mountPath,
	}
}

func GetImage(version string) string {
	if version == "2.9.2" {
		return QydwdDatanodeImage
	}
	return UhopperDatanodeImage
}
