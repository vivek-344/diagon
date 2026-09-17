package security

import "testing"

func TestPasswordHashAndVerify(t *testing.T) {
	password := "correct horse battery staple"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	ok, err := VerifyPassword(password, hash)
	if err != nil {
		t.Fatalf("VerifyPassword() error = %v", err)
	}

	if !ok {
		t.Fatal("VerifyPassword() returned false for correct password")
	}

	ok, err = VerifyPassword("wrong password", hash)
	if err != nil {
		t.Fatalf("VerifyPassword() error = %v", err)
	}

	if ok {
		t.Fatal("VerifyPassword() returned true for incorrect password")
	}
}
