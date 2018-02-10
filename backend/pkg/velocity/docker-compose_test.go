package velocity

import (
	"reflect"
	"testing"
)

func Test_getServiceOrder(t *testing.T) {
	type args struct {
		services     map[string]dockerComposeService
		serviceOrder []string
	}

	services := map[string]dockerComposeService{
		"proxy": {
			Links: []string{
				"backend",
				"frontend",
			},
		},
		"database": {
			Links: []string{
				"redis",
			},
		},
		"redis": {
			Links: []string{
				"frontend",
			},
		},
		"frontend": {
			Links: []string{},
		},
		"backend": {
			Links: []string{"database"},
		},
	}

	services2 := map[string]dockerComposeService{
		"database": {
			Links: []string{
				"redis",
			},
		},
		"proxy": {
			Links: []string{
				"backend",
				"frontend",
			},
		},
		"redis": {
			Links: []string{
				"frontend",
			},
		},
		"frontend": {
			Links: []string{},
		},
		"backend": {
			Links: []string{"database"},
		},
	}

	services3 := map[string]dockerComposeService{
		"redis": {
			Links: []string{
				"frontend",
			},
		},
		"database": {
			Links: []string{
				"redis",
			},
		},
		"frontend": {
			Links: []string{},
		},
		"backend": {
			Links: []string{"database"},
		},
		"proxy": {
			Links: []string{
				"backend",
				"frontend",
			},
		},
	}

	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "1",
			args: args{
				services:     services,
				serviceOrder: []string{},
			},

			want: []string{"frontend", "redis", "database", "backend", "proxy"},
		},
		{
			name: "2",
			args: args{
				services:     services2,
				serviceOrder: []string{},
			},

			want: []string{"frontend", "redis", "database", "backend", "proxy"},
		},
		{
			name: "3",
			args: args{
				services:     services3,
				serviceOrder: []string{},
			},

			want: []string{"frontend", "redis", "database", "backend", "proxy"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getServiceOrder(tt.args.services, tt.args.serviceOrder); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getServiceOrder() = %v, want %v", got, tt.want)
			}
		})
	}
}
