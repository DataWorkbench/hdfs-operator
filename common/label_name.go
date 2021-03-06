package common

import (
	v1 "github.com/dataworkbench/hdfs-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"strings"
)

const (
	TypeLabelName        = "dataomnis.io/type"
	ClusterNameLabelName = "dataomnis.io/cluster-name"
	Type                 = "hdfs"
	StatefulSetLabel     = "dataomnis.io/statefulset-name"
)

// ExtractNamespacedName returns an NamespacedName based on the given Object.
func ExtractNamespacedName(object metav1.Object) types.NamespacedName {
	return types.NamespacedName{
		Namespace: object.GetNamespace(),
		Name:      object.GetName(),
	}
}

func NewStatefulSetLabels(hdfs types.NamespacedName, ssetName string) map[string]string {
	lbls := NewLabels(hdfs)
	lbls[StatefulSetLabel] = ssetName
	return lbls
}

// NewLabels constructs a new set of labels from an HDFS definition.
func NewLabels(hdfs types.NamespacedName) map[string]string {
	return map[string]string{
		ClusterNameLabelName: hdfs.Name,
		TypeLabelName:        Type,
	}
}

// HeadlessService returns a headless service for the given StatefulSet
func HeadlessService(hdfs v1.HDFS, ssetName string, ports []corev1.ServicePort) corev1.Service {
	nsn := ExtractNamespacedName(&hdfs)
	return corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: nsn.Namespace,
			Name:      ssetName,
			Labels:    NewStatefulSetLabels(nsn, ssetName),
			OwnerReferences: GetOwnerReference(hdfs),
		},
		Spec: corev1.ServiceSpec{
			Type:      corev1.ServiceTypeClusterIP,
			ClusterIP: corev1.ClusterIPNone,
			Selector:  NewStatefulSetLabels(nsn, ssetName),
			Ports:     ports,
		},
	}
}

func GetName(hn, name string) string {
	var result strings.Builder
	result.WriteString(hn)
	result.WriteString("-")
	result.WriteString(name)
	return result.String()
}

func GetOwnerReference(hdfs v1.HDFS) []metav1.OwnerReference {
	return []metav1.OwnerReference{
		*metav1.NewControllerRef(&hdfs, schema.GroupVersionKind{
			Group:   v1.GroupVersion.Group,
			Version: v1.GroupVersion.Version,
			Kind:    "HDFS",
		}),
	}
}
