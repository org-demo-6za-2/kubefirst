package k8s

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/ssh/terminal"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
)

func CreateSecretV2(kubeConfigPath string, secret *v1.Secret) error {
	clientset, err := GetClientSet(false, kubeConfigPath)
	if err != nil {
		return err
	}

	_, err = clientset.CoreV1().Secrets(secret.Namespace).Create(
		context.Background(),
		secret,
		metaV1.CreateOptions{},
	)
	if err != nil {
		return err
	}
	log.Info().Msgf("Created Secret %s in Namespace %s\n", secret.Name, secret.Namespace)
	return nil
}

func ReadSecretV2(kubeConfigPath string, namespace string, secretName string) (map[string]string, error) {
	clientset, err := GetClientSet(false, kubeConfigPath)
	if err != nil {
		return map[string]string{}, err
	}

	secret, err := clientset.CoreV1().Secrets(namespace).Get(context.Background(), secretName, metaV1.GetOptions{})
	if err != nil {
		log.Error().Msgf("Error getting secret: %s\n", err)
		return map[string]string{}, nil
	}

	parsedSecretData := make(map[string]string)
	for key, value := range secret.Data {
		parsedSecretData[key] = string(value)
	}

	return parsedSecretData, nil
}

// PodExecSession executes a command against a Pod
func PodExecSession(kubeConfigPath string, p *PodSessionOptions, silent bool) error {
	// v1.PodExecOptions is passed to the rest client to form the req URL
	podExecOptions := v1.PodExecOptions{
		Stdin:   p.Stdin,
		Stdout:  p.Stdout,
		Stderr:  p.Stderr,
		TTY:     p.TtyEnabled,
		Command: p.Command,
	}

	err := podExec(kubeConfigPath, p, podExecOptions, silent)
	if err != nil {
		return err
	}
	return nil
}

// podExec performs kube-exec on a Pod with a given command
func podExec(kubeConfigPath string, ps *PodSessionOptions, pe v1.PodExecOptions, silent bool) error {
	clientset, err := GetClientSet(false, kubeConfigPath)
	if err != nil {
		return err
	}

	config, err := GetClientConfig(false, kubeConfigPath)
	if err != nil {
		return err
	}

	// Format the request to be sent to the API
	req := clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(ps.PodName).
		Namespace(ps.Namespace).
		SubResource("exec")
	req.VersionedParams(&pe, scheme.ParameterCodec)

	// POST op against Kubernetes API to initiate remote command
	exec, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err != nil {
		log.Fatal().Msgf("Error executing command on Pod: %s", err)
		return err
	}

	// Put the terminal into raw mode to prevent it echoing characters twice
	oldState, err := terminal.MakeRaw(0)
	if err != nil {
		log.Fatal().Msgf("Error when attempting to start terminal: %s", err)
		return err
	}
	defer terminal.Restore(0, oldState)

	var showOutput io.Writer
	if silent {
		showOutput = io.Discard
	} else {
		showOutput = os.Stdout
	}
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  os.Stdin,
		Stdout: showOutput,
		Stderr: os.Stderr,
		Tty:    ps.TtyEnabled,
	})
	if err != nil {
		log.Fatal().Msgf("Error running command on Pod: %s", err)
	}
	return nil
}

// ReturnDeploymentObject returns a matching appsv1.Deployment object based on the filters
func ReturnDeploymentObject(kubeConfigPath string, matchLabel string, matchLabelValue string, namespace string, timeoutSeconds float64) (*appsv1.Deployment, error) {
	clientset, err := GetClientSet(false, kubeConfigPath)
	if err != nil {
		return nil, err
	}

	// Filter
	deploymentListOptions := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", matchLabel, matchLabelValue),
	}

	// Create watch operation
	objWatch, err := clientset.
		AppsV1().
		Deployments(namespace).
		Watch(context.Background(), deploymentListOptions)
	if err != nil {
		log.Fatal().Msgf("Error when attempting to search for Deployment: %s", err)
	}
	log.Info().Msgf("Waiting for %s Deployment to be created.", matchLabelValue)

	objChan := objWatch.ResultChan()
	for {
		select {
		case event, ok := <-objChan:
			time.Sleep(time.Second * 1)
			if !ok {
				// Error if the channel closes
				log.Fatal().Msgf("Error waiting for %s Deployment to be created: %s", matchLabelValue, err)
			}
			if event.
				Object.(*appsv1.Deployment).Status.Replicas > 0 {
				spec, err := clientset.AppsV1().Deployments(namespace).List(context.Background(), deploymentListOptions)
				if err != nil {
					log.Fatal().Msgf("Error when searching for Deployment: %s", err)
					return nil, err
				}
				return &spec.Items[0], nil
			}
		case <-time.After(time.Duration(timeoutSeconds) * time.Second):
			log.Error().Msg("The Deployment was not created within the timeout period.")
			return nil, errors.New("The Deployment was not created within the timeout period.")
		}
	}
}

// ReturnPodObject returns a matching v1.Pod object based on the filters
func ReturnPodObject(kubeConfigPath string, matchLabel string, matchLabelValue string, namespace string, timeoutSeconds float64) (*v1.Pod, error) {
	clientset, err := GetClientSet(false, kubeConfigPath)
	if err != nil {
		return nil, err
	}

	// Filter
	podListOptions := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", matchLabel, matchLabelValue),
	}

	// Create watch operation
	objWatch, err := clientset.
		CoreV1().
		Pods(namespace).
		Watch(context.Background(), podListOptions)
	if err != nil {
		log.Fatal().Msgf("Error when attempting to search for Pod: %s", err)
	}
	log.Info().Msgf("Waiting for %s Pod to be created.", matchLabelValue)

	objChan := objWatch.ResultChan()
	for {
		select {
		case event, ok := <-objChan:
			time.Sleep(time.Second * 1)
			if !ok {
				// Error if the channel closes
				log.Fatal().Msgf("Error waiting for %s Pod to be created: %s", matchLabelValue, err)
			}
			if event.
				Object.(*v1.Pod).Status.Phase == "Pending" {
				spec, err := clientset.CoreV1().Pods(namespace).List(context.Background(), podListOptions)
				if err != nil {
					log.Fatal().Msgf("Error when searching for Pod: %s", err)
					return nil, err
				}
				return &spec.Items[0], nil
			}
			if event.
				Object.(*v1.Pod).Status.Phase == "Running" {
				spec, err := clientset.CoreV1().Pods(namespace).List(context.Background(), podListOptions)
				if err != nil {
					log.Fatal().Msgf("Error when searching for Pod: %s", err)
					return nil, err
				}
				return &spec.Items[0], nil
			}
		case <-time.After(time.Duration(timeoutSeconds) * time.Second):
			log.Error().Msg("The Pod was not created within the timeout period.")
			return nil, errors.New("The Pod was not created within the timeout period.")
		}
	}
}

// ReturnStatefulSetObject returns a matching appsv1.StatefulSet object based on the filters
func ReturnStatefulSetObject(kubeConfigPath string, matchLabel string, matchLabelValue string, namespace string, timeoutSeconds float64) (*appsv1.StatefulSet, error) {
	clientset, err := GetClientSet(false, kubeConfigPath)
	if err != nil {
		return nil, err
	}

	// Filter
	statefulSetListOptions := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", matchLabel, matchLabelValue),
	}

	// Create watch operation
	objWatch, err := clientset.
		AppsV1().
		StatefulSets(namespace).
		Watch(context.Background(), statefulSetListOptions)
	if err != nil {
		log.Fatal().Msgf("Error when attempting to search for StatefulSet: %s", err)
	}
	log.Info().Msgf("Waiting for %s StatefulSet to be created.", matchLabelValue)

	objChan := objWatch.ResultChan()
	for {
		select {
		case event, ok := <-objChan:
			time.Sleep(time.Second * 1)
			if !ok {
				// Error if the channel closes
				log.Fatal().Msgf("Error waiting for %s StatefulSet to be created: %s", matchLabelValue, err)
			}
			if event.
				Object.(*appsv1.StatefulSet).Status.Replicas > 0 {
				spec, err := clientset.AppsV1().StatefulSets(namespace).List(context.Background(), statefulSetListOptions)
				if err != nil {
					log.Fatal().Msgf("Error when searching for StatefulSet: %s", err)
					return nil, err
				}
				return &spec.Items[0], nil
			}
		case <-time.After(time.Duration(timeoutSeconds) * time.Second):
			log.Error().Msg("The StatefulSet was not created within the timeout period.")
			return nil, errors.New("The StatefulSet was not created within the timeout period.")
		}
	}
}

// WaitForDeploymentReady waits for a target Deployment to become ready
func WaitForDeploymentReady(kubeConfigPath string, deployment *appsv1.Deployment, timeoutSeconds int64) (bool, error) {
	clientset, err := GetClientSet(false, kubeConfigPath)
	if err != nil {
		return false, err
	}

	// Format list for metav1.ListOptions for watch
	configuredReplicas := deployment.Status.Replicas
	watchOptions := metav1.ListOptions{
		FieldSelector: fmt.Sprintf(
			"metadata.name=%s", deployment.Name),
	}

	// Create watch operation
	objWatch, err := clientset.
		AppsV1().
		Deployments(deployment.ObjectMeta.Namespace).
		Watch(context.Background(), watchOptions)
	if err != nil {
		log.Fatal().Msgf("Error when attempting to wait for Deployment: %s", err)
	}
	log.Info().Msgf("Waiting for %s Deployment to be ready. This could take up to %v seconds.", deployment.Name, timeoutSeconds)

	objChan := objWatch.ResultChan()
	for {
		select {
		case event, ok := <-objChan:
			time.Sleep(time.Second * 1)
			if !ok {
				// Error if the channel closes
				log.Fatal().Msgf("Error waiting for Deployment: %s", err)
			}
			if event.
				Object.(*appsv1.Deployment).
				Status.ReadyReplicas == configuredReplicas {
				log.Info().Msgf("All Pods in Deployment %s are ready.", deployment.Name)
				return true, nil
			}
		case <-time.After(time.Duration(timeoutSeconds) * time.Second):
			log.Error().Msg("The Deployment was not ready within the timeout period.")
			return false, errors.New("The Deployment was not ready within the timeout period.")
		}
	}
}

// WaitForPodReady waits for a target Pod to become ready
func WaitForPodReady(kubeConfigPath string, pod *v1.Pod, timeoutSeconds int64) (bool, error) {
	clientset, err := GetClientSet(false, kubeConfigPath)
	if err != nil {
		return false, err
	}

	// Format list for metav1.ListOptions for watch
	watchOptions := metav1.ListOptions{
		FieldSelector: fmt.Sprintf(
			"metadata.name=%s", pod.Name),
	}

	// Create watch operation
	objWatch, err := clientset.
		CoreV1().
		Pods(pod.ObjectMeta.Namespace).
		Watch(context.Background(), watchOptions)
	if err != nil {
		log.Fatal().Msgf("Error when attempting to wait for Pod: %s", err)
	}
	log.Info().Msgf("Waiting for %s Pod to be ready. This could take up to %v seconds.", pod.Name, timeoutSeconds)

	// Feed events using provided channel
	objChan := objWatch.ResultChan()

	// Listen until the Pod is ready
	// Timeout if it isn't ready within timeoutSeconds
	for {
		select {
		case event, ok := <-objChan:
			if !ok {
				// Error if the channel closes
				log.Error().Msg("fail")
			}
			if event.
				Object.(*v1.Pod).
				Status.
				Phase == "Running" {
				log.Info().Msgf("Pod %s is %s.", pod.Name, event.Object.(*v1.Pod).Status.Phase)
				return true, nil
			}
		case <-time.After(time.Duration(timeoutSeconds) * time.Second):
			log.Error().Msg("The operation timed out while waiting for the Pod to become ready.")
			return false, errors.New("The operation timed out while waiting for the Pod to become ready.")
		}
	}
}

// WaitForStatefulSetReady waits for a target StatefulSet to become ready
func WaitForStatefulSetReady(kubeConfigPath string, statefulset *appsv1.StatefulSet, timeoutSeconds int64, ignoreReady bool) (bool, error) {
	clientset, err := GetClientSet(false, kubeConfigPath)
	if err != nil {
		return false, err
	}

	// Format list for metav1.ListOptions for watch
	configuredReplicas := statefulset.Status.Replicas
	watchOptions := metav1.ListOptions{
		FieldSelector: fmt.Sprintf(
			"metadata.name=%s", statefulset.Name),
	}

	// Create watch operation
	objWatch, err := clientset.
		AppsV1().
		StatefulSets(statefulset.ObjectMeta.Namespace).
		Watch(context.Background(), watchOptions)
	if err != nil {
		log.Fatal().Msgf("Error when attempting to wait for StatefulSet: %s", err)
	}
	log.Info().Msgf("Waiting for %s StatefulSet to be ready. This could take up to %v seconds.", statefulset.Name, timeoutSeconds)

	objChan := objWatch.ResultChan()
	for {
		select {
		case event, ok := <-objChan:
			time.Sleep(time.Second * 1)
			if !ok {
				// Error if the channel closes
				log.Fatal().Msgf("Error waiting for StatefulSet: %s", err)
			}
			if ignoreReady {
				if event.
					Object.(*appsv1.StatefulSet).
					Status.CurrentReplicas == configuredReplicas {
					log.Info().Msgf("All Pods in StatefulSet %s have been created.", statefulset.Name)
					return true, nil
				}
			} else {
				if event.
					Object.(*appsv1.StatefulSet).
					Status.ReadyReplicas == configuredReplicas {
					log.Info().Msgf("All Pods in StatefulSet %s are ready.", statefulset.Name)
					return true, nil
				}
			}

		case <-time.After(time.Duration(timeoutSeconds) * time.Second):
			log.Error().Msg("The StatefulSet was not ready within the timeout period.")
			return false, errors.New("The StatefulSet was not ready within the timeout period.")
		}
	}
}
