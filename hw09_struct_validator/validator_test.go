package hw09structvalidator

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

type UserRole string

// Test the function on different structures and other types.
type (
	User struct {
		ID     string `json:"id" validate:"len:36"`
		Name   string
		Age    int             `validate:"min:18|max:50"`
		Email  string          `validate:"regexp:^\\w+@\\w+\\.\\w+$"`
		Role   UserRole        `validate:"in:admin,stuff"`
		Phones []string        `validate:"len:11"`
		meta   json.RawMessage //nolint:unused
	}

	App struct {
		Version string `validate:"len:5"`
	}

	Token struct {
		Header    []byte
		Payload   []byte
		Signature []byte
	}

	Response struct {
		Code int    `validate:"in:200,404,500"`
		Body string `json:"omitempty"`
	}

	AppUsers struct {
		ID        string `json:"id" validate:"len:2"`
		Users     []User `validate:"nested"`
		NotNested User
		meta      json.RawMessage //nolint:unused
	}

	Private struct {
		private User `validate:"nested"` //nolint:unused
	}

	NestedApp struct {
		Users NestedUser `validate:"nested"`
	}

	NestedUser struct {
		User User `validate:"nested"`
	}

	ComplexCheck struct {
		Value int `validate:"min:18|max:50|in:20,30"`
	}
)

func TestValidate(t *testing.T) { //nolint:funlen
	tests := []struct {
		in          interface{}
		expectedErr error
	}{
		{
			in: User{},
			expectedErr: ValidationErrors([]ValidationError{
				{
					Field: "/ID",
					Err:   ErrCheckStringLength,
				},
				{
					Field: "/Age",
					Err:   ErrCheckInt64Min,
				},
				{
					Field: "/Email",
					Err:   ErrCheckStringRegexp,
				},
				{
					Field: "/Role",
					Err:   ErrCheckStringEnum,
				},
			}),
		},
		{
			in: App{},
			expectedErr: ValidationErrors{
				{
					Field: "/Version",
					Err:   ErrCheckStringLength,
				},
			},
		},
		{
			in: AppUsers{},
			expectedErr: ValidationErrors{
				{
					Field: "/ID",
					Err:   ErrCheckStringLength,
				},
			},
		},
		{
			in: AppUsers{
				ID: "12",
				Users: []User{
					{}, {
						Age:    20,
						Email:  "iv@test.com",
						Role:   "stuff",
						Phones: []string{"123-456-789"},
					},
				},
			},
			expectedErr: ValidationErrors{
				{
					Field: "/Users[0]/ID",
					Err:   ErrCheckStringLength,
				},
				{
					Field: "/Users[0]/Age",
					Err:   ErrCheckInt64Min,
				},
				{
					Field: "/Users[0]/Email",
					Err:   ErrCheckStringRegexp,
				},
				{
					Field: "/Users[0]/Role",
					Err:   ErrCheckStringEnum,
				},
				{
					Field: "/Users[1]/ID",
					Err:   ErrCheckStringLength,
				},
			},
		},

		{
			in: NestedApp{
				NestedUser{
					User{
						ID:     "Иванов Иван Иванович  - 1234567890世界",
						Age:    20,
						Role:   "stuff",
						Phones: []string{"123-456-789"},
					},
				},
			},
			expectedErr: ValidationErrors{
				{
					Field: "/Users/User/Email",
					Err:   ErrCheckStringRegexp,
				},
			},
		},
		{
			in: ComplexCheck{},
			expectedErr: ValidationErrors{
				{
					Field: "/Value",
					Err: ErrCheckErrorList{
						ErrCheckInt64Min, ErrCheckInt64Enum,
					},
				},
			},
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			tt := tt
			t.Parallel()

			err := Validate(tt.in)
			require.Error(t, err)

			var expectedValidatoinErr ValidationErrors
			require.True(t, errors.As(tt.expectedErr, &expectedValidatoinErr))

			var actualValidatoinErr ValidationErrors
			require.True(t, errors.As(err, &actualValidatoinErr))

			require.Equal(t, len(expectedValidatoinErr), len(actualValidatoinErr))

			for i, expectedErr := range expectedValidatoinErr {
				actualErr := actualValidatoinErr[i]
				require.Equal(t, expectedErr.Field, actualErr.Field)

				var expectedCheckErrList ErrCheckErrorList

				if errors.As(expectedErr.Err, &expectedCheckErrList) {
					var actualCheckErrList ErrCheckErrorList
					require.True(t, errors.As(actualErr.Err, &actualCheckErrList))

					for j, expectedChecErr := range expectedCheckErrList {
						actualErr := actualCheckErrList[j]
						require.ErrorIs(t, actualErr, expectedChecErr)
					}
				} else {
					require.ErrorIs(t, actualErr.Err, expectedErr.Err)
				}
			}
		})
	}
}

func TestValidateNoError(t *testing.T) {
	tests := []struct {
		in interface{}
	}{
		{
			in: User{
				ID:     "Иванов Иван Иванович  - 1234567890世界",
				Age:    20,
				Email:  "iv@test.com",
				Role:   "stuff",
				Phones: []string{"123-456-789"},
			},
		},
		{
			in: Token{},
		},
		{
			in: Token{
				Header: []byte{0x01, 0x02, 0x03},
			},
		},
		{
			in: Response{
				Code: 404,
			},
		},
		{
			in: Private{},
		},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			tt := tt
			t.Parallel()

			err := Validate(tt.in)
			require.NoError(t, err)
		})
	}
}
