package controllers

import (
	v1 "github.com/dataworkbench/hdfs-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// BuildPodTemplateSpec builds a new PodTemplateSpec for an Elasticsearch node.
func BuildPodTemplateSpec(c client.Client, hdfs v1.HDFS) (corev1.PodTemplateSpec, error) {
	volumes, volumeMounts := buildVolumes(hdfs.Name, hdfs.Spec.Namenode)

	builder := NewPodTemplateBuilder(hdfs.Spec.Namenode.PodTemplate, hdfs)

	// builde Containers
	builder.WithContainers(hdfs,volumeMounts, "uhopper/hadoop-namenode:2.7.2").
	    WithSpecVolumes(volumes...)

	return builder.PodTemplate, nil
}


func buildVolumes(hdfsName string, nodeSpec v1.NamenodeSet) (volumes []corev1.Volume,volumeMounts []corev1.VolumeMount) {

	configVolume := NewConfigMapVolume(
		hdfsName, //configMapName  eg:hdfs-config
		ConfigVolumeName,
		ConfigVolumeMountPath)

	scriptsVolume := NewConfigMapVolumeWithMode(
		hdfsName, //configMapName  eg:my-hdfs-namenode-scripts
		ScriptsVolumeName,
		ScriptsVolumeMountPath,
		0744)

	// append future volumes from PVCs (not resolved to a claim yet) eg: metadatadir
	persistentVolumes := make([]corev1.Volume, 0, len(nodeSpec.VolumeClaimTemplates))
	for _, claimTemplate := range nodeSpec.VolumeClaimTemplates {
		persistentVolumes = append(persistentVolumes, corev1.Volume{
			Name: claimTemplate.Name,
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					// actual claim name will be resolved and fixed right before pod creation
					ClaimName: "claim-name-placeholder",
				},
			},
		})
	}

	volumes = append(
		persistentVolumes, // includes the data volume, unless specified differently in the pod template
		append(
			volumes,
			scriptsVolume.Volume(),
			configVolume.Volume(),
		)...)

	volumeMounts = append(
		volumeMounts,
		scriptsVolume.VolumeMount(),
		configVolume.VolumeMount(),
	)

	return volumes, volumeMounts
}


func getDefaultContainerPorts() []corev1.ContainerPort {
	return []corev1.ContainerPort{
		{Name: "http", ContainerPort:50070},
		{Name: "fs", ContainerPort: 8020},
	}
}

func getDefaultServicePorts() []corev1.ServicePort {
	return []corev1.ServicePort{
		{Name: "http", Port:50070},
		{Name: "fs", Port: 8020},
	}
}


