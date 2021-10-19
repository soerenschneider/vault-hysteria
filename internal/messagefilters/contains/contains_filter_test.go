package contains

import (
	"reflect"
	"testing"
)

func TestNewContainsFilterFromMap(t *testing.T) {
	type args struct {
		args map[string]interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    *ContainsFilter
		wantErr bool
	}{
		{
			name: "empty map",
			args: args{
				map[string]interface{}{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "all ok",
			args: args{
				map[string]interface{}{
					configKeywordContains:   "substring",
					configKeywordIgnoreCase: true,
				},
			},
			want: &ContainsFilter{
				contains:   "substring",
				ignoreCase: true,
			},
			wantErr: false,
		},
		{
			name: "missing",
			args: args{
				map[string]interface{}{
					configKeywordContains: "substring",
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "wrong type",
			args: args{
				map[string]interface{}{
					configKeywordContains:   "substring",
					configKeywordIgnoreCase: "true",
				},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewContainsFilterFromMap(tt.args.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewContainsFilterFromMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewContainsFilterFromMap() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContainsFilter_Accept(t *testing.T) {
	type fields struct {
		contains   string
		ignoreCase bool
	}
	type args struct {
		msg string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "simple",
			fields: fields{
				contains:   "substring",
				ignoreCase: false,
			},
			args: args{"substringsubstringsubstring"},
			want: true,
		},
		{
			name: "no substring",
			fields: fields{
				contains:   "substring",
				ignoreCase: false,
			},
			args: args{"lorem ipsum"},
			want: false,
		},
		{
			name: "case sensitive",
			fields: fields{
				contains:   "substring",
				ignoreCase: false,
			},
			args: args{"SUBSTRING"},
			want: false,
		},
		{
			name: "case insensitive",
			fields: fields{
				contains:   "substring",
				ignoreCase: true,
			},
			args: args{"SUBSTRING"},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := NewContainsFilter(tt.fields.contains, tt.fields.ignoreCase)
			if got := c.Accept(tt.args.msg); got != tt.want {
				t.Errorf("Accept() = %v, want %v", got, tt.want)
			}
		})
	}
}
