package server

import (
	"reflect"
	"testing"

	clienterrorspb "github.com/requiemofthesouls/client-errors/pb"
)

func Test_newHTTPErrorDetails(t *testing.T) {
	type args struct {
		pbErrorDetails *clienterrorspb.ErrorDetails
	}
	tests := []struct {
		name string
		args args
		want *httpErrorDetails
	}{
		{
			name: "without nothing",
			args: args{
				pbErrorDetails: nil,
			},
			want: nil,
		},
		{
			name: "without validation errors",
			args: args{
				pbErrorDetails: &clienterrorspb.ErrorDetails{
					Msg: "error message",
				},
			},
			want: &httpErrorDetails{
				Msg: "error message",
			},
		},
		{
			name: "with empty validation errors items",
			args: args{
				pbErrorDetails: &clienterrorspb.ErrorDetails{
					Msg:              "error message",
					ValidationErrors: &clienterrorspb.ErrorDetails_ValidationErrors{},
				},
			},
			want: &httpErrorDetails{
				Msg:              "error message",
				ValidationErrors: &validationErrors{Items: []*validationErrorItem{}},
			},
		},
		{
			name: "full filled",
			args: args{
				pbErrorDetails: &clienterrorspb.ErrorDetails{
					Msg: "error message",
					ValidationErrors: &clienterrorspb.ErrorDetails_ValidationErrors{
						Items: []*clienterrorspb.ErrorDetails_ValidationErrorItem{
							{
								Field: "field_1",
								Msg:   "message_1",
							},
							{
								Field: "field_2",
								Msg:   "message_2",
							},
							{
								Field: "field_3",
								Msg:   "message_3",
							},
						},
					},
				},
			},
			want: &httpErrorDetails{
				Msg: "error message",
				ValidationErrors: &validationErrors{
					Items: []*validationErrorItem{
						{
							Field: "field_1",
							Msg:   "message_1",
						},
						{
							Field: "field_2",
							Msg:   "message_2",
						},
						{
							Field: "field_3",
							Msg:   "message_3",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newHTTPErrorDetails(tt.args.pbErrorDetails); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newHTTPErrorDetails() = %v, want %v", got, tt.want)
			}
		})
	}
}
