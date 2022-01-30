package pokedialog

import (
	_ "embed"
	"reflect"
	"testing"
	"time"
)

func TestFrameAt(t *testing.T) {
	type args struct {
		lines          []string
		charsPerSecond float64
		duration       time.Duration
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			args: args{
				lines:          []string{"hello world"},
				charsPerSecond: 3,
				duration:       time.Second,
			},
			want: []string{"hel"},
		},
		{
			args: args{
				lines:          []string{"hello world"},
				charsPerSecond: 3,
				duration:       time.Second * 3,
			},
			want: []string{"hello wor"},
		},
		{
			args: args{
				lines:          []string{"hello", "world"},
				charsPerSecond: 3,
				duration:       time.Second * 3,
			},
			want: []string{"hello", "worl"},
		},
		{
			args: args{
				lines:          []string{"hello", "world", "again"},
				charsPerSecond: 3,
				duration:       time.Second * 4,
			},
			want: []string{"world", "ag"},
		},
		{
			args: args{
				lines:          []string{"hello", "world", "again"},
				charsPerSecond: 5,
				duration:       time.Second * 2,
			},
			want: []string{"hello", "world"},
		},
		{
			args: args{
				lines:          []string{"hello", "world", "again"},
				charsPerSecond: 5,
				duration:       time.Second * 3,
			},
			want: []string{"world", "again"},
		},
		{
			args: args{
				lines:          []string{"hello", "world", "again"},
				charsPerSecond: 5,
				duration:       time.Second * 4,
			},
			want: []string{"world", "again"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FrameAt(tt.args.lines, tt.args.charsPerSecond, tt.args.duration); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FrameAt() = %v, want %v", got, tt.want)
			}
		})
	}
}
