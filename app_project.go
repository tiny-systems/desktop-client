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
	jsonpatch "gomodules.xyz/jsonpatch/v2"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/duration"
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
			Description: fmt.Sprintf("Created %s ago", duration.ShortHumanDuration(time.Since(project.CreationTimestamp.Time))),
		})

	}

	return projectsApi, nil
}

func (a *App) GetProject(contextName string, namespace string, projectName string) (*Project, error) {
	return &Project{}, nil
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

	scheme := runtime.NewScheme()
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("failed to add scheme: %w", err)
	}

	kubeClient, err := client.NewWithWatch(config, client.Options{Scheme: scheme})
	if err != nil {
		return nil, fmt.Errorf("unable to create client: %w", err)
	}

	return resource.NewManagerFromClient(kubeClient, namespace)
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

// GetFlowGraph fetches the graph data for a flow (for preview)
func (a *App) GetFlowGraph(contextName string, namespace string, projectName string, flowResourceName string) (map[string]interface{}, error) {
	mgr, err := a.getManager(contextName, namespace)
	if err != nil {
		return nil, err
	}

	nodes, err := mgr.GetProjectFlowNodes(a.ctx, projectName, flowResourceName)
	if err != nil {
		return nil, fmt.Errorf("unable to get flow nodes: %w", err)
	}

	// Build graph structure for Vue Flow
	graphNodes := make([]map[string]interface{}, 0)
	graphEdges := make([]map[string]interface{}, 0)

	for _, node := range nodes {
		// Parse position from annotations
		posX, _ := strconv.ParseFloat(node.Annotations[v1alpha1.ComponentPosXAnnotation], 64)
		posY, _ := strconv.ParseFloat(node.Annotations[v1alpha1.ComponentPosYAnnotation], 64)

		// Build handles from ports
		handles := make([]map[string]interface{}, 0)
		for _, port := range node.Status.Ports {
			if port.Name == v1alpha1.ControlPort || port.Name == v1alpha1.SettingsPort {
				continue
			}
			handleType := "target"
			if port.Source {
				handleType = "source"
			}
			handles = append(handles, map[string]interface{}{
				"id":               port.Name,
				"type":             handleType,
				"label":            port.Label,
				"position":         port.Position,
				"rotated_position": port.Position,
			})
		}

		// Get label from annotation or use component name
		label := node.Annotations[v1alpha1.NodeLabelAnnotation]
		if label == "" {
			label = node.Spec.Component
		}

		graphNodes = append(graphNodes, map[string]interface{}{
			"id":   node.Name,
			"type": "tinyNode",
			"position": map[string]interface{}{
				"x": posX,
				"y": posY,
			},
			"data": map[string]interface{}{
				"label":     label,
				"handles":   handles,
				"component": node.Spec.Component,
				"status":    node.Status.Status,
				"error":     node.Status.Error,
				"dashboard": node.Labels[v1alpha1.DashboardLabel],
			},
		})

		// Build edges from node spec
		for _, edge := range node.Spec.Edges {
			// edge.To is in format "node-name:port-name"
			toParts := strings.Split(edge.To, ":")
			targetNode := edge.To
			targetHandle := ""
			if len(toParts) == 2 {
				targetNode = toParts[0]
				targetHandle = toParts[1]
			}

			graphEdges = append(graphEdges, map[string]interface{}{
				"id":           edge.ID,
				"source":       node.Name,
				"sourceHandle": edge.Port,
				"target":       targetNode,
				"targetHandle": targetHandle,
				"type":         "tinyEdge",
			})
		}
	}

	return map[string]interface{}{
		"nodes": graphNodes,
		"edges": graphEdges,
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
	mgr, err := a.getManager(contextName, namespace)
	if err != nil {
		return err
	}

	if err := mgr.DeleteProject(a.ctx, projectName); err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	return nil
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
