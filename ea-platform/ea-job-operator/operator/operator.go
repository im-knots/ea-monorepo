package operator

import (
	"context"
	"time"

	"ea-job-operator/logger"
	"ea-job-operator/metrics"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// WatchNewAgentJobs monitors AgentJob CRs and updates blank statuses to "inactive".
func WatchNewAgentJobs() {
	logger.Slog.Info("Starting AgentJob watcher")

	// Create Kubernetes in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		logger.Slog.Error("Failed to create in-cluster config", "error", err)
		metrics.OperatorStepCounter.WithLabelValues("config_load", "error").Inc()
		return
	}

	// Create dynamic client
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		logger.Slog.Error("Failed to create dynamic Kubernetes client", "error", err)
		metrics.OperatorStepCounter.WithLabelValues("client_create", "error").Inc()
		return
	}

	metrics.OperatorStepCounter.WithLabelValues("start", "success").Inc()

	// Define the GroupVersionResource (GVR) for AgentJob CRD
	agentJobGVR := schema.GroupVersionResource{
		Group:    "ea.erulabs.ai",
		Version:  "v1",
		Resource: "agentjobs",
	}

	namespace := "ea-platform" // Change as needed, get from config!!!

	for {
		// List all AgentJob CRs
		metrics.OperatorStepCounter.WithLabelValues("list_jobs", "start").Inc()
		jobs, err := dynamicClient.Resource(agentJobGVR).Namespace(namespace).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			logger.Slog.Error("Failed to list AgentJobs", "error", err)
			metrics.OperatorStepCounter.WithLabelValues("list_jobs", "error").Inc()
			time.Sleep(10 * time.Second)
			continue
		} else {
			metrics.OperatorStepCounter.WithLabelValues("list_jobs", "success").Inc()
		}

		// Iterate over jobs and update those with blank status
		for _, job := range jobs.Items {
			status, found, _ := unstructured.NestedString(job.Object, "status", "state")
			if !found || status == "" {
				logger.Slog.Info("Updating job status to inactive", "job", job.GetName())
				metrics.OperatorStepCounter.WithLabelValues("update_status", "start").Inc()

				// Prepare the status update
				job.Object["status"] = map[string]interface{}{
					"state":   "inactive",
					"message": "Job detected but not started yet",
				}

				// Apply the update
				_, err := dynamicClient.Resource(agentJobGVR).Namespace(namespace).
					UpdateStatus(context.TODO(), &job, metav1.UpdateOptions{})

				if err != nil {
					logger.Slog.Error("Failed to update job status", "job", job.GetName(), "error", err)
					metrics.OperatorStepCounter.WithLabelValues("update_status", "error").Inc()
				} else {
					logger.Slog.Info("Successfully updated job status", "job", job.GetName())
					metrics.OperatorStepCounter.WithLabelValues("update_status", "success").Inc()
				}
			}
		}

		// Sleep before next iteration
		time.Sleep(10 * time.Second)
	}
}

// WatchInactiveAgentJobs monitors inactive AgentJob CRs and spawns Kubernetes Jobs.
func WatchInactiveAgentJobs() {
	logger.Slog.Info("Starting Inactive AgentJob watcher")

	config, err := rest.InClusterConfig()
	if err != nil {
		logger.Slog.Error("Failed to create in-cluster config", "error", err)
		return
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		logger.Slog.Error("Failed to create dynamic Kubernetes client", "error", err)
		return
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		logger.Slog.Error("Failed to create Kubernetes clientset", "error", err)
		return
	}

	namespace := "ea-platform"
	agentJobGVR := schema.GroupVersionResource{
		Group:    "ea.erulabs.ai",
		Version:  "v1",
		Resource: "agentjobs",
	}

	for {
		jobs, err := dynamicClient.Resource(agentJobGVR).Namespace(namespace).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			logger.Slog.Error("Failed to list AgentJobs", "error", err)
			time.Sleep(10 * time.Second)
			continue
		}

		for _, job := range jobs.Items {
			status, _, _ := unstructured.NestedString(job.Object, "status", "state")
			if status == "inactive" {
				logger.Slog.Info("Updating job status to pending", "job", job.GetName())
				updateAgentJobStatus(dynamicClient, job.GetName(), "pending", "Job is being scheduled")

				logger.Slog.Info("Spawning Kubernetes Job for agent job", "job", job.GetName())

				k8sJob := &batchv1.Job{
					ObjectMeta: metav1.ObjectMeta{
						Name:      job.GetName(),
						Namespace: namespace,
					},
					Spec: batchv1.JobSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								RestartPolicy: corev1.RestartPolicyNever,
								Containers: []corev1.Container{{
									Name:    "hello-world",
									Image:   "busybox",
									Command: []string{"echo", "Hello, World!"},
								}},
							},
						},
					},
				}

				_, err = clientset.BatchV1().Jobs(namespace).Create(context.TODO(), k8sJob, metav1.CreateOptions{})
				if err != nil {
					logger.Slog.Error("Failed to create Kubernetes Job", "job", job.GetName(), "error", err)
					continue
				}

				updateAgentJobStatus(dynamicClient, job.GetName(), "executing", "Job is now executing")
			}
		}
		time.Sleep(10 * time.Second)
	}
}

// WatchCompletedJobs monitors Kubernetes Jobs and updates AgentJob status upon completion.
func WatchCompletedJobs() {
	logger.Slog.Info("Starting Job Completion Watcher")

	config, err := rest.InClusterConfig()
	if err != nil {
		logger.Slog.Error("Failed to create in-cluster config", "error", err)
		return
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		logger.Slog.Error("Failed to create dynamic Kubernetes client", "error", err)
		return
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		logger.Slog.Error("Failed to create Kubernetes clientset", "error", err)
		return
	}

	namespace := "ea-platform"

	for {
		jobs, err := clientset.BatchV1().Jobs(namespace).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			logger.Slog.Error("Failed to list Kubernetes Jobs", "error", err)
			time.Sleep(10 * time.Second)
			continue
		}

		for _, job := range jobs.Items {
			if job.Status.Succeeded > 0 {
				logger.Slog.Info("Marking AgentJob as completed", "job", job.Name)
				updateAgentJobStatus(dynamicClient, job.Name, "completed", "Job execution successful")
			} else if job.Status.Failed > 0 {
				logger.Slog.Info("Marking AgentJob as failed", "job", job.Name)
				updateAgentJobStatus(dynamicClient, job.Name, "failed", "Job execution failed")
			}
		}

		time.Sleep(10 * time.Second)
	}
}

// updateAgentJobStatus updates the status of an AgentJob.
func updateAgentJobStatus(dynamicClient dynamic.Interface, jobName, state, message string) {
	agentJobGVR := schema.GroupVersionResource{
		Group:    "ea.erulabs.ai",
		Version:  "v1",
		Resource: "agentjobs",
	}

	latestJob, err := dynamicClient.Resource(agentJobGVR).Namespace("ea-platform").
		Get(context.TODO(), jobName, metav1.GetOptions{})
	if err != nil {
		logger.Slog.Error("Failed to get latest AgentJob before update", "job", jobName, "error", err)
		return
	}

	latestJob.Object["status"] = map[string]interface{}{
		"state":   state,
		"message": message,
	}

	_, err = dynamicClient.Resource(agentJobGVR).Namespace("ea-platform").
		UpdateStatus(context.TODO(), latestJob, metav1.UpdateOptions{})

	if err != nil {
		logger.Slog.Error("Failed to update AgentJob status", "job", jobName, "error", err)
	}
}
