package auth

import "testing"

func TestHashPassword(t *testing.T) {
	hash, err := HashPassword("senha123")
	if err != nil {
		t.Fatalf("expected no error, received: %v", err)
	}
	if hash == "" {
		t.Fatal("expected a no empty hash")
	}
	if hash == "senha123" {
		t.Fatal("hash can not be equal to original password")
	}
}

func TestCheckPassword_CorrectPassword(t *testing.T) {
	hash, _ := HashPassword("senha123")

	err := CheckPassword("senha123", hash)
	if err != nil {
		t.Fatalf("expected invalid password, returned error: %v", err)
	}
}

func TestCheckPassword_WrongPassword(t *testing.T) {
	hash, _ := HashPassword("senha123")

	err := CheckPassword("WrongPassword", hash)
	if err == nil {
		t.Fatal("expected wrong password error,  but no error was returned")
	}
}