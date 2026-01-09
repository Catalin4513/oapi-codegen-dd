package gen

import (
	"testing"
)

func TestResponseValidation(t *testing.T) {
	tests := []struct {
		name    string
		resp    Response
		wantErr bool
	}{
		{
			name: "valid - all fields valid",
			resp: Response{
				Msn1:                     ptrMsn("12345"),
				Msn2:                     ptrString("anything"),
				MsnReqWithConstraints:    "valid",
				MsnReqWithoutConstraints: "anything",
				Msn3:                     ptrInt(50),
				MsnFloat:                 3.14,
				MsnBool:                  true,
				UserRequired:             User{},
			},
			wantErr: false,
		},
		{
			name: "invalid - msn1 too short",
			resp: Response{
				Msn1:                     ptrMsn("123"),
				MsnReqWithConstraints:    "valid",
				MsnReqWithoutConstraints: "anything",
				MsnFloat:                 1.0,
				MsnBool:                  false,
				UserRequired:             User{},
			},
			wantErr: true,
		},
		{
			name: "invalid - msn1 too long",
			resp: Response{
				Msn1:                     ptrMsn("12345678"),
				MsnReqWithConstraints:    "valid",
				MsnReqWithoutConstraints: "anything",
				MsnFloat:                 1.0,
				MsnBool:                  false,
				UserRequired:             User{},
			},
			wantErr: true,
		},
		{
			name: "invalid - msn3 too small",
			resp: Response{
				Msn3:                     ptrInt(0),
				MsnReqWithConstraints:    "valid",
				MsnReqWithoutConstraints: "anything",
				MsnFloat:                 1.0,
				MsnBool:                  false,
				UserRequired:             User{},
			},
			wantErr: true,
		},
		{
			name: "invalid - msn3 too large",
			resp: Response{
				Msn3:                     ptrInt(101),
				MsnReqWithConstraints:    "valid",
				MsnReqWithoutConstraints: "anything",
				MsnFloat:                 1.0,
				MsnBool:                  false,
				UserRequired:             User{},
			},
			wantErr: true,
		},
		{
			name: "invalid - missing required MsnFloat",
			resp: Response{
				// MsnFloat is required but missing (zero value)
				MsnReqWithConstraints:    "valid",
				MsnReqWithoutConstraints: "anything",
				MsnBool:                  true,
				UserRequired:             User{},
			},
			wantErr: true,
		},
		{
			name: "valid - MsnBool false is allowed (booleans can't be required)",
			resp: Response{
				MsnReqWithConstraints:    "valid",
				MsnReqWithoutConstraints: "anything",
				MsnFloat:                 1.0,
				MsnBool:                  false, // false is a valid value for booleans
				UserRequired:             User{},
			},
			wantErr: false,
		},
		{
			name: "valid - MsnFloat with zero value fails required check",
			resp: Response{
				MsnReqWithConstraints:    "valid",
				MsnReqWithoutConstraints: "anything",
				MsnFloat:                 0.0, // zero value should fail required validation
				MsnBool:                  true,
				UserRequired:             User{},
			},
			wantErr: true,
		},
		{
			name: "valid - UserRequired with empty struct",
			resp: Response{
				MsnReqWithConstraints:    "valid",
				MsnReqWithoutConstraints: "anything",
				MsnFloat:                 1.0,
				MsnBool:                  true,
				UserRequired:             User{}, // Empty user struct - should this be valid?
			},
			wantErr: false, // Currently passes because "required" only checks non-nil
		},
		{
			name: "valid - UserOptional is nil",
			resp: Response{
				MsnReqWithConstraints:    "valid",
				MsnReqWithoutConstraints: "anything",
				MsnFloat:                 1.0,
				MsnBool:                  true,
				UserRequired:             User{Name: ptrString2("John"), Age: ptrInt(30)},
				UserOptional:             nil, // Optional, so nil is fine
			},
			wantErr: false,
		},
		{
			name: "valid - UserOptional with values",
			resp: Response{
				MsnReqWithConstraints:    "valid",
				MsnReqWithoutConstraints: "anything",
				MsnFloat:                 1.0,
				MsnBool:                  true,
				UserRequired:             User{Name: ptrString2("John"), Age: ptrInt(30)},
				UserOptional:             &User{Name: ptrString2("Jane"), Age: ptrInt(25)},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.resp.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Response.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func ptrMsn(s string) *MsnWithConstraints {
	m := MsnWithConstraints(s)
	return &m
}

func ptrString(s string) *MsnWithoutConstraints {
	m := MsnWithoutConstraints(s)
	return &m
}

func ptrInt(i int) *int {
	return &i
}

func ptrString2(s string) *string {
	return &s
}

func TestPredefinedValue_Validate(t *testing.T) {
	tests := []struct {
		name    string
		pv      PredefinedValue
		wantErr bool
	}{
		{
			name: "valid - valid Predefined value",
			pv: PredefinedValue{
				Value: ptrPredefined(PredefinedA2),
				Type:  ptrString2("test"),
			},
			wantErr: false,
		},
		{
			name: "valid - nil Value",
			pv: PredefinedValue{
				Value: nil,
				Type:  ptrString2("test"),
			},
			wantErr: false,
		},
		{
			name: "invalid - invalid Predefined value",
			pv: PredefinedValue{
				Value: ptrPredefined("INVALID"),
				Type:  ptrString2("test"),
			},
			wantErr: true, // Should fail because "INVALID" is not a valid Predefined value
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.pv.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("PredefinedValue.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func ptrPredefined(p Predefined) *Predefined {
	return &p
}
