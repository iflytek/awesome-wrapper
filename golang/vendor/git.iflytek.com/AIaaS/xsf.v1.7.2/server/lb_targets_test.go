package xsf

import "testing"

func Test_checkLbTargets(t *testing.T) {
	type args struct {
		lbTargets string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"test1", args{`AIPaaS,hu,lb-niche,1.0.0,sub,subsvc`}, true},
		{"test2", args{`pro1,gro1,svc1,api1,sub1,ent1;pro2,gro2,svc2,api2,sub2,ent2`}, true},
		{"test3", args{`pro1,gro1,svc1,api1,sub1,ent1|ent2;pro2,gro2,svc2,api2,sub2,ent2|ent3`}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := checkLbTargets(tt.args.lbTargets); got != tt.want {
				t.Errorf("checkLbTargets() = %v, want %v", got, tt.want)
			}
		})
	}
}
