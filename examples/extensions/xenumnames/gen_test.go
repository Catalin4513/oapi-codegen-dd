package xenumnames

import (
	"testing"
)

func TestClientType_Validate(t *testing.T) {
	tests := []struct {
		name    string
		value   ClientType
		wantErr bool
	}{
		{
			name:    "valid ACT",
			value:   ACT,
			wantErr: false,
		},
		{
			name:    "valid EXP",
			value:   EXP,
			wantErr: false,
		},
		{
			name:    "invalid value",
			value:   ClientType("INVALID"),
			wantErr: true,
		},
		{
			name:    "empty value",
			value:   ClientType(""),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.value.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("ClientType.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClientTypeWithNamesExtension_Validate(t *testing.T) {
	tests := []struct {
		name    string
		value   ClientTypeWithNamesExtension
		wantErr bool
	}{
		{
			name:    "valid Active",
			value:   Active,
			wantErr: false,
		},
		{
			name:    "valid Expired",
			value:   Expired,
			wantErr: false,
		},
		{
			name:    "invalid value",
			value:   ClientTypeWithNamesExtension("INVALID"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.value.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("ClientTypeWithNamesExtension.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
