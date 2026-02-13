package main

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
	"time"

	jsonpatchapply "github.com/evanphx/json-patch"
	"github.com/tiny-systems/module/api/v1alpha1"
	"github.com/tiny-systems/module/pkg/resource"
	"github.com/tiny-systems/module/pkg/utils"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
	"gomodules.xyz/jsonpatch/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Project struct {
  Name        string `json:"name"`
  Title       string `json:"title"`
  Description string `json:"description"`
}

// ProjectDetails - Complete project information for the project page
type ProjectDetails struct {
  Name         string `json:"name"`
  Title        string `json:"title"`
  Description  string `json:"description"`
  ResourceName string `json:"resourceName"`
  ClusterName  string `json:"clusterName"`
}

// ProjectStats - Statistics for the project
type ProjectStats struct {
  WidgetsCount int `json:"widgetsCount"`
  FlowsCount   int `json:"flowsCount"`
  NodesCount   int `json:"nodesCount"`
}

// Flow - Flow information for listing
type Flow struct {
  Name         string                 `json:"name"`
  ResourceName string                 `json:"resourceName"`
  NodeCount    int                    `json:"nodeCount"`
  Graph        map[string]interface{} `json:"graph,omitempty"`
}

// Widget - Widget data for dashboard
type Widget struct {
  ID            string                 `json:"id"`
  Title         string                 `json:"title"`
  NodeName      string                 `json:"nodeName"`
  Port          string                 `json:"port"`
  DefaultSchema map[string]interface{} `json:"defaultSchema"`
  Schema        map[string]interface{} `json:"schema,omitempty"`
  Data          map[string]interface{} `json:"data"`
  GridX         int                    `json:"gridX"`
  GridY         int                    `json:"gridY"`
  GridW         int                    `json:"gridW"`
  GridH         int                    `json:"gridH"`
  Pages         []string               `json:"pages,omitempty"`
}

// WidgetPage - Dashboard page
type WidgetPage struct {
  Name         string `json:"name"`
  Title        string `json:"title"`
  ResourceName string `json:"resourceName"`
  SortIdx      int    `json:"sortIdx"`
}

// Module - Installed module information
type Module struct {
  Name       string            `json:"name"`
  Version    string            `json:"version"`
  SDKVersion string            `json:"sdkVersion"`
  Components []ModuleComponent `json:"components"`
}

// ModuleComponent - Component within a module
type ModuleComponent struct {
  Name        string   `json:"name"`
  Description string   `json:"description"`
  Info        string   `json:"info"`
  Tags        []string `json:"tags"`
}

// NodeUpdate - Real-time update event
type NodeUpdate struct {
  EventType string  `json:"eventType"` // "ADDED", "MODIFIED", "DELETED"
  NodeName  string  `json:"nodeName"`
  Widget    *Widget `json:"widget,omitempty"`
}

type UpdateEvent struct {
  // Type indicates what kind of data is in the Payload (e.g., "status", "progress", "error")
  Type string `json:"type"`
  // Payload holds the actual data, which can be any Go struct, map, or primitive type.
  Payload interface{} `json:"payload"`
}

// StatusPayload is a simple struct for a text status message.
type StatusPayload struct {
  Message string `json:"message"`
}

// ProgressPayload is a struct for percentage updates.
type ProgressPayload struct {
  Percentage int    `json:"percentage"`
  File       string `json:"file"`
}

// ErrorPayload is a struct for detailed error reporting.
type ErrorPayload struct {
  Code    int    `json:"code"`
  Details string `json:"details"`
}

func (a *App) GetProjects(contextName string, namespace string) ([]Project, error) {
  defer func() {
    if r := recover(); r != nil {
      a.logger.Error(nil, "panic in GetProjects",
        "panic", r,
        "stacktrace", string(debug.Stack()))
    }
  }()

  a.logger.Info("getting projects", "context", contextName, "namespace", namespace)
  var projectsApi []Project

  config, err := loadContextConfig(contextName)
  if err != nil {
    return nil, err
  }

  scheme := runtime.NewScheme()
  if err := v1alpha1.AddToScheme(scheme); err != nil {
    return nil, fmt.Errorf("failed to add scheme: %w", err)
  }

  kubeClient, err := client.NewWithWatch(config, client.Options{Scheme: scheme})
  if err != nil {
    return nil, fmt.Errorf("unable to create client: %w", err)
  }

  mgr, err := resource.NewManagerFromClient(kubeClient, namespace)
  if err != nil {
    return nil, err
  }
  projects, err := mgr.GetProjectList(a.ctx)
  if err != nil {
    return nil, err
  }
  a.logger.Info("projects found", "count", len(projects))

  for _, project := range projects {
    // Use title from annotation, fallback to resource name if not set
    title := project.Annotations[v1alpha1.ProjectNameAnnotation]
    if title == "" {
      title = project.Name
    }

    projectsApi = append(projectsApi, Project{
      Name:        project.Name,
      Title:       title,
      Description: project.Spec.Description,
    })

  }

  return projectsApi, nil
}

func (a *App) GetProject(contextName string, namespace string, projectName string) (*Project, error) {
  return &Project{}, nil
}

// GetModules returns a list of installed modules in the namespace
func (a *App) GetModules(contextName string, namespace string) ([]Module, error) {
  defer func() {
    if r := recover(); r != nil {
      a.logger.Error(nil, "panic in GetModules",
        "panic", r,
        "stacktrace", string(debug.Stack()))
    }
  }()

  a.logger.Info("getting modules", "context", contextName, "namespace", namespace)

  mgr, err := a.getManager(contextName, namespace)
  if err != nil {
    return nil, err
  }

  // Use SDK function to get installed modules
  installedModules, err := mgr.GetInstalledComponents(a.ctx)
  if err != nil {
    return nil, fmt.Errorf("failed to get modules: %w", err)
  }

  a.logger.Info("modules found", "count", len(installedModules))

  var modules []Module
  for _, mod := range installedModules {
    // Convert SDK module.ComponentInfo to our ModuleComponent
    var components []ModuleComponent
    for _, comp := range mod.Components {
      components = append(components, ModuleComponent{
        Name:        comp.Name,
        Description: comp.Description,
        Info:        comp.Info,
        Tags:        comp.Tags,
      })
    }

    modules = append(modules, Module{
      Name:       mod.Name,
      Version:    mod.Version,
      SDKVersion: mod.SDKVersion,
      Components: components,
    })
  }

  // Sort modules by name for consistent display
  sort.Slice(modules, func(i, j int) bool {
    return modules[i].Name < modules[j].Name
  })

  return modules, nil
}

// CreateProject creates a new project in the cluster
func (a *App) CreateProject(contextName string, namespace string, name string) (*Project, error) {
  if name == "" {
    return nil, fmt.Errorf("project name is required")
  }

  mgr, err := a.getManager(contextName, namespace)
  if err != nil {
    return nil, err
  }

  project, err := mgr.CreateProject(a.ctx, namespace, name)
  if err != nil {
    return nil, fmt.Errorf("failed to create project: %w", err)
  }

  // Return the created project
  title := project.Annotations[v1alpha1.ProjectNameAnnotation]
  if title == "" {
    title = project.Name
  }

  return &Project{
    Name:        project.Name,
    Title:       title,
    Description: "Just created",
  }, nil
}

// getManager creates a resource manager for the given context and namespace
func (a *App) getManager(contextName string, namespace string) (*resource.Manager, error) {
  config, err := loadContextConfig(contextName)
  if err != nil {
    return nil, err
  }

  return resource.NewManagerFromConfig(config, namespace)
}

// GetProjectDetails fetches complete project information
func (a *App) GetProjectDetails(contextName string, namespace string, projectName string) (*ProjectDetails, error) {
  mgr, err := a.getManager(contextName, namespace)
  if err != nil {
    return nil, err
  }

  project, err := mgr.GetProject(a.ctx, projectName, namespace)
  if err != nil {
    return nil, fmt.Errorf("unable to get project: %w", err)
  }

  return &ProjectDetails{
    Name:         project.Annotations[v1alpha1.ProjectNameAnnotation],
    Title:        project.Annotations[v1alpha1.ProjectNameAnnotation],
    Description:  project.Spec.Description,
    ResourceName: project.Name,
    ClusterName:  contextName,
  }, nil
}

// GetProjectStats fetches project statistics
func (a *App) GetProjectStats(contextName string, namespace string, projectName string) (*ProjectStats, error) {
  mgr, err := a.getManager(contextName, namespace)
  if err != nil {
    return nil, err
  }

  // Get flows count
  flows, err := mgr.GetFlowList(a.ctx, projectName)
  if err != nil {
    return nil, fmt.Errorf("unable to get flows: %w", err)
  }

  // Get nodes count
  nodes, err := mgr.GetProjectNodes(a.ctx, projectName)
  if err != nil {
    return nil, fmt.Errorf("unable to get nodes: %w", err)
  }

  // Count widgets (nodes with dashboard label and _control port)
  widgetsCount := 0
  for _, node := range nodes {
    if node.Labels[v1alpha1.DashboardLabel] == "true" {
      for _, port := range node.Status.Ports {
        if port.Name == v1alpha1.ControlPort {
          widgetsCount++
          break
        }
      }
    }
  }

  return &ProjectStats{
    WidgetsCount: widgetsCount,
    FlowsCount:   len(flows),
    NodesCount:   len(nodes),
  }, nil
}

// GetFlows fetches all flows for a project with node counts
func (a *App) GetFlows(contextName string, namespace string, projectName string) ([]Flow, error) {
  mgr, err := a.getManager(contextName, namespace)
  if err != nil {
    return nil, err
  }

  flows, err := mgr.GetFlowList(a.ctx, projectName)
  if err != nil {
    return nil, fmt.Errorf("unable to get flows: %w", err)
  }

  var result []Flow
  for _, flow := range flows {
    // Get node count for this flow
    nodes, err := mgr.GetProjectFlowNodes(a.ctx, projectName, flow.Name)
    nodeCount := 0
    if err == nil {
      nodeCount = len(nodes)
    }

    // Get flow name from annotation or use resource name
    name := flow.Annotations[v1alpha1.FlowDescriptionAnnotation]
    if name == "" {
      name = flow.Name
    }

    result = append(result, Flow{
      Name:         name,
      ResourceName: flow.Name,
      NodeCount:    nodeCount,
    })
  }

  return result, nil
}

// GetFlowsWithGraphs fetches all flows for a project with their graphs for preview
func (a *App) GetFlowsWithGraphs(contextName string, namespace string, projectName string) ([]Flow, error) {
  mgr, err := a.getManager(contextName, namespace)
  if err != nil {
    return nil, err
  }

  flows, err := mgr.GetFlowList(a.ctx, projectName)
  if err != nil {
    return nil, fmt.Errorf("unable to get flows: %w", err)
  }

  // Fetch ALL project nodes once (to include shared nodes)
  allNodes, err := mgr.GetProjectNodes(a.ctx, projectName)
  if err != nil {
    return nil, fmt.Errorf("unable to get project nodes: %w", err)
  }

  allNodesMap := make(map[string]v1alpha1.TinyNode, len(allNodes))
  for _, node := range allNodes {
    allNodesMap[node.Name] = node
  }

  var result []Flow
  for _, flow := range flows {
    // Get flow name from annotation or use resource name
    name := flow.Annotations[v1alpha1.FlowDescriptionAnnotation]
    if name == "" {
      name = flow.Name
    }

    // Build graph for preview using NodesToGraphWithOptions which filters by flow
    flowName := flow.Name
    nodeElements, edgeElements, _ := utils.NodesToGraphWithOptions(allNodesMap, &flowName, true)

    // Count nodes for this flow
    nodeCount := len(nodeElements)

    var graph map[string]interface{}
    if nodeCount > 0 {
      graph = map[string]interface{}{
        "nodes": nodeElements,
        "edges": edgeElements,
      }
    }

    result = append(result, Flow{
      Name:         name,
      ResourceName: flow.Name,
      NodeCount:    nodeCount,
      Graph:        graph,
    })
  }

  return result, nil
}

// CreateFlow creates a new flow in a project
func (a *App) CreateFlow(contextName string, namespace string, projectName string, flowName string) (*Flow, error) {
  if flowName == "" {
    return nil, fmt.Errorf("flow name is required")
  }

  mgr, err := a.getManager(contextName, namespace)
  if err != nil {
    return nil, err
  }

  resourceName, err := mgr.CreateFlow(a.ctx, namespace, projectName, flowName)
  if err != nil {
    return nil, fmt.Errorf("failed to create flow: %w", err)
  }

  return &Flow{
    Name:         flowName,
    ResourceName: *resourceName,
    NodeCount:    0,
  }, nil
}

// UndeployFlow deletes a flow and all its nodes from the cluster
func (a *App) UndeployFlow(contextName string, namespace string, flowResourceName string) error {
  if flowResourceName == "" {
    return fmt.Errorf("flow resource name is required")
  }

  mgr, err := a.getManager(contextName, namespace)
  if err != nil {
    return err
  }

  if err := mgr.DeleteFlow(a.ctx, flowResourceName); err != nil {
    return fmt.Errorf("failed to undeploy flow: %w", err)
  }

  return nil
}

// RenameFlow renames a flow
func (a *App) RenameFlow(contextName string, namespace string, flowResourceName string, newName string) error {
  if flowResourceName == "" {
    return fmt.Errorf("flow resource name is required")
  }
  if newName == "" {
    return fmt.Errorf("new name is required")
  }

  mgr, err := a.getManager(contextName, namespace)
  if err != nil {
    return err
  }

  if err := mgr.RenameFlow(a.ctx, flowResourceName, namespace, newName); err != nil {
    return fmt.Errorf("failed to rename flow: %w", err)
  }

  return nil
}

// GetFlowGraph fetches the graph data for a flow (for preview)
func (a *App) GetFlowGraph(contextName string, namespace string, projectName string, flowResourceName string) (map[string]interface{}, error) {
  mgr, err := a.getManager(contextName, namespace)
  if err != nil {
    return nil, err
  }

  // Fetch ALL project nodes to include shared nodes from other flows
  allNodes, err := mgr.GetProjectNodes(a.ctx, projectName)
  if err != nil {
    return nil, fmt.Errorf("unable to get project nodes: %w", err)
  }

  allNodesMap := make(map[string]v1alpha1.TinyNode, len(allNodes))
  for _, node := range allNodes {
    allNodesMap[node.Name] = node
  }

  // Use NodesToGraphWithOptions to filter nodes for this flow (including shared nodes)
  // Pass minimal=true for preview (less data)
  nodeElements, edgeElements, err := utils.NodesToGraphWithOptions(allNodesMap, &flowResourceName, true)
  if err != nil {
    return nil, fmt.Errorf("failed to convert nodes to graph: %w", err)
  }

  return map[string]interface{}{
    "nodes": nodeElements,
    "edges": edgeElements,
  }, nil
}

// GetWidgetPages fetches dashboard pages for a project
func (a *App) GetWidgetPages(contextName string, namespace string, projectName string) ([]WidgetPage, error) {
  mgr, err := a.getManager(contextName, namespace)
  if err != nil {
    return nil, err
  }

  pages, err := mgr.GetProjectPageWidgets(a.ctx, projectName)
  if err != nil {
    return nil, fmt.Errorf("unable to get widget pages: %w", err)
  }

  var result []WidgetPage
  for _, page := range pages {
    sortIdx, _ := strconv.Atoi(page.Annotations[v1alpha1.PageSortIdxAnnotation])
    title := page.Annotations[v1alpha1.PageTitleAnnotation]
    if title == "" {
      title = page.Name
    }

    result = append(result, WidgetPage{
      Name:         title,
      Title:        title,
      ResourceName: page.Name,
      SortIdx:      sortIdx,
    })
  }

  // Sort by sort index
  sort.Slice(result, func(i, j int) bool {
    return result[i].SortIdx < result[j].SortIdx
  })

  return result, nil
}

// GetWidgets fetches widgets for a specific page
func (a *App) GetWidgets(contextName string, namespace string, projectName string, pageResourceName string) ([]Widget, error) {
  mgr, err := a.getManager(contextName, namespace)
  if err != nil {
    return nil, err
  }

  // Get all nodes with dashboard label
  nodes, err := mgr.GetProjectNodes(a.ctx, projectName)
  if err != nil {
    return nil, fmt.Errorf("unable to get nodes: %w", err)
  }

  // Get widget pages to find grid positions
  pages, err := mgr.GetProjectPageWidgets(a.ctx, projectName)
  if err != nil {
    return nil, fmt.Errorf("unable to get widget pages: %w", err)
  }

  // Determine if current page is the default page (first page by sort index)
  isDefaultPage := false
  if len(pages) > 0 {
    // Sort pages by sort index to find the default
    sortedPages := make([]v1alpha1.TinyWidgetPage, len(pages))
    copy(sortedPages, pages)
    sort.Slice(sortedPages, func(i, j int) bool {
      idxI, _ := strconv.Atoi(sortedPages[i].Annotations[v1alpha1.PageSortIdxAnnotation])
      idxJ, _ := strconv.Atoi(sortedPages[j].Annotations[v1alpha1.PageSortIdxAnnotation])
      return idxI < idxJ
    })
    isDefaultPage = sortedPages[0].Name == pageResourceName
  }

  // Build a map of widget positions from the current page
  // Key is widget.Port which stores "nodeName:portName" format
  widgetPositions := make(map[string]v1alpha1.TinyWidget)
  // Also track which pages each widget appears on (across all pages)
  widgetPages := make(map[string][]string)
  for _, page := range pages {
    for _, widget := range page.Spec.Widgets {
      // Track pages for each widget
      widgetPages[widget.Port] = append(widgetPages[widget.Port], page.Name)
      // Store positions from the current page
      if page.Name == pageResourceName {
        widgetPositions[widget.Port] = widget
      }
    }
  }

  var result []Widget
  widgetIndex := 0
  for _, node := range nodes {
    if node.Labels[v1alpha1.DashboardLabel] != "true" {
      continue
    }

    // Use ApiNodeToMap to get properly resolved schema with configurable definitions
    nodeMap := utils.ApiNodeToMap(node, nil, false)
    nodeData, ok := nodeMap["data"].(map[string]interface{})
    if !ok {
      continue
    }

    handles, ok := nodeData["handles"].([]interface{})
    if !ok {
      continue
    }

    // Find _control port handle
    var controlSchema map[string]interface{}
    var controlData map[string]interface{}
    hasControlPort := false

    for _, handle := range handles {
      handleMap, ok := handle.(map[string]interface{})
      if !ok {
        continue
      }
      if handleMap["id"] != v1alpha1.ControlPort {
        continue
      }
      hasControlPort = true

      // Parse schema - handle various types that might come from ApiNodeToMap
      var schemaBytes []byte
      switch v := handleMap["schema"].(type) {
      case json.RawMessage:
        schemaBytes = v
      case []byte:
        schemaBytes = v
      case string:
        schemaBytes = []byte(v)
      }

      if len(schemaBytes) > 0 {
        controlSchema = make(map[string]interface{})
        _ = json.Unmarshal(schemaBytes, &controlSchema)
      }

      // Parse configuration/data - handle various types that might come from ApiNodeToMap
      var configBytes []byte
      switch v := handleMap["configuration"].(type) {
      case json.RawMessage:
        configBytes = v
      case []byte:
        configBytes = v
      case string:
        configBytes = []byte(v)
      }

      if len(configBytes) > 0 {
        controlData = make(map[string]interface{})
        _ = json.Unmarshal(configBytes, &controlData)
      }
      break
    }

    if !hasControlPort {
      continue
    }

    // Check if this widget should be shown on the current page
    portFullName := utils.GetPortFullName(node.Name, v1alpha1.ControlPort)
    assignedPages := widgetPages[portFullName]
    isOnCurrentPage := false
    for _, p := range assignedPages {
      if p == pageResourceName {
        isOnCurrentPage = true
        break
      }
    }

    // Widget visibility rules:
    // 1. If widget is assigned to current page -> show
    // 2. If widget is NOT assigned to ANY page AND current page is default -> show
    // 3. Otherwise -> don't show
    if !isOnCurrentPage {
      if len(assignedPages) > 0 || !isDefaultPage {
        // Widget is assigned to other pages, or we're not on default page
        continue
      }
      // Widget is unassigned and we're on the default page - show it
    }

    if controlSchema == nil {
      controlSchema = make(map[string]interface{})
    }
    if controlData == nil {
      controlData = make(map[string]interface{})
    }

    // Get grid position from widget page or calculate default positions
    // Widget positions are stored by port full name "nodeName:portName"
    // Default layout: 2 widgets per row (each width 3 in a 6-column grid)
    // Default height 4 units to fit form fields without scrolling
    gridX, gridY, gridW, gridH := 0, 0, 3, 4
    var customSchema map[string]interface{}
    var customTitle string
    if pos, ok := widgetPositions[portFullName]; ok {
      gridX, gridY, gridW, gridH = pos.GridX, pos.GridY, pos.GridW, pos.GridH
      // Use custom title from widget page if set
      customTitle = pos.Name
      // Apply SchemaPatch (JSON Patch RFC 6902) to original schema if present
      if len(pos.SchemaPatch) > 0 && controlSchema != nil {
        // SchemaPatch is stored as JSON Patch operations array
        patch, err := jsonpatchapply.DecodePatch(pos.SchemaPatch)
        if err == nil {
          // Apply patch to original schema
          originalBytes, err := json.Marshal(controlSchema)
          if err == nil {
            patchedBytes, err := patch.Apply(originalBytes)
            if err == nil {
              customSchema = make(map[string]interface{})
              _ = json.Unmarshal(patchedBytes, &customSchema)
            }
          }
        }
      }
    } else {
      // Calculate default position based on widget index
      // Alternating x: 0, 3, 0, 3, ...
      // Incrementing y every 2 widgets: 0, 0, 4, 4, 8, 8, ...
      gridX = (widgetIndex % 2) * 3
      gridY = (widgetIndex / 2) * 4
    }

    // Get title: use custom title from widget page, fall back to node annotation, then component name
    title := customTitle
    if title == "" {
      title = node.Annotations[v1alpha1.NodeLabelAnnotation]
    }
    if title == "" {
      title = node.Spec.Component
    }

    // Get pages this widget appears on (use already computed assignedPages)
    widgetPagesList := assignedPages
    if len(widgetPagesList) == 0 {
      // If widget isn't on any page yet, show it as belonging to current (default) page
      widgetPagesList = []string{pageResourceName}
    }

    result = append(result, Widget{
      ID:            node.Name,
      Title:         title,
      NodeName:      node.Name,
      Port:          v1alpha1.ControlPort,
      DefaultSchema: controlSchema,
      Schema:        customSchema,
      Data:          controlData,
      GridX:         gridX,
      GridY:         gridY,
      GridW:         gridW,
      GridH:         gridH,
      Pages:         widgetPagesList,
    })
    widgetIndex++
  }

  return result, nil
}

// DeleteProject deletes a project and all its resources
func (a *App) DeleteProject(contextName string, namespace string, projectName string) error {
  a.logger.Info("deleting project", "context", contextName, "namespace", namespace, "project", projectName)

  mgr, err := a.getManager(contextName, namespace)
  if err != nil {
    a.logger.Error(err, "failed to get manager for delete")
    return err
  }

  // First delete all flows and their nodes
  flows, err := mgr.GetFlowList(a.ctx, projectName)
  if err != nil {
    a.logger.Error(err, "failed to get flows for deletion", "project", projectName)
  } else {
    for _, flow := range flows {
      a.logger.Info("deleting flow nodes", "flow", flow.Name)
      // Delete nodes in this flow first
      if err := mgr.DeleteFlowNodes(a.ctx, projectName, flow.Name); err != nil {
        a.logger.Error(err, "failed to delete flow nodes", "flow", flow.Name)
      }
      a.logger.Info("deleting flow", "flow", flow.Name)
      if err := mgr.DeleteFlow(a.ctx, flow.Name); err != nil {
        a.logger.Error(err, "failed to delete flow", "flow", flow.Name)
      }
    }
  }

  // Delete all pages
  pages, err := mgr.GetProjectPageWidgets(a.ctx, projectName)
  if err != nil {
    a.logger.Error(err, "failed to get pages for deletion", "project", projectName)
  } else {
    for _, page := range pages {
      a.logger.Info("deleting page", "page", page.Name)
      if err := mgr.DeletePage(a.ctx, &page); err != nil {
        a.logger.Error(err, "failed to delete page", "page", page.Name)
      }
    }
  }

  // Finally delete the project itself
  if err := mgr.DeleteProject(a.ctx, projectName); err != nil {
    a.logger.Error(err, "failed to delete project resource", "project", projectName)
    return fmt.Errorf("failed to delete project: %w", err)
  }

  a.logger.Info("project deleted successfully", "project", projectName)
  return nil
}

// RenameProject renames a project by updating its name annotation
func (a *App) RenameProject(contextName string, namespace string, projectName string, newName string) error {
  if newName == "" {
    return fmt.Errorf("new name is required")
  }

  mgr, err := a.getManager(contextName, namespace)
  if err != nil {
    return err
  }

  return mgr.RenameProject(a.ctx, projectName, namespace, newName)
}

// SaveProjectDescription saves a project's description to the CRD
func (a *App) SaveProjectDescription(contextName string, namespace string, projectName string, description string) error {
  mgr, err := a.getManager(contextName, namespace)
  if err != nil {
    return err
  }

  return mgr.UpdateProjectDescription(a.ctx, projectName, namespace, description)
}

// CreateDashboardPage creates a new dashboard page for a project
func (a *App) CreateDashboardPage(contextName string, namespace string, projectName string, title string) (*WidgetPage, error) {
  if title == "" {
    return nil, fmt.Errorf("page title is required")
  }

  mgr, err := a.getManager(contextName, namespace)
  if err != nil {
    return nil, err
  }

  // Get existing pages to determine sort index
  pages, err := mgr.GetProjectPageWidgets(a.ctx, projectName)
  if err != nil {
    return nil, fmt.Errorf("unable to get widget pages: %w", err)
  }

  // Create new page with sort index = number of existing pages
  resourceName, err := mgr.CreatePage(a.ctx, title, projectName, namespace, len(pages))
  if err != nil {
    return nil, fmt.Errorf("failed to create page: %w", err)
  }

  return &WidgetPage{
    Name:         title,
    Title:        title,
    ResourceName: *resourceName,
    SortIdx:      len(pages),
  }, nil
}

// DeleteDashboardPage deletes a dashboard page from a project
func (a *App) DeleteDashboardPage(contextName string, namespace string, pageResourceName string) error {
  mgr, err := a.getManager(contextName, namespace)
  if err != nil {
    return err
  }

  page := &v1alpha1.TinyWidgetPage{}
  page.Name = pageResourceName
  page.Namespace = namespace

  if err := mgr.DeletePage(a.ctx, page); err != nil {
    return fmt.Errorf("failed to delete page: %w", err)
  }

  return nil
}

// SaveWidgets saves widget grid positions and page assignments
func (a *App) SaveWidgets(contextName string, namespace string, projectName string, pageResourceName string, widgets []Widget) error {
  mgr, err := a.getManager(contextName, namespace)
  if err != nil {
    return err
  }

  // Get all pages for the project
  pages, err := mgr.GetProjectPageWidgets(a.ctx, projectName)
  if err != nil {
    return fmt.Errorf("unable to get widget pages: %w", err)
  }

  // Build a map of pages by name for easy lookup
  pageMap := make(map[string]*v1alpha1.TinyWidgetPage)
  for i := range pages {
    pageMap[pages[i].Name] = &pages[i]
  }

  // Build existing widgets map per page to preserve SchemaPatch and other data
  existingWidgetsPerPage := make(map[string]map[string]v1alpha1.TinyWidget)
  for _, page := range pages {
    existingWidgetsPerPage[page.Name] = make(map[string]v1alpha1.TinyWidget)
    for _, w := range page.Spec.Widgets {
      existingWidgetsPerPage[page.Name][w.Port] = w
    }
  }

  // Track which widgets should be added/removed from other pages
  // Key: page name, Value: map of port -> widget (widgets to add/keep)
  otherPagesWidgets := make(map[string]map[string]v1alpha1.TinyWidget)
  // Track widgets to remove from other pages
  otherPagesRemove := make(map[string]map[string]bool)

  // Initialize other pages with their existing widgets
  for pageName, existingWidgets := range existingWidgetsPerPage {
    if pageName == pageResourceName {
      continue // Current page is handled separately
    }
    otherPagesWidgets[pageName] = make(map[string]v1alpha1.TinyWidget)
    otherPagesRemove[pageName] = make(map[string]bool)
    for port, w := range existingWidgets {
      otherPagesWidgets[pageName][port] = w
    }
  }

  // Build widgets for current page
  currentPageWidgets := make([]v1alpha1.TinyWidget, 0, len(widgets))

  // Process each widget from the save request
  for _, widget := range widgets {
    portFullName := utils.GetPortFullName(widget.NodeName, widget.Port)

    // Determine which pages this widget should be on
    targetPages := widget.Pages
    if len(targetPages) == 0 {
      targetPages = []string{pageResourceName}
    }

    // Check if widget should be on current page
    shouldBeOnCurrentPage := false
    for _, p := range targetPages {
      if p == pageResourceName {
        shouldBeOnCurrentPage = true
        break
      }
    }

    if shouldBeOnCurrentPage {
      // Build widget for current page with full position/schema data
      var w v1alpha1.TinyWidget
      if existing, ok := existingWidgetsPerPage[pageResourceName][portFullName]; ok {
        w = existing
      }
      w.Name = widget.Title
      w.Port = portFullName
      w.GridX = widget.GridX
      w.GridY = widget.GridY
      w.GridW = widget.GridW
      w.GridH = widget.GridH

      // Create SchemaPatch as JSON Patch (RFC 6902) comparing DefaultSchema to Schema
      // This matches how the platform stores schema customizations
      if widget.Schema != nil && len(widget.Schema) > 0 && widget.DefaultSchema != nil && len(widget.DefaultSchema) > 0 {
        // Marshal both schemas
        originalBytes, err1 := json.Marshal(widget.DefaultSchema)
        modifiedBytes, err2 := json.Marshal(widget.Schema)
        if err1 == nil && err2 == nil && string(originalBytes) != string(modifiedBytes) {
          // Create JSON Patch
          patch, err := jsonpatch.CreatePatch(originalBytes, modifiedBytes)
          if err == nil && len(patch) > 0 {
            patchBytes, err := json.Marshal(patch)
            if err == nil {
              w.SchemaPatch = patchBytes
            }
          }
        } else if string(originalBytes) == string(modifiedBytes) {
          // Schemas are the same, clear patch
          w.SchemaPatch = nil
        }
      } else if widget.Schema == nil || len(widget.Schema) == 0 {
        // Clear SchemaPatch if Schema is empty (reset to default)
        w.SchemaPatch = nil
      }

      currentPageWidgets = append(currentPageWidgets, w)
    }

    // Handle other pages - add or remove widget as needed
    for pageName := range otherPagesWidgets {
      shouldBeOnThisPage := false
      for _, p := range targetPages {
        if p == pageName {
          shouldBeOnThisPage = true
          break
        }
      }

      if shouldBeOnThisPage {
        // Widget should be on this page - add or update it
        var w v1alpha1.TinyWidget
        if existing, ok := existingWidgetsPerPage[pageName][portFullName]; ok {
          w = existing
        } else {
          // New widget on this page - use default position
          w.GridX = 0
          w.GridY = 0
          w.GridW = widget.GridW
          w.GridH = widget.GridH
        }
        w.Name = widget.Title
        w.Port = portFullName
        otherPagesWidgets[pageName][portFullName] = w
      } else {
        // Widget should NOT be on this page - mark for removal
        otherPagesRemove[pageName][portFullName] = true
      }
    }
  }

  // Update current page
  currentPage := pageMap[pageResourceName]
  if currentPage == nil {
    return fmt.Errorf("page not found: %s", pageResourceName)
  }
  currentPage.Spec.Widgets = currentPageWidgets
  if err := mgr.UpdatePage(a.ctx, currentPage); err != nil {
    return fmt.Errorf("failed to update page %s: %w", pageResourceName, err)
  }

  // Update other pages only if there are changes
  for pageName, widgetsMap := range otherPagesWidgets {
    page := pageMap[pageName]
    if page == nil {
      continue
    }

    // Remove widgets marked for removal
    for port := range otherPagesRemove[pageName] {
      delete(widgetsMap, port)
    }

    // Check if this page actually changed
    existingPorts := make(map[string]bool)
    for _, w := range page.Spec.Widgets {
      existingPorts[w.Port] = true
    }
    newPorts := make(map[string]bool)
    for port := range widgetsMap {
      newPorts[port] = true
    }

    // Only update if widgets changed
    needsUpdate := len(existingPorts) != len(newPorts)
    if !needsUpdate {
      for port := range existingPorts {
        if !newPorts[port] {
          needsUpdate = true
          break
        }
      }
    }
    if !needsUpdate {
      for port := range newPorts {
        if !existingPorts[port] {
          needsUpdate = true
          break
        }
      }
    }

    if needsUpdate {
      // Convert map to slice
      newWidgets := make([]v1alpha1.TinyWidget, 0, len(widgetsMap))
      for _, w := range widgetsMap {
        newWidgets = append(newWidgets, w)
      }
      page.Spec.Widgets = newWidgets
      if err := mgr.UpdatePage(a.ctx, page); err != nil {
        return fmt.Errorf("failed to update page %s: %w", pageName, err)
      }
    }
  }

  return nil
}

// SendSignal sends a signal to a node's control port
func (a *App) SendSignal(contextName string, namespace string, nodeName string, port string, data string) error {
  mgr, err := a.getManager(contextName, namespace)
  if err != nil {
    return err
  }

  return mgr.CreateSignal(a.ctx, nodeName, namespace, port, []byte(data))
}

// WatchProjectNodes starts watching nodes for real-time updates
func (a *App) WatchProjectNodes(contextName string, namespace string, projectName string) error {
  a.watchMu.Lock()
  defer a.watchMu.Unlock()

  // Stop any existing watcher
  if a.watchCancel != nil {
    a.watchCancel()
  }

  mgr, err := a.getManager(contextName, namespace)
  if err != nil {
    return err
  }

  watchCtx, cancel := context.WithCancel(a.ctx)
  a.watchCancel = cancel

  watcher, err := mgr.WatchNodes(watchCtx, projectName)
  if err != nil {
    return fmt.Errorf("unable to start watch: %w", err)
  }

  // Process events in goroutine with automatic reconnection
  go func() {
    currentWatcher := watcher
    defer func() {
      if currentWatcher != nil {
        currentWatcher.Stop()
      }
    }()

    for {
      select {
      case <-watchCtx.Done():
        return
      case event, ok := <-currentWatcher.ResultChan():
        if !ok {
          // Channel closed - try to reconnect
          currentWatcher.Stop()

          select {
          case <-watchCtx.Done():
            return
          case <-time.After(time.Second):
          }

          newWatcher, err := mgr.WatchNodes(watchCtx, projectName)
          if err != nil {
            return
          }
          currentWatcher = newWatcher
          continue
        }

        node, ok := event.Object.(*v1alpha1.TinyNode)
        if !ok {
          continue
        }

        // Only process dashboard nodes
        if node.Labels[v1alpha1.DashboardLabel] != "true" {
          continue
        }

        update := NodeUpdate{
          EventType: string(event.Type),
          NodeName:  node.Name,
        }

        if event.Type != watch.Deleted {
          update.Widget = parseNodeToWidget(node)
        }

        wailsruntime.EventsEmit(a.ctx, "nodeUpdate", update)
      }
    }
  }()

  return nil
}

// StopWatchProjectNodes stops the node watcher
func (a *App) StopWatchProjectNodes() error {
  a.watchMu.Lock()
  defer a.watchMu.Unlock()

  if a.watchCancel != nil {
    a.watchCancel()
    a.watchCancel = nil
  }
  return nil
}

// Type aliases for SDK export types (for backward compatibility)

// ExportProject exports a project to JSON format
func (a *App) ExportProject(contextName string, namespace string, projectName string) (string, error) {
	mgr, err := a.getManager(contextName, namespace)
	if err != nil {
		return "", err
	}

	// Get all flows
	flows, err := mgr.GetFlowList(a.ctx, projectName)
	if err != nil {
		return "", fmt.Errorf("unable to get flows: %w", err)
	}

	// Get all nodes
	allNodes, err := mgr.GetProjectNodes(a.ctx, projectName)
	if err != nil {
		return "", fmt.Errorf("unable to get nodes: %w", err)
	}

	allNodesMap := make(map[string]v1alpha1.TinyNode, len(allNodes))
	for _, node := range allNodes {
		allNodesMap[node.Name] = node
	}

	// Build export flows
	var exportFlows []utils.ExportFlow
	for _, flow := range flows {
		name := flow.Annotations[v1alpha1.FlowDescriptionAnnotation]
		if name == "" {
			name = flow.Name
		}
		exportFlows = append(exportFlows, utils.ExportFlow{
			ResourceName: flow.Name,
			Name:         name,
		})
	}

	// Convert nodes to elements
	nodeElements, edgeElements, err := utils.NodesToGraphWithOptions(allNodesMap, nil, false)
	if err != nil {
		return "", fmt.Errorf("failed to convert nodes to graph: %w", err)
	}

	// Combine elements and add flow field
	elements := make([]map[string]interface{}, 0, len(nodeElements)+len(edgeElements))

	for _, elem := range nodeElements {
		if m, ok := elem.(map[string]interface{}); ok {
			nodeID, _ := m["id"].(string)
			if node, exists := allNodesMap[nodeID]; exists {
				m["flow"] = node.Labels[v1alpha1.FlowNameLabel]
			}
			// Strip _control port configuration — it contains runtime state
			// that may include sensitive data (tokens, secrets from user input)
			stripControlPortConfig(m)
			elements = append(elements, m)
		}
	}

	for _, elem := range edgeElements {
		if m, ok := elem.(map[string]interface{}); ok {
			sourceID, _ := m["source"].(string)
			if node, exists := allNodesMap[sourceID]; exists {
				m["flow"] = node.Labels[v1alpha1.FlowNameLabel]
			}
			elements = append(elements, m)
		}
	}

	// Get dashboard pages
	pages, err := mgr.GetProjectPageWidgets(a.ctx, projectName)
	if err != nil {
		return "", fmt.Errorf("unable to get widget pages: %w", err)
	}

	// Build export pages
	var exportPages []utils.ExportPage
	for _, page := range pages {
		sortIdx, _ := strconv.Atoi(page.Annotations[v1alpha1.PageSortIdxAnnotation])
		title := page.Annotations[v1alpha1.PageTitleAnnotation]
		if title == "" {
			title = page.Name
		}

		var widgets []utils.ExportWidget
		for _, w := range page.Spec.Widgets {
			ew := utils.ExportWidget{
				Port:        w.Port,
				Name:        w.Name,
				GridX:       w.GridX,
				GridY:       w.GridY,
				GridW:       w.GridW,
				GridH:       w.GridH,
				SchemaPatch: w.SchemaPatch,
			}
			widgets = append(widgets, ew)
		}

		exportPages = append(exportPages, utils.ExportPage{
			Name:    page.Name,
			Title:   title,
			SortIdx: sortIdx,
			Widgets: widgets,
		})
	}

	// Get project description from CRD
	var projectDescription string
	project, err := mgr.GetProject(a.ctx, projectName, namespace)
	if err == nil {
		projectDescription = project.Spec.Description
	}

	// Build export object
	export := utils.ProjectExport{
		Version:     utils.CurrentExportVersion,
		Description: projectDescription,
		TinyFlows:   exportFlows,
		Elements:    elements,
		Pages:       exportPages,
	}

	data, err := json.MarshalIndent(export, "", "  ")
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// ImportProject imports JSON data into an existing project
func (a *App) ImportProject(contextName string, namespace string, projectName string, jsonData string) error {
	// Create a dedicated context with longer timeout for import operations
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Parse import data
	var importData utils.ProjectExport
	if err := json.Unmarshal([]byte(jsonData), &importData); err != nil {
		return fmt.Errorf("invalid import data: %v", err)
	}

	if importData.Version != utils.CurrentExportVersion {
		return fmt.Errorf("unsupported import version: %d", importData.Version)
	}

	// Validate import data and log warnings
	utils.ValidateProjectExport(&importData)

	mgr, err := a.getManager(contextName, namespace)
	if err != nil {
		return err
	}

	// Save project description from import if present
	if importData.Description != "" {
		if descErr := mgr.UpdateProjectDescription(ctx, projectName, namespace, importData.Description); descErr != nil {
			a.logger.Error(descErr, "failed to save project description from import")
		}
	}

	// Get existing flows
	existingFlows, err := mgr.GetFlowList(ctx, projectName)
	if err != nil {
		return fmt.Errorf("unable to get existing flows: %w", err)
	}

	existingFlowNames := make(map[string]bool)
	for _, flow := range existingFlows {
		existingFlowNames[flow.Name] = true
	}

	// Get existing nodes
	existingNodes, err := mgr.GetProjectNodes(ctx, projectName)
	if err != nil {
		return fmt.Errorf("unable to get existing nodes: %w", err)
	}

	existingNodesMap := make(map[string]v1alpha1.TinyNode)
	for _, node := range existingNodes {
		existingNodesMap[node.Name] = node
	}

	// Create flows that don't exist
	flowResourceNameMap := make(map[string]string) // old name -> new name
	for _, importFlow := range importData.TinyFlows {
		a.logger.Info("processing flow", "oldResourceName", importFlow.ResourceName, "displayName", importFlow.Name)
		if existingFlowNames[importFlow.ResourceName] {
			flowResourceNameMap[importFlow.ResourceName] = importFlow.ResourceName
			a.logger.Info("flow already exists", "resourceName", importFlow.ResourceName)
			continue
		}
		newResourceName, err := mgr.CreateFlow(ctx, namespace, projectName, importFlow.Name)
		if err != nil {
			a.logger.Error(err, "failed to create flow", "name", importFlow.Name)
			continue
		}
		flowResourceNameMap[importFlow.ResourceName] = *newResourceName
		a.logger.Info("created new flow", "oldResourceName", importFlow.ResourceName, "newResourceName", *newResourceName)
	}
	a.logger.Info("flow mapping complete", "mappings", flowResourceNameMap)

	// Track errors for reporting
	var failedNodes []string
	var importedNodes int

	// Map old node IDs to new node names (for edge translation)
	nodeIDMap := make(map[string]string) // oldID -> newName

	// Import nodes
	for _, elem := range importData.Elements {
		elemType, _ := elem["type"].(string)
		if elemType == "edge" || elemType == "tinyEdge" || elemType == "" {
			continue
		}

		oldNodeID, _ := elem["id"].(string)
		if oldNodeID == "" {
			continue
		}

		// Check if node already exists — update it instead of skipping
		if existing, exists := existingNodesMap[oldNodeID]; exists {
			nodeIDMap[oldNodeID] = oldNodeID
			a.updateExistingNode(ctx, &existing, elem, mgr)
			continue
		}

		// Get flow for this node
		flowName, _ := elem["flow"].(string)
		if flowName == "" {
			a.logger.Info("skipping node - no flow name", "nodeID", oldNodeID)
			continue
		}
		newFlowName := flowResourceNameMap[flowName]
		if newFlowName == "" {
			a.logger.Info("skipping node - flow not in map", "nodeID", oldNodeID, "flowName", flowName, "availableFlows", flowResourceNameMap)
			continue
		}
		a.logger.Info("creating node", "nodeID", oldNodeID, "oldFlow", flowName, "newFlow", newFlowName)

		data, _ := elem["data"].(map[string]interface{})
		if data == nil {
			continue
		}

		component, _ := data["component"].(string)
		module, _ := data["module"].(string)
		version, _ := data["module_version"].(string)
		if version == "" {
			version, _ = data["version"].(string)
		}
		if component == "" || module == "" {
			continue
		}

		// Get position
		position, _ := elem["position"].(map[string]interface{})
		posX, posY := 0, 0
		if position != nil {
			if x, ok := position["x"].(float64); ok {
				posX = int(x)
			}
			if y, ok := position["y"].(float64); ok {
				posY = int(y)
			}
		}

		// Create node with proper naming: {hash}.{module}.{component}-{suffix}
		// Use suffix from original node ID if available to preserve uniqueness
		// when multiple nodes share the same component type (e.g., two Tickers)
		nodeGenerateName := utils.GetNodeGenerateName(projectName, newFlowName, module, component)
		suffix := strconv.FormatInt(time.Now().UnixNano(), 36)[:5]
		if oldNodeID != "" {
			if idx := strings.LastIndex(oldNodeID, "-"); idx >= 0 && idx < len(oldNodeID)-1 {
				suffix = oldNodeID[idx+1:]
			}
		}
		nodeName := nodeGenerateName + suffix

		// Get label from data
		label, _ := data["label"].(string)
		if label == "" {
			label = component
		}

		// Get spin value
		spin := 0
		if spinVal, ok := data["spin"].(float64); ok {
			spin = int(spinVal)
		}

		// Get dashboard flag
		dashboard, _ := data["dashboard"].(string)

		labels := map[string]string{
			v1alpha1.FlowNameLabel:    newFlowName,
			v1alpha1.ProjectNameLabel: projectName,
		}
		if dashboard == "true" {
			labels[v1alpha1.DashboardLabel] = "true"
		}

		// Extract port configurations from handles
		var ports []v1alpha1.TinyNodePortConfig
		if handles, ok := data["handles"].([]interface{}); ok {
			a.logger.Info("processing handles for node", "component", component, "handleCount", len(handles))
			for _, h := range handles {
				handle, ok := h.(map[string]interface{})
				if !ok {
					continue
				}
				portID, _ := handle["id"].(string)
				if portID == "" {
					continue
				}

				// Get configuration and schema from handle
				config := handle["configuration"]
				schema := handle["schema"]

				// Marshal config to JSON for storage
				var configBytes []byte
				if config != nil {
					var err error
					configBytes, err = json.Marshal(config)
					if err != nil {
						a.logger.Error(err, "failed to marshal port config", "port", portID)
						continue
					}
					// Check if empty
					if string(configBytes) == "{}" || string(configBytes) == "null" {
						configBytes = nil
					}
				}

				// Marshal schema to JSON for storage
				var schemaBytes []byte
				if schema != nil {
					var err error
					schemaBytes, err = json.Marshal(schema)
					if err != nil {
						a.logger.Error(err, "failed to marshal port schema", "port", portID)
					}
				}

				// Skip if both config and schema are empty
				if len(configBytes) == 0 && len(schemaBytes) == 0 {
					continue
				}

				ports = append(ports, v1alpha1.TinyNodePortConfig{
					Port:          portID,
					Configuration: configBytes,
					Schema:        schemaBytes,
				})
				a.logger.Info("added port config", "port", portID, "configLen", len(configBytes), "schemaLen", len(schemaBytes))
			}
		} else {
			a.logger.Info("no handles found for node", "component", component, "dataKeys", getMapKeys(data))
		}

		node := &v1alpha1.TinyNode{
			ObjectMeta: metav1.ObjectMeta{
				Name:      nodeName,
				Namespace: namespace,
				Labels:    labels,
				Annotations: map[string]string{
					v1alpha1.ComponentPosXAnnotation:   strconv.Itoa(posX),
					v1alpha1.ComponentPosYAnnotation:   strconv.Itoa(posY),
					v1alpha1.ComponentPosSpinAnnotation: strconv.Itoa(spin),
					v1alpha1.NodeLabelAnnotation:       label,
				},
			},
			Spec: v1alpha1.TinyNodeSpec{
				Module:        module,
				ModuleVersion: version,
				Component:     component,
				Ports:         ports,
			},
		}

		// Use async CreateNode - don't wait for sync (too slow)
		if err := mgr.CreateNode(ctx, node); err != nil {
			a.logger.Error(err, "failed to create node", "component", component)
			failedNodes = append(failedNodes, component)
		} else {
			a.logger.Info("imported node", "component", component, "name", nodeName, "portsCount", len(ports))
			nodeIDMap[oldNodeID] = nodeName
			importedNodes++
		}
	}

	// Wait for nodes to be created before adding edges
	if len(nodeIDMap) > 0 {
		a.logger.Info("waiting for nodes to be ready before adding edges", "nodeCount", len(nodeIDMap))
		time.Sleep(3 * time.Second)
	}

	// Import edges - collect all updates per node to do a single update
	// This prevents race conditions where controller reconciliation between updates could reset data
	edgesBySourceNode := make(map[string][]v1alpha1.TinyNodeEdge)
	portConfigsByTargetNode := make(map[string][]v1alpha1.TinyNodePortConfig)

	for _, elem := range importData.Elements {
		elemType, _ := elem["type"].(string)
		if elemType != "edge" && elemType != "tinyEdge" {
			continue
		}

		oldSourceID, _ := elem["source"].(string)
		sourceHandle, _ := elem["sourceHandle"].(string)
		oldTargetID, _ := elem["target"].(string)
		targetHandle, _ := elem["targetHandle"].(string)
		flowName, _ := elem["flow"].(string)

		// Log the raw edge data for debugging
		edgeDataRaw, _ := json.Marshal(elem["data"])
		a.logger.Info("processing edge", "source", oldSourceID, "sourceHandle", sourceHandle, "target", oldTargetID, "targetHandle", targetHandle, "dataRaw", string(edgeDataRaw))

		// Translate old IDs to new names
		newSourceName := nodeIDMap[oldSourceID]
		newTargetName := nodeIDMap[oldTargetID]
		newFlowName := flowResourceNameMap[flowName]

		if newSourceName == "" || newTargetName == "" || newFlowName == "" {
			a.logger.Info("skipping edge - missing node mapping", "source", oldSourceID, "target", oldTargetID)
			continue
		}

		// Generate new edge ID with new node names: {source}_{sourcePort}-{target}_{targetPort}
		newEdgeID := fmt.Sprintf("%s_%s-%s_%s", newSourceName, sourceHandle, newTargetName, targetHandle)

		edge := v1alpha1.TinyNodeEdge{
			ID:     newEdgeID,
			Port:   sourceHandle,
			To:     newTargetName + ":" + targetHandle,
			FlowID: newFlowName,
		}
		edgesBySourceNode[newSourceName] = append(edgesBySourceNode[newSourceName], edge)

		// Extract edge configuration from elem["data"]["configuration"]
		// This needs to be added as a port config on the TARGET node
		edgeData, hasData := elem["data"].(map[string]interface{})
		if !hasData {
			a.logger.Info("edge has no data map", "edge", newEdgeID)
			continue
		}

		var configBytes []byte
		var schemaBytes []byte

		config := edgeData["configuration"]
		a.logger.Info("edge configuration check", "edge", newEdgeID, "hasConfig", config != nil, "configType", fmt.Sprintf("%T", config))

		if config != nil {
			var err error
			configBytes, err = json.Marshal(config)
			if err != nil {
				a.logger.Error(err, "failed to marshal edge config", "edge", newEdgeID)
			} else {
				a.logger.Info("edge config marshaled", "edge", newEdgeID, "configLen", len(configBytes), "configPreview", truncateString(string(configBytes), 200))
			}
		}
		if edgeSchema := edgeData["schema"]; edgeSchema != nil {
			var err error
			schemaBytes, err = json.Marshal(edgeSchema)
			if err != nil {
				a.logger.Error(err, "failed to marshal edge schema", "edge", newEdgeID)
			}
		}

		// Only add port config if there's actual configuration data
		if len(configBytes) > 0 || len(schemaBytes) > 0 {
			// Create port config for target node
			// From = source port full name, Port = target port name
			sourcePortFullName := newSourceName + ":" + sourceHandle
			portConfig := v1alpha1.TinyNodePortConfig{
				From:          sourcePortFullName,
				Port:          targetHandle,
				Configuration: configBytes,
				Schema:        schemaBytes,
				FlowID:        newFlowName,
			}
			portConfigsByTargetNode[newTargetName] = append(portConfigsByTargetNode[newTargetName], portConfig)
			a.logger.Info("prepared edge port config", "target", newTargetName, "port", targetHandle, "from", sourcePortFullName, "configLen", len(configBytes), "schemaLen", len(schemaBytes))
		}
	}

	// Collect all nodes that need updates
	allNodesToUpdate := make(map[string]bool)
	for nodeName := range edgesBySourceNode {
		allNodesToUpdate[nodeName] = true
	}
	for nodeName := range portConfigsByTargetNode {
		allNodesToUpdate[nodeName] = true
	}

	a.logger.Info("edge port configs summary",
		"totalEdges", len(edgesBySourceNode),
		"totalPortConfigs", len(portConfigsByTargetNode),
		"nodesToUpdate", len(allNodesToUpdate))

	// Update each node ONCE with all its changes (both edges and port configs)
	// This prevents race conditions with controller reconciliation
	for nodeName := range allNodesToUpdate {
		node, err := mgr.GetNode(ctx, nodeName, namespace)
		if err != nil {
			a.logger.Error(err, "failed to get node for update", "node", nodeName)
			continue
		}

		existingPortsCount := len(node.Spec.Ports)
		existingEdgesCount := len(node.Spec.Edges)

		// Merge edges: replace existing by ID, append new
		if edges, ok := edgesBySourceNode[nodeName]; ok {
			existingEdgeMap := make(map[string]int) // edgeID -> index
			for i, e := range node.Spec.Edges {
				existingEdgeMap[e.ID] = i
			}
			for _, edge := range edges {
				if idx, exists := existingEdgeMap[edge.ID]; exists {
					node.Spec.Edges[idx] = edge
				} else {
					node.Spec.Edges = append(node.Spec.Edges, edge)
				}
			}
			a.logger.Info("merged edges on node", "node", nodeName, "importedEdgeCount", len(edges))
		}

		// Merge port configs: replace existing by From+Port key, append new
		if portConfigs, ok := portConfigsByTargetNode[nodeName]; ok {
			existingPortMap := make(map[string]int) // "from|port" -> index
			for i, pc := range node.Spec.Ports {
				if pc.From != "" {
					existingPortMap[pc.From+"|"+pc.Port] = i
				}
			}
			for _, pc := range portConfigs {
				key := pc.From + "|" + pc.Port
				if idx, exists := existingPortMap[key]; exists {
					node.Spec.Ports[idx] = pc
				} else {
					node.Spec.Ports = append(node.Spec.Ports, pc)
				}
			}
			a.logger.Info("merged port configs on node", "node", nodeName, "importedPortConfigCount", len(portConfigs))
		}

		if err := mgr.UpdateNode(ctx, node); err != nil {
			a.logger.Error(err, "failed to update node", "node", nodeName)
		} else {
			a.logger.Info("updated node successfully", "node", nodeName,
				"existingEdges", existingEdgesCount, "totalEdges", len(node.Spec.Edges),
				"existingPorts", existingPortsCount, "totalPorts", len(node.Spec.Ports))
		}
	}

	// Verification: Read back nodes and check if port configs were persisted
	a.logger.Info("verifying port configs were persisted...")
	time.Sleep(1 * time.Second) // Small delay to let k8s process
	for nodeName := range portConfigsByTargetNode {
		verifyNode, err := mgr.GetNode(ctx, nodeName, namespace)
		if err != nil {
			a.logger.Error(err, "failed to verify node", "node", nodeName)
			continue
		}
		edgePortConfigCount := 0
		for _, pc := range verifyNode.Spec.Ports {
			if pc.From != "" {
				edgePortConfigCount++
				a.logger.Info("verified edge port config",
					"node", nodeName,
					"from", pc.From,
					"port", pc.Port,
					"configLen", len(pc.Configuration))
			}
		}
		a.logger.Info("verification result", "node", nodeName,
			"totalPorts", len(verifyNode.Spec.Ports),
			"edgePortConfigs", edgePortConfigCount)
	}

	// Import pages with widgets
	var failedPages []string
	existingPages, err := mgr.GetProjectPageWidgets(ctx, projectName)
	if err != nil {
		return fmt.Errorf("unable to get existing pages: %w", err)
	}
	a.logger.Info("existing pages before import", "count", len(existingPages), "projectName", projectName)

	// Build map of existing page titles to avoid duplicates
	existingPageTitles := make(map[string]bool)
	for _, page := range existingPages {
		title := page.Annotations[v1alpha1.PageTitleAnnotation]
		if title == "" {
			title = page.Name
		}
		existingPageTitles[title] = true
	}

	for _, importPage := range importData.Pages {
		// Check by title, not by resource name (resource names are regenerated)
		if existingPageTitles[importPage.Title] {
			a.logger.Info("skipping page - already exists", "title", importPage.Title)
			continue
		}

		// Build widgets with translated port references first
		var widgets []v1alpha1.TinyWidget
		for _, importWidget := range importPage.Widgets {
			// Translate port reference: "oldNodeID:portName" -> "newNodeName:portName"
			portParts := strings.SplitN(importWidget.Port, ":", 2)
			if len(portParts) != 2 {
				a.logger.Info("skipping widget - invalid port format", "port", importWidget.Port)
				continue
			}
			oldNodeID := portParts[0]
			portName := portParts[1]

			newNodeName := nodeIDMap[oldNodeID]
			if newNodeName == "" {
				a.logger.Info("skipping widget - node not found in map", "oldNodeID", oldNodeID, "availableNodes", len(nodeIDMap))
				continue
			}

			newPort := newNodeName + ":" + portName
			widget := v1alpha1.TinyWidget{
				Port:  newPort,
				Name:  importWidget.Name,
				GridX: importWidget.GridX,
				GridY: importWidget.GridY,
				GridW: importWidget.GridW,
				GridH: importWidget.GridH,
			}
			if len(importWidget.SchemaPatch) > 0 {
				widget.SchemaPatch = []byte(importWidget.SchemaPatch)
			}
			widgets = append(widgets, widget)
			a.logger.Info("translated widget port", "oldPort", importWidget.Port, "newPort", newPort, "widgetName", importWidget.Name)
		}

		// Create the page
		a.logger.Info("creating page", "title", importPage.Title, "project", projectName, "namespace", namespace, "sortIdx", importPage.SortIdx)
		newPageName, err := mgr.CreatePage(ctx, importPage.Title, projectName, namespace, importPage.SortIdx)
		if err != nil {
			a.logger.Error(err, "failed to create page", "title", importPage.Title)
			failedPages = append(failedPages, importPage.Title)
			continue
		}
		if newPageName == nil {
			a.logger.Error(nil, "CreatePage returned nil name without error", "title", importPage.Title)
			failedPages = append(failedPages, importPage.Title)
			continue
		}
		a.logger.Info("page created successfully", "title", importPage.Title, "resourceName", *newPageName)

		// Update page with widgets if any
		if len(widgets) > 0 {
			// Small delay to ensure page is queryable
			time.Sleep(500 * time.Millisecond)

			// Fetch the page we just created to update it
			allPages, err := mgr.GetProjectPageWidgets(ctx, projectName)
			if err != nil {
				a.logger.Error(err, "failed to get pages after creation", "project", projectName)
				continue
			}
			a.logger.Info("fetched pages after creation", "count", len(allPages))

			found := false
			for i := range allPages {
				if allPages[i].Name == *newPageName {
					found = true
					allPages[i].Spec.Widgets = widgets
					if err := mgr.UpdatePage(ctx, &allPages[i]); err != nil {
						a.logger.Error(err, "failed to update page with widgets", "page", *newPageName)
					} else {
						a.logger.Info("added widgets to page", "page", *newPageName, "widgetCount", len(widgets))
					}
					break
				}
			}
			if !found {
				a.logger.Error(nil, "created page not found when trying to add widgets", "expectedName", *newPageName, "availablePages", len(allPages))
			}
		} else {
			a.logger.Info("no widgets to add to page", "page", *newPageName)
		}
	}

	// Final verification
	finalPages, _ := mgr.GetProjectPageWidgets(ctx, projectName)
	a.logger.Info("pages after import complete", "count", len(finalPages), "projectName", projectName)

	// Build detailed summary
	var summary strings.Builder
	summary.WriteString(fmt.Sprintf("Import complete: %d nodes imported", importedNodes))
	if len(failedNodes) > 0 {
		summary.WriteString(fmt.Sprintf(", %d nodes failed (%v)", len(failedNodes), failedNodes))
	}
	summary.WriteString(fmt.Sprintf(", %d pages imported", len(finalPages)))
	if len(failedPages) > 0 {
		summary.WriteString(fmt.Sprintf(", %d pages failed (%v)", len(failedPages), failedPages))
	}
	summary.WriteString(fmt.Sprintf(", nodeIDMap has %d entries", len(nodeIDMap)))
	a.logger.Info(summary.String())

	// Return error summary if any failures
	if len(failedNodes) > 0 || len(failedPages) > 0 {
		return fmt.Errorf("import completed with errors: %d nodes imported, %d nodes failed (%v), %d pages failed (%v)",
			importedNodes, len(failedNodes), failedNodes, len(failedPages), failedPages)
	}

	return nil
}

// updateExistingNode updates an existing node with imported data (position, label, module version, port configs from handles)
func (a *App) updateExistingNode(ctx context.Context, node *v1alpha1.TinyNode, elem map[string]interface{}, mgr resource.ManagerInterface) {
	data, _ := elem["data"].(map[string]interface{})
	if data == nil {
		a.logger.Info("import element has no data, skipping update", "node", node.Name)
		return
	}

	if node.Annotations == nil {
		node.Annotations = make(map[string]string)
	}

	// Update position
	if position, _ := elem["position"].(map[string]interface{}); position != nil {
		if x, ok := position["x"].(float64); ok {
			node.Annotations[v1alpha1.ComponentPosXAnnotation] = strconv.Itoa(int(x))
		}
		if y, ok := position["y"].(float64); ok {
			node.Annotations[v1alpha1.ComponentPosYAnnotation] = strconv.Itoa(int(y))
		}
	}

	// Update spin
	if spinVal, ok := data["spin"].(float64); ok {
		node.Annotations[v1alpha1.ComponentPosSpinAnnotation] = strconv.Itoa(int(spinVal))
	}

	// Update label
	if label, ok := data["label"].(string); ok && label != "" {
		node.Annotations[v1alpha1.NodeLabelAnnotation] = label
	}

	// Update module version
	if version, ok := data["module_version"].(string); ok && version != "" {
		node.Spec.ModuleVersion = version
	}

	// Rebuild port configs from handles (same logic as node creation)
	var handlePorts []v1alpha1.TinyNodePortConfig
	if handles, ok := data["handles"].([]interface{}); ok {
		for _, h := range handles {
			handle, ok := h.(map[string]interface{})
			if !ok {
				continue
			}
			portID, _ := handle["id"].(string)
			if portID == "" {
				continue
			}

			config := handle["configuration"]
			schema := handle["schema"]

			var configBytes []byte
			if config != nil {
				var err error
				configBytes, err = json.Marshal(config)
				if err != nil {
					a.logger.Error(err, "failed to marshal port config", "port", portID)
					continue
				}
				if string(configBytes) == "{}" || string(configBytes) == "null" {
					configBytes = nil
				}
			}

			var schemaBytes []byte
			if schema != nil {
				var err error
				schemaBytes, err = json.Marshal(schema)
				if err != nil {
					a.logger.Error(err, "failed to marshal port schema", "port", portID)
				}
			}

			if len(configBytes) == 0 && len(schemaBytes) == 0 {
				continue
			}

			handlePorts = append(handlePorts, v1alpha1.TinyNodePortConfig{
				Port:          portID,
				Configuration: configBytes,
				Schema:        schemaBytes,
			})
		}
	}

	// Replace handle-level port configs (From=""), keep edge-level port configs (From!="")
	var edgePorts []v1alpha1.TinyNodePortConfig
	for _, pc := range node.Spec.Ports {
		if pc.From != "" {
			edgePorts = append(edgePorts, pc)
		}
	}
	node.Spec.Ports = append(handlePorts, edgePorts...)

	if err := mgr.UpdateNode(ctx, node); err != nil {
		a.logger.Error(err, "failed to update existing node", "node", node.Name)
	} else {
		a.logger.Info("updated existing node", "node", node.Name, "handlePorts", len(handlePorts))
	}
}

// getMapKeys returns the keys of a map for logging purposes
func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// truncateString truncates a string to maxLen characters
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// stripControlPortConfig removes configuration from _control port handles
// in a node element map. _control ports reflect runtime state and may contain
// sensitive data (e.g. tokens, secrets passed through the flow).
func stripControlPortConfig(nodeMap map[string]interface{}) {
	data, ok := nodeMap["data"].(map[string]interface{})
	if !ok {
		return
	}
	handles, ok := data["handles"].([]interface{})
	if !ok {
		return
	}
	for _, h := range handles {
		handle, ok := h.(map[string]interface{})
		if !ok {
			continue
		}
		if handle["id"] == v1alpha1.ControlPort {
			delete(handle, "configuration")
		}
	}
}

// parseNodeToWidget converts a TinyNode to a Widget
// Uses ApiNodeToMap to properly resolve schema definitions (same as GetWidgets)
func parseNodeToWidget(node *v1alpha1.TinyNode) *Widget {
  // Use ApiNodeToMap to get properly resolved schema with configurable definitions
  nodeMap := utils.ApiNodeToMap(*node, nil, false)
  nodeData, ok := nodeMap["data"].(map[string]interface{})
  if !ok {
    return nil
  }

  handles, ok := nodeData["handles"].([]interface{})
  if !ok {
    return nil
  }

  // Find _control port handle
  var controlSchema map[string]interface{}
  var controlData map[string]interface{}
  hasControlPort := false

  for _, handle := range handles {
    handleMap, ok := handle.(map[string]interface{})
    if !ok {
      continue
    }
    if handleMap["id"] != v1alpha1.ControlPort {
      continue
    }
    hasControlPort = true

    // Parse schema - handle various types that might come from ApiNodeToMap
    var schemaBytes []byte
    switch v := handleMap["schema"].(type) {
    case json.RawMessage:
      schemaBytes = v
    case []byte:
      schemaBytes = v
    case string:
      schemaBytes = []byte(v)
    }

    if len(schemaBytes) > 0 {
      controlSchema = make(map[string]interface{})
      _ = json.Unmarshal(schemaBytes, &controlSchema)
    }

    // Parse configuration/data - handle various types that might come from ApiNodeToMap
    var configBytes []byte
    switch v := handleMap["configuration"].(type) {
    case json.RawMessage:
      configBytes = v
    case []byte:
      configBytes = v
    case string:
      configBytes = []byte(v)
    }

    if len(configBytes) > 0 {
      controlData = make(map[string]interface{})
      _ = json.Unmarshal(configBytes, &controlData)
    }
    break
  }

  if !hasControlPort {
    return nil
  }

  if controlSchema == nil {
    controlSchema = make(map[string]interface{})
  }
  if controlData == nil {
    controlData = make(map[string]interface{})
  }

  // Get title from annotation or use node label
  title := node.Annotations[v1alpha1.NodeLabelAnnotation]
  if title == "" {
    title = node.Spec.Component
  }

  return &Widget{
    ID:            node.Name,
    Title:         title,
    NodeName:      node.Name,
    Port:          v1alpha1.ControlPort,
    DefaultSchema: controlSchema,
    Data:          controlData,
    // Default grid size for new widgets (position will be calculated by frontend)
    GridW: 3,
    GridH: 4,
  }
}
