package operator

import (
	"context"
	"ea-job-operator/config"
	"ea-job-operator/logger"
	"time"

	"golang.org/x/exp/rand"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
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
	jobQueue       = workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	inactiveQueue  = workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	completedQueue = workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter()) // ✅ Added queue for completed jobs
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
		go processNewJobQueue(dynamicClient, stopCh) // Worker for new jobs
	}

	if cfg.FeatureInactiveAgentJobs == "true" {
		go watchInactiveAgentJobs(dynamicFactory, clientset, stopCh)
		go processInactiveQueue(dynamicClient, clientset, stopCh) // Worker for inactive jobs
	}

	if cfg.FeatureCompletedJobs == "true" {
		go watchCompletedJobs(k8sFactory, dynamicClient, stopCh) // Pass the correct factory
		go processCompletedQueue(dynamicClient, stopCh)          // Worker for completed jobs
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

			jobName := job.GetName()
			status, found, _ := unstructured.NestedString(job.Object, "status", "state")

			if !found || status == "" {
				logger.Slog.Info("Queuing job update", "job", jobName)
				jobQueue.Add(job) // Add to queue instead of updating immediately
			}
		},
	})

	// Start the informer loop
	informer.Run(stopCh)
}

// watchInactiveAgentJobs detects inactive AgentJobs and queues them for execution
func watchInactiveAgentJobs(factory dynamicinformer.DynamicSharedInformerFactory, clientset *kubernetes.Clientset, stopCh <-chan struct{}) {
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

			jobName := job.GetName()
			status, _, _ := unstructured.NestedString(job.Object, "status", "state")

			if status == "inactive" {
				logger.Slog.Info("Queuing inactive job for execution", "job", jobName)
				inactiveQueue.Add(job) // Add to queue instead of executing immediately
			}
		},
	})

	// Start the informer loop
	informer.Run(stopCh)
}

// watchCompletedJobs detects completed Kubernetes Jobs and updates AgentJobs
func watchCompletedJobs(factory informers.SharedInformerFactory, dynamicClient dynamic.Interface, stopCh <-chan struct{}) {
	logger.Slog.Info("Starting Kubernetes Job Completion Informer")

	informer := factory.Batch().V1().Jobs().Informer()

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(_, newObj interface{}) {
			job, ok := newObj.(*batchv1.Job)
			if !ok {
				logger.Slog.Error("Failed to parse Kubernetes Job object")
				return
			}

			jobName := job.Name

			if job.Status.Succeeded > 0 {
				logger.Slog.Info("Queuing completed Kubernetes Job for processing", "job", jobName)
				completedQueue.Add(job)
			}
		},
	})

	informer.Run(stopCh)
}

//
// PROCESS QUEUE FUNCTIONS
//

// processNewJobQueue processes the job queue asynchronously
func processNewJobQueue(dynamicClient dynamic.Interface, stopCh <-chan struct{}) {
	for {
		item, shutdown := jobQueue.Get()
		if shutdown {
			return
		}

		job := item.(*unstructured.Unstructured)
		jobName := job.GetName()

		go func() {
			defer jobQueue.Done(job)
			updateAgentJobStatus(dynamicClient, job, jobName, "inactive", "Job detected but not started yet")
		}()
	}
}

// processInactiveQueue processes the inactive job queue and spawns Kubernetes Jobs
func processInactiveQueue(dynamicClient dynamic.Interface, clientset *kubernetes.Clientset, stopCh <-chan struct{}) {
	for {
		item, shutdown := inactiveQueue.Get()
		if shutdown {
			return
		}

		job := item.(*unstructured.Unstructured)
		jobName := job.GetName()

		go func() {
			defer inactiveQueue.Done(job)

			// Update AgentJob status
			updateAgentJobStatus(dynamicClient, job, jobName, "executing", "Job is now executing")

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
								Name:    "executor",
								Image:   "busybox",
								Command: []string{"sh", "-c", "echo 'hello world'"},
							}},
						},
					},
				},
			}

			_, err := clientset.BatchV1().Jobs(namespace).Create(context.TODO(), k8sJob, metav1.CreateOptions{})
			if err != nil {
				logger.Slog.Error("Failed to create Kubernetes Job", "job", jobName, "error", err)
				return
			}

			logger.Slog.Info("Successfully created Kubernetes Job", "job", jobName)
		}()
	}
}

// processCompletedQueue now properly looks up the AgentJob
func processCompletedQueue(dynamicClient dynamic.Interface, stopCh <-chan struct{}) {
	for {
		item, shutdown := completedQueue.Get()
		if shutdown {
			return
		}

		k8sJob := item.(*batchv1.Job)
		jobName := k8sJob.Name

		go func() {
			defer completedQueue.Done(k8sJob)

			// ✅ Look up the associated AgentJob before updating
			agentJob, err := findAgentJobByK8sJob(dynamicClient, jobName)
			if err != nil {
				logger.Slog.Error("Failed to find AgentJob for Kubernetes Job", "job", jobName, "error", err)
				return
			}

			// ✅ Ensure agentJob is non-nil
			if agentJob == nil {
				logger.Slog.Error("AgentJob not found for Kubernetes Job", "job", jobName)
				return
			}

			// ✅ Update associated AgentJob status
			updateAgentJobStatus(dynamicClient, agentJob, agentJob.GetName(), "completed", "Kubernetes Job execution successful")
		}()
	}
}

//
// HELPER FUNCTIONS
//

// updateAgentJobStatus updates an AgentJob's status with exponential backoff
func updateAgentJobStatus(dynamicClient dynamic.Interface, job *unstructured.Unstructured, jobName, state, message string) {
	baseDelay := 100 * time.Millisecond // Start with 100ms
	maxDelay := 2 * time.Second         // Maximum backoff delay

	for retries := 0; retries < 5; retries++ {
		updatedJob := job.DeepCopy()

		// Modify only the status field
		err := unstructured.SetNestedMap(updatedJob.Object, map[string]interface{}{
			"state":   state,
			"message": message,
		}, "status")

		if err != nil {
			logger.Slog.Error("Failed to set status field in AgentJob", "job", jobName, "error", err)
			return
		}

		// Update the job status in Kubernetes
		_, err = dynamicClient.Resource(agentJobGVR).Namespace(namespace).
			UpdateStatus(context.TODO(), updatedJob, metav1.UpdateOptions{})

		if err == nil {
			logger.Slog.Info("Successfully updated job status", "job", jobName, "state", state)
			return
		}

		if apierrors.IsConflict(err) {
			delay := time.Duration(rand.Intn(int(baseDelay))) + baseDelay
			logger.Slog.Warn("Conflict detected while updating AgentJob, retrying...", "job", jobName, "retry_delay", delay)
			time.Sleep(delay)
			baseDelay *= 2 // Exponential backoff
			if baseDelay > maxDelay {
				baseDelay = maxDelay
			}
			continue
		}

		logger.Slog.Error("Failed to update AgentJob status", "job", jobName, "error", err)
		return
	}
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
