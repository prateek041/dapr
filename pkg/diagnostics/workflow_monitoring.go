/*
Copyright 2023 The Dapr Authors
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package diagnostics

import (
	"context"

	diagUtils "github.com/dapr/dapr/pkg/diagnostics/utils"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
)

var (
	executionTypeKey = tag.MustNewKey("execution_type")
)

const (
	StatusSuccess     = "success"
	StatusFailed      = "failed"
	StatusRecoverable = "recoverable"
	CreateWorkflow    = "create_workflow"
	GetWorkflow       = "get_workflow"
	AddEvent          = "add_event"
	PurgeWorkflow     = "purge_workflow"
	Activity          = "activity"
	Workflow          = "workflow"

	WorkflowEvent = "event"
	Timer         = "timer"

	ComponentName = "dapr"
)

type workflowMetrics struct {
	// workflowOperationCount records count of Successful/Failed requests to Create/Get/Purge Workflow and Add Events.
	workflowOperationCount *stats.Int64Measure
	// workflowOperationLatency records latency of response for workflow operation requests.
	workflowOperationLatency *stats.Float64Measure
	// workflowExecutionCount records count of Successful/Failed/Recoverable workflow/activity executions.
	workflowExecutionCount *stats.Int64Measure
	// workflowExecutionLatency records latency of workflow/activity executions.
	workflowExecutionLatency *stats.Float64Measure
	appID                    string
	enabled                  bool
	namespace                string
}

func newWorkflowMetrics() *workflowMetrics {
	return &workflowMetrics{
		workflowOperationCount: stats.Int64(
			"runtime/workflow/operation/count",
			"The number of successful/failed workflow operation requests.",
			stats.UnitDimensionless),
		workflowOperationLatency: stats.Float64(
			"runtime/workflow/operation/latency",
			"The latencies of responses for workflow operation requests.",
			stats.UnitMilliseconds),
		workflowExecutionCount: stats.Int64(
			"runtime/workflow/execution/count",
			"The number of successful/failed/recoverable workflow/activity executions.",
			stats.UnitDimensionless),
		workflowExecutionLatency: stats.Float64(
			"runtime/workflow/execution/latency",
			"The total time taken to run a workflow/activity to completion.",
			stats.UnitMilliseconds),
	}
}

func (w *workflowMetrics) IsEnabled() bool {
	return w != nil && w.enabled
}

// Init registers the workflow metrics views.
func (w *workflowMetrics) Init(appID, namespace string) error {
	w.appID = appID
	w.enabled = true
	w.namespace = namespace

	return view.Register(
		diagUtils.NewMeasureView(w.workflowOperationCount, []tag.Key{appIDKey, componentKey, namespaceKey, operationKey, statusKey}, view.Count()),
		diagUtils.NewMeasureView(w.workflowOperationLatency, []tag.Key{appIDKey, componentKey, namespaceKey, operationKey, statusKey}, defaultLatencyDistribution),
		diagUtils.NewMeasureView(w.workflowExecutionCount, []tag.Key{appIDKey, componentKey, namespaceKey, executionTypeKey, statusKey}, view.Count()),
		diagUtils.NewMeasureView(w.workflowExecutionLatency, []tag.Key{appIDKey, componentKey, namespaceKey, executionTypeKey, statusKey}, defaultLatencyDistribution))
}

// WorkflowOperationEvent records total number of Successful/Failed workflow Operations requests. It also records latency for those requests.
func (w *workflowMetrics) WorkflowOperationEvent(ctx context.Context, operation, component, status string, elapsed float64) {
	if !w.IsEnabled() {
		return
	}

	stats.RecordWithTags(ctx, diagUtils.WithTags(w.workflowOperationCount.Name(), appIDKey, w.appID, componentKey, component, namespaceKey, w.namespace, operationKey, operation, statusKey, status), w.workflowOperationCount.M(1))

	if elapsed > 0 {
		stats.RecordWithTags(ctx, diagUtils.WithTags(w.workflowOperationLatency.Name(), appIDKey, w.appID, componentKey, component, namespaceKey, w.namespace, operationKey, operation, statusKey, status), w.workflowOperationLatency.M(elapsed))
	}

}

// ExecutionEvent records total number of successful/failed workflow/activity executions. It also records latency for executions.
func (w *workflowMetrics) ExecutionEvent(ctx context.Context, component, executionType, status string, elapsed float64) {
	if !w.IsEnabled() {
		return
	}

	stats.RecordWithTags(ctx, diagUtils.WithTags(w.workflowExecutionCount.Name(), appIDKey, w.appID, componentKey, component, namespaceKey, w.namespace, executionTypeKey, executionType, statusKey, status), w.workflowExecutionCount.M(1))

	if elapsed > 0 {
		stats.RecordWithTags(ctx, diagUtils.WithTags(w.workflowExecutionLatency.Name(), appIDKey, w.appID, componentKey, component, namespaceKey, w.namespace, executionTypeKey, executionType, statusKey, status), w.workflowExecutionLatency.M(elapsed))
	}
}