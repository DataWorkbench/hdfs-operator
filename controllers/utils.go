package controllers

import (
	v1 "github.com/dataworkbench/hdfs-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sort"
)
var (
	defaultOptional = false
)
const (
	TypeLabelName = "qy.dataworkbench.com/type"
	ClusterNameLabelName = "qy.dataworkbench.com/cluster-name"
	StatefulSetNameLabelName = "qy.dataworkbench.com/statefulset-name"
	Type = "hdfs"
	ConfigVolumeName = "hdfs-config"
	ConfigVolumeMountPath = "/etc/hadoop-custom-conf"
	ScriptsVolumeName = "nn-scripts"
	ScriptsVolumeMountPath = "/nn-scripts"
	)

// NewLabels constructs a new set of labels from an Elasticsearch definition.
func NewLabels(hdfs types.NamespacedName) map[string]string {
	return map[string]string{
		ClusterNameLabelName: hdfs.Name,
		TypeLabelName: Type,
	}
}

// AppendDefaultPVCs appends defaults PVCs to a set of existing ones.
func AppendDefaultPVCs(
	existing []corev1.PersistentVolumeClaim,
	podSpec corev1.PodSpec,
	defaults ...corev1.PersistentVolumeClaim,
) []corev1.PersistentVolumeClaim {
	// any user defined PVC shortcuts the defaulting attempt
	if len(existing) > 0 {
		return existing
	}

	// create a set of volume names that are not PVC-volumes
	nonPVCvolumes := Make()

	for _, volume := range podSpec.Volumes {
		if volume.PersistentVolumeClaim == nil {
			// this volume is not a PVC
			nonPVCvolumes.Add(volume.Name)
		}
	}

	for _, defaultPVC := range defaults {
		if nonPVCvolumes.Has(defaultPVC.Name) {
			continue
		}
		existing = append(existing, defaultPVC)
	}
	return existing
}

type StringSet map[string]struct{}

func Make(strings ...string) StringSet {
	set := make(map[string]struct{}, len(strings))
	for _, str := range strings {
		set[str] = struct{}{}
	}
	return set
}

func (set StringSet) Add(s string) {
	set[s] = struct{}{}
}

func (set StringSet) Has(s string) (exists bool) {
	if set != nil {
		_, exists = set[s]
	}
	return
}

// ConfigMapVolume defines a volume to expose a configmap
type ConfigMapVolume struct {
	configMapName string
	name          string
	mountPath     string
	items         []corev1.KeyToPath
	defaultMode   int32
}

// Volume returns the k8s volume.
func (cm ConfigMapVolume) Volume() corev1.Volume {
	return corev1.Volume{
		Name: cm.name,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: cm.configMapName,
				},
				Items:       cm.items,
				Optional:    &defaultOptional,
				DefaultMode: &cm.defaultMode,
			},
		},
	}
}

// VolumeMount returns the k8s volume mount.
func (cm ConfigMapVolume) VolumeMount() corev1.VolumeMount {
	return corev1.VolumeMount{
		Name:      cm.name,
		MountPath: cm.mountPath,
		ReadOnly:  true,
	}
}

// NewConfigMapVolumeWithMode creates a new ConfigMapVolume struct with default mode
func NewConfigMapVolumeWithMode(configMapName, name, mountPath string, defaultMode int32) ConfigMapVolume {
	return ConfigMapVolume{
		configMapName: configMapName,
		name:          name,
		mountPath:     mountPath,
		defaultMode:   defaultMode,
	}
}

// NewConfigMapVolume creates a new ConfigMapVolume
func NewConfigMapVolume(configMapName, name, mountPath string) ConfigMapVolume {
	return ConfigMapVolume{
		configMapName: configMapName,
		name:          name,
		mountPath:     mountPath,
	}
}

// PodTemplateBuilder helps with building a pod template inheriting values
// from a user-provided pod template. It focuses on building a pod with
// one main Container.
type PodTemplateBuilder struct {
	PodTemplate        corev1.PodTemplateSpec
	containerDefaulter *corev1.Container
}

// NewPodTemplateBuilder returns an initialized PodTemplateBuilder
func NewPodTemplateBuilder(base corev1.PodTemplateSpec, hdfs v1.HDFS) *PodTemplateBuilder {
	builder := &PodTemplateBuilder{
		PodTemplate:   *base.DeepCopy(),
	}
	return builder
}

func (b *PodTemplateBuilder) WithContainers(hdfs v1.HDFS,volumeMounts []corev1.VolumeMount,image string) *PodTemplateBuilder {

	defaultContainerPorts := getDefaultContainerPorts()

	container := corev1.Container{
		ImagePullPolicy: corev1.PullIfNotPresent,
		Image:           image,
		Name:            hdfs.Spec.Namenode.Name,
		Env:             NameNodeEnvVars(hdfs.Spec.Namenode.Name),
		Command:         []string{"/bin/sh", "-c"},
		Ports:           defaultContainerPorts,
		VolumeMounts:    volumeMounts,
		//Resources: defaultResources,
	}

	b.PodTemplate.Spec.Containers = append(b.PodTemplate.Spec.Containers,container)

	return b

}

// WithSpecVolumes appends the given volumes to the Container, unless already provided in the template.
func (b *PodTemplateBuilder) WithSpecVolumes(volumes ...corev1.Volume) *PodTemplateBuilder {
	for _, v := range volumes {
			b.PodTemplate.Spec.Volumes = append(b.PodTemplate.Spec.Volumes, v)
	}
	// order volumes by name to ensure stable pod spec comparison
	sort.SliceStable(b.PodTemplate.Spec.Volumes, func(i, j int) bool {
		return b.PodTemplate.Spec.Volumes[i].Name < b.PodTemplate.Spec.Volumes[j].Name
	})
	return b
}

func NameNodeEnvVars(name string) []corev1.EnvVar {
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

// ExtractNamespacedName returns an NamespacedName based on the given Object.
func ExtractNamespacedName(object metav1.Object) types.NamespacedName {
	return types.NamespacedName{
		Namespace: object.GetNamespace(),
		Name:      object.GetName(),
	}
}

func NewStatefulSetLabels(es types.NamespacedName, ssetName string) map[string]string {
	lbls := NewLabels(es)
	lbls[StatefulSetNameLabelName] = ssetName
	return lbls
}