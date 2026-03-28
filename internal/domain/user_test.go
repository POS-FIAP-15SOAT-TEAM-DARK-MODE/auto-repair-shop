package domain

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUser_Validate(t *testing.T) {
	tests := []struct {
		name    string
		user    *User
		wantErr bool
	}{
		{"Valid", NewUser("John Doe", "john@example.com", "Secret@123"), false},
		{"EmptyName", NewUser("", "john@example.com", "Secret@123"), true},
		{"ShortName", NewUser("Jo", "john@example.com", "Secret@123"), true},
		{"InvalidName", NewUser("John123", "john@example.com", "Secret@123"), true},
		{"EmptyEmail", NewUser("John Doe", "", "Secret@123"), true},
		{"InvalidEmail_NoAt", NewUser("John Doe", "johnexample.com", "Secret@123"), true},
		{"InvalidEmail_Space", NewUser("John Doe", "john @example.com", "Secret@123"), true},
		{"InvalidEmail_NoDot", NewUser("John Doe", "john@example", "Secret@123"), true},
		{"EmptyPassword", NewUser("John Doe", "john@example.com", ""), true},
		{"ShortPassword", NewUser("John Doe", "john@example.com", "S@1"), true},
		{"LongPassword", NewUser("John Doe", "john@example.com", string(make([]byte, 100))), true},
		{"WeakPassword_NoUpper", NewUser("John Doe", "john@example.com", "secret@123"), true},
		{"WeakPassword_NoSpecial", NewUser("John Doe", "john@example.com", "Secret123"), true},
		{"InvalidEmail_DotPrefix", NewUser("John Doe", "john@.example.com", "Secret@123"), true},
		{"InvalidEmail_DotSuffix", NewUser("John Doe", "john@example.com.", "Secret@123"), true},
		{"InvalidEmail_EmptyLocal", NewUser("John Doe", "@example.com", "Secret@123"), true},
		{"InvalidEmail_EmptyDomain", NewUser("John Doe", "john@", "Secret@123"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.user.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUser_HashPassword(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		u := NewUser("John Doe", "john@example.com", "Secret@123")
		originalPass := u.Password

		err := u.HashPassword(context.Background())
		assert.NoError(t, err)
		assert.NotEqual(t, originalPass, u.Password)
		assert.True(t, len(u.Password) > 20)
	})

	t.Run("PasswordTooLong", func(t *testing.T) {
		u := NewUser("John Doe", "john@example.com", string(make([]byte, 73)))
		err := u.HashPassword(context.Background())
		assert.Error(t, err)
		assert.Equal(t, ErrUserPasswordTooLong, err)
	})
}

func TestCreateUserToDomain(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		u, err := CreateUserToDomain("John Doe", "john@example.com", "Secret@1223")
		assert.NoError(t, err)
		assert.NotNil(t, u)
	})

	t.Run("Invalid", func(t *testing.T) {
		u, err := CreateUserToDomain("", "", "")
		assert.Error(t, err)
		assert.Nil(t, u)
	})
}

func TestUser_HashPassword_GenericError(t *testing.T) {
	// Temporarily change hashingCost to an invalid value to trigger a bcrypt error
	oldCost := hashingCost
	defer func() { hashingCost = oldCost }()
	hashingCost = 32 // bcrypt.MaxCost is 31

	u := NewUser("John Doe", "john@example.com", "Secret@123")
	err := u.HashPassword(context.Background())
	assert.Error(t, err)
}

func TestCustomer_CNPJ_CheckDigitCalculation_RemainderLessThan2(t *testing.T) {
	// First check digit returns 0 if remainder < 2.
	// sum % 11 == 0
	assert.Error(t, validateCNPJ("00000000110100"))
	// sum % 11 == 1
	// weights1 := []int{5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2}
	// weights1[10]=3. 4*3 = 12. 12 % 11 = 1.
	assert.Error(t, validateCNPJ("00000000004000"))
}

func TestValidateCPF_Direct(t *testing.T) {
	t.Run("InvalidLength", func(t *testing.T) {
		assert.Equal(t, ErrCPFLength, validateCPF("123"))
	})
	t.Run("AllSame", func(t *testing.T) {
		assert.Equal(t, ErrInvalidCPF, validateCPF("11111111111"))
	})
	t.Run("FirstCheckDigitFails", func(t *testing.T) {
		// 111.444.777-35 is valid.
		// 11144477745 -> first check digit (3) changed to 4.
		assert.Equal(t, ErrInvalidCPF, validateCPF("11144477745"))
	})
	t.Run("SecondCheckDigitFails", func(t *testing.T) {
		// 111.444.777-35 is valid.
		// 11144477730 -> first check digit (3) is correct, second (5) is wrong.
		assert.Equal(t, ErrInvalidCPF, validateCPF("11144477730"))
	})
}

func TestValidateCNPJ_Direct(t *testing.T) {
	t.Run("InvalidLength", func(t *testing.T) {
		assert.Equal(t, ErrCNPJLength, validateCNPJ("123"))
	})
	t.Run("InvalidChars", func(t *testing.T) {
		// Needs to be in the first 12 chars to avoid triggering the numeric check digit check first.
		assert.Equal(t, ErrInvalidCNPJ, validateCNPJ("12345678901!00"))
	})
	t.Run("FirstCheckDigitFails", func(t *testing.T) {
		// ABCDEFGH000195 is valid.
		// ABCDEFGH000105 -> first check digit (9) changed to 0.
		assert.Equal(t, ErrInvalidCNPJ, validateCNPJ("ABCDEFGH000105"))
	})
	t.Run("SecondCheckDigitFails", func(t *testing.T) {
		// ABCDEFGH000195 is valid.
		// ABCDEFGH000190 -> first check digit (9) is correct, second (5) is wrong.
		assert.Equal(t, ErrInvalidCNPJ, validateCNPJ("ABCDEFGH000190"))
	})
}

func TestHashPassword_Errors(t *testing.T) {
	t.Run("PasswordTooLong", func(t *testing.T) {
		u := NewUser("John", "john@e.com", strings.Repeat("a", 73))
		err := u.HashPassword(context.Background())
		assert.Equal(t, ErrUserPasswordTooLong, err)
	})
}

func TestNewCustomer_ErrorPaths(t *testing.T) {
	t.Run("CreateUserToDomainFails", func(t *testing.T) {
		_, err := NewCustomer("", "invalid", "", IndividualCustomerType, "11144477735", "", "11999999999")
		assert.Error(t, err)
	})
}

func TestCreateUserToDomain_HashPasswordFails(t *testing.T) {
	// Temporarily change hashingCost to an invalid value
	oldCost := hashingCost
	defer func() { hashingCost = oldCost }()
	hashingCost = 32

	_, err := CreateUserToDomain("John Doe", "john@example.com", "Secret@123")
	assert.Error(t, err)
}
