package captcha

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSolveArythmeticCaptcha(t *testing.T) {
	type args struct {
		captcha string
	}
	tests := []struct {
		name       string
		args       args
		wantAnswer string
	}{
		{"1", args{"24 + ? =  31"}, "7"},
		{"2", args{"48 + ? =  49"}, "1"},
		{"3", args{"? + 20 =  26"}, "6"},
		{"4", args{" 13 + 8 =  ??  "}, "21"},
		{"5", args{" 13 - 8 =  x  "}, "5"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.wantAnswer, SolveArithmeticCaptcha(tt.args.captcha), "SolveArythmeticCaptcha(%v)", tt.args.captcha)
		})
	}
}
