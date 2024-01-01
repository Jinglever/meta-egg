package helper

import "testing"

func TestGetEnvName(t *testing.T) {
	type args struct {
		projName string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "TestGetEnvName",
			args: args{
				projName: "test_project",
			},
			want: "TP",
		},
		{
			name: "TestGetEnvName",
			args: args{
				projName: "test",
			},
			want: "TEST",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetEnvPrefix(tt.args.projName); got != tt.want {
				t.Errorf("GetEnvName() = %v, want %v", got, tt.want)
			}
		})
	}
}
