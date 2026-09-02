package auth

import "testing"

func TestGenerateAndValidateToken(t *testing.T) {
	token, err := GenerateToken("user-123", "admin", "test-secret")
	if err != nil {
		t.Fatalf("Expected no error, returned: %v", err)
	}

	claims, err := ValidateToken(token, "test-secret")
	if err != nil {
		t.Fatalf("Expected invalid token, returned: %v", err)
	}

	if claims.Subject != "user-123" {
		t.Errorf("expected sub 'user-123', returned '%s'", claims.Subject)
	}
	if claims.Role != "admin" {
		t.Errorf("expected role 'admin', returned '%s'", claims.Role)
	}
}

func TestValidateToken_WrongSecret(t *testing.T) {
	token, _ := GenerateToken("user-123", "admin", "test-secret")

	_, err := ValidateToken(token, "different-secret")
	if err == nil {
		t.Fatal("expected error validanting with the wrong secret")
	}
}