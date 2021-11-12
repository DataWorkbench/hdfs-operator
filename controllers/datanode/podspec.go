package datanode

import (
	v1 "github.com/dataworkbench/hdfs-operator/api/v1"
	com "github.com/dataworkbench/hdfs-operator/common"
	corev1 "k8s.io/api/core/v1"
)

//const (
//	UhopperDatanodeImage   = "uhopper/hadoop-datanode:2.7.2"
//	QydwdDatanodeImage     = "qydwd/hadoop-datanode:2.9.2"
//)

// BuildPodTemplateSpec builds a new PodTemplateSpec for DataNode.
func BuildPodTemplateSpec(hdfs v1.HDFS, labels map[string]string) (corev1.PodTemplateSpec, error) {
	volumes, volumeMounts := buildVolumes(hdfs.Name,hdfs.Spec.Datanode)

	container := buildContainer(hdfs.Spec.Datanode.Name, volumeMounts, hdfs.Spec.Version,hdfs.Spec.Image)

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

func buildVolumes(name string, nodeSpec v1.Datanode) (volumes []corev1.Volume, volumeMounts []corev1.VolumeMount) {

	configVolume := com.NewConfigMapVolume(com.GetName(name, com.CommonConfigName), com.VolumesConfigMapName, com.HdfsConfigMountPath)

	scriptsVolume := com.NewConfigMapVolumeWithMode(com.GetName(name, DatanodeScripts), DNScriptsVolumeName, DNScriptsVolumeMountPath, 0744)

	// append container volumeMounts from PVCs
	persistentVolumes := make([]corev1.VolumeMount, 0, len(nodeSpec.Datadirs))
	for _, dir := range nodeSpec.Datadirs {
		persistentVolumes = append(persistentVolumes, corev1.VolumeMount{
			Name:      DNDataVolumeName,
			MountPath: DNDataVolumeMountPath+dir,
			SubPath: dir,
		})
	}

	//SSetSpec.Template.Spec.Volume
	volumes = append(volumes,
		scriptsVolume.Volume(),
		configVolume.Volume(),
	)
	//SSetSpec.Template.Spec.containers.volumeMounts
	volumeMounts = append(persistentVolumes,
		scriptsVolume.VolumeMount(),
		configVolume.VolumeMount(),
	)

	return volumes, volumeMounts
}

func buildContainer(name string, volumeMounts []corev1.VolumeMount, version string,image string) corev1.Container {

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
		Command:         []string{"/entrypoint.sh"},
		Args:            []string{"/opt/hadoop-"+version+"/bin/hdfs", "--config", "/etc/hadoop", "datanode"},
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
