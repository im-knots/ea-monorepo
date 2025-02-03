package operator

import (
	"context"
	"os"
	"time"

	"ea-job-operator/logger"
	"ea-job-operator/metrics"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
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
		metrics.OperatorStepCounter.WithLabelValues("list_jobs", "start").Inc()
		jobs, err := dynamicClient.Resource(agentJobGVR).Namespace(namespace).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			logger.Slog.Error("Failed to list AgentJobs", "error", err)
			metrics.OperatorStepCounter.WithLabelValues("list_jobs", "error").Inc()
			time.Sleep(10 * time.Second)
			continue
		}
		metrics.OperatorStepCounter.WithLabelValues("list_jobs", "success").Inc()

		// Iterate over jobs and update those with blank status
		for _, job := range jobs.Items {
			status, found, _ := unstructured.NestedString(job.Object, "status", "state")
			lock, _, _ := unstructured.NestedString(job.Object, "metadata", "annotations", "ea-job-operator-lock")

			// Skip if job is already locked by another pod
			if status == "" && (lock == "" || lockExpired(lock)) {
				// Try to lock the job for processing
				if !tryLockAgentJob(dynamicClient, job.GetName()) {
					continue // Another pod already locked it
				}

				// Use the proper function to update status
				if !found || status == "" {
					logger.Slog.Info("Updating job status to inactive", "job", job.GetName())
					metrics.OperatorStepCounter.WithLabelValues("update_status", "start").Inc()

					// 🚀 Use `updateAgentJobStatus()` instead of modifying `job.Object` directly!
					updateAgentJobStatus(dynamicClient, job.GetName(), "inactive", "Job detected but not started yet")

					metrics.OperatorStepCounter.WithLabelValues("update_status", "success").Inc()
				}
			}
		}

		// Sleep before next iteration
		time.Sleep(1 * time.Second)
	}
}

func WatchInactiveAgentJobs() {
	logger.Slog.Info("Starting Inactive AgentJob watcher")

	config, err := rest.InClusterConfig()
	if err != nil {
		logger.Slog.Error("Failed to create in-cluster config", "error", err)
		metrics.OperatorStepCounter.WithLabelValues("config_load", "error").Inc()
		return
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		logger.Slog.Error("Failed to create dynamic Kubernetes client", "error", err)
		metrics.OperatorStepCounter.WithLabelValues("client_create", "error").Inc()
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
			lock, _, _ := unstructured.NestedString(job.Object, "metadata", "annotations", "ea-job-operator-lock")

			// Skip if job is already locked by another pod
			if status == "inactive" && lock == getPodName() {
				// Try to lock the job for processing
				// if !tryLockAgentJob(dynamicClient, job.GetName()) {
				// 	continue // Another pod already locked it
				// }

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
									Name:    "executor",
									Image:   "busybox",
									Command: []string{"sh", "-c", "echo 'hello world'"},
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
		time.Sleep(1 * time.Second)
	}
}

// WatchCompletedJobs monitors Kubernetes Jobs and updates AgentJob status upon completion.
func WatchCompletedJobs() {
	logger.Slog.Info("Starting Job Completion Watcher")

	config, err := rest.InClusterConfig()
	if err != nil {
		logger.Slog.Error("Failed to create in-cluster config", "error", err)
		metrics.OperatorStepCounter.WithLabelValues("config_load", "error").Inc()
		return
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		logger.Slog.Error("Failed to create dynamic Kubernetes client", "error", err)
		metrics.OperatorStepCounter.WithLabelValues("client_create", "error").Inc()
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
		// List all Kubernetes Jobs
		k8sJobs, err := clientset.BatchV1().Jobs(namespace).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			logger.Slog.Error("Failed to list Kubernetes Jobs", "error", err)
			time.Sleep(10 * time.Second)
			continue
		}

		for _, k8sJob := range k8sJobs.Items {
			// Fetch the corresponding AgentJob
			agentJob, err := dynamicClient.Resource(agentJobGVR).Namespace(namespace).
				Get(context.TODO(), k8sJob.Name, metav1.GetOptions{})

			if err != nil {
				logger.Slog.Error("Failed to get AgentJob for Kubernetes Job", "job", k8sJob.Name, "error", err)
				continue
			}

			// Get the latest AgentJob status
			status, _, _ := unstructured.NestedString(agentJob.Object, "status", "state")
			lock, _, _ := unstructured.NestedString(agentJob.Object, "metadata", "annotations", "ea-job-operator-lock")

			// Skip if job is already locked by another pod
			if status == "executing" && lock == getPodName() {
				// Try to lock the job for processing
				// if !tryLockAgentJob(dynamicClient, agentJob.GetName()) {
				// 	continue // Another pod already locked it
				// }

				// If already marked as completed or failed, skip
				if status == "completed" || status == "failed" {
					continue
				}

				// Determine job outcome and update AgentJob status
				if k8sJob.Status.Succeeded > 0 {
					logger.Slog.Info("Marking AgentJob as completed", "job", k8sJob.Name)
					updateAgentJobStatus(dynamicClient, k8sJob.Name, "completed", "Job execution successful")
				} else if k8sJob.Status.Failed > 0 {
					logger.Slog.Info("Marking AgentJob as failed", "job", k8sJob.Name)
					updateAgentJobStatus(dynamicClient, k8sJob.Name, "failed", "Job execution failed")
				}
			}
		}

		time.Sleep(1 * time.Second)
	}
}

// WatchCompletedAgentJobs monitors completed AgentJob CRs older than 5 minutes and cleans them up.
func WatchCompletedAgentJobs() {
	logger.Slog.Info("Starting Completed AgentJob watcher")

	config, err := rest.InClusterConfig()
	if err != nil {
		logger.Slog.Error("Failed to create in-cluster config", "error", err)
		metrics.OperatorStepCounter.WithLabelValues("config_load", "error").Inc()
		return
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		logger.Slog.Error("Failed to create dynamic Kubernetes client", "error", err)
		metrics.OperatorStepCounter.WithLabelValues("client_create", "error").Inc()
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
			lock, _, _ := unstructured.NestedString(job.Object, "metadata", "annotations", "ea-job-operator-lock")

			// Skip if job is already locked by another pod
			// if status == "completed" && (lock == "" || lockExpired(lock)) {
			if status == "completed" && lock == getPodName() {

				// Try to lock the job for processing
				// if !tryLockAgentJob(dynamicClient, job.GetName()) {
				// 	continue // Another pod already locked it
				// }

				if status == "completed" || status == "failed" {
					metadata, found, _ := unstructured.NestedMap(job.Object, "spec", "metadata")
					if found {
						createdAtStr, _ := metadata["created_at"].(string)
						createdAt, err := time.Parse(time.RFC3339, createdAtStr)
						if err == nil && time.Since(createdAt) > 1*time.Minute {
							logger.Slog.Info("Deleting old completed AgentJob", "job", job.GetName())

							// Delete AgentJob CR
							err = dynamicClient.Resource(agentJobGVR).Namespace(namespace).Delete(context.TODO(), job.GetName(), metav1.DeleteOptions{})
							if err != nil {
								logger.Slog.Error("Failed to delete AgentJob CR", "job", job.GetName(), "error", err)
							} else {
								logger.Slog.Info("Successfully deleted AgentJob CR", "job", job.GetName())
							}

							// Delete associated Kubernetes Job and the job pods
							deletePolicy := metav1.DeletePropagationBackground
							err = clientset.BatchV1().Jobs(namespace).Delete(context.TODO(), job.GetName(), metav1.DeleteOptions{
								PropagationPolicy: &deletePolicy,
							})
							if err != nil {
								logger.Slog.Error("Failed to delete associated Kubernetes Job", "job", job.GetName(), "error", err)
							} else {
								logger.Slog.Info("Successfully deleted associated Kubernetes Job", "job", job.GetName())
							}

							// Delete associated Pods created by the Job
							// podList, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
							// 	LabelSelector: fmt.Sprintf("job-name=%s", job.GetName()),
							// })
							// if err != nil {
							// 	logger.Slog.Error("Failed to list associated Pods for Job", "job", job.GetName(), "error", err)
							// } else {
							// 	for _, pod := range podList.Items {
							// 		err := clientset.CoreV1().Pods(namespace).Delete(context.TODO(), pod.Name, metav1.DeleteOptions{})
							// 		if err != nil {
							// 			logger.Slog.Error("Failed to delete associated Pod", "pod", pod.Name, "job", job.GetName(), "error", err)
							// 		} else {
							// 			logger.Slog.Info("Successfully deleted associated Pod", "pod", pod.Name, "job", job.GetName())
							// 		}
							// 	}
							// }
						}
					}
				}
			}
		}

		time.Sleep(1 * time.Second)
	}
}

// We need this for high availablility IE if operator pods start dying off but i cant get it working properly without race conditions out the ass at the moment. TODO
// disabled with env FEATURE_CLEAN_ORPHANS=true
func WatchCleanOrphans() {
	logger.Slog.Info("Starting Orphaned Lock Cleaner")

	config, err := rest.InClusterConfig()
	if err != nil {
		logger.Slog.Error("Failed to create in-cluster config", "error", err)
		metrics.OperatorStepCounter.WithLabelValues("config_load", "error").Inc()
		return
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		logger.Slog.Error("Failed to create dynamic Kubernetes client", "error", err)
		metrics.OperatorStepCounter.WithLabelValues("client_create", "error").Inc()
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

	// gracePeriod := 10 * time.Second // Allow jobs at least 10 seconds before considering them orphaned

	for {
		// List all active operator pods
		podList, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
			LabelSelector: "app=ea-job-operator",
		})
		if err != nil {
			logger.Slog.Error("Failed to list operator pods", "error", err)
			time.Sleep(5 * time.Second)
			continue
		}

		// Build a set of active pod names
		activePods := make(map[string]bool)
		for _, pod := range podList.Items {
			activePods[pod.Name] = true
		}

		// List all AgentJobs
		jobs, err := dynamicClient.Resource(agentJobGVR).Namespace(namespace).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			logger.Slog.Error("Failed to list AgentJobs", "error", err)
			time.Sleep(5 * time.Second)
			continue
		}

		for _, job := range jobs.Items {
			jobName := job.GetName()

			// Get the latest version of the job to avoid stale data
			latestJob, err := dynamicClient.Resource(agentJobGVR).Namespace(namespace).Get(context.TODO(), jobName, metav1.GetOptions{})
			if err != nil {
				logger.Slog.Error("Failed to get latest AgentJob", "job", jobName, "error", err)
				continue
			}

			// Check if there's a lock annotation
			lock, lockExists, _ := unstructured.NestedString(latestJob.Object, "metadata", "annotations", "ea-job-operator-lock")

			// If no lock, or the lock belongs to an active pod, skip
			if !lockExists || lock == "" || activePods[lock] {
				continue
			}

			// Before proceeding, double-check the latest job version to ensure another pod hasn't taken it
			updatedLatestJob, err := dynamicClient.Resource(agentJobGVR).Namespace(namespace).Get(context.TODO(), jobName, metav1.GetOptions{})
			if err != nil {
				logger.Slog.Error("Failed to double-check AgentJob lock", "job", jobName, "error", err)
				continue
			}

			// If the lock was updated and now belongs to an active pod, skip
			newLock, newLockExists, _ := unstructured.NestedString(updatedLatestJob.Object, "metadata", "annotations", "ea-job-operator-lock")
			if newLockExists && activePods[newLock] {
				logger.Slog.Info("Skipping cleanup: job was re-locked by another active pod", "job", jobName, "new_lock_owner", newLock)
				continue
			}

			// If the lock still belongs to a non-existent pod, proceed with cleaning up
			logger.Slog.Info("Detected orphaned lock, removing and resetting job status", "job", jobName, "orphaned_pod", lock)

			// Delete associated Kubernetes Job and the job pods
			deletePolicy := metav1.DeletePropagationBackground
			err = clientset.BatchV1().Jobs(namespace).Delete(context.TODO(), jobName, metav1.DeleteOptions{
				PropagationPolicy: &deletePolicy,
			})
			if err != nil && !apierrors.IsNotFound(err) {
				logger.Slog.Error("Failed to delete associated Kubernetes Job", "job", jobName, "error", err)
				continue
			}
			logger.Slog.Info("Deleted associated Kubernetes Job for orphan unlock event", "job", jobName)

			// Fetch the latest version of the job before modifying
			for retries := 0; retries < 5; retries++ {
				latestJob, err := dynamicClient.Resource(agentJobGVR).Namespace(namespace).Get(context.TODO(), jobName, metav1.GetOptions{})
				if err != nil {
					logger.Slog.Error("Failed to get latest AgentJob before updating status for orphan unlock", "job", jobName, "error", err)
					continue
				}

				// Deep copy to avoid modifying the original object
				updatedJob := latestJob.DeepCopy()

				// Remove lock annotation
				annotations, _, _ := unstructured.NestedMap(updatedJob.Object, "metadata", "annotations")
				if annotations == nil {
					annotations = make(map[string]interface{})
				}
				delete(annotations, "ea-job-operator-lock")
				unstructured.SetNestedMap(updatedJob.Object, annotations, "metadata", "annotations")

				// First update the job to remove the lock annotation
				_, err = dynamicClient.Resource(agentJobGVR).Namespace(namespace).Update(context.TODO(), updatedJob, metav1.UpdateOptions{})
				if err != nil {
					logger.Slog.Error("Failed to remove orphaned lock annotation", "job", jobName, "error", err)
					continue
				}

				// Fetch the latest version of the job again before resetting status
				latestJob, err = dynamicClient.Resource(agentJobGVR).Namespace(namespace).Get(context.TODO(), jobName, metav1.GetOptions{})
				if err != nil {
					logger.Slog.Error("Failed to get latest AgentJob before resetting status", "job", jobName, "error", err)
					continue
				}

				// Deep copy to modify status
				updatedJob = latestJob.DeepCopy()

				// Now update the status to reset it
				statusMap := map[string]interface{}{
					"state":   "",
					"message": "Lock removed due to orphaned operator pod",
				}
				unstructured.SetNestedMap(updatedJob.Object, statusMap, "status")

				_, err = dynamicClient.Resource(agentJobGVR).Namespace(namespace).UpdateStatus(context.TODO(), updatedJob, metav1.UpdateOptions{})
				if err == nil {
					logger.Slog.Info("Successfully removed orphaned lock, deleted Kubernetes Job, and reset job status", "job", jobName)
					break // Exit retry loop on success
				}

				// Handle resource version conflict
				if apierrors.IsConflict(err) {
					logger.Slog.Warn("Conflict detected while updating AgentJob status, retrying...", "job", jobName)
					time.Sleep(250 * time.Millisecond)
					continue
				}

				// Other errors
				logger.Slog.Error("Failed to reset status after removing orphaned lock", "job", jobName, "error", err)
				break
			}

			// Introduce a small delay to prevent immediate reprocessing loops
			time.Sleep(200 * time.Millisecond)
		}

		time.Sleep(5 * time.Second)
	}
}

// updateAgentJobStatus updates the status of an AgentJob.
func updateAgentJobStatus(dynamicClient dynamic.Interface, jobName, state, message string) {
	agentJobGVR := schema.GroupVersionResource{
		Group:    "ea.erulabs.ai",
		Version:  "v1",
		Resource: "agentjobs",
	}

	// Retry logic for handling conflicts
	for retries := 0; retries < 5; retries++ {
		// 🚀 Always fetch the latest version before modifying status
		latestJob, err := dynamicClient.Resource(agentJobGVR).Namespace("ea-platform").
			Get(context.TODO(), jobName, metav1.GetOptions{})
		if err != nil {
			logger.Slog.Error("Failed to get latest AgentJob before updating status", "job", jobName, "error", err)
			return
		}

		// Deep copy to avoid modifying the original object
		updatedJob := latestJob.DeepCopy()

		// Modify only the status field
		err = unstructured.SetNestedMap(updatedJob.Object, map[string]interface{}{
			"state":   state,
			"message": message,
		}, "status")
		if err != nil {
			logger.Slog.Error("Failed to set status field in AgentJob", "job", jobName, "error", err)
			return
		}

		// 🚀 Attempt to update using the latest resource version
		_, err = dynamicClient.Resource(agentJobGVR).Namespace("ea-platform").
			UpdateStatus(context.TODO(), updatedJob, metav1.UpdateOptions{})

		if err == nil {
			logger.Slog.Info("Successfully updated job status", "job", jobName, "state", state)
			return
		}

		// Handle resource version conflict
		if apierrors.IsConflict(err) {
			logger.Slog.Warn("Conflict detected while updating AgentJob, retrying...", "job", jobName)
			time.Sleep(250 * time.Millisecond) // Slight delay before retrying
			continue
		}

		// Other errors
		logger.Slog.Error("Failed to update AgentJob status", "job", jobName, "error", err)
		return
	}

	logger.Slog.Error("Failed to update AgentJob after multiple retries", "job", jobName)
}

func tryLockAgentJob(dynamicClient dynamic.Interface, jobName string) bool {
	agentJobGVR := schema.GroupVersionResource{
		Group:    "ea.erulabs.ai",
		Version:  "v1",
		Resource: "agentjobs",
	}

	// Get the latest version of the AgentJob
	latestJob, err := dynamicClient.Resource(agentJobGVR).Namespace("ea-platform").
		Get(context.TODO(), jobName, metav1.GetOptions{})
	if err != nil {
		logger.Slog.Error("Failed to get latest AgentJob before locking", "job", jobName, "error", err)
		return false
	}

	// Check if another pod already locked it
	annotations, _, _ := unstructured.NestedMap(latestJob.Object, "metadata", "annotations")
	if annotations == nil {
		annotations = make(map[string]interface{})
	}

	currentPod := getPodName()
	if lock, exists := annotations["ea-job-operator-lock"]; exists {
		// If another pod owns the lock, do nothing
		if lock.(string) != currentPod {
			logger.Slog.Warn("AgentJob is already locked by another pod", "job", jobName, "lock_owner", lock)
			return false
		}
	}

	// Lock the job to the current pod (without timestamp)
	annotations["ea-job-operator-lock"] = currentPod
	unstructured.SetNestedMap(latestJob.Object, annotations, "metadata", "annotations")

	// Attempt to update the job with the lock
	_, err = dynamicClient.Resource(agentJobGVR).Namespace("ea-platform").
		Update(context.TODO(), latestJob, metav1.UpdateOptions{})
	if err != nil {
		logger.Slog.Error("Failed to lock AgentJob", "job", jobName, "error", err)
		return false
	}

	logger.Slog.Info("Successfully locked AgentJob to pod", "job", jobName, "pod", currentPod)
	return true
}

func lockExpired(lock string) bool {
	currentPod := getPodName()

	// If the lock is empty, consider it expired (i.e., job is not locked)
	if lock == "" {
		return true
	}

	// If the lock is owned by the current pod, it's not expired
	if lock == currentPod {
		return false
	}

	// Otherwise, another pod owns the lock → do not process
	logger.Slog.Warn("AgentJob is locked by another pod", "current_pod", currentPod, "lock_owner", lock)
	return false
}

func getPodName() string {
	if podName := os.Getenv("POD_NAME"); podName != "" {
		return podName
	}
	return "unknown"
}
