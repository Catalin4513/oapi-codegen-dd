package types

import (
	"testing"
)

func TestStatusCode_Validate(t *testing.T) {
	tests := []struct {
		name    string
		value   StatusCode
		wantErr bool
	}{
		{
			name:    "valid 200",
			value:   StatusCodeN200,
			wantErr: false,
		},
		{
			name:    "valid 404",
			value:   StatusCodeN404,
			wantErr: false,
		},
		{
			name:    "valid 500",
			value:   StatusCodeN500,
			wantErr: false,
		},
		{
			name:    "invalid 201",
			value:   StatusCode(201),
			wantErr: true,
		},
		{
			name:    "invalid 0",
			value:   StatusCode(0),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.value.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("StatusCode.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPriority_Validate(t *testing.T) {
	tests := []struct {
		name    string
		value   Priority
		wantErr bool
	}{
		{
			name:    "valid 1.0",
			value:   PriorityN10,
			wantErr: false,
		},
		{
			name:    "valid 2.5",
			value:   PriorityN25,
			wantErr: false,
		},
		{
			name:    "valid 5.0",
			value:   PriorityN50,
			wantErr: false,
		},
		{
			name:    "invalid 3.0",
			value:   Priority(3.0),
			wantErr: true,
		},
		{
			name:    "invalid 0.0",
			value:   Priority(0.0),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.value.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Priority.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestColor_Validate(t *testing.T) {
	tests := []struct {
		name    string
		value   Color
		wantErr bool
	}{
		{
			name:    "valid red",
			value:   ColorRed,
			wantErr: false,
		},
		{
			name:    "valid green",
			value:   ColorGreen,
			wantErr: false,
		},
		{
			name:    "valid blue",
			value:   ColorBlue,
			wantErr: false,
		},
		{
			name:    "invalid yellow",
			value:   Color("yellow"),
			wantErr: true,
		},
		{
			name:    "empty value",
			value:   Color(""),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.value.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Color.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
