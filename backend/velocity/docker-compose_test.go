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
		"proxy": dockerComposeService{
			Links: []string{
				"backend",
				"frontend",
			},
		},
		"database": dockerComposeService{
			Links: []string{
				"redis",
			},
		},
		"redis": dockerComposeService{
			Links: []string{
				"frontend",
			},
		},
		"frontend": dockerComposeService{
			Links: []string{},
		},
		"backend": dockerComposeService{
			Links: []string{"database"},
		},
	}

	services2 := map[string]dockerComposeService{
		"database": dockerComposeService{
			Links: []string{
				"redis",
			},
		},
		"proxy": dockerComposeService{
			Links: []string{
				"backend",
				"frontend",
			},
		},
		"redis": dockerComposeService{
			Links: []string{
				"frontend",
			},
		},
		"frontend": dockerComposeService{
			Links: []string{},
		},
		"backend": dockerComposeService{
			Links: []string{"database"},
		},
	}

	services3 := map[string]dockerComposeService{
		"redis": dockerComposeService{
			Links: []string{
				"frontend",
			},
		},
		"database": dockerComposeService{
			Links: []string{
				"redis",
			},
		},
		"frontend": dockerComposeService{
			Links: []string{},
		},
		"backend": dockerComposeService{
			Links: []string{"database"},
		},
		"proxy": dockerComposeService{
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
