package operator

import (
	"ea-ainu-operator/config"
	"ea-ainu-operator/logger"
	"ea-ainu-operator/mongo"
	"time"

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

// dbClient is the shared MongoDB client for handlers.
var dbClient *mongo.MongoClient

// SetDBClient sets the MongoDB client for handlers.
func SetDBClient(client *mongo.MongoClient) {
	if client == nil {
		logger.Slog.Error("SetDBClient called with nil client")
	}
	dbClient = client
	logger.Slog.Info("Database client successfully initialized in operators")
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
	completedAgentJobQueue = workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	errorAgentJobQueue     = workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
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
		go processNewJobQueue(stopCh)
	}

	if cfg.FeatureInactiveAgentJobs == "true" {
		go watchInactiveAgentJobs(dynamicFactory, stopCh)
		go processInactiveQueue(stopCh)
	}

	if cfg.FeatureCompletedAgentJobs == "true" {
		go watchCompletedAgentJobs(dynamicFactory, stopCh)
		go processCompletedAgentJobQueue(stopCh)
	}

	if cfg.FeatureErrorAgentJobs == "true" {
		go watchErrorAgentJobs(dynamicFactory, stopCh)
		go processErrorAgentJobQueue(stopCh)
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

// watchNewAgentJobs detects new AgentJobs and queues updates to mongo TO NEW
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

// watchInactiveAgentJobs detects inactive AgentJobs and updates the mongodb entry in the ea-ainu-engine database TO PENDING
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

// watchCompletedAgentJobs detects completed AgentJobs TO COMPLETE
func watchCompletedAgentJobs(factory dynamicinformer.DynamicSharedInformerFactory, stopCh <-chan struct{}) {
	logger.Slog.Info("Starting Completed AgentJob Informer")

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

// watchErrorAgentJobs detects errored AgentJobs TO ERROR
func watchErrorAgentJobs(factory dynamicinformer.DynamicSharedInformerFactory, stopCh <-chan struct{}) {
	logger.Slog.Info("Starting Error AgentJob Informer")

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
			if status == "error" {
				// logger.Slog.Info("Queuing completed AgentJob for cleanup", "job", jobName)
				errorAgentJobQueue.Add(job)
			}
		},
	})

	informer.Run(stopCh)
}

//
// PROCESS QUEUE FUNCTIONS
//

// processNewJobQueue watches for new AgentJobs and adds them to the user's job array in MongoDB
func processNewJobQueue(stopCh <-chan struct{}) {
	wait.Until(func() {
		if newAgentJobQueue.Len() == 0 {
			return
		}

		batchSize := min(10, newAgentJobQueue.Len())

		for i := 0; i < batchSize; i++ {
			item, shutdown := newAgentJobQueue.Get()
			if shutdown {
				return
			}

			job, ok := item.(*unstructured.Unstructured)
			if !ok {
				logger.Slog.Error("Failed to cast item to AgentJob")
				continue
			}

			jobName := job.GetName()

			go func() {
				defer newAgentJobQueue.Done(job) // Mark as processed

				// Extract user ID from spec (not metadata)
				userID, found, _ := unstructured.NestedString(job.Object, "spec", "user")
				if !found || userID == "" {
					logger.Slog.Error("User ID not found in AgentJob spec", "job", jobName)
					return
				}

				// Extract agent ID and other metadata
				agentID, _, _ := unstructured.NestedString(job.Object, "spec", "agentID")
				creator, _, _ := unstructured.NestedString(job.Object, "spec", "creator")

				// Create job entry for MongoDB
				jobEntry := map[string]interface{}{
					"job_name":     jobName,
					"job_type":     "AgentJob",
					"status":       "New",
					"last_active":  time.Now(),
					"id":           job.GetUID(),
					"created_time": job.GetCreationTimestamp().Time,
					"agent_id":     agentID,
					"creator":      creator,
				}

				// Update MongoDB: Add job entry to user's job array
				filter := map[string]interface{}{"id": userID}
				update := map[string]interface{}{
					"$push": map[string]interface{}{"jobs": jobEntry},
				}

				_, err := dbClient.UpdateRecord("ainuUsers", "users", filter, update)
				if err != nil {
					logger.Slog.Error("Failed to update user jobs array in MongoDB", "user_id", userID, "job", jobName, "error", err)
				} else {
					logger.Slog.Info("Added new job to user's job array", "user_id", userID, "job", jobName)
				}
			}()
		}
	}, time.Second, stopCh)
}

func processInactiveQueue(stopCh <-chan struct{}) {
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

			job, ok := item.(*unstructured.Unstructured)
			if !ok {
				logger.Slog.Error("Failed to cast item to AgentJob")
				continue
			}

			jobName := job.GetName()

			go func() {
				defer inactiveAgentJobQueue.Done(job) // Mark as processed

				// Extract user ID from spec
				userID, found, _ := unstructured.NestedString(job.Object, "spec", "user")
				if !found || userID == "" {
					logger.Slog.Error("User ID not found in AgentJob spec", "job", jobName)
					return
				}

				// Extract job ID
				jobID := string(job.GetUID()) // Ensure it's a string
				if jobID == "" {
					logger.Slog.Error("Job ID not found in AgentJob spec", "job", jobName)
					return
				}

				// Update MongoDB: Find the user by userID and update the job by jobID
				filter := map[string]interface{}{
					"id":      userID, // Find user
					"jobs.id": jobID,  // Find job inside the user's jobs array
				}

				update := map[string]interface{}{
					"$set": map[string]interface{}{
						"jobs.$.status":      "Pending",  // Update status
						"jobs.$.last_active": time.Now(), // Update last_active timestamp
					},
				}

				_, err := dbClient.UpdateRecord("ainuUsers", "users", filter, update)
				if err != nil {
					logger.Slog.Error("Failed to update job status in MongoDB", "user_id", userID, "job_id", jobID, "error", err)
				} else {
					logger.Slog.Info("Updated job to Pending in user's job array", "user_id", userID, "job_id", jobID)
				}
			}()
		}
	}, time.Second, stopCh)
}

func processCompletedAgentJobQueue(stopCh <-chan struct{}) {
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

				// Extract user ID from spec
				userID, found, _ := unstructured.NestedString(job.Object, "spec", "user")
				if !found || userID == "" {
					logger.Slog.Error("User ID not found in AgentJob spec", "job", jobName)
					return
				}

				// Extract job ID
				jobID := string(job.GetUID()) // Ensure it's a string
				if jobID == "" {
					logger.Slog.Error("Job ID not found in AgentJob spec", "job", jobName)
					return
				}

				// Update MongoDB: Find the user by userID and update the job by jobID
				filter := map[string]interface{}{
					"id":      userID, // Find user
					"jobs.id": jobID,  // Find job inside the user's jobs array
				}

				update := map[string]interface{}{
					"$set": map[string]interface{}{
						"jobs.$.status":      "Complete", // Update status
						"jobs.$.last_active": time.Now(), // Update last_active timestamp
					},
				}

				_, err := dbClient.UpdateRecord("ainuUsers", "users", filter, update)
				if err != nil {
					logger.Slog.Error("Failed to update job status in MongoDB", "user_id", userID, "job_id", jobID, "error", err)
				} else {
					logger.Slog.Info("Updated job to Complete in user's job array", "user_id", userID, "job_id", jobID)
				}
			}()
		}
	}, time.Second, stopCh) // Runs continuously while stopCh is open
}

func processErrorAgentJobQueue(stopCh <-chan struct{}) {
	wait.Until(func() {
		if errorAgentJobQueue.Len() == 0 {
			return
		}

		//
		batchSize := min(10, errorAgentJobQueue.Len())

		// Process jobs directly from the queue
		for i := 0; i < batchSize; i++ {
			item, shutdown := errorAgentJobQueue.Get()
			if shutdown {
				return
			}

			job := item.(*unstructured.Unstructured)
			jobName := job.GetName()

			go func() {
				defer errorAgentJobQueue.Done(job) // Mark as processed

				// Extract user ID from spec
				userID, found, _ := unstructured.NestedString(job.Object, "spec", "user")
				if !found || userID == "" {
					logger.Slog.Error("User ID not found in AgentJob spec", "job", jobName)
					return
				}

				// Extract job ID
				jobID := string(job.GetUID()) // Ensure it's a string
				if jobID == "" {
					logger.Slog.Error("Job ID not found in AgentJob spec", "job", jobName)
					return
				}

				// Update MongoDB: Find the user by userID and update the job by jobID
				filter := map[string]interface{}{
					"id":      userID, // Find user
					"jobs.id": jobID,  // Find job inside the user's jobs array
				}

				update := map[string]interface{}{
					"$set": map[string]interface{}{
						"jobs.$.status":      "Error",    // Update status
						"jobs.$.last_active": time.Now(), // Update last_active timestamp
					},
				}

				_, err := dbClient.UpdateRecord("ainuUsers", "users", filter, update)
				if err != nil {
					logger.Slog.Error("Failed to update job status in MongoDB", "user_id", userID, "job_id", jobID, "error", err)
				} else {
					logger.Slog.Info("Updated job to Error in user's job array", "user_id", userID, "job_id", jobID)
				}
			}()
		}
	}, time.Second, stopCh) // Runs continuously while stopCh is open
}

//
// HELPER FUNCTIONS
//
