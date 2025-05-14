package depman

import (
	"testing"
)

func TestCheckVersionUpdate(t *testing.T) {
	testCases := []struct {
		name            string
		currentVersion  string
		requiredVersion string
		expectedUpdate  UpdateType
		expectError     bool
	}{
		{
			name:            "No update needed",
			currentVersion:  "1.2.3",
			requiredVersion: "1.2.3",
			expectedUpdate:  NoUpdate,
			expectError:     false,
		},
		{
			name:            "Patch update needed",
			currentVersion:  "1.2.3",
			requiredVersion: "1.2.4",
			expectedUpdate:  PatchUpdate,
			expectError:     false,
		},
		{
			name:            "Minor update needed",
			currentVersion:  "1.2.3",
			requiredVersion: "1.3.0",
			expectedUpdate:  MinorUpdate,
			expectError:     false,
		},
		{
			name:            "Major update needed",
			currentVersion:  "1.2.3",
			requiredVersion: "2.0.0",
			expectedUpdate:  MajorUpdate,
			expectError:     false,
		},
		{
			name:            "Current version newer",
			currentVersion:  "2.0.0",
			requiredVersion: "1.0.0",
			expectedUpdate:  NoUpdate,
			expectError:     false,
		},
		{
			name:            "Invalid current version",
			currentVersion:  "not-a-version",
			requiredVersion: "1.0.0",
			expectedUpdate:  NoUpdate,
			expectError:     true,
		},
		{
			name:            "Invalid required version",
			currentVersion:  "1.0.0",
			requiredVersion: "not-a-version",
			expectedUpdate:  NoUpdate,
			expectError:     true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			updateType, err := CheckVersionUpdate(tc.currentVersion, tc.requiredVersion)

			// Check error expectation
			if tc.expectError && err == nil {
				t.Errorf("Expected an error but got none")
			}
			if !tc.expectError && err != nil {
				t.Errorf("Did not expect an error but got: %v", err)
			}

			// If we don't expect an error, check the update type
			if !tc.expectError {
				if updateType != tc.expectedUpdate {
					t.Errorf("Expected update type %s but got %s", tc.expectedUpdate, updateType)
				}
			}
		})
	}
}

func TestIsVersionCompatible(t *testing.T) {
	testCases := []struct {
		name           string
		currentVersion string
		constraint     string
		expected       bool
		expectError    bool
	}{
		{
			name:           "Compatible exact version",
			currentVersion: "1.2.3",
			constraint:     "=1.2.3",
			expected:       true,
			expectError:    false,
		},
		{
			name:           "Compatible with caret constraint",
			currentVersion: "1.2.5",
			constraint:     "^1.2.0",
			expected:       true,
			expectError:    false,
		},
		{
			name:           "Compatible with tilde constraint",
			currentVersion: "1.2.5",
			constraint:     "~1.2.0",
			expected:       true,
			expectError:    false,
		},
		{
			name:           "Compatible with range",
			currentVersion: "1.2.3",
			constraint:     ">1.0.0 <2.0.0",
			expected:       true,
			expectError:    false,
		},
		{
			name:           "Incompatible version",
			currentVersion: "2.0.0",
			constraint:     "^1.0.0",
			expected:       false,
			expectError:    false,
		},
		{
			name:           "Invalid current version",
			currentVersion: "not-a-version",
			constraint:     "^1.0.0",
			expected:       false,
			expectError:    true,
		},
		{
			name:           "Invalid constraint",
			currentVersion: "1.0.0",
			constraint:     "invalid-constraint",
			expected:       false,
			expectError:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			compatible, err := IsVersionCompatible(tc.currentVersion, tc.constraint)

			// Check error expectation
			if tc.expectError && err == nil {
				t.Errorf("Expected an error but got none")
			}
			if !tc.expectError && err != nil {
				t.Errorf("Did not expect an error but got: %v", err)
			}

			// If we don't expect an error, check compatibility
			if !tc.expectError {
				if compatible != tc.expected {
					t.Errorf("Expected compatibility %v but got %v", tc.expected, compatible)
				}
			}
		})
	}
}
