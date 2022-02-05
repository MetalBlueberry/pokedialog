package pokedialog

import (
	_ "embed"
	"reflect"
	"strings"
	"testing"
)

func TestLinesAt(t *testing.T) {
	type args struct {
		lines    []string
		position int
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			args: args{
				lines:    []string{"hello world"},
				position: 3,
			},
			want: []string{"hel"},
		},
		{
			args: args{
				lines:    []string{"hello world"},
				position: 9,
			},
			want: []string{"hello wor"},
		},
		{
			args: args{
				lines:    []string{"hello", "world"},
				position: 9,
			},
			want: []string{"hello", "worl"},
		},
		{
			args: args{
				lines:    []string{"hello", "world", "again"},
				position: 12,
			},
			want: []string{"world", "ag"},
		},
		{
			args: args{
				lines:    []string{"hello", "world", "again"},
				position: 10,
			},
			want: []string{"hello", "world"},
		},
		{
			args: args{
				lines:    []string{"hello", "world", "again"},
				position: 15,
			},
			want: []string{"world", "again"},
		},
		{
			args: args{
				lines:    []string{"hello", "world", "again"},
				position: 20,
			},
			want: []string{"world", "again"},
		},
		{
			args: args{
				lines:    []string{"Test long text", "with other"},
				position: 25,
			},
			want: []string{"Test long text", "with other"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := LinesAt(tt.args.lines, tt.args.position); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FrameAt() = [%v], want [%v]", strings.Join(got, ","), strings.Join(tt.want, ","))
			}
		})
	}
}
