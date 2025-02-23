package operator

import (
	"context"
	"ea-job-operator/config"
	"ea-job-operator/logger"
	"encoding/json"
	"fmt"
	"time"

	"github.com/mitchellh/mapstructure"
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

// NodeAPI describes how to call an API (base URL, endpoint, etc.).
type NodeAPI struct {
	BaseURL  string            `json:"base_url"`
	Endpoint string            `json:"endpoint"`
	Method   string            `json:"method"`
	Headers  map[string]string `json:"headers,omitempty"`
}

// NodeParameter describes each parameter for a NodeDefinition.
type NodeParameter struct {
	Key         string        `json:"key"`
	Type        string        `json:"type"`
	Description string        `json:"description,omitempty"`
	Default     interface{}   `json:"default,omitempty"`
	Enum        []interface{} `json:"enum,omitempty"`
}

// NodeDefinitionMetadata holds metadata about the node definition.
type NodeDefinitionMetadata struct {
	Description string                 `json:"description,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	Additional  map[string]interface{} `json:"additional,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// NodeDefinition represents the "template" for a node.
type NodeDefinition struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	Name       string                 `json:"name,omitempty"`
	Creator    string                 `json:"creator,omitempty"`
	API        *NodeAPI               `json:"api,omitempty"`
	Parameters []NodeParameter        `json:"parameters,omitempty"`
	Outputs    []NodeParameter        `json:"outputs,omitempty"`
	Metadata   NodeDefinitionMetadata `json:"metadata"`
}

// NodeInstance represents a reference to a node definition.
type NodeInstance struct {
	Alias      string                 `json:"alias,omitempty"`
	Type       string                 `json:"type"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}

// Edge represents a connection between nodes in an agent workflow.
type Edge struct {
	From MultiString `json:"from"`
	To   MultiString `json:"to"`
}

// Metadata holds timestamps for Agents.
type Metadata struct {
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	AgentJobID string    `json:"agent_job_id"`
}

// Agent represents an AI workflow with interconnected nodes.
type Agent struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Creator     string         `json:"creator"`
	Description string         `json:"description"`
	Nodes       []NodeInstance `json:"nodes"`
	Edges       []Edge         `json:"edges"`
	Metadata    Metadata       `json:"metadata"`
}

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
	errorJobQueue          = workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	nodeStatusQueue        = workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
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
		go watchCompletedJobs(k8sFactory, dynamicClient, stopCh)
		go processCompletedQueue(dynamicClient, clientset, stopCh)
		go processErrorJobQueue(dynamicClient, clientset, stopCh)
	}

	if cfg.FeatureCompletedAgentJobs == "true" { //  NEW: Watch & clean up completed AgentJobs
		go watchCompletedAgentJobs(dynamicFactory, stopCh)
		go processCompletedAgentJobQueue(dynamicClient, stopCh)
	}

	if cfg.FeatureNodeStatusUpdates == "true" {
		go watchNodeStatus(k8sFactory, stopCh)
		go processNodeStatusQueue(dynamicClient, stopCh)
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

			// jobName := job.Name

			if job.Status.Succeeded > 0 {
				completedJobQueue.Add(job)
			} else if job.Status.Active > 0 {
				logger.Slog.Info("Job is still active, skipping error queue", "job", job.Name)
			} else if isJobFailed(job) {
				agentJob, err := findAgentJobByK8sJob(dynamicClient, job.Name)
				if err == nil && agentJob != nil {
					errorJobQueue.Add(agentJob)
				}
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

func watchNodeStatus(factory informers.SharedInformerFactory, stopCh <-chan struct{}) {
	logger.Slog.Info("Starting Node Status Event Informer")

	eventInformer := factory.Core().V1().Events().Informer()

	eventInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			event, ok := obj.(*corev1.Event)
			if !ok {
				logger.Slog.Error("Failed to parse Event object")
				return
			}

			if event.Reason == "NodeStatusUpdate" {
				logger.Slog.Info("Detected Node Status Update", "event", event.Name)
				nodeStatusQueue.Add(event)
			}
		},
	})

	eventInformer.Run(stopCh)
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

		for i := 0; i < batchSize; i++ {
			item, shutdown := inactiveAgentJobQueue.Get()
			if shutdown {
				return
			}

			job := item.(*unstructured.Unstructured)
			jobName := job.GetName()

			go func() {
				defer inactiveAgentJobQueue.Done(job)

				logger.Slog.Info("Processing inactive AgentJob", "job", jobName)

				// Update AgentJob status to "executing"
				err := updateAgentJobStatus(dynamicClient, job, jobName, "executing", "Job is now executing")
				if err != nil {
					logger.Slog.Error("Failed to update AgentJob status", "job", jobName, "error", err)
					inactiveAgentJobQueue.AddAfter(job, 10*time.Second)
					return
				}

				// Extract the 'spec' section
				spec, found, err := unstructured.NestedMap(job.Object, "spec")
				if err != nil || !found {
					logger.Slog.Error("Failed to extract spec from AgentJob", "job", jobName, "error", err)
					return
				}

				// Decode the spec into the Agent struct
				var agent Agent
				if err := mapstructure.Decode(spec, &agent); err != nil {
					logger.Slog.Error("Failed to decode Agent spec", "job", jobName, "error", err)
					return
				}

				// Ensure all nodes have aliases
				for i, node := range agent.Nodes {
					if node.Alias == "" {
						alias := fmt.Sprintf("node-%d", i)
						logger.Slog.Warn("Missing alias in node, assigning default alias", "node_type", node.Type, "alias", alias)
						agent.Nodes[i].Alias = alias
					}
				}

				// Set additional metadata
				agentID, found, err := unstructured.NestedString(job.Object, "spec", "agentID")
				if err != nil || !found {
					logger.Slog.Warn("Agent ID not found in job spec, defaulting to job UID", "job", jobName)
					return
				} else {
					agent.ID = agentID
				}

				agentJobID, found, err := unstructured.NestedString(job.Object, "metadata", "name")
				if err != nil || !found {
					logger.Slog.Warn("AgentJob ID not found in metadata, defaulting to job name", "job", jobName)
					agentJobID = jobName // Fallback to jobName if not found
				}

				agent.Metadata.CreatedAt = job.GetCreationTimestamp().Time
				agent.Metadata.UpdatedAt = time.Now()

				// Add AgentJob ID to metadata
				agent.Metadata.AgentJobID = agentJobID

				// Marshal the cleaned AgentJob JSON
				agentJobJSON, err := json.Marshal(agent)
				if err != nil {
					logger.Slog.Error("Failed to marshal AgentJob JSON", "job", jobName, "error", err)
					return
				}

				// Create ConfigMap with AgentJob JSON
				configMapName := fmt.Sprintf("%s-config", jobName)
				configMap := &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      configMapName,
						Namespace: namespace,
					},
					Data: map[string]string{
						"agentjob.json": string(agentJobJSON),
					},
				}

				_, err = clientset.CoreV1().ConfigMaps(namespace).Create(context.TODO(), configMap, metav1.CreateOptions{})
				if err != nil && !apierrors.IsAlreadyExists(err) {
					logger.Slog.Error("Failed to create ConfigMap for AgentJob", "job", jobName, "error", err)
					return
				}

				// Create Kubernetes Job for ea-job-executor
				backoffLimit := int32(5) // Limit retries to 5
				k8sJob := &batchv1.Job{
					ObjectMeta: metav1.ObjectMeta{
						Name:      jobName,
						Namespace: namespace,
					},
					Spec: batchv1.JobSpec{
						BackoffLimit: &backoffLimit,
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								ServiceAccountName: "ea-job-executor-sa",
								RestartPolicy:      corev1.RestartPolicyNever,
								Containers: []corev1.Container{{
									Name:            "executor",
									Image:           "ea-job-executor:latest",
									Command:         []string{"/app/ea-job-executor"},
									ImagePullPolicy: corev1.PullIfNotPresent,
									VolumeMounts: []corev1.VolumeMount{
										{
											Name:      "agentjob-json",
											MountPath: "/app/agentjob.json",
											SubPath:   "agentjob.json",
										},
									},
								}},
								Volumes: []corev1.Volume{
									{
										Name: "agentjob-json",
										VolumeSource: corev1.VolumeSource{
											ConfigMap: &corev1.ConfigMapVolumeSource{
												LocalObjectReference: corev1.LocalObjectReference{
													Name: configMapName,
												},
											},
										},
									},
								},
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
					inactiveAgentJobQueue.AddAfter(job, 10*time.Second)
					return
				}

				logger.Slog.Info("Successfully created Kubernetes Job", "job", jobName)
			}()
		}
	}, time.Second, stopCh)
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
			configMapName := fmt.Sprintf("%s-config", jobName)

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

				// Delete the associated ConfigMap
				err = clientset.CoreV1().ConfigMaps(namespace).Delete(context.TODO(), configMapName, metav1.DeleteOptions{PropagationPolicy: &deletePolicy})
				if err != nil && !apierrors.IsNotFound(err) {
					logger.Slog.Error("Failed to delete ConfigMap", "configMap", configMapName, "error", err)
				} else {
					logger.Slog.Info("Successfully deleted ConfigMap", "configMap", configMapName)
				}

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

func processErrorJobQueue(dynamicClient dynamic.Interface, clientset *kubernetes.Clientset, stopCh <-chan struct{}) {
	wait.Until(func() {
		if errorJobQueue.Len() == 0 {
			return
		}

		//
		batchSize := min(10, errorJobQueue.Len())

		// Process jobs directly from the queue
		for i := 0; i < batchSize; i++ {
			item, shutdown := errorJobQueue.Get()
			if shutdown {
				return
			}

			job := item.(*unstructured.Unstructured)
			jobName := job.GetName()
			k8sJob, err := findK8sJobByAgentJob(clientset, jobName)
			if err != nil || k8sJob == nil {
				logger.Slog.Error("Failed to find Kubernetes Job for AgentJob", "job", jobName, "error", err)
				return
			}

			go func() {
				defer errorJobQueue.Done(job) // Mark as processed

				logger.Slog.Info("Processing errored AgentJob", "job", jobName)

				// Update AgentJob status to "executing"
				err := updateAgentJobStatus(dynamicClient, job, jobName, "error", "An error occured")
				if err != nil {
					logger.Slog.Error("Failed to update AgentJob status", "job", jobName, "error", err)
					inactiveAgentJobQueue.AddAfter(job, 10*time.Second) // Retry if update fails
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
	}, time.Second, stopCh) // Runs continuously while stopCh is open
}

func processNodeStatusQueue(dynamicClient dynamic.Interface, stopCh <-chan struct{}) {
	wait.Until(func() {
		for {
			item, shutdown := nodeStatusQueue.Get()
			if shutdown {
				return
			}

			event := item.(*corev1.Event)
			jobID := event.InvolvedObject.Name

			nodeAlias := event.Annotations["nodeAlias"]
			status := event.Annotations["status"]
			outputJSON := event.Annotations["output"]

			var output map[string]interface{}
			if err := json.Unmarshal([]byte(outputJSON), &output); err != nil {
				fmt.Printf("Failed to unmarshal output JSON: %v\n", err)
				nodeStatusQueue.Done(event)
				continue
			}

			agentJob, err := dynamicClient.Resource(agentJobGVR).Namespace(namespace).Get(context.TODO(), jobID, metav1.GetOptions{})
			if err != nil {
				fmt.Printf("Failed to get AgentJob: %v\n", err)
				nodeStatusQueue.Done(event)
				continue
			}

			nodeStatus := map[string]interface{}{
				"alias":       nodeAlias,
				"status":      status,
				"output":      outputJSON,
				"lastUpdated": time.Now().Format(time.RFC3339),
			}

			existingNodes, _, _ := unstructured.NestedSlice(agentJob.Object, "status", "nodes")
			found := false
			for i, n := range existingNodes {
				node := n.(map[string]interface{})
				if node["alias"] == nodeAlias {
					existingNodes[i] = nodeStatus
					found = true
					break
				}
			}
			if !found {
				existingNodes = append(existingNodes, nodeStatus)
			}

			if err := unstructured.SetNestedSlice(agentJob.Object, existingNodes, "status", "nodes"); err != nil {
				fmt.Printf("Failed to set node status: %v\n", err)
				nodeStatusQueue.Done(event)
				continue
			}

			_, err = dynamicClient.Resource(agentJobGVR).Namespace(namespace).UpdateStatus(context.TODO(), agentJob, metav1.UpdateOptions{})
			if err != nil {
				fmt.Printf("Failed to update AgentJob status: %v\n", err)
			}

			nodeStatusQueue.Done(event)
		}
	}, time.Second, stopCh)
}

//
// HELPER FUNCTIONS
//

// updateAgentJobStatus updates an AgentJob's status
func updateAgentJobStatus(dynamicClient dynamic.Interface, job *unstructured.Unstructured, jobName, state, message string) error {
	updatedJob := job.DeepCopy()

	// Get existing status to preserve node status updates
	existingStatus, found, err := unstructured.NestedMap(updatedJob.Object, "status")
	if err != nil || !found {
		existingStatus = make(map[string]interface{}) // Initialize if not found
	}

	// Merge the new status fields without overwriting existing data
	existingStatus["state"] = state
	existingStatus["message"] = message

	// Apply the merged status back to the object
	if err := unstructured.SetNestedMap(updatedJob.Object, existingStatus, "status"); err != nil {
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

	// Handle conflict errors gracefully
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

// findK8sJobByAgentJob finds the Kubernetes Job associated with an AgentJob.
func findK8sJobByAgentJob(clientset *kubernetes.Clientset, agentJobName string) (*batchv1.Job, error) {
	// List all Kubernetes Jobs in the namespace
	jobList, err := clientset.BatchV1().Jobs(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	// Iterate over Jobs and find the one linked to the AgentJob
	for _, job := range jobList.Items {
		if job.Name == agentJobName {
			return &job, nil
		}
	}

	// No matching Kubernetes Job found
	return nil, nil
}

func isJobFailed(job *batchv1.Job) bool {
	for _, cond := range job.Status.Conditions {
		if cond.Type == batchv1.JobFailed && cond.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}

// MultiString allows a field to be either a single string or an array of strings.
type MultiString []string
