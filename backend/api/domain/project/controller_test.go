package project

import (
	"log"
	"net/http"
	"reflect"
	"testing"

	"github.com/gorilla/mux"
	"github.com/unrolled/render"
)

func TestNewController(t *testing.T) {
	type args struct {
		projectManager  *Manager
		projectResolver *Resolver
	}
	tests := []struct {
		name string
		args args
		want *Controller
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewController(tt.args.projectManager, tt.args.projectResolver); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewController() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestController_Setup(t *testing.T) {
	type fields struct {
		logger   *log.Logger
		render   *render.Render
		manager  *Manager
		resolver *Resolver
	}
	type args struct {
		router *mux.Router
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Controller{
				logger:   tt.fields.logger,
				render:   tt.fields.render,
				manager:  tt.fields.manager,
				resolver: tt.fields.resolver,
			}
			c.Setup(tt.args.router)
		})
	}
}

func TestController_getProjectHandler(t *testing.T) {
	type fields struct {
		logger   *log.Logger
		render   *render.Render
		manager  *Manager
		resolver *Resolver
	}
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Controller{
				logger:   tt.fields.logger,
				render:   tt.fields.render,
				manager:  tt.fields.manager,
				resolver: tt.fields.resolver,
			}
			c.getProjectHandler(tt.args.w, tt.args.r)
		})
	}
}

func TestController_getProjectsHandler(t *testing.T) {
	type fields struct {
		logger   *log.Logger
		render   *render.Render
		manager  *Manager
		resolver *Resolver
	}
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Controller{
				logger:   tt.fields.logger,
				render:   tt.fields.render,
				manager:  tt.fields.manager,
				resolver: tt.fields.resolver,
			}
			c.getProjectsHandler(tt.args.w, tt.args.r)
		})
	}
}

func TestController_deleteProjectHandler(t *testing.T) {
	type fields struct {
		logger   *log.Logger
		render   *render.Render
		manager  *Manager
		resolver *Resolver
	}
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Controller{
				logger:   tt.fields.logger,
				render:   tt.fields.render,
				manager:  tt.fields.manager,
				resolver: tt.fields.resolver,
			}
			c.deleteProjectHandler(tt.args.w, tt.args.r)
		})
	}
}

func TestController_postProjectsHandler(t *testing.T) {
	type fields struct {
		logger   *log.Logger
		render   *render.Render
		manager  *Manager
		resolver *Resolver
	}
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Controller{
				logger:   tt.fields.logger,
				render:   tt.fields.render,
				manager:  tt.fields.manager,
				resolver: tt.fields.resolver,
			}
			c.postProjectsHandler(tt.args.w, tt.args.r)
		})
	}
}
