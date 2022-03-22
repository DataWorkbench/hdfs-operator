package common

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sort"
)

var defaultOptional = false

func AppendPVCs( name string, sc string, ca string) ( pvcs []corev1.PersistentVolumeClaim ) {

	defaultVolumeClaim := corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			StorageClassName: &sc,
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse(ca),
				},
			},
		},
	}
	pvcs = append(pvcs, defaultVolumeClaim)
	return pvcs
}

func AppendDefaultPVCs(existing []corev1.PersistentVolumeClaim, name string, sc string) []corev1.PersistentVolumeClaim {
	if len(existing) > 0 {
		return existing
	}
	defaultVolumeClaim := corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			StorageClassName: &sc,
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse("10Gi"),
				},
			},
		},
	}
	existing = append(existing, defaultVolumeClaim)
	return existing
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
	PodTemplate corev1.PodTemplateSpec
	//containerDefaulter *corev1.Container
}


func (b *PodTemplateBuilder) WithContainers(container corev1.Container) *PodTemplateBuilder {
	b.PodTemplate.Spec.Containers = append(b.PodTemplate.Spec.Containers, container)
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

func (b *PodTemplateBuilder) WithImagePullSecrets(secrets ...string) *PodTemplateBuilder {
	for _, v := range secrets {
		b.PodTemplate.Spec.ImagePullSecrets = append(b.PodTemplate.Spec.ImagePullSecrets,
			corev1.LocalObjectReference{
			   Name: v,
			})
	}
	return b
}

func (b *PodTemplateBuilder) WithDNSPolicy(dnsPolicy corev1.DNSPolicy) *PodTemplateBuilder {
	if b.PodTemplate.Spec.DNSPolicy == "" {
		b.PodTemplate.Spec.DNSPolicy = dnsPolicy
	}
	return b
}

func (b *PodTemplateBuilder) WithRestartPolicy(restartPolicy corev1.RestartPolicy) *PodTemplateBuilder {
	if b.PodTemplate.Spec.RestartPolicy == "" {
		b.PodTemplate.Spec.RestartPolicy = restartPolicy
	}
	return b
}

func (b *PodTemplateBuilder) WithHostNetwork(hostNetwork bool) *PodTemplateBuilder {
	b.PodTemplate.Spec.HostNetwork = hostNetwork
	return b
}

func (b *PodTemplateBuilder) WithHostPID(hostPID bool) *PodTemplateBuilder {
	b.PodTemplate.Spec.HostPID = hostPID
	return b
}

func (b *PodTemplateBuilder) WithTemplateMetadata(labels map[string]string) *PodTemplateBuilder {
	b.PodTemplate.Labels = labels
	return b
}
