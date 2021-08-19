package namenode

import (
	v1 "github.com/dataworkbench/hdfs-operator/api/v1"
	com "github.com/dataworkbench/hdfs-operator/controllers/common"
	corev1 "k8s.io/api/core/v1"
)

const (
	ScriptsVolumeName = "nn-scripts"
	ScriptsVolumeMountPath = "/nn-scripts"
	//StatefulSetLabel  = "github.dataworkbench.com/statefulset-name"
)

// BuildPodTemplateSpec builds a new PodTemplateSpec for NameNode.
func BuildPodTemplateSpec( hdfs v1.HDFS) (corev1.PodTemplateSpec, error) {
	volumes, volumeMounts := buildVolumes(hdfs.Name, hdfs.Spec.Namenode)

	container := buildContainer(hdfs.Spec.Namenode.Name,volumeMounts,"uhopper/hadoop-namenode:2.7.2") //hdfs.version

	builder := com.NewPodTemplateBuilder(hdfs.Spec.Namenode.PodTemplate)

	builder.WithContainers(container).
		WithSpecVolumes(volumes...).
		WithDNSPolicy(corev1.DNSClusterFirst).
		WithRestartPolicy(corev1.RestartPolicyAlways)

	return builder.PodTemplate, nil
}

func buildVolumes(name string, nodeSpec v1.NamenodeSet) (volumes []corev1.Volume,volumeMounts []corev1.VolumeMount) {

	configVolume := com.NewConfigMapVolume(com.GetName(name,com.CommonConfigName),
		com.VolumesConfigMapName,
		com.HdfsConfigMountPath)

	scriptsVolume := com.NewConfigMapVolumeWithMode(com.GetName(name,NamenodeScripts),
		ScriptsVolumeName,
		ScriptsVolumeMountPath,
		0744)

	// append container volumeMounts from PVCs
	persistentVolumes := make([]corev1.VolumeMount, 0, len(nodeSpec.VolumeClaimTemplates))
	for _, claimTemplate := range nodeSpec.VolumeClaimTemplates {
		persistentVolumes = append(persistentVolumes, corev1.VolumeMount{
			Name: claimTemplate.Name,
			MountPath: "/hadoop/dfs/name",
			SubPath: "name",
		})
	}

	volumes = append(volumes, scriptsVolume.Volume(), configVolume.Volume())
	volumeMounts = append(persistentVolumes,
		append(volumeMounts, scriptsVolume.VolumeMount(),configVolume.VolumeMount())...)


	return volumes, volumeMounts
}

func buildContainer( name string,volumeMounts []corev1.VolumeMount,image string ) corev1.Container {
	defaultContainerPorts := getDefaultContainerPorts()
	return corev1.Container{
		ImagePullPolicy: corev1.PullIfNotPresent,
		Image:           image,
		Name:            name,
		Env:             envVars(name),
		Command:         []string{"/bin/sh", "-c"},
		Args:            []string{"/entrypoint.sh\"/nn-scripts/format-and-run.sh\""},
		Ports:           defaultContainerPorts,
		VolumeMounts:    volumeMounts,
	}
}

func GetDefaultServicePorts() []corev1.ServicePort {
	return []corev1.ServicePort{
		{Name: "http", Port:50070},
		{Name: "fs", Port: 8020},
	}
}

func getDefaultContainerPorts() []corev1.ContainerPort {
	return []corev1.ContainerPort{
		{Name: "http", ContainerPort:50070},
		{Name: "fs", ContainerPort: 8020},
	}
}

func envVars(name string) []corev1.EnvVar {
	return []corev1.EnvVar{
		{Name: "HADOOP_CUSTOM_CONF_DIR", Value: "/etc/hadoop-custom-conf"},
		{Name: "MULTIHOMED_NETWORK", Value: "0"},
		{Name: "MY_POD", Value: "", ValueFrom: &corev1.EnvVarSource{
			FieldRef: &corev1.ObjectFieldSelector{ FieldPath: "metadata.name"},
		}},
		{Name: "NAMENODE_POD_0", Value: name+"-0"},
		{Name: "NAMENODE_POD_1", Value: name+"-1"},
	}
}
