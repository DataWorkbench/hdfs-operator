package yarn

import (
	v1 "github.com/dataworkbench/hdfs-operator/api/v1"
	com "github.com/dataworkbench/hdfs-operator/common"
	corev1 "k8s.io/api/core/v1"
)

var defaultOptional = true

// BuildRMPodTemplate builds a new PodTemplateSpec for NameNode.
func BuildRMPodTemplate(hdfs v1.HDFS, labels map[string]string) (corev1.PodTemplateSpec, error) {
	volumes, volumeMounts := buildVolumes(hdfs.Name)

	container := buildRMContainer(com.GetName(hdfs.Name, hdfs.Spec.Yarn.Name), volumeMounts,hdfs.Spec.Version, hdfs.Spec.Image)

	builder := &com.PodTemplateBuilder{} //NewPodTemplateBuilder()
	builder.WithContainers(container).
		WithSpecVolumes(volumes...).
		WithRestartPolicy(corev1.RestartPolicyAlways).
		//WithHostNetwork(defaultOptional).
		//WithDNSPolicy(corev1.DNSClusterFirstWithHostNet).
		WithTemplateMetadata(labels)

	return builder.PodTemplate, nil
}

// BuildNMPodTemplate builds a new PodTemplateSpec for NameNode.
func BuildNMPodTemplate(hdfs v1.HDFS, labels map[string]string) (corev1.PodTemplateSpec, error) {
	volumes, volumeMounts := buildVolumes(hdfs.Name)

	container := buildNMContainer(com.GetName(hdfs.Name, hdfs.Spec.Yarn.Name), volumeMounts,hdfs.Spec.Version, hdfs.Spec.Image)

	builder := &com.PodTemplateBuilder{} //NewPodTemplateBuilder()
	builder.WithContainers(container).
		WithSpecVolumes(volumes...).
		WithRestartPolicy(corev1.RestartPolicyAlways).
		//WithHostNetwork(defaultOptional).
		//WithDNSPolicy(corev1.DNSClusterFirstWithHostNet).
		WithTemplateMetadata(labels)

	return builder.PodTemplate, nil
}

func buildVolumes(name string) (volumes []corev1.Volume, volumeMounts []corev1.VolumeMount) {

	configVolume := com.NewConfigMapVolume(com.GetName(name, YarnConfigName),
		YarnConfigName,
		com.HdfsConfigMountPath)

	volumes = append(volumes, configVolume.Volume())
	volumeMounts = append(volumeMounts,configVolume.VolumeMount())

	return volumes, volumeMounts
}

func buildRMContainer(name string, volumeMounts []corev1.VolumeMount, version string,image string) corev1.Container {
	defaultContainerPorts := getRMContainerPorts()
	return corev1.Container{
		ImagePullPolicy: corev1.PullIfNotPresent,
		Image:           image,
		Name:            name,
		Env:             envVars(),
		Command:         []string{"/entrypoint.sh"},
		Args:            []string{"/opt/hadoop-"+version+"/bin/yarn", "--config", "/etc/hadoop", "resourcemanager"},
		Ports:           defaultContainerPorts,
		VolumeMounts:    volumeMounts,
	}
}

func buildNMContainer(name string, volumeMounts []corev1.VolumeMount, version string,image string) corev1.Container {
	defaultContainerPorts := getNMContainerPorts()
	return corev1.Container{
		ImagePullPolicy: corev1.PullIfNotPresent,
		Image:           image,
		Name:            name,
		Env:             envVars(),
		Command:         []string{"/entrypoint.sh"},
		Args:            []string{"/opt/hadoop-"+version+"/bin/yarn", "--config", "/etc/hadoop", "nodemanager"},
		Ports:           defaultContainerPorts,
		VolumeMounts:    volumeMounts,
	}
}

func GetRMServicePorts() []corev1.ServicePort {
	return []corev1.ServicePort{
		{Name: "web", Port: int32(8088)},
		{Name: "scheduler", Port: int32(8030)},
		{Name: "resource", Port: int32(8031)},
		{Name: "address", Port: int32(8032)},
		{Name: "admin", Port: int32(8033)},
	}
}

func getRMContainerPorts() []corev1.ContainerPort {
	return []corev1.ContainerPort{
		{Name: "web", ContainerPort: int32(8088)},
		{Name: "scheduler", ContainerPort: int32(8030)},
		{Name: "resource", ContainerPort: int32(8031)},
		{Name: "address", ContainerPort: int32(8032)},
		{Name: "admin", ContainerPort: int32(8033)},
	}
}

func GetNMServicePorts() []corev1.ServicePort {
	return []corev1.ServicePort{
		{Name: "api", Port: int32(8040)},
		{Name: "web", Port: int32(8042)},
	}
}

func getNMContainerPorts() []corev1.ContainerPort {
	return []corev1.ContainerPort{
		{Name: "api", ContainerPort: int32(8040)},
		{Name: "web", ContainerPort: int32(8042)},
	}
}

func envVars() []corev1.EnvVar {
	return []corev1.EnvVar{
		{Name: "HADOOP_CUSTOM_CONF_DIR", Value: "/etc/hadoop-custom-conf"},
		{Name: "MULTIHOMED_NETWORK", Value: "0"},
	}
}
