package main

import (
	"fmt"
	"github.com/tiny-systems/module/api/v1alpha1"
	"github.com/tiny-systems/module/pkg/resource"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/duration"
	"runtime/debug"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

type Project struct {
	Name        string `json:"name"`
	Title       string `json:"title"`
	Description string `json:"description"`
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
			// Log the panic value and the full stack trace
			a.logger.Error(nil, "A panic occurred in GetProjects!",
				"panic", r,
				"stacktrace", string(debug.Stack()))

			// You can optionally re-panic if you want the app to crash
			// or just return an error to the frontend if you prefer recovery.
			// For now, let's just log it and let the Wails binding layer handle the return error.
		}
	}()

	a.logger.Info(fmt.Sprintf("getting projects for context: %s; namespace: %s", contextName, namespace))
	var projectsApi []Project

	config, err := loadContextConfig(contextName)
	if err != nil {
		return nil, err
	}

	scheme := runtime.NewScheme()
	_ = v1alpha1.AddToScheme(scheme)

	kubeClient, err := client.NewWithWatch(config, client.Options{Scheme: scheme})
	if err != nil {
		return nil, fmt.Errorf("unable to create client %s", err)
	}

	//NewWailsLogr
	projects, err := resource.NewManager(kubeClient, a.logger, namespace).GetProjectList(a.ctx)
	if err != nil {
		return nil, err
	}
	a.logger.Info(fmt.Sprintf("%d projects found", len(projects)))

	for _, project := range projects {

		projectsApi = append(projectsApi, Project{
			Name:        project.Name,
			Title:       project.Annotations[v1alpha1.ProjectNameAnnotation],
			Description: fmt.Sprintf("Created %s ago", duration.ShortHumanDuration(time.Now().Sub(project.CreationTimestamp.Time))),
		})

	}

	return projectsApi, nil
}

func (a *App) GetProject(contextName string, namespace string, projectName string) (*Project, error) {

	go func() {
		wailsruntime.EventsEmit(a.ctx, "projectStream", UpdateEvent{
			Type: "status",
			Payload: StatusPayload{
				Message: fmt.Sprintf("Starting analysis for '%s'...", projectName),
			},
		})
		time.Sleep(1 * time.Second)

		// 1. Send a progress update
		wailsruntime.EventsEmit(a.ctx, "projectStream", UpdateEvent{
			Type: "progress",
			Payload: ProgressPayload{
				Percentage: 45,
				File:       "config.yaml",
			},
		})
		time.Sleep(2 * time.Second)

		// 2. Send an error
		wailsruntime.EventsEmit(a.ctx, "projectStream", UpdateEvent{
			Type: "error",
			Payload: ErrorPayload{
				Code:    500,
				Details: "Database connection failed. Retrying...",
			},
		})
		time.Sleep(3 * time.Second)

		// 3. Send a final status
		wailsruntime.EventsEmit(a.ctx, "projectStream", UpdateEvent{
			Type: "status",
			Payload: StatusPayload{
				Message: "Project analysis completed successfully!",
			},
		})
	}()

	return &Project{}, nil
}
