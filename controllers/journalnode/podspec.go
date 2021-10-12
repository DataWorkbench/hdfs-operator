package journalnode

import (
	v1 "github.com/dataworkbench/hdfs-operator/api/v1"
	com "github.com/dataworkbench/hdfs-operator/common"
	corev1 "k8s.io/api/core/v1"
)

const (
	UhopperImage   = "uhopper/hadoop-namenode:2.7.2"
	QydwdImage     = "qydwd/hadoop-namenode:2.9.2"
)

var defaultOptional = true

// BuildPodTemplateSpec builds a new PodTemplateSpec for  NameNode.
func BuildPodTemplateSpec(hdfs v1.HDFS, labels map[string]string) (corev1.PodTemplateSpec, error) {
	volumes, volumeMounts := buildVolumes(hdfs.Name, hdfs.Spec.Namenode)
	// builde Containers
	container := buildContainer(hdfs.Spec.Journalnode.Name, volumeMounts, hdfs.Spec.Version)

	builder := &com.PodTemplateBuilder{}
	builder.WithContainers(container).
		WithSpecVolumes(volumes...).
		WithRestartPolicy(corev1.RestartPolicyAlways).
		WithHostNetwork(defaultOptional).
		WithDNSPolicy(corev1.DNSClusterFirstWithHostNet).
		WithTemplateMetadata(labels)

	return builder.PodTemplate, nil
}

func buildVolumes(name string, nodeSpec v1.NamenodeSet) (volumes []corev1.Volume, volumeMounts []corev1.VolumeMount) {

	configVolume := com.NewConfigMapVolume(com.GetName(name, com.CommonConfigName),
		com.VolumesConfigMapName,
		com.HdfsConfigMountPath)
	// append container volumeMounts from PVCs eg: metadatadir
	persistentVolumes := make([]corev1.VolumeMount, 0, len(nodeSpec.VolumeClaimTemplates))
	for _, claimTemplate := range nodeSpec.VolumeClaimTemplates {
		persistentVolumes = append(persistentVolumes,
			corev1.VolumeMount{
				Name:      claimTemplate.Name,
				SubPath:   "journal",
				MountPath: "/hadoop/dfs/journal",
			},
			corev1.VolumeMount{
				Name:      claimTemplate.Name,
				SubPath:   "name",
				MountPath: "/hadoop/dfs/name",
			})
	}

	volumes = append(volumes, configVolume.Volume())
	volumeMounts = append(persistentVolumes,
		append(volumeMounts, configVolume.VolumeMount())...)

	return volumes, volumeMounts
}

func buildContainer(name string, volumeMounts []corev1.VolumeMount, version string) corev1.Container {
	defaultContainerPorts := getDefaultContainerPorts()
	return corev1.Container{
		ImagePullPolicy: corev1.PullIfNotPresent,
		Image:           GetImage(version),
		Name:            name,
		Env:             envVars(),
		Command:         []string{"/entrypoint.sh"},
		//Args:            []string{"/opt/hadoop-2.7.2/bin/hdfs", "--config", "/etc/hadoop", "journalnode"},
		Args:            []string{"/opt/hadoop-"+version+"/bin/hdfs", "--config", "/etc/hadoop", "journalnode"},
		Ports:           defaultContainerPorts,
		VolumeMounts:    volumeMounts,
	}
}

func envVars() []corev1.EnvVar {
	return []corev1.EnvVar{
		{Name: "HADOOP_CUSTOM_CONF_DIR", Value: "/etc/hadoop-custom-conf"},
	}
}

func GetDefaultServicePorts() []corev1.ServicePort {
	return []corev1.ServicePort{
		{Name: "jn", Port: 8485},
		{Name: "http", Port: 8480},
	}
}

func getDefaultContainerPorts() []corev1.ContainerPort {
	return []corev1.ContainerPort{
		{Name: "jn", ContainerPort: 8485},
		{Name: "http", ContainerPort: 8480},
	}
}

func GetImage(version string) string {
	if version == "2.9.2" {
		return QydwdImage
	}
	return UhopperImage
}
