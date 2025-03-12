package tee_backend

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/manatee-project/manatee/app/api/biz/dal/db"
	"github.com/pkg/errors"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type MockTeeBackend struct {
	ctx       context.Context
	clientSet *kubernetes.Clientset
	namespace string
}

func NewMockTeeBackend(ctx context.Context) (*MockTeeBackend, error) {
	clusterConfig, err := rest.InClusterConfig()
	if err != nil {
		return nil, errors.Wrap(err, "failed to init cluster config")
	}

	clientSet, err := kubernetes.NewForConfig(clusterConfig)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create client")
	}

	RunningNameSpaceByte, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		return nil, errors.Wrap(err, "failed to get namespace")
	}
	namespace := string(RunningNameSpaceByte)

	return &MockTeeBackend{
		ctx:       ctx,
		clientSet: clientSet,
		namespace: namespace,
	}, nil
}

func (m *MockTeeBackend) LaunchInstance(instanceName string, j *db.Job) error {
	ttlSecondsAfterFinished := int32(3600 * 3)
	var envs []corev1.EnvVar
	for key, value := range j.ExtraEnvs {
		envs = append(envs, corev1.EnvVar{
			Name:  key,
			Value: value,
		})
	}
	envs = append(envs, corev1.EnvVar{
		Name:  "TEE_BACKEND",
		Value: os.Getenv("TEE_BACKEND"),
	},
	)
	mockTeeJob := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instanceName,
			Namespace: m.namespace,
		},
		Spec: batchv1.JobSpec{
			TTLSecondsAfterFinished: &ttlSecondsAfterFinished,
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					ServiceAccountName: "dcr-k8s-pod-sa",
					Containers: []corev1.Container{
						{
							Name:  "mock-tee",
							Image: convertImageToLocal(j.DockerImage),
							Env:   envs,
						},
					},
					RestartPolicy: "Never",
				},
			},
		},
	}
	_, err := m.clientSet.BatchV1().Jobs(m.namespace).Create(m.ctx, mockTeeJob, metav1.CreateOptions{})
	if err != nil {
		return errors.Wrap(err, "failed to create kubernetes job")
	}
	return nil
}

func (m *MockTeeBackend) CleanUpInstance(instanceName string) error {
	deletePolicy := metav1.DeletePropagationForeground
	if err := m.clientSet.BatchV1().Jobs(m.namespace).Delete(m.ctx, instanceName, metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}); err != nil {
		return errors.Wrap(err, "failed to delete job")
	}
	return nil
}

func (m *MockTeeBackend) GetInstanceStatus(instanceName string) (string, error) {
	teeJob, err := m.clientSet.BatchV1().Jobs(m.namespace).Get(m.ctx, instanceName, metav1.GetOptions{})
	if err != nil {
		hlog.Errorf("[MockTeeBackend]failed to get mock tee job: %v", err)
		return "", errors.Wrap(err, "failed to get job")
	}
	hlog.Infof("[MockTeeBackend]mock tee name: %v, status: %v", teeJob.Name, teeJob.Status)
	if teeJob.Status.Active > 0 {
		return "RUNNING", nil
	} else {
		return "TERMINATED", nil
	}
}

func convertImageToLocal(imageName string) string {
	index := strings.Index(imageName, "/")
	if index == -1 {
		hlog.Errorf("[MockTeeBackend]failed to find / in image name")
		return ""
	}
	return fmt.Sprintf("localhost:5000/%s", imageName[index+1:])
}
