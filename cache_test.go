package main

import (
	"testing"

)

func TestNote_EncodeBasicAuth(t *testing.T) {
	type fields struct {
		UserName  string
		BasicAuth string

	}
	tests := []struct {
		name         string
		fields       fields
		wantLogin    string
		wantPassword string
	}{
		{
			name: "kananov.yan",
			fields: fields{UserName: "kananov.yan",BasicAuth:"0JrQsNC90LDQvdC+0LIg0K/QvToxMzUyNDYNCg==", },
			wantLogin:"Кананов Ян",
			wantPassword:"135246",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &Note{
				UserName:  tt.fields.UserName,
				BasicAuth: tt.fields.BasicAuth,
				
			}
			gotLogin, gotPassword := n.EncodeBasicAuth()
			if gotLogin != tt.wantLogin {
				t.Errorf("Note.EncodeBasicAuth() gotLogin = %v, want %v", gotLogin, tt.wantLogin)
			}
			if gotPassword != tt.wantPassword {
				t.Errorf("Note.EncodeBasicAuth() gotPassword = %v, want %v", gotPassword, tt.wantPassword)
			}
		})
	}
}
