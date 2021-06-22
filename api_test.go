package main

import (
	"io"
	"testing"
)

func Test_httpService_makeRequest(t *testing.T) {
	
	type args struct {
		point  string
		method string
		body   io.ReadCloser
	}
	type test struct {
		name    string
		args    args
		wantErr bool
	}
	
	var tests []test

	addtest := func(name string, argss args, wantErr bool){
		tests = append(tests, test{
			name: name,
			args: argss,
			wantErr: wantErr,
		})
	}

	addtest("1", args{"/service/token?domainuser=adushkin.vasiliy", "GET", nil}, false)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
		
			_, got, err := views.makeRequest(tt.args.point, tt.args.method, tt.args.body, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("httpService.makeRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("Result %+v", string(got))
		})
	}
}
