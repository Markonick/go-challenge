## INTRO

The hookbro solution is a simple service that listens to events and forwards them to the SVIX Webhook API. In an attempt to decouple the receiving part to the sending part we introduced a workerPool. This renders these two parts asynchrounous to each other or at least tries to.


## BASIC DOMAIN KNOWLEDGE
Looking at the Gigs Projects documentation, what I understood of the hierarchy:
```
Organization
    └── Projects
         └── Users
             └── Subscriptions, Plans, Devices, SIMs
```

Key points:

1. Organizations can have multiple projects
2. Each project has:
- Unique ID (e.g., "gigs", "dev", "stage")
- Organization reference
- Configuration for billing, payments, etc.
- Users and their subscriptions
  
What this means for our webhook service:

We should create one Svix application per project, not per organization or user, because:
- Projects are the main isolation boundary
- All user/subscription data belongs to a project
- Projects have their own configuration and settings
  
This ensures events are properly isolated per project and customers (project owners) receive only their relevant events. 


## Task System Architecture

This document explains the relationship between the task interface and webhook implementation in our system.

### Core Components

#### Task Interface
Located in `internal/worker/worker.go`, this defines the contract for all tasks in the system:

```go
type Task interface {
    Execute(ctx context.Context) error
    ID() string
}
```

Every task must implement:
- `Execute()`: Performs the actual work
- `ID()`: Returns a unique identifier for the task

#### WebhookTask Implementation
Located in `internal/tasks/webhook_task.go`, this provides a concrete implementation of the Task interface:

```go
type WebhookTask struct {
    event         models.BaseEvent
    svixClient    svix.Client
    projectAppIDs map[string]string
}
```

So the relationship is that WebhookTask satisfies the Task interface by implementing:
```
Execute(ctx context.Context) error
``` 
which processes webhook events and sends them via the svix client
```
ID() string
```
which returns the event ID as the task identifier

### Design Pattern

This architecture follows the Interface Pattern in Go, where:

1. The `worker` package defines a generic interface for tasks
2. The `webhook` package provides a concrete implementation
3. The worker system processes tasks without knowing their specific implementation details


### Usage

Tasks are processed through the task service, which manages worker goroutines and handles task results. See `internal/services/task_service.go` for implementation details.

### Flow Diagram
```
[Client] → [Controller] → [TaskService] → [WorkerPool] → [WebhookTask] → [Svix API]
   |            |             |               |               |              |
   |         Validate     Create Task    Manage Pool     Execute Task    HTTP Call
   |            |             |               |               |              |
   |            |             |          [Worker 1]           |         [Success/Error]
   |            |             |          [Worker 2]           |              |
   |            |             |          [Worker 3]           |              |
   |            |             |                               |              |
   |            |             |          [Worker N]           |              |
   |            |             |               |               |              |
   |            |             |               |               |              |
   |         202 Accept       |               |               |              |
   '------------'             |               |               |              |
                              |               |               |              |
                              |        [Task Complete]        |              |
                              |              or               |              |
                              |        [Error Logged]         |              |
                              |                ↑              |              |
                              |                |              |              |
                              '----------------'--------------'              |
                                          Error                              |
                                       Propagation  <------------------------'
```

### Flow


1. **Task Submission**
   - Client sends webhook to notification controller
   - Controller parses the event data
   - Controller calls TaskService's ProcessEvent
   - Controller immediately returns 202 Accepted
   - TaskService uses injected factory to create new WebhookTask
   - TaskService submits task to worker pool

2. **Processing**
   - Worker pool maintains a fixed number of workers
   - Pool's Submit method queues task for processing
   - Available worker picks up task from queue
   - Worker executes task asynchronously
   - Error channel created for this specific task execution
   - Worker monitors task execution

3. **Task Execution**
   - WebhookTask validates project ID
   - WebhookTask looks up corresponding Svix app ID
   - WebhookTask calls Svix client to send message
   - Any errors from Svix are captured
   - Success/failure is logged

4. **Error Handling**
   - Errors propagate from Svix → WebhookTask → Worker
   - Worker sends error to task's error channel
   - Error channel is closed after task completion
   - Errors are logged with task ID and details
   - Pool returns any errors to TaskService

Key Points:
- Single worker pool instance for application lifetime
- Multiple workers process tasks concurrently
- Async processing after quick acknowledgment
- Error propagation through channels
- Structured logging at each step



# Implementation Notes

## The Problem and Our Approach

When designing this webhook delivery service, I focused on reliability and scalability while keeping the implementation straightforward. The core challenge was ensuring reliable delivery of events to Svix while handling various failure scenarios gracefully.

## Key Design Decisions

I chose to implement an asynchronous processing model using a worker pool pattern. Here's why:

When a notification arrives, we want to:
1. Quickly acknowledge receipt (202 Accepted)
2. Process it reliably in the background
3. Handle failures without affecting other deliveries

The worker pool (using `gammazero/workerpool`) manages this:
```go
func (p *Pool) ProcessTask(task Task) error {
	// Create a new background context for the task
	ctx := context.Background()
	errChan := make(chan error, 1)

	p.wp.Submit(func() {
		err := task.Execute(ctx)
		errChan <- err // Send the error (or nil)
		close(errChan) // Always close the channel
	})

	return <-errChan // Wait for result
}

func (p *Pool) Close() {
	p.wp.StopWait()
}
```

## Error Handling Philosophy

The system implements a layered error handling approach: at the worker pool level, errors from task execution are captured through error channels and propagated upward; at the task service level (task_service.go), these errors are logged with contextual information (event ID, task ID) and passed to the caller; and at the API level, errors are translated into appropriate HTTP responses. The use of errChan in the worker pool ensures errors aren't lost during concurrent execution, while the synchronous wait for results provides reliable error feedback to clients.

The synchronous wait is a problem though and needs to be fixed. 
The following would be the async way:
```
func (p *Pool) ProcessTask(task Task) error {
    p.wp.Submit(func() {
        if err := task.Execute(context.Background()); err != nil {
            logger.Log.Error().
                Err(err).
                Str("task_id", task.ID()).
                Msg("Task execution failed")
        }
    })

    return nil // Return immediately
}
```

According to my investigation this needs to be implemented differenty in order to be non-blocking, the pool in the above example will return asynchronously but will not bubble up errors eg. 409, 422 and we need to rely on logging infrastructure to debug behaviour.

## Dependency Injection

Decided to use dig library for di, as a way to handle all factories at top level and to inject in isolation.
```
// This is a factory function registration
must(container.Provide(func(svixClient svix.Client, projectAppIDs map[string]string) func(models.BaseEvent) worker.Task {
    return func(event models.BaseEvent) worker.Task {
        return task.NewWebhookTask(event, svixClient, projectAppIDs)
    }
}))
```

What's happening here is:
1. We're not creating a single task
2. We're registering a factory function that creates new tasks
3. This factory is used by the TaskService to create new tasks for each event:

```
type taskServiceImpl struct {
    workerPool *worker.Pool
    createTask func(models.BaseEvent) worker.Task  // This is our factory
}

func (t *taskServiceImpl) ProcessEvent(event models.BaseEvent) error {
    task := t.createTask(event)  // New task created for each event
    return t.workerPool.ProcessTask(task)
}
```
The flow is:
1. DI container provides the factory function
2. TaskService uses this factory to create new tasks
3. Each event gets its own new task instance
4. Worker pool processes these individual tasks

## Production Considerations

While this implementation works, there are several things I'd add for production:

1. **Observability**: We need to track:
   - Webhook delivery success rates
   - Processing latencies
   - Queue depths
   - Error patterns
   - CPU/MEM loading

2. **Scaling**: The worker pool is configurable, but we need:
   - Load-based auto-scaling
   - Better resource utilization metrics
   - Cross-region deployment options

3. **Health Checks**
   - Worker pool status
   - Svix connectivity
   - Queue health
   - System resources
  
## Customer Experience

For customers, webhook delivery isn't just about technical reliability. I'd prioritize:

1. **Transparency**: Let customers see:
   - Delivery attempts
   - Failure reasons
   - Retry status

2. **Control**: Give customers ability to:
   - Configure retry policies
   - Filter event types
   - Test webhook endpoints

3. **Webhook UI**: Create our own Svix
   - Instead of relying on external 3rd party API create our own version of SVIX
   - Customers don't need to pivot between Gigs and Svix, everything lives in Gigs.com
  

## What I'd Do Differently

With more time, I would:

1. Add persistent storage for event tracking
2. I would consider using AWS Lambdas
3. Implement proper dead-letter queuing
4. Add comprehensive metrics collection
5. Build customer-facing debugging tools

## Testing Strategy

The current tests cover the basics:
```go
func TestTaskService_ProcessEvent(t *testing.T) {
    // Tests task submission and basic error cases
}
```

But we need:
1. Integration tests with Svix
2. Load testing under various conditions
3. Failure injection testing
4. End-to-end delivery verification

## Alternative Architecture: AWS Lambda Consideration

I would consider using AWS Lambdas or similar in GCP or Azure as this service is a perfect fit for a serverless architecture:

1. **Event-Driven Nature**
   - Each webhook delivery is an independent event
   - No state management required between requests
   - Natural fit for Lambda's event-handling model

2. **Cost Efficiency**
   ```
   Current Architecture:     Always-running server(s) regardless of load
   Lambda Architecture:      Pay only for actual webhook processing time
   ```

3. **Auto-Scaling Benefits**
   - Lambda automatically scales with incoming requests
   - No need to manage worker pools
   - Better handling of sudden traffic spikes
   - Natural rate limiting through concurrent execution limits

4. **Simplified Error Handling**
   ```go
   // Instead of worker pools:
   func HandleWebhook(ctx context.Context, event Event) error {
       if err := sendToSvix(event); err != nil {
           // Lambda will automatically retry based on configuration
           return err
       }
       return nil
   }
   ```

5. **Built-in Retry Mechanisms**
   - AWS Lambda retry policies for failed executions
   - DLQ (Dead Letter Queue) support out of the box
   - CloudWatch metrics and logging included

6. **Infrastructure Benefits**
   - No server management
   - Built-in monitoring
   - Regional deployment
   - High availability

7. **Cost Example**
   ```
   1M webhooks/month @ 256MB memory, 1s average duration
   Lambda cost:     ~$0.20
   EC2 t3.micro:    ~$8.50
   ```

8. **Potential Architecture**
   ```
   API Gateway → Lambda → Svix
        ↓
   CloudWatch Logs
        ↓
   Lambda DLQ (for failed deliveries)
   ```

This would eliminate our current worker pool complexity while providing better scalability and operational visibility.


## Cons of AWS Lambda approach
1. **Vendor lock-in** -> Add as much abstraction as possible eg. wrappers
2. **Cost Unpredictability** -> spikes bring up costs -> Put concurrency limits + cost montoring/alerts
3. **Lambda 15-min timeout** -> Chunk messages + use SQS
4. **Cold Starts** -> Use provisioned concurrency + implement warm-up strategies + optimise code init
5. **Stateless** -> Leverage dynamoDB or redis for state but keep stateless as much as possible

## Final Thoughts

This implementation is an attempt to prioritize reliability over complexity. 
As I am new to Golang, I struggled a bit with various notions such as implementing interfaces on types, understanding how to implement channels in the context of goroutines and deciding what is acceptable, in the Golang community - eg. is dependency injection a thing as it is in Java or C#? Or what type of libraries are more common these days due to reliability but also product fit (eg, workerPool, logger, air, dig, gin etc).

Another thing I found challenging was maintaining focus while writing the code as there seems to be 
a lot of clutter and I feel that this is not just inherent to the language itself but rather my lack of experience with the language. This is something I would defintely want to improve myself on, in Golang. Remove clutter and better refactoring through mastering the language basics but also more advanced conecpts

I found myself following these steps:
1. Learn about the Go environment, setup IDE for it, how to setup a Go project, what are typical
   project structures usedin the Go community. Always had in mind not to over-complicate things. I could be wrong.
2. Start with simple implementation: call->api(/notifications) -> parse and send event directly to client synchronously. This was my initial happy path to get to learn the basic business logic of GIGS first but also SVIX. This took me some time.
3. Meanwhile, while trying to understand the SVIX dahsboard and environment in general, I decided to create a script for deleting Svix apps, as I often found myself debugging and starting from a clean slate, so added a delete-svix-apps.sh script
4. Also realised kind of late during development effort that there was already a "test" folder available with 50 json events, along with a run.sh script, so I started using this as well.
5. This helped me realise how to optimise the process of initialising Apps + WH endpoints and refactoring this
6. Then ensure that the message.Create() to send events to svix follows (run.sh)
7. Then I put some thought about how to implement this in production. Probably something involving the decoupling of the api and the svix client. We dont want a 3rd party external API with possibly unexpected errors to drive our performance. Usually in this kind of situation usage of a event queue or a worker queue is common, so perhaps adding some sort of message queue would be in order. Then I thought why not actually create a system that uses a thread pool or even better a worker pool? So I picked a library (github.com/gammazero/workerpool v1.1.3). I did not want to implement a redis queue or rabbtimq queue at this point.
8. Then I realised that there was too much magic encapsulated in the workerpool - realised this when starting to write some tests, too much abstraction I did not understand - so I did a step back and looked into goroutines and channels. Actually workerpool employs goroutines and channels but I need more control.
9. After realising that implementing a worker pool myself was beyond my capabilities with Go at this point and to not perpetuate my effort, I decided to revert back to workerpool
10. One assumption I made early and should've fixed is I am hardcoding a map from projects -> appID (projectAppIDS). 
11. Another assumption is that a svix client is initialised at startup (in order not to spawn many connections everytime we make calls). This however has the negative effect of not allowing the server to spin up if the token is incorrect. It should instead allow the server to run but just deny API calls to Svix.

I would've liked to get a better understanding handling errors. All my decisions were made based on previous experience in other language environments so although the principles remain the same, there were enough questions to hinder my progress with confidence. Eg. how do interfaces work and when to use them in this project and was it an overkill? Or do I need DI here? Do I even need DI in Go? 
What about code naming conventions and code cleansiness in general. I installed a linter for this.

Tried not to make things more complicated than they should be but I have the impression that this could've gone better.
