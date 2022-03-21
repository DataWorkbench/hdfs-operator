package journalnode

import (
	v1 "github.com/dataworkbench/hdfs-operator/api/v1"
	com "github.com/dataworkbench/hdfs-operator/common"
	corev1 "k8s.io/api/core/v1"
)


var defaultOptional = true

// BuildPodTemplateSpec builds a new PodTemplateSpec for  NameNode.
func BuildPodTemplateSpec(hdfs v1.HDFS, labels map[string]string) (corev1.PodTemplateSpec, error) {
	volumes, volumeMounts := buildVolumes(hdfs.Name, hdfs.Spec.Namenode)
	// builde Containers
	container := buildContainer(hdfs.Spec.Journalnode.Name, volumeMounts, hdfs)

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

	volumes = append(volumes, configVolume.Volume())
	volumeMounts = append(volumeMounts, configVolume.VolumeMount(),
		corev1.VolumeMount{
			Name:      JNEditDataPvcName,
			SubPath:   "journal",
			MountPath: "/hadoop/dfs/journal",
		},corev1.VolumeMount{
			Name:      JNEditDataPvcName,
			SubPath:   "name",
			MountPath: "/hadoop/dfs/name",
		})


	return volumes, volumeMounts
}

func buildContainer(name string, volumeMounts []corev1.VolumeMount, hdfs v1.HDFS) corev1.Container {
	defaultContainerPorts := getDefaultContainerPorts()
	return corev1.Container{
		ImagePullPolicy: corev1.PullPolicy(hdfs.Spec.ImagePullPolicy),
		Image:           hdfs.Spec.Image,
		Name:            name,
		Env:             envVars(),
		Command:         []string{"/entrypoint.sh"},
		Args:            []string{"/opt/hadoop-"+hdfs.Spec.Version+"/bin/hdfs", "--config", "/etc/hadoop", "journalnode"},
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
