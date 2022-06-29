package controller

import "testing"

func Test_convertMetricsKind(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test-1",
			args: args{name: "s1-QPM"},
			want: "qpm",
		},
		{
			name: "test-2",
			args: args{name: "s1-DUBBO-QPM"},
			want: "dubbo-qpm",
		},
		{
			name: "test-2",
			args: args{name: "s11111-DUBBO-QPM"},
			want: "dubbo-qpm",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convertMetricsKind(tt.args.name); got != tt.want {
				t.Errorf("convertMetricsKind() = %v, want %v", got, tt.want)
			}
		})
	}
}
