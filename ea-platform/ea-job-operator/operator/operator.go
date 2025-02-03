package operator

import (
	"context"
	"ea-job-operator/config"
	"ea-job-operator/logger"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

// AgentJob GVR
var agentJobGVR = schema.GroupVersionResource{
	Group:    "ea.erulabs.ai",
	Version:  "v1",
	Resource: "agentjobs",
}

const namespace = "ea-platform" // Namespace where jobs exist

// Workqueues to throttle updates and avoid API spam
var (
	newAgentJobQueue       = workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	inactiveAgentJobQueue  = workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	completedJobQueue      = workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	completedAgentJobQueue = workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
)

// StartOperators initializes informers and starts watching resources
func StartOperators(stopCh <-chan struct{}) {
	cfg := config.LoadConfig()

	logger.Slog.Info("Starting Informer-based Operators")

	// Set up Kubernetes client with higher rate limits
	k8sConfig, err := rest.InClusterConfig()
	if err != nil {
		logger.Slog.Error("Failed to create in-cluster config", "error", err)
		return
	}
	k8sConfig.QPS = 50    // Increase Queries Per Second
	k8sConfig.Burst = 100 // Increase Burst Capacity

	dynamicClient, err := dynamic.NewForConfig(k8sConfig)
	if err != nil {
		logger.Slog.Error("Failed to create dynamic Kubernetes client", "error", err)
		return
	}

	clientset, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		logger.Slog.Error("Failed to create Kubernetes client", "error", err)
		return
	}

	// Use Separate Informer Factories
	dynamicFactory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(dynamicClient, 0, namespace, nil)
	k8sFactory := informers.NewSharedInformerFactoryWithOptions(clientset, 0, informers.WithNamespace(namespace))

	// Run all enabled informers concurrently
	if cfg.FeatureNewAgentJobs == "true" {
		go watchNewAgentJobs(dynamicFactory, stopCh)
		go processNewJobQueue(dynamicClient, stopCh)
	}

	if cfg.FeatureInactiveAgentJobs == "true" {
		go watchInactiveAgentJobs(dynamicFactory, stopCh)
		go processInactiveQueue(dynamicClient, clientset, stopCh)
	}

	if cfg.FeatureCompletedJobs == "true" {
		go watchCompletedJobs(k8sFactory, stopCh)
		go processCompletedQueue(dynamicClient, clientset, stopCh)
	}

	if cfg.FeatureCompletedAgentJobs == "true" { //  NEW: Watch & clean up completed AgentJobs
		go watchCompletedAgentJobs(dynamicFactory, stopCh)
		go processCompletedAgentJobQueue(dynamicClient, stopCh)
	}

	// Start and sync factories
	dynamicFactory.Start(stopCh)
	k8sFactory.Start(stopCh)

	dynamicFactory.WaitForCacheSync(stopCh)
	k8sFactory.WaitForCacheSync(stopCh)

	<-stopCh // Wait for shutdown signal
}

//
// INFORMER WATCHER FUNCTIONS
//

// watchNewAgentJobs detects new AgentJobs and queues updates
func watchNewAgentJobs(factory dynamicinformer.DynamicSharedInformerFactory, stopCh <-chan struct{}) {
	logger.Slog.Info("Starting AgentJob Creation Informer")

	// Create a dynamic informer for the CRD
	informer := factory.ForResource(agentJobGVR).Informer()

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			job, ok := obj.(*unstructured.Unstructured)
			if !ok {
				logger.Slog.Error("Failed to parse AgentJob object")
				return
			}

			// jobName := job.GetName()
			status, found, _ := unstructured.NestedString(job.Object, "status", "state")

			if !found || status == "" {
				// logger.Slog.Info("Queuing job update", "job", jobName)
				newAgentJobQueue.Add(job) // Add to queue instead of updating immediately
			}
		},
	})

	// Start the informer loop
	informer.Run(stopCh)
}

// watchInactiveAgentJobs detects inactive AgentJobs and queues them for execution
func watchInactiveAgentJobs(factory dynamicinformer.DynamicSharedInformerFactory, stopCh <-chan struct{}) {
	logger.Slog.Info("Starting AgentJob Execution Informer")

	// Create a dynamic informer for the CRD
	informer := factory.ForResource(agentJobGVR).Informer()

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(_, newObj interface{}) {
			job, ok := newObj.(*unstructured.Unstructured)
			if !ok {
				logger.Slog.Error("Failed to parse AgentJob object")
				return
			}

			// jobName := job.GetName()
			status, _, _ := unstructured.NestedString(job.Object, "status", "state")

			if status == "inactive" {
				// logger.Slog.Info("Queuing inactive job for execution", "job", jobName)
				inactiveAgentJobQueue.Add(job) // Add to queue instead of executing immediately
			}
		},
	})

	// Start the informer loop
	informer.Run(stopCh)
}

// watchCompletedJobs detects completed Kubernetes Jobs and updates AgentJobs
func watchCompletedJobs(factory informers.SharedInformerFactory, stopCh <-chan struct{}) {
	logger.Slog.Info("Starting Kubernetes Job Completion Informer")

	informer := factory.Batch().V1().Jobs().Informer()

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(_, newObj interface{}) {
			job, ok := newObj.(*batchv1.Job)
			if !ok {
				logger.Slog.Error("Failed to parse Kubernetes Job object")
				return
			}

			// jobName := job.Name

			if job.Status.Succeeded > 0 {
				// logger.Slog.Info("Queuing completed Kubernetes Job for processing", "job", jobName)
				completedJobQueue.Add(job)
			}
		},
	})

	informer.Run(stopCh)
}

// watchCompletedAgentJobs detects completed AgentJobs and queues them for deletion after 1 minute
func watchCompletedAgentJobs(factory dynamicinformer.DynamicSharedInformerFactory, stopCh <-chan struct{}) {
	logger.Slog.Info("Starting Completed AgentJob Cleanup Informer")

	informer := factory.ForResource(agentJobGVR).Informer()

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(_, newObj interface{}) {
			job, ok := newObj.(*unstructured.Unstructured)
			if !ok {
				logger.Slog.Error("Failed to parse AgentJob object")
				return
			}

			// jobName := job.GetName()
			status, _, _ := unstructured.NestedString(job.Object, "status", "state")

			// If job is completed, queue it for deletion
			if status == "completed" {
				// logger.Slog.Info("Queuing completed AgentJob for cleanup", "job", jobName)
				completedAgentJobQueue.Add(job)
			}
		},
	})

	informer.Run(stopCh)
}

//
// PROCESS QUEUE FUNCTIONS
//

func processNewJobQueue(dynamicClient dynamic.Interface, stopCh <-chan struct{}) {
	wait.Until(func() {
		if newAgentJobQueue.Len() == 0 {
			return
		}

		batchSize := min(10, newAgentJobQueue.Len())

		// Process jobs directly from the queue
		for i := 0; i < batchSize; i++ {
			item, shutdown := newAgentJobQueue.Get()
			if shutdown {
				return
			}

			job := item.(*unstructured.Unstructured)
			jobName := job.GetName()

			go func() {
				defer newAgentJobQueue.Done(job) // Mark as processed

				logger.Slog.Info("Processing new AgentJob", "job", jobName)

				// Update AgentJob status to "inactive"
				err := updateAgentJobStatus(dynamicClient, job, jobName, "inactive", "Job detected but not started yet")
				if err != nil {
					logger.Slog.Error("Failed to update AgentJob status", "job", jobName, "error", err)
					newAgentJobQueue.AddAfter(job, 10*time.Second) // Retry if update fails
				}
			}()
		}
	}, time.Second, stopCh) // Runs continuously while stopCh is open
}

func processInactiveQueue(dynamicClient dynamic.Interface, clientset *kubernetes.Clientset, stopCh <-chan struct{}) {
	wait.Until(func() {
		if inactiveAgentJobQueue.Len() == 0 {
			return
		}

		batchSize := min(10, inactiveAgentJobQueue.Len())

		// Process jobs directly from the queue
		for i := 0; i < batchSize; i++ {
			item, shutdown := inactiveAgentJobQueue.Get()
			if shutdown {
				return
			}

			job := item.(*unstructured.Unstructured)
			jobName := job.GetName()

			go func() {
				defer inactiveAgentJobQueue.Done(job) // Mark as processed

				logger.Slog.Info("Processing inactive AgentJob", "job", jobName)

				// Update AgentJob status to "executing"
				err := updateAgentJobStatus(dynamicClient, job, jobName, "executing", "Job is now executing")
				if err != nil {
					logger.Slog.Error("Failed to update AgentJob status", "job", jobName, "error", err)
					inactiveAgentJobQueue.AddAfter(job, 10*time.Second) // Retry if update fails
					return
				}

				// Create Kubernetes Job
				k8sJob := &batchv1.Job{
					ObjectMeta: metav1.ObjectMeta{
						Name:      jobName,
						Namespace: namespace,
					},
					Spec: batchv1.JobSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								RestartPolicy: corev1.RestartPolicyNever,
								Containers: []corev1.Container{{
									Name:            "executor",
									Image:           "busybox",
									Command:         []string{"sh", "-c", "echo 'hello world'"},
									ImagePullPolicy: corev1.PullIfNotPresent,
								}},
							},
						},
					},
				}

				_, err = clientset.BatchV1().Jobs(namespace).Create(context.TODO(), k8sJob, metav1.CreateOptions{})
				if err != nil {
					if apierrors.IsAlreadyExists(err) {
						logger.Slog.Warn("Kubernetes Job already exists, skipping", "job", jobName)
						return
					}
					logger.Slog.Error("Failed to create Kubernetes Job", "job", jobName, "error", err)
					inactiveAgentJobQueue.AddAfter(job, 10*time.Second) // Retry if creation fails
					return
				}

				logger.Slog.Info("Successfully created Kubernetes Job", "job", jobName)
			}()
		}
	}, time.Second, stopCh) // Runs continuously while stopCh is open
}

func processCompletedQueue(dynamicClient dynamic.Interface, clientset *kubernetes.Clientset, stopCh <-chan struct{}) {
	wait.Until(func() {
		if completedJobQueue.Len() == 0 {
			return
		}

		batchSize := min(10, completedJobQueue.Len())

		for i := 0; i < batchSize; i++ {
			item, shutdown := completedJobQueue.Get()
			if shutdown {
				return
			}

			k8sJob := item.(*batchv1.Job)
			jobName := k8sJob.Name

			go func() {
				defer completedJobQueue.Done(k8sJob)

				agentJob, err := findAgentJobByK8sJob(dynamicClient, jobName)
				if err != nil || agentJob == nil {
					logger.Slog.Error("Failed to find AgentJob for Kubernetes Job", "job", jobName, "error", err)
					return
				}

				err = updateAgentJobStatus(dynamicClient, agentJob, agentJob.GetName(), "completed", "Kubernetes Job execution successful")
				if err != nil {
					logger.Slog.Error("Failed to update AgentJob status", "job", jobName, "error", err)
					completedJobQueue.AddAfter(k8sJob, 10*time.Second) // Retry
					return
				}

				deletePolicy := metav1.DeletePropagationBackground
				err = clientset.BatchV1().Jobs(namespace).Delete(context.TODO(), jobName, metav1.DeleteOptions{PropagationPolicy: &deletePolicy})
				if err != nil && !apierrors.IsNotFound(err) {
					logger.Slog.Error("Failed to delete Kubernetes Job", "job", jobName, "error", err)
					completedJobQueue.AddAfter(k8sJob, 10*time.Second) // Retry
				}
			}()
		}
	}, time.Second, stopCh) // Run continuously instead of waiting
}

func processCompletedAgentJobQueue(dynamicClient dynamic.Interface, stopCh <-chan struct{}) {
	cfg := config.LoadConfig()

	wait.Until(func() {
		if completedAgentJobQueue.Len() == 0 {
			return
		}

		//
		batchSize := min(10, completedAgentJobQueue.Len())

		// Process jobs directly from the queue
		for i := 0; i < batchSize; i++ {
			item, shutdown := completedAgentJobQueue.Get()
			if shutdown {
				return
			}

			job := item.(*unstructured.Unstructured)
			jobName := job.GetName()

			go func() {
				defer completedAgentJobQueue.Done(job) // Mark as processed

				// Fetch cleanup grace period from config
				gracePeriodMinutes := cfg.CompletedCleanupGracePeriod
				gracePeriodDuration := time.Duration(gracePeriodMinutes) * time.Minute

				// Only delete if the AgentJob has exceeded the grace period
				creationTimestamp := job.GetCreationTimestamp()
				if gracePeriodMinutes > 0 && time.Since(creationTimestamp.Time) < gracePeriodDuration {
					logger.Slog.Info("Skipping deletion, AgentJob is within grace period",
						"job", jobName, "grace_period", gracePeriodMinutes)
					completedAgentJobQueue.AddAfter(job, 30*time.Second) // Requeue after 30s
					return
				}

				// Attempt to delete the AgentJob
				err := dynamicClient.Resource(agentJobGVR).Namespace(namespace).Delete(
					context.TODO(),
					jobName,
					metav1.DeleteOptions{},
				)
				if err != nil {
					if apierrors.IsNotFound(err) {
						logger.Slog.Warn("AgentJob not found, skipping deletion", "job", jobName)
						return
					}
					logger.Slog.Error("Failed to delete completed AgentJob", "job", jobName, "error", err)

					// If error is due to conflict, requeue it for retry
					if apierrors.IsConflict(err) {
						completedAgentJobQueue.AddAfter(job, 10*time.Second) // Retry later
					}
					return
				}

				logger.Slog.Info("Successfully deleted completed AgentJob", "job", jobName)
			}()
		}
	}, time.Second, stopCh) // Runs continuously while stopCh is open
}

//
// HELPER FUNCTIONS
//

// updateAgentJobStatus updates an AgentJob's status
func updateAgentJobStatus(dynamicClient dynamic.Interface, job *unstructured.Unstructured, jobName, state, message string) error {
	updatedJob := job.DeepCopy()

	// Modify only the status field
	err := unstructured.SetNestedMap(updatedJob.Object, map[string]interface{}{
		"state":   state,
		"message": message,
	}, "status")

	if err != nil {
		logger.Slog.Error("Failed to set status field in AgentJob", "job", jobName, "error", err)
		return err
	}

	// Try updating the job status in Kubernetes
	_, err = dynamicClient.Resource(agentJobGVR).Namespace(namespace).
		UpdateStatus(context.TODO(), updatedJob, metav1.UpdateOptions{})

	if err == nil {
		logger.Slog.Info("Successfully updated job status", "job", jobName, "state", state)
		return nil
	}

	// If there's a conflict, assume another pod already handled it and move on
	if apierrors.IsConflict(err) {
		logger.Slog.Warn("Conflict detected while updating AgentJob, skipping retry", "job", jobName)
		return nil
	}

	logger.Slog.Error("Failed to update AgentJob status", "job", jobName, "error", err)
	return err
}

// Find the AgentJob linked to a completed Kubernetes Job
func findAgentJobByK8sJob(dynamicClient dynamic.Interface, jobName string) (*unstructured.Unstructured, error) {
	jobList, err := dynamicClient.Resource(agentJobGVR).Namespace(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, job := range jobList.Items {
		if job.GetName() == jobName {
			return &job, nil
		}
	}

	return nil, nil
}
