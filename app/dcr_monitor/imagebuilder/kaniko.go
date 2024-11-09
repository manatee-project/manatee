package imagebuilder

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"regexp"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/manatee-project/manatee/app/dcr_api/biz/dal/db"
	"github.com/pkg/errors"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type ImageInfo struct {
	Image  string
	Digest string
}

type ImageBuilder interface {
	BuildImage(*db.Job, string, string, string) error
	CheckImageBuilderStatusAndGetInfo(string) (bool, *ImageInfo, error)
}

type KanikoImageBuilder struct {
	ctx       context.Context
	clientSet *kubernetes.Clientset
	namespace string
}

func NewKanikoImageBuilder() (*KanikoImageBuilder, error) {
	var err error
	clusterConfig, err := rest.InClusterConfig()
	if err != nil {
		panic(err)
	}

	clientSet, err := kubernetes.NewForConfig(clusterConfig)
	if err != nil {
		panic(err)
	}

	RunningNameSpaceByte, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		panic(err)
	}
	namespace := string(RunningNameSpaceByte)

	return &KanikoImageBuilder{
		ctx:       context.Background(),
		clientSet: clientSet,
		namespace: namespace,
	}, nil
}

func (b *KanikoImageBuilder) CheckImageBuilderStatusAndGetInfo(uuid string) (bool, *ImageInfo, error) {

	k8sJobName := "kaniko-" + uuid
	k8sJob, err := b.clientSet.BatchV1().Jobs(b.namespace).Get(b.ctx, k8sJobName, metav1.GetOptions{})
	if err != nil {
		hlog.Errorf("[KanikoJobMonitor]failed to get job: %v", err)
		return false, nil, errors.Wrap(err, "failed to get job")
	}

	if len(k8sJob.Status.Conditions) == 0 {
		hlog.Infof("[KanikoJobMonitor]job %v is still running", k8sJob.Name)
		return false, nil, nil
	}

	hlog.Infof("[KanikoJobMonitor]job name: %v, job status: %v", k8sJob.Name, k8sJob.Status.Conditions[0].Type)

	if k8sJob.Status.Conditions[0].Type == batchv1.JobComplete {
		image, digest, err := b.getImageDigest(k8sJob.Name)
		if err != nil {
			hlog.Errorf("[KanikoJobMonitor] failed to get image digest: %+v", err)
			return false, nil, err
		}
		hlog.Infof("Image build done: %s@sha256:%s", image, digest)

		b.deleteJob(k8sJob.Name)

		return true, &ImageInfo{Image: image, Digest: digest}, nil
	} else if k8sJob.Status.Conditions[0].Type == batchv1.JobFailed {
		return true, nil, nil
	}
	return false, nil, nil
}

func (b *KanikoImageBuilder) getImageDigest(jobName string) (string, string, error) {
	pods, err := b.clientSet.CoreV1().Pods(b.namespace).List(b.ctx, metav1.ListOptions{
		LabelSelector: "job-name=" + jobName,
	})
	if err != nil {
		return "", "", errors.Wrap(err, "failed to list pod")
	}
	hlog.Infof("[KanikoJobMonitor] pods num: %d", len(pods.Items))
	for _, pod := range pods.Items {
		req := b.clientSet.CoreV1().Pods(b.namespace).GetLogs(pod.Name, &corev1.PodLogOptions{})
		logs, err := req.Stream(b.ctx)
		if err != nil {
			return "", "", errors.Wrap(err, "failed to read log stream")
		}

		image, digest, err := b.getImageAndDigestFromLog(logs)
		if err != nil {
			return "", "", err
		}
		_ = logs.Close()
		if digest == "" {
			continue
		}
		hlog.Infof("[KanikoJobMonitor]got image digest %v", digest)

		return image, digest, nil
	}
	return "", "", errors.New("failed to read digest")
}

func (b *KanikoImageBuilder) getImageAndDigestFromLog(reader io.Reader) (string, string, error) {
	var lastLine string
	scanner := bufio.NewScanner(reader)
	// get the last line
	for scanner.Scan() {
		lastLine = scanner.Text()
	}
	if scanner.Err() != nil {
		return "", "", errors.Wrap(scanner.Err(), "failed to get last line")
	}

	// Regular expression to match any URL format before the digest
	r := regexp.MustCompile(`([^@\s]+)@sha256:([a-z0-9]+)`)
	matches := r.FindStringSubmatch(lastLine)

	if len(matches) < 3 {
		return "", "", errors.New("failed to parse image and digest from log")
	}

	return matches[1] + "@sha256:" + matches[2], matches[2], nil
}
func (b *KanikoImageBuilder) deleteJob(name string) error {
	// hlog.Infof("[KanikoJobMonitor]delete job: %v", name)
	deletePolicy := metav1.DeletePropagationForeground
	if err := b.clientSet.BatchV1().Jobs(b.namespace).Delete(b.ctx, name, metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}); err != nil {
		return errors.Wrap(err, "failed to delete job")
	}
	return nil
}

func (b *KanikoImageBuilder) BuildImage(j *db.Job, bucket string, baseImage string, image string) error {

	buildArgs := []string{
		fmt.Sprintf("--context=%s", j.BuildContextPath),
		fmt.Sprintf("--destination=%s", image),
		// TODO: this should be signed URL (See https://github.com/manatee-project/manatee/issues/23)
		fmt.Sprintf("--build-arg=OUTPUTPATH=%s", fmt.Sprintf("gs://%s/%s/output/out-%s-%s", bucket, j.Creator, j.UUID[:8], j.JupyterFileName)),
		fmt.Sprintf("--build-arg=JUPYTER_FILENAME=%s", j.JupyterFileName),
		fmt.Sprintf("--build-arg=USER_WORKSPACE=%s", fmt.Sprintf("%s-workspace", j.Creator)),
		fmt.Sprintf("--build-arg=BASE_IMAGE=%s", baseImage),
		// TODO: this should be signed URL (See https://github.com/manatee-project/manatee/issues/23)
		fmt.Sprintf("--build-arg=CUSTOMTOKEN_CLOUDSTORAGE_PATH=%s", fmt.Sprintf("gs://%s/%s/output/%s-token", j.Creator, j.UUID)),
	}
	kanikoJobName := fmt.Sprintf("kaniko-%s", j.UUID)
	err := b.createBuildJob(kanikoJobName, buildArgs)
	if err != nil {
		return err
	}
	return nil
}

func (b *KanikoImageBuilder) createBuildJob(jobName string, buildArgs []string) error {
	memQuantity, err := resource.ParseQuantity("6000M")
	if err != nil {
		return errors.Wrap(err, "failed to parse mem quantity")
	}
	ttlSecondsAfterFinished := int32(3600 * 24)
	kanikoJob := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: b.namespace,
		},
		Spec: batchv1.JobSpec{
			TTLSecondsAfterFinished: &ttlSecondsAfterFinished,
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					ServiceAccountName: "dcr-k8s-pod-sa",
					Containers: []corev1.Container{
						{
							Name:  "kaniko",
							Image: "gcr.io/kaniko-project/executor:latest",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceMemory: memQuantity,
								},
							},
							Args: append([]string{
								"--dockerfile=Dockerfile",
								"--reproducible",
								"--compressed-caching=false",
								"--cache=true",
								"--cache-ttl=72h",
							}, buildArgs...),
						},
					},
					RestartPolicy: "Never",
				},
			},
		},
	}
	_, err = b.clientSet.BatchV1().Jobs(b.namespace).Create(b.ctx, kanikoJob, metav1.CreateOptions{})
	if err != nil {
		return errors.Wrap(err, "failed to create kubernetes job")
	}
	return nil
}
