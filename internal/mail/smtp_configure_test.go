package mail

import (
	"os"
	"testing"
)

func smtpVarsClean() {
	for _, env := range []string{"SMTP_USERNAME", "SMTP_PASSWORD", "SMTP_HOST", "SMTP_PORT", "SMTP_FROM_ADDRESS"} {
		_ = os.Setenv(env, "")
	}
}

func TestSmtpConfigureBasics(t *testing.T) {
	FailTestOnError(t, SetupTestEnv())
	smtpVarsClean()

	_ = os.Setenv("SMTP_USERNAME", "test@example.com")
	_ = os.Setenv("SMTP_PASSWORD", "hunter2")
	_ = os.Setenv("SMTP_HOST", "smtp.comentario.app")
	_ = os.Setenv("SMTP_FROM_ADDRESS", "no-reply@comentario.app")

	if err := SMTPConfigure(); err != nil {
		t.Errorf("unexpected error when configuring SMTP: %v", err)
		return
	}
}

func TestSmtpConfigureEmptyHost(t *testing.T) {
	FailTestOnError(t, SetupTestEnv())
	smtpVarsClean()

	_ = os.Setenv("SMTP_USERNAME", "test@example.com")
	_ = os.Setenv("SMTP_PASSWORD", "hunter2")
	_ = os.Setenv("SMTP_FROM_ADDRESS", "no-reply@comentario.app")

	if err := SMTPConfigure(); err != nil {
		t.Errorf("unexpected error when configuring SMTP: %v", err)
		return
	}

	if SMTPConfigured {
		t.Errorf("SMTP configured when it should not be due to empty COMENTARIO_SMTP_HOST")
		return
	}
}

func TestSmtpConfigureEmptyAddress(t *testing.T) {
	FailTestOnError(t, SetupTestEnv())
	smtpVarsClean()

	_ = os.Setenv("SMTP_USERNAME", "test@example.com")
	_ = os.Setenv("SMTP_PASSWORD", "hunter2")
	_ = os.Setenv("SMTP_HOST", "smtp.comentario.app")
	_ = os.Setenv("SMTP_PORT", "25")

	if err := SMTPConfigure(); err == nil {
		t.Errorf("expected error not found; SMTP should not be configured when COMENTARIO_SMTP_FROM_ADDRESS is empty")
		return
	}
}
