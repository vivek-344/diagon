package security

import "testing"

func TestHashPassword(t *testing.T) {
	password := "correct horse battery staple"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hash == password {
		t.Fatal("password was stored in plaintext")
	}

	if len(hash) == 0 {
		t.Fatal("expected password hash")
	}
}

func TestVerifyPassword(t *testing.T) {
	password := "correct horse battery staple"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	tests := []struct {
		name     string
		password string
		want     bool
	}{
		{
			name:     "correct password",
			password: password,
			want:     true,
		},
		{
			name:     "wrong password",
			password: "wrong password",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := VerifyPassword(
				tt.password,
				hash,
			)
			if err != nil {
				t.Fatalf(
					"unexpected error: %v",
					err,
				)
			}

			if got != tt.want {
				t.Fatalf(
					"expected %v, got %v",
					tt.want,
					got,
				)
			}
		})
	}
}

func TestVerifyPassword_InvalidHash(t *testing.T) {
	_, err := VerifyPassword(
		"password",
		"not-a-valid-hash",
	)

	if err == nil {
		t.Fatal("expected error")
	}
}
