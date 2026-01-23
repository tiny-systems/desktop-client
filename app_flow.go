package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/tiny-systems/module/api/v1alpha1"
	"github.com/tiny-systems/module/pkg/utils"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

// FlowEditorData contains complete flow data for the editor.
type FlowEditorData struct {
	Flow     FlowInfo                 `json:"flow"`
	Project  ProjectInfo              `json:"project"`
	Elements []map[string]interface{} `json:"elements"`
	Meta     map[string]interface{}   `json:"meta"`
}

// FlowInfo contains basic flow information.
type FlowInfo struct {
	Name         string `json:"name"`
	ResourceName string `json:"resourceName"`
	Description  string `json:"description"`
}

// ProjectInfo contains basic project information.
type ProjectInfo struct {
	Name         string `json:"name"`
	ResourceName string `json:"resourceName"`
}

// FlowNodeEvent represents a real-time update event for flow nodes.
type FlowNodeEvent struct {
	Type  string                 `json:"type"`
	ID    string                 `json:"id"`
	Graph map[string]interface{} `json:"graph,omitempty"`
}

// ComponentInfo represents an available component for adding to a flow.
type ComponentInfo struct {
	Name        string   `json:"name"`
	Module      string   `json:"module"`
	Version     string   `json:"version"`
	Description string   `json:"description"`
	Info        string   `json:"info"`
	Tags        []string `json:"tags"`
}

// NodePosition represents position update data.
type NodePosition struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

var (
	flowWatchMu     sync.Mutex
	flowWatchCancel context.CancelFunc

	// currentTracker stores the active tracker for cleanup
	currentTrackerMu   sync.Mutex
	currentTrackerName string
)


// startStatsStreaming starts streaming stats from the otel-collector for edge animations.
func (a *App) startStatsStreaming(ctx context.Context, contextName, namespace, projectName, flowResourceName string) {
	config, err := loadContextConfig(contextName)
	if err != nil {
		a.logger.Error(err, "failed to load context config for stats streaming")
		return
	}

	pfClient := NewPortForwardClient(config, namespace)

	traceService := utils.NewTraceService(utils.TraceServiceConfig{
		Client: pfClient,
	})

	go func() {
		defer pfClient.Close()
		defer traceService.Close()

		err := traceService.SubscribeToStats(ctx, namespace, projectName, flowResourceName, func(events []utils.StatsEvent) {
			statsBatch := make(map[string]interface{})
			for _, event := range events {
				if !strings.HasPrefix(event.Metric, "tiny_edge_") || event.Element == "" {
					continue
				}

				stats, ok := statsBatch[event.Element].(map[string]interface{})
				if !ok {
					stats = make(map[string]interface{})
				}
				stats[event.Metric] = event.Value
				statsBatch[event.Element] = stats
			}

			if len(statsBatch) > 0 {
				wailsruntime.EventsEmit(a.ctx, "flowNodeUpdate", FlowNodeEvent{
					Type:  "STATS",
					Graph: statsBatch,
				})
			}
		})
		if err != nil && ctx.Err() == nil {
			a.logger.Error(err, "stats subscription failed")
		}
	}()
}

// GetFlowForEditor fetches complete flow data for the editor.
func (a *App) GetFlowForEditor(contextName, namespace, projectName, flowResourceName string) (*FlowEditorData, error) {
	mgr, err := a.getManager(contextName, namespace)
	if err != nil {
		return nil, err
	}

	flow, err := mgr.GetFLow(a.ctx, flowResourceName, namespace)
	if err != nil {
		return nil, fmt.Errorf("get flow: %w", err)
	}

	project, err := mgr.GetProject(a.ctx, projectName, namespace)
	if err != nil {
		return nil, fmt.Errorf("get project: %w", err)
	}

	projectDisplayName := project.Annotations[v1alpha1.ProjectNameAnnotation]
	if projectDisplayName == "" {
		projectDisplayName = projectName
	}

	nodes, err := mgr.GetProjectFlowNodes(a.ctx, projectName, flowResourceName)
	if err != nil {
		return nil, fmt.Errorf("get flow nodes: %w", err)
	}

	nodesMap := make(map[string]v1alpha1.TinyNode, len(nodes))
	for _, node := range nodes {
		nodesMap[node.Name] = node
	}

	allElements, err := utils.ExportNodes(nodesMap)
	if err != nil {
		return nil, fmt.Errorf("export nodes: %w", err)
	}

	elements := make([]map[string]interface{}, 0, len(allElements))
	for _, el := range allElements {
		if m, ok := el.(map[string]interface{}); ok {
			elements = append(elements, m)
		}
	}

	flowName := flow.Annotations[v1alpha1.FlowDescriptionAnnotation]
	if flowName == "" {
		flowName = flow.Name
	}

	meta := parseViewportMeta(flow.Annotations)

	return &FlowEditorData{
		Flow: FlowInfo{
			Name:         flowName,
			ResourceName: flow.Name,
			Description:  flowName,
		},
		Project: ProjectInfo{
			Name:         projectDisplayName,
			ResourceName: projectName,
		},
		Elements: elements,
		Meta:     meta,
	}, nil
}

func parseViewportMeta(annotations map[string]string) map[string]interface{} {
	meta := make(map[string]interface{})

	if x, err := strconv.ParseFloat(annotations["tinysystems.io/viewport-x"], 64); err == nil {
		meta["x"] = x
	}
	if y, err := strconv.ParseFloat(annotations["tinysystems.io/viewport-y"], 64); err == nil {
		meta["y"] = y
	}
	if zoom, err := strconv.ParseFloat(annotations["tinysystems.io/viewport-zoom"], 64); err == nil {
		meta["zoom"] = zoom
	}

	return meta
}

func buildNodeElement(node *v1alpha1.TinyNode) map[string]interface{} {
	return utils.ApiNodeToMap(*node, map[string]interface{}{}, false)
}

func buildEdgeElement(sourceNode *v1alpha1.TinyNode, edge *v1alpha1.TinyNodeEdge) map[string]interface{} {
	data := map[string]interface{}{
		"valid":  true,
		"flowID": edge.FlowID,
	}

	targetPort := utils.GetPortFullName(sourceNode.Name, edge.Port)
	for _, portConfig := range sourceNode.Spec.Ports {
		if portConfig.From == targetPort || (portConfig.Port == edge.Port && strings.Contains(edge.To, portConfig.From)) {
			if len(portConfig.Configuration) > 0 {
				data["configuration"] = json.RawMessage(portConfig.Configuration)
			}
			if len(portConfig.Schema) > 0 {
				data["schema"] = json.RawMessage(portConfig.Schema)
			}
			break
		}
	}

	edgeMap, err := utils.ApiEdgeToProtoMap(sourceNode, edge, data)
	if err != nil {
		toParts := strings.Split(edge.To, ":")
		targetNode := edge.To
		targetHandle := ""
		if len(toParts) == 2 {
			targetNode = toParts[0]
			targetHandle = toParts[1]
		}
		return map[string]interface{}{
			"id":           edge.ID,
			"source":       sourceNode.Name,
			"sourceHandle": edge.Port,
			"target":       targetNode,
			"targetHandle": targetHandle,
			"type":         "tinyEdge",
			"data":         data,
		}
	}
	return edgeMap
}

// WatchFlowNodes starts watching nodes for a specific flow and emits events.
func (a *App) WatchFlowNodes(contextName, namespace, projectName, flowResourceName string) error {
	flowWatchMu.Lock()
	defer flowWatchMu.Unlock()

	if flowWatchCancel != nil {
		flowWatchCancel()
	}

	mgr, err := a.getManager(contextName, namespace)
	if err != nil {
		return err
	}

	watchCtx, cancel := context.WithCancel(a.ctx)
	flowWatchCancel = cancel

	watcher, err := mgr.WatchNodes(watchCtx, projectName)
	if err != nil {
		cancel()
		return fmt.Errorf("start watch: %w", err)
	}

	go a.startStatsStreaming(watchCtx, contextName, namespace, projectName, flowResourceName)

	heartbeatTicker := time.NewTicker(2 * time.Second)

	go func() {
		currentWatcher := watcher
		defer func() {
			heartbeatTicker.Stop()
			if currentWatcher != nil {
				currentWatcher.Stop()
			}
		}()

		for {
			select {
			case <-watchCtx.Done():
				return

			case <-heartbeatTicker.C:
				wailsruntime.EventsEmit(a.ctx, "flowNodeUpdate", FlowNodeEvent{Type: "TICK"})

			case event, ok := <-currentWatcher.ResultChan():
				if !ok {
					currentWatcher.Stop()
					select {
					case <-watchCtx.Done():
						return
					case <-time.After(time.Second):
					}

					newWatcher, err := mgr.WatchNodes(watchCtx, projectName)
					if err != nil {
						a.logger.Error(err, "reconnect watch failed")
						return
					}
					currentWatcher = newWatcher
					continue
				}

				node, ok := event.Object.(*v1alpha1.TinyNode)
				if !ok || node.Labels[v1alpha1.FlowNameLabel] != flowResourceName {
					continue
				}

				update := FlowNodeEvent{
					Type: string(event.Type),
					ID:   node.Name,
				}

				if event.Type != watch.Deleted {
					update.Graph = buildNodeElement(node)

					for i := range node.Spec.Edges {
						edge := &node.Spec.Edges[i]
						wailsruntime.EventsEmit(a.ctx, "flowNodeUpdate", FlowNodeEvent{
							Type:  string(event.Type),
							ID:    edge.ID,
							Graph: buildEdgeElement(node, edge),
						})
					}
				}

				wailsruntime.EventsEmit(a.ctx, "flowNodeUpdate", update)
			}
		}
	}()

	return nil
}

// StopWatchFlowNodes stops the flow node watcher.
func (a *App) StopWatchFlowNodes() error {
	flowWatchMu.Lock()
	defer flowWatchMu.Unlock()

	if flowWatchCancel != nil {
		flowWatchCancel()
		flowWatchCancel = nil
	}
	return nil
}

// GetAvailableComponents returns all available components that can be added to a flow.
func (a *App) GetAvailableComponents(contextName, namespace string) ([]ComponentInfo, error) {
	mgr, err := a.getManager(contextName, namespace)
	if err != nil {
		return nil, err
	}

	modules, err := mgr.GetInstalledComponents(a.ctx)
	if err != nil {
		return nil, fmt.Errorf("get installed components: %w", err)
	}

	var components []ComponentInfo
	for _, mod := range modules {
		for _, comp := range mod.Components {
			components = append(components, ComponentInfo{
				Name:        comp.Name,
				Module:      mod.Name,
				Version:     mod.Version,
				Description: comp.Description,
				Info:        comp.Info,
				Tags:        comp.Tags,
			})
		}
	}

	return components, nil
}

// AddNode adds a new component node to a flow.
func (a *App) AddNode(contextName, namespace, projectName, flowResourceName, componentName, moduleName, moduleVersion string, posX, posY float64) (map[string]interface{}, error) {
	mgr, err := a.getManager(contextName, namespace)
	if err != nil {
		return nil, err
	}

	nodeName := utils.SanitizeResourceName(componentName) + "-" + uuid.New().String()[:8]

	node := &v1alpha1.TinyNode{
		ObjectMeta: metav1.ObjectMeta{
			Name:      nodeName,
			Namespace: namespace,
			Labels: map[string]string{
				v1alpha1.FlowNameLabel:    flowResourceName,
				v1alpha1.ProjectNameLabel: projectName,
			},
			Annotations: map[string]string{
				v1alpha1.ComponentPosXAnnotation: strconv.Itoa(int(posX)),
				v1alpha1.ComponentPosYAnnotation: strconv.Itoa(int(posY)),
				v1alpha1.NodeLabelAnnotation:     componentName,
			},
		},
		Spec: v1alpha1.TinyNodeSpec{
			Module:        moduleName,
			ModuleVersion: moduleVersion,
			Component:     componentName,
		},
	}

	if err := mgr.CreateNodeSync(a.ctx, node, 30*time.Second); err != nil {
		return nil, fmt.Errorf("create node: %w", err)
	}

	createdNode, err := mgr.GetNode(a.ctx, nodeName, namespace)
	if err != nil {
		return nil, fmt.Errorf("get created node: %w", err)
	}

	return buildNodeElement(createdNode), nil
}

// DeleteNode deletes a node from a flow.
func (a *App) DeleteNode(contextName, namespace, nodeResourceName string) error {
	mgr, err := a.getManager(contextName, namespace)
	if err != nil {
		return err
	}

	node, err := mgr.GetNode(a.ctx, nodeResourceName, namespace)
	if err != nil {
		return fmt.Errorf("get node: %w", err)
	}

	if err := mgr.DeleteNode(a.ctx, node); err != nil {
		return fmt.Errorf("delete node: %w", err)
	}

	return nil
}

// UpdateNodePosition updates a node's position in the flow.
func (a *App) UpdateNodePosition(contextName, namespace, nodeResourceName string, posX, posY float64) error {
	mgr, err := a.getManager(contextName, namespace)
	if err != nil {
		return err
	}

	node, err := mgr.GetNode(a.ctx, nodeResourceName, namespace)
	if err != nil {
		return fmt.Errorf("get node: %w", err)
	}

	if node.Annotations == nil {
		node.Annotations = make(map[string]string)
	}
	node.Annotations[v1alpha1.ComponentPosXAnnotation] = strconv.Itoa(int(posX))
	node.Annotations[v1alpha1.ComponentPosYAnnotation] = strconv.Itoa(int(posY))

	if err := mgr.UpdateNode(a.ctx, node); err != nil {
		return fmt.Errorf("update node position: %w", err)
	}

	return nil
}

// UpdateNodeLabel updates a node's display label.
func (a *App) UpdateNodeLabel(contextName, namespace, nodeResourceName, label string) error {
	mgr, err := a.getManager(contextName, namespace)
	if err != nil {
		return err
	}

	node, err := mgr.GetNode(a.ctx, nodeResourceName, namespace)
	if err != nil {
		return fmt.Errorf("get node: %w", err)
	}

	if node.Annotations == nil {
		node.Annotations = make(map[string]string)
	}
	node.Annotations[v1alpha1.NodeLabelAnnotation] = label

	if err := mgr.UpdateNode(a.ctx, node); err != nil {
		return fmt.Errorf("update node label: %w", err)
	}

	return nil
}

// UpdateNodeComment updates a node's comment.
func (a *App) UpdateNodeComment(contextName, namespace, nodeResourceName, comment string) error {
	mgr, err := a.getManager(contextName, namespace)
	if err != nil {
		return err
	}

	node, err := mgr.GetNode(a.ctx, nodeResourceName, namespace)
	if err != nil {
		return fmt.Errorf("get node: %w", err)
	}

	if node.Annotations == nil {
		node.Annotations = make(map[string]string)
	}
	node.Annotations[v1alpha1.NodeCommentAnnotation] = comment

	if err := mgr.UpdateNode(a.ctx, node); err != nil {
		return fmt.Errorf("update node comment: %w", err)
	}

	return nil
}

// RotateNode rotates a node by incrementing its spin value.
func (a *App) RotateNode(contextName, namespace, nodeResourceName string) error {
	mgr, err := a.getManager(contextName, namespace)
	if err != nil {
		return err
	}

	node, err := mgr.GetNode(a.ctx, nodeResourceName, namespace)
	if err != nil {
		return fmt.Errorf("get node: %w", err)
	}

	if node.Annotations == nil {
		node.Annotations = make(map[string]string)
	}

	currentSpin, _ := strconv.Atoi(node.Annotations[v1alpha1.ComponentPosSpinAnnotation])
	node.Annotations[v1alpha1.ComponentPosSpinAnnotation] = strconv.Itoa((currentSpin + 1) % 4)

	if err := mgr.UpdateNode(a.ctx, node); err != nil {
		return fmt.Errorf("rotate node: %w", err)
	}

	return nil
}

// ToggleNodeDashboard toggles the dashboard visibility for a node.
func (a *App) ToggleNodeDashboard(contextName, namespace, nodeResourceName string, enabled bool) error {
	mgr, err := a.getManager(contextName, namespace)
	if err != nil {
		return err
	}

	node, err := mgr.GetNode(a.ctx, nodeResourceName, namespace)
	if err != nil {
		return fmt.Errorf("get node: %w", err)
	}

	if node.Labels == nil {
		node.Labels = make(map[string]string)
	}

	if enabled {
		node.Labels[v1alpha1.DashboardLabel] = "true"
	} else {
		delete(node.Labels, v1alpha1.DashboardLabel)
	}

	if err := mgr.UpdateNode(a.ctx, node); err != nil {
		return fmt.Errorf("update node dashboard setting: %w", err)
	}

	return nil
}

// UpdateNodeConfiguration updates a node's port configuration.
func (a *App) UpdateNodeConfiguration(contextName, namespace, nodeResourceName, port, configuration, schema string) error {
	mgr, err := a.getManager(contextName, namespace)
	if err != nil {
		return err
	}

	node, err := mgr.GetNode(a.ctx, nodeResourceName, namespace)
	if err != nil {
		return fmt.Errorf("get node: %w", err)
	}

	found := false
	for i, portConfig := range node.Spec.Ports {
		if portConfig.Port == port && portConfig.From == "" {
			node.Spec.Ports[i].Configuration = []byte(configuration)
			if schema != "" {
				node.Spec.Ports[i].Schema = []byte(schema)
			}
			found = true
			break
		}
	}

	if !found {
		newPortConfig := v1alpha1.TinyNodePortConfig{
			Port:          port,
			Configuration: []byte(configuration),
		}
		if schema != "" {
			newPortConfig.Schema = []byte(schema)
		}
		node.Spec.Ports = append(node.Spec.Ports, newPortConfig)
	}

	if err := mgr.UpdateNodeSync(a.ctx, node, 30*time.Second); err != nil {
		return fmt.Errorf("update node configuration: %w", err)
	}

	return nil
}

// ConnectNodes creates an edge between two nodes.
func (a *App) ConnectNodes(contextName, namespace, flowResourceName, sourceNode, sourcePort, targetNode, targetPort, configuration string) error {
	mgr, err := a.getManager(contextName, namespace)
	if err != nil {
		return err
	}

	node, err := mgr.GetNode(a.ctx, sourceNode, namespace)
	if err != nil {
		return fmt.Errorf("get source node: %w", err)
	}

	newEdge := v1alpha1.TinyNodeEdge{
		ID:     uuid.New().String(),
		Port:   sourcePort,
		To:     fmt.Sprintf("%s:%s", targetNode, targetPort),
		FlowID: flowResourceName,
	}
	node.Spec.Edges = append(node.Spec.Edges, newEdge)

	if configuration != "" {
		portConfig := v1alpha1.TinyNodePortConfig{
			Port:          sourcePort,
			From:          fmt.Sprintf("%s:%s", targetNode, targetPort),
			Configuration: []byte(configuration),
			FlowID:        flowResourceName,
		}
		node.Spec.Ports = append(node.Spec.Ports, portConfig)
	}

	if err := mgr.UpdateNodeSync(a.ctx, node, 30*time.Second); err != nil {
		return fmt.Errorf("connect nodes: %w", err)
	}

	return nil
}

// DisconnectNodes removes an edge between two nodes.
func (a *App) DisconnectNodes(contextName, namespace, sourceNode, edgeID string) error {
	mgr, err := a.getManager(contextName, namespace)
	if err != nil {
		return err
	}

	node, err := mgr.GetNode(a.ctx, sourceNode, namespace)
	if err != nil {
		return fmt.Errorf("get source node: %w", err)
	}

	var targetTo string
	newEdges := make([]v1alpha1.TinyNodeEdge, 0, len(node.Spec.Edges))
	for _, edge := range node.Spec.Edges {
		if edge.ID == edgeID {
			targetTo = edge.To
			continue
		}
		newEdges = append(newEdges, edge)
	}
	node.Spec.Edges = newEdges

	if targetTo != "" {
		newPorts := make([]v1alpha1.TinyNodePortConfig, 0, len(node.Spec.Ports))
		for _, portConfig := range node.Spec.Ports {
			if portConfig.From == targetTo {
				continue
			}
			newPorts = append(newPorts, portConfig)
		}
		node.Spec.Ports = newPorts
	}

	if err := mgr.UpdateNodeSync(a.ctx, node, 30*time.Second); err != nil {
		return fmt.Errorf("disconnect nodes: %w", err)
	}

	return nil
}

// UpdateEdgeConfiguration updates an edge's configuration.
func (a *App) UpdateEdgeConfiguration(contextName, namespace, sourceNode, sourcePort, targetTo, configuration, flowID string) error {
	mgr, err := a.getManager(contextName, namespace)
	if err != nil {
		return err
	}

	node, err := mgr.GetNode(a.ctx, sourceNode, namespace)
	if err != nil {
		return fmt.Errorf("get source node: %w", err)
	}

	found := false
	for i, portConfig := range node.Spec.Ports {
		if portConfig.Port == sourcePort && portConfig.From == targetTo {
			node.Spec.Ports[i].Configuration = []byte(configuration)
			found = true
			break
		}
	}

	if !found {
		node.Spec.Ports = append(node.Spec.Ports, v1alpha1.TinyNodePortConfig{
			Port:          sourcePort,
			From:          targetTo,
			Configuration: []byte(configuration),
			FlowID:        flowID,
		})
	}

	if err := mgr.UpdateNodeSync(a.ctx, node, 30*time.Second); err != nil {
		return fmt.Errorf("update edge configuration: %w", err)
	}

	return nil
}

// SaveFlowMeta saves flow viewport metadata.
func (a *App) SaveFlowMeta(contextName, namespace, flowResourceName string, viewportX, viewportY, zoom float64) error {
	mgr, err := a.getManager(contextName, namespace)
	if err != nil {
		return err
	}

	flow, err := mgr.GetFLow(a.ctx, flowResourceName, namespace)
	if err != nil {
		return fmt.Errorf("get flow: %w", err)
	}

	if flow.Annotations == nil {
		flow.Annotations = make(map[string]string)
	}
	flow.Annotations["tinysystems.io/viewport-x"] = strconv.FormatFloat(viewportX, 'f', 2, 64)
	flow.Annotations["tinysystems.io/viewport-y"] = strconv.FormatFloat(viewportY, 'f', 2, 64)
	flow.Annotations["tinysystems.io/viewport-zoom"] = strconv.FormatFloat(zoom, 'f', 2, 64)

	if err := mgr.GetK8sClient().Update(a.ctx, flow); err != nil {
		return fmt.Errorf("save flow metadata: %w", err)
	}

	return nil
}

// BatchUpdateNodePositions updates multiple node positions at once.
func (a *App) BatchUpdateNodePositions(contextName, namespace string, positions map[string]NodePosition) error {
	mgr, err := a.getManager(contextName, namespace)
	if err != nil {
		return err
	}

	for nodeResourceName, pos := range positions {
		node, err := mgr.GetNode(a.ctx, nodeResourceName, namespace)
		if err != nil {
			continue
		}

		if node.Annotations == nil {
			node.Annotations = make(map[string]string)
		}
		node.Annotations[v1alpha1.ComponentPosXAnnotation] = strconv.Itoa(int(pos.X))
		node.Annotations[v1alpha1.ComponentPosYAnnotation] = strconv.Itoa(int(pos.Y))

		if err := mgr.UpdateNode(a.ctx, node); err != nil {
			a.logger.Error(err, "update node position failed", "node", nodeResourceName)
		}
	}

	return nil
}

// InspectNodePort returns the simulated data for a specific port.
// If traceID is provided, it uses real runtime data from the trace instead of simulated data.
func (a *App) InspectNodePort(contextName, namespace, projectName, nodeResourceName, portName, traceID string) (map[string]interface{}, error) {
	mgr, err := a.getManager(contextName, namespace)
	if err != nil {
		return nil, err
	}

	nodes, err := mgr.GetProjectNodes(a.ctx, projectName)
	if err != nil {
		return nil, fmt.Errorf("get project nodes: %w", err)
	}

	nodesMap := make(map[string]v1alpha1.TinyNode, len(nodes))
	var targetNode *v1alpha1.TinyNode
	for _, n := range nodes {
		nodesMap[n.Name] = n
		if n.Name == nodeResourceName {
			nodeCopy := n
			targetNode = &nodeCopy
		}
	}

	if targetNode == nil {
		return nil, fmt.Errorf("node not found: %s", nodeResourceName)
	}

	handles := utils.GetAllPortHandles(*targetNode)

	var targetHandle map[string]interface{}
	for _, handle := range handles {
		if handle["id"] == portName {
			targetHandle = handle
			break
		}
	}

	if targetHandle == nil {
		return nil, fmt.Errorf("port not found: %s", portName)
	}

	result := map[string]interface{}{
		"id":    targetHandle["id"],
		"label": targetHandle["label"],
		"type":  targetHandle["type"],
	}

	if schemaBytes := extractBytes(targetHandle["schema"]); len(schemaBytes) > 0 {
		var schema map[string]interface{}
		if err := json.Unmarshal(schemaBytes, &schema); err == nil {
			result["schema"] = schema
		}
	}

	if configBytes := extractBytes(targetHandle["configuration"]); len(configBytes) > 0 {
		var config map[string]interface{}
		if err := json.Unmarshal(configBytes, &config); err == nil {
			result["configuration"] = config
		}
	}

	ctx, cancel := context.WithTimeout(a.ctx, 3*time.Second)
	defer cancel()

	portFullName := utils.GetPortFullName(nodeResourceName, portName)

	// Load runtime data from trace if traceID is provided
	var runtimeData map[string][]byte
	if traceID != "" {
		config, err := loadContextConfig(contextName)
		if err == nil {
			pfClient := NewPortForwardClient(config, namespace)
			defer pfClient.Close()

			traceService := utils.NewTraceService(utils.TraceServiceConfig{
				Client: pfClient,
			})
			defer traceService.Close()

			// Fetch trace and extract runtime data using SDK functions (same as platform)
			trace, err := traceService.GetTraceByID(ctx, namespace, projectName, traceID)
			if err == nil && trace != nil {
				_, runtimeData = utils.ExtractTraceStatistics(trace)
				result["dataSource"] = "trace"
			}
		}
	}

	if runtimeData == nil {
		result["dataSource"] = "simulated"
	}

	// Use SimulatePortData with runtime data (same as platform)
	// If runtimeData is nil, it behaves like SimulatePortDataSimple
	simulatedData, err := utils.SimulatePortData(ctx, nodesMap, portFullName, runtimeData)
	if err != nil {
		result["data"] = nil
		result["dataError"] = err.Error()
	} else {
		result["data"] = simulatedData
	}

	return result, nil
}

func extractBytes(v interface{}) []byte {
	switch val := v.(type) {
	case json.RawMessage:
		return val
	case []byte:
		return val
	case string:
		return []byte(val)
	default:
		return nil
	}
}

// RunNodeAction triggers an action on a node port.
func (a *App) RunNodeAction(contextName, namespace, nodeResourceName, port, data string) error {
	mgr, err := a.getManager(contextName, namespace)
	if err != nil {
		return err
	}
	return mgr.CreateSignal(a.ctx, nodeResourceName, namespace, port, []byte(data))
}

// GetNodeHandles returns all handles for a specific node with their current data.
func (a *App) GetNodeHandles(contextName, namespace, nodeResourceName string) ([]map[string]interface{}, error) {
	mgr, err := a.getManager(contextName, namespace)
	if err != nil {
		return nil, err
	}

	node, err := mgr.GetNode(a.ctx, nodeResourceName, namespace)
	if err != nil {
		return nil, fmt.Errorf("get node: %w", err)
	}

	nodeMap := utils.ApiNodeToMap(*node, nil, false)
	nodeData, ok := nodeMap["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid node data")
	}

	handles, ok := nodeData["handles"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid handles data")
	}

	result := make([]map[string]interface{}, 0, len(handles))
	for _, handle := range handles {
		handleMap, ok := handle.(map[string]interface{})
		if !ok {
			continue
		}

		portData := map[string]interface{}{
			"id":       handleMap["id"],
			"label":    handleMap["label"],
			"type":     handleMap["type"],
			"position": handleMap["position"],
		}

		if schemaBytes := extractBytes(handleMap["schema"]); len(schemaBytes) > 0 {
			var schema map[string]interface{}
			if err := json.Unmarshal(schemaBytes, &schema); err == nil {
				portData["schema"] = schema
			}
		}

		if configBytes := extractBytes(handleMap["configuration"]); len(configBytes) > 0 {
			var config map[string]interface{}
			if err := json.Unmarshal(configBytes, &config); err == nil {
				portData["configuration"] = config
			}
		}

		result = append(result, portData)
	}

	return result, nil
}

// RunExpressionResult is the response from RunExpression.
type RunExpressionResult struct {
	Result          string `json:"result"`
	ValidSchema     bool   `json:"validSchema"`
	ValidationError string `json:"validationError"`
}

// RunExpression evaluates a JSONPath expression against JSON data and validates the result against a schema.
func (a *App) RunExpression(expression, data, schema string) (*RunExpressionResult, error) {
	resp, err := utils.RunExpression(&utils.RunExpressionRequest{
		Expression: expression,
		Data:       data,
		Schema:     schema,
	})
	if err != nil {
		return nil, err
	}

	return &RunExpressionResult{
		Result:          resp.Result,
		ValidSchema:     resp.ValidSchema,
		ValidationError: resp.ValidationError,
	}, nil
}

// TracesResponse is the response from GetTraces
type TracesResponse struct {
	Traces []utils.TraceInfo `json:"traces"`
	Total  int64             `json:"total"`
	Offset int64             `json:"offset"`
}

// GetTraces fetches traces for a specific flow
func (a *App) GetTraces(contextName, namespace, projectName, flowName string, start, end, offset int64) (*TracesResponse, error) {
	config, err := loadContextConfig(contextName)
	if err != nil {
		return nil, err
	}

	pfClient := NewPortForwardClient(config, namespace)
	defer pfClient.Close()

	traceService := utils.NewTraceService(utils.TraceServiceConfig{
		Client: pfClient,
	})
	defer traceService.Close()

	// Convert int64 timestamps to time.Time
	var startTime, endTime time.Time
	if start > 0 {
		startTime = time.Unix(start, 0)
	}
	if end > 0 {
		endTime = time.Unix(end, 0)
	}

	resp, err := traceService.GetTraces(a.ctx, namespace, projectName, flowName, startTime, endTime, offset)
	if err != nil {
		return nil, err
	}

	return &TracesResponse{
		Traces: resp.Traces,
		Total:  resp.Total,
		Offset: resp.Offset,
	}, nil
}

// TraceDataResponse is the response from GetTraceByID
type TraceDataResponse struct {
	TraceID string       `json:"traceId"`
	Spans   []utils.Span `json:"spans"`
}

// GetTraceByID fetches a trace by its ID
func (a *App) GetTraceByID(contextName, namespace, projectName, traceID string) (*TraceDataResponse, error) {
	config, err := loadContextConfig(contextName)
	if err != nil {
		return nil, err
	}

	pfClient := NewPortForwardClient(config, namespace)
	defer pfClient.Close()

	traceService := utils.NewTraceService(utils.TraceServiceConfig{
		Client: pfClient,
	})
	defer traceService.Close()

	trace, err := traceService.GetTraceByID(a.ctx, namespace, projectName, traceID)
	if err != nil {
		return nil, err
	}

	return &TraceDataResponse{
		TraceID: trace.TraceID,
		Spans:   trace.Spans,
	}, nil
}

// ApplyTraceToFlowResponse contains graph elements with trace stats applied
type ApplyTraceToFlowResponse struct {
	Nodes []map[string]interface{} `json:"nodes"`
	Edges []map[string]interface{} `json:"edges"`
}

// ApplyTraceToFlow fetches trace stats and applies them to graph elements
func (a *App) ApplyTraceToFlow(contextName, namespace, projectName, flowResourceName, traceID string) (*ApplyTraceToFlowResponse, error) {
	mgr, err := a.getManager(contextName, namespace)
	if err != nil {
		return nil, err
	}

	// Get all nodes for the flow
	nodes, err := mgr.GetProjectFlowNodes(a.ctx, projectName, flowResourceName)
	if err != nil {
		return nil, fmt.Errorf("get flow nodes: %w", err)
	}

	nodesMap := make(map[string]v1alpha1.TinyNode, len(nodes))
	for _, node := range nodes {
		nodesMap[node.Name] = node
	}

	// Fetch trace data and extract statistics
	var traceStats *utils.TraceStatistics
	if traceID != "" {
		config, err := loadContextConfig(contextName)
		if err != nil {
			return nil, err
		}

		pfClient := NewPortForwardClient(config, namespace)
		defer pfClient.Close()

		traceService := utils.NewTraceService(utils.TraceServiceConfig{
			Client: pfClient,
		})
		defer traceService.Close()

		trace, err := traceService.GetTraceByID(a.ctx, namespace, projectName, traceID)
		if err != nil {
			return nil, fmt.Errorf("get trace: %w", err)
		}

		traceStats, _ = utils.ExtractTraceStatistics(trace)
	}

	// Build response with trace stats applied
	response := &ApplyTraceToFlowResponse{
		Nodes: make([]map[string]interface{}, 0),
		Edges: make([]map[string]interface{}, 0),
	}

	for _, node := range nodes {
		nodeAsMap := utils.ApiNodeToMap(node, map[string]interface{}{}, false)

		if traceStats != nil {
			utils.ApplyTraceStatToNode(nodeAsMap, traceStats)
		}

		response.Nodes = append(response.Nodes, nodeAsMap)

		// Process edges
		for i := range node.Spec.Edges {
			edge := &node.Spec.Edges[i]
			edgeAsMap, err := utils.ApiEdgeToProtoMap(&node, edge, map[string]interface{}{
				"valid": true,
			})
			if err != nil {
				continue
			}

			if traceStats != nil {
				utils.ApplyTraceStatToEdge(edgeAsMap, traceStats)
			}

			response.Edges = append(response.Edges, edgeAsMap)
		}
	}

	return response, nil
}

// CreateTracker creates a TinyTracker resource for the current project/flow session.
// This enables telemetry collection while the user is viewing the flow.
func (a *App) CreateTracker(contextName, namespace, projectName string) error {
	currentTrackerMu.Lock()
	defer currentTrackerMu.Unlock()

	// Delete existing tracker first if any
	if currentTrackerName != "" {
		a.deleteTrackerInternal(contextName, namespace)
	}

	mgr, err := a.getManager(contextName, namespace)
	if err != nil {
		return err
	}

	tracker, err := mgr.PutTracker(a.ctx, namespace, projectName)
	if err != nil {
		return fmt.Errorf("create tracker: %w", err)
	}

	currentTrackerName = tracker.Name
	a.logger.Info("Created tracker", "name", tracker.Name, "project", projectName)
	return nil
}

// DeleteTracker removes the current TinyTracker resource when leaving a flow/project.
func (a *App) DeleteTracker(contextName, namespace string) error {
	currentTrackerMu.Lock()
	defer currentTrackerMu.Unlock()

	return a.deleteTrackerInternal(contextName, namespace)
}

// deleteTrackerInternal deletes the tracker without locking (caller must hold lock)
func (a *App) deleteTrackerInternal(contextName, namespace string) error {
	if currentTrackerName == "" {
		return nil
	}

	mgr, err := a.getManager(contextName, namespace)
	if err != nil {
		return err
	}

	tracker := &v1alpha1.TinyTracker{
		ObjectMeta: metav1.ObjectMeta{
			Name:      currentTrackerName,
			Namespace: namespace,
		},
	}

	if err := mgr.DeleteTracker(a.ctx, tracker); err != nil {
		a.logger.Error(err, "Failed to delete tracker", "name", currentTrackerName)
		return fmt.Errorf("delete tracker: %w", err)
	}

	a.logger.Info("Deleted tracker", "name", currentTrackerName)
	currentTrackerName = ""
	return nil
}
