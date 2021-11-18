package namenode

import (
	v1 "github.com/dataworkbench/hdfs-operator/api/v1"
	com "github.com/dataworkbench/hdfs-operator/common"
	corev1 "k8s.io/api/core/v1"
)

const (
	ScriptsVolumeName      = "nn-scripts"
	ScriptsVolumeMountPath = "/nn-scripts"
)

var defaultOptional = true

// BuildPodTemplateSpec builds a new PodTemplateSpec for NameNode.
func BuildPodTemplateSpec(hdfs v1.HDFS, labels map[string]string) (corev1.PodTemplateSpec, error) {
	volumes, volumeMounts := buildVolumes(hdfs.Name)

	container := buildContainer(com.GetName(hdfs.Name, hdfs.Spec.Namenode.Name), volumeMounts, hdfs.Spec.Image)

	builder := &com.PodTemplateBuilder{} //NewPodTemplateBuilder()
	builder.WithContainers(container).
		WithSpecVolumes(volumes...).
		WithRestartPolicy(corev1.RestartPolicyAlways).
		WithHostNetwork(defaultOptional).
		WithDNSPolicy(corev1.DNSClusterFirstWithHostNet).
		WithTemplateMetadata(labels)

	return builder.PodTemplate, nil
}

func buildVolumes(name string) (volumes []corev1.Volume, volumeMounts []corev1.VolumeMount) {

	configVolume := com.NewConfigMapVolume(com.GetName(name, com.CommonConfigName),
		com.VolumesConfigMapName,
		com.HdfsConfigMountPath)

	scriptsVolume := com.NewConfigMapVolumeWithMode(com.GetName(name, NamenodeScripts),
		ScriptsVolumeName,
		ScriptsVolumeMountPath,
		0744)

	volumes = append(volumes, scriptsVolume.Volume(), configVolume.Volume())
	volumeMounts = append(volumeMounts, scriptsVolume.VolumeMount(),
		configVolume.VolumeMount(),
		corev1.VolumeMount{
			Name:      NNMetaDataPvcName,
			MountPath: NNMetaDataVolumeMountPath,
			SubPath:   "name",
		})

	return volumes, volumeMounts
}

func buildContainer(name string, volumeMounts []corev1.VolumeMount, image string) corev1.Container {
	defaultContainerPorts := getDefaultContainerPorts()
	return corev1.Container{
		ImagePullPolicy: corev1.PullIfNotPresent,
		Image:           image,
		Name:            name,
		Env:             envVars(name),
		Command:         []string{"/bin/sh", "-c"},
		//Args:            []string{"while true; do echo hello; sleep 10;done"},
		Args:            []string{"/entrypoint.sh \"/nn-scripts/format-and-run.sh\"" },
		Ports:           defaultContainerPorts,
		VolumeMounts:    volumeMounts,
	}
}

func GetDefaultServicePorts() []corev1.ServicePort {
	return []corev1.ServicePort{
		{Name: "http", Port: int32(com.NamenodeHttpPort)},
		{Name: "fs", Port: int32(com.NamenodeRpcPort)},
	}
}

func getDefaultContainerPorts() []corev1.ContainerPort {
	return []corev1.ContainerPort{
		{Name: "http", ContainerPort: int32(com.NamenodeHttpPort)},
		{Name: "fs", ContainerPort: int32(com.NamenodeRpcPort)},
	}
}

func envVars(name string) []corev1.EnvVar {
	return []corev1.EnvVar{
		{Name: "HADOOP_CUSTOM_CONF_DIR", Value: "/etc/hadoop-custom-conf"},
		{Name: "MULTIHOMED_NETWORK", Value: "0"},
		{Name: "MY_POD", Value: "", ValueFrom: &corev1.EnvVarSource{
			FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.name"},
		}},
		{Name: "NAMENODE_POD_0", Value: name + "-0"},
		{Name: "NAMENODE_POD_1", Value: name + "-1"},
	}
}
