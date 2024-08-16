package utils

import (
	"testing"
)

func Test_addExtraTag(t *testing.T) {
	type args struct {
		base string
		kvs  map[string]string
	}
	tests := []struct {
		name string
		args args
	}{
		{name: "test1", args: args{base: "", kvs: map[string]string{"k": "v"}}},
		{name: "test2", args: args{base: `{"k":"v"}`, kvs: map[string]string{"k": "v"}}},
		{name: "test3", args: args{base: `{"k1":"v1"}`, kvs: map[string]string{"k": "v"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotExtra, err := addExtraTag(tt.args.base, tt.args.kvs)
			t.Logf("gotExtra:%v,err:%v", gotExtra, err)
		})
	}
}

func Test_extractExtraTag(t *testing.T) {
	gotExtra, err := addExtraTag(`{"k1":"v1"}`, map[string]string{"k2": "v2"})
	if err != nil {
		t.Fatal(err)
	}

	type args struct {
		base string
		tag  string
	}
	tests := []struct {
		name    string
		args    args
		wantVal string
		wantErr bool
	}{
		{name: "test1", args: args{base: gotExtra, tag: "k1"}, wantVal: "v1", wantErr: false},
		{name: "test2", args: args{base: gotExtra, tag: "k"}, wantVal: "", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotVal, err := extractExtraTag(tt.args.base, tt.args.tag)
			t.Logf("gotVal:%v,err:%v", gotVal, err)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractExtraTag() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotVal != tt.wantVal {
				t.Errorf("extractExtraTag() gotVal = %v, want %v", gotVal, tt.wantVal)
			}
		})
	}
}
