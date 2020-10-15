package main

import (
	"context"
	"fmt"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"log"
	"os"
	"sort"
	"strings"
)

type NamespaceRecord struct {
	Namespace     string
	AppLabels     map[string]string
	ServiceLabels map[string]string
}

func map2str(m map[string]string) string {
	keys := make([]string, 0, len(m))
	sb := &strings.Builder{}
	for k := range m {
		if k == "" {
			continue
		}
		if m[k] == "" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if sb.Len() > 0 {
			sb.WriteRune(',')
		}
		sb.WriteString(k)
		sb.WriteRune('=')
		sb.WriteString(m[k])
	}
	return sb.String()
}

func (r *NamespaceRecord) AddServiceLabels(labels map[string]string, v string) (err error) {
	if labels == nil {
		return
	}
	if len(labels) == 0 {
		return
	}
	l := map2str(labels)
	if r.ServiceLabels[l] != "" {
		err = fmt.Errorf("service already existed: %s -> %s", l, r.ServiceLabels[l])
		return
	}
	r.ServiceLabels[l] = v
	return
}

func (r *NamespaceRecord) AddAppLabels(labels map[string]string, v string) (err error) {
	if labels == nil {
		return
	}
	if len(labels) == 0 {
		return
	}
	l := map2str(labels)
	if r.AppLabels[l] != "" {
		err = fmt.Errorf("app already existed: %s -> %s", l, r.AppLabels[l])
		return
	}
	r.AppLabels[l] = v
	return
}

type NamespaceRecordStore struct {
	data map[string]*NamespaceRecord
}

func NewNamespaceRecordStore() *NamespaceRecordStore {
	return &NamespaceRecordStore{
		data: map[string]*NamespaceRecord{},
	}
}

func (s *NamespaceRecordStore) Find(namespace string) (r *NamespaceRecord) {
	r = s.data[namespace]
	if r == nil {
		r = &NamespaceRecord{
			Namespace:     namespace,
			ServiceLabels: map[string]string{},
			AppLabels:     map[string]string{},
		}
		s.data[namespace] = r
	}
	return r
}

func exit(err *error) {
	if *err != nil {
		log.Println("exited with error:", (*err).Error())
		os.Exit(1)
	} else {
		log.Println("exited")
	}
}

func main() {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ltime | log.Lmsgprefix)

	var err error
	defer exit(&err)

	var cfg *rest.Config
	if cfg, err = rest.InClusterConfig(); err != nil {
		return
	}
	var client *kubernetes.Clientset
	if client, err = kubernetes.NewForConfig(cfg); err != nil {
		return
	}

	store := NewNamespaceRecordStore()

	var services *corev1.ServiceList
	if services, err = client.CoreV1().Services("").List(context.Background(), metav1.ListOptions{}); err != nil {
		return
	}

	for _, service := range services.Items {
		r := store.Find(service.Namespace)
		if errLocal := r.AddServiceLabels(service.Spec.Selector, service.Name); errLocal != nil {
			log.Printf("%s: service %s: %s", service.Namespace, service.Name, errLocal.Error())
		}
	}

	var deployments *appv1.DeploymentList
	if deployments, err = client.AppsV1().Deployments("").List(context.Background(), metav1.ListOptions{}); err != nil {
		return
	}

	var statefulsets *appv1.StatefulSetList
	if statefulsets, err = client.AppsV1().StatefulSets("").List(context.Background(), metav1.ListOptions{}); err != nil {
		return
	}

	for _, deployment := range deployments.Items {
		r := store.Find(deployment.Namespace)
		if errLocal := r.AddAppLabels(deployment.Spec.Selector.MatchLabels, "deployment/"+deployment.Name); errLocal != nil {
			log.Printf("%s: deployment %s: %s", deployment.Namespace, deployment.Name, errLocal.Error())
		}
	}

	for _, statefulset := range statefulsets.Items {
		r := store.Find(statefulset.Namespace)
		if errLocal := r.AddAppLabels(statefulset.Spec.Selector.MatchLabels, "statefulset/"+statefulset.Name); errLocal != nil {
			log.Printf("%s: statefulset %s: %s", statefulset.Namespace, statefulset.Name, errLocal.Error())
		}
	}
}
