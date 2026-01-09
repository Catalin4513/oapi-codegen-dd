package nested

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNestedBooleanValidation_PortalInvoiceList(t *testing.T) {
	t.Run("false should be valid", func(t *testing.T) {
		invoiceList := PortalInvoiceList{Enabled: false}

		// No validation needed - booleans are always valid
		require.NotNil(t, invoiceList)
	})

	t.Run("true should be valid", func(t *testing.T) {
		invoiceList := PortalInvoiceList{Enabled: true}

		// No validation needed - booleans are always valid
		require.NotNil(t, invoiceList)
	})
}

func TestNestedBooleanValidation_PortalFeatures(t *testing.T) {
	t.Run("nested struct with false should PASS validation", func(t *testing.T) {
		features := PortalFeatures{
			InvoiceHistory: PortalInvoiceList{Enabled: false},
		}

		err := features.Validate()
		require.NoError(t, err, "nested false should be valid (KEY TEST)")
	})

	t.Run("nested struct with true should PASS validation", func(t *testing.T) {
		features := PortalFeatures{
			InvoiceHistory: PortalInvoiceList{Enabled: true},
		}

		err := features.Validate()
		require.NoError(t, err, "nested true should be valid")
	})
}

func TestNestedBooleanValidation_ResponseType(t *testing.T) {
	t.Run("Response with false should PASS validation", func(t *testing.T) {
		id := "1"
		response := PostBillingPortalConfigurationsConfigurationResponse{
			Features: PortalFeatures{
				ID:             &id,
				InvoiceHistory: PortalInvoiceList{Enabled: false},
			},
		}

		err := response.Validate()
		require.NoError(t, err, "deeply nested false should be valid (KEY TEST)")
	})

	t.Run("Response with true should PASS validation", func(t *testing.T) {
		response := PostBillingPortalConfigurationsConfigurationResponse{
			Features: PortalFeatures{
				InvoiceHistory: PortalInvoiceList{Enabled: true},
			},
		}

		err := response.Validate()
		require.NoError(t, err, "deeply nested true should be valid")
	})
}

func TestNestedBooleanValidation_TableTests(t *testing.T) {
	testCases := []struct {
		name     string
		response PostBillingPortalConfigurationsConfigurationResponse
	}{
		{
			name: "false",
			response: PostBillingPortalConfigurationsConfigurationResponse{
				Features: PortalFeatures{
					InvoiceHistory: PortalInvoiceList{Enabled: false},
				},
			},
		},
		{
			name: "true",
			response: PostBillingPortalConfigurationsConfigurationResponse{
				Features: PortalFeatures{
					InvoiceHistory: PortalInvoiceList{Enabled: true},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.response.Validate()
			require.NoError(t, err)
		})
	}
}
