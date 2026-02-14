package service

import (
	"context"
	"crypto/tls"
	"database/sql"
	"fmt"
	"math/rand"
	"net/smtp"
	"time"
)

type EmailService struct {
	db          *sql.DB
	smtpHost    string
	smtpPort    string
	senderEmail string
	senderPass  string
}

func NewEmailService(db *sql.DB, smtpHost, smtpPort, senderEmail, senderPass string) *EmailService {
	return &EmailService{
		db:          db,
		smtpHost:    smtpHost,
		smtpPort:    smtpPort,
		senderEmail: senderEmail,
		senderPass:  senderPass,
	}
}

type SendMailRequest struct {
	Mail string `json:"mail"`
}

type MailCheckRequest struct {
	Mail   string `json:"mail"`
	Number int    `json:"number"`
}

type GlobalResponseDto struct {
	Result  bool   `json:"result"`
	Message string `json:"message"`
}

func (s *EmailService) SendSignUpMail(ctx context.Context, email string) (*GlobalResponseDto, error) {
	// Check if member already exists
	var exists bool
	err := s.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM member WHERE username = ?)`, email).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("failed to check member: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("이미 가입된 회원입니다.")
	}

	number := generateVerificationCode()

	subject := "[LOA TODO] 이메일 인증"
	body := `<h1>LoaTodo에 가입해주셔서 감사합니다.</h1>` +
		`<h3>인증 번호입니다.</h3>` +
		`<p>3분내로 입력해주세요.</p>` +
		fmt.Sprintf(`<h4> 인증번호 : %d</h4>`, number)

	if err := s.sendMail(email, subject, body); err != nil {
		return nil, fmt.Errorf("메일 전송에 실패하였습니다: %w", err)
	}

	// Save auth mail record
	_, err = s.db.ExecContext(ctx,
		`INSERT INTO auth_mail (mail, number, is_auth, created_date, modified_date) VALUES (?, ?, false, NOW(), NOW())`,
		email, number)
	if err != nil {
		return nil, fmt.Errorf("인증번호 저장에 실패하였습니다: %w", err)
	}

	return &GlobalResponseDto{Result: true, Message: "인증번호 전송이 정상처리 되었습니다."}, nil
}

func (s *EmailService) SendResetPasswordMail(ctx context.Context, email string) error {
	// Check member exists and is not SNS user
	var authProvider string
	err := s.db.QueryRowContext(ctx, `SELECT auth_provider FROM member WHERE username = ?`, email).Scan(&authProvider)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("존재하지 않는 회원입니다.")
		}
		return fmt.Errorf("회원 조회에 실패하였습니다: %w", err)
	}
	if authProvider != "none" {
		return fmt.Errorf("SNS 가입자는 비밀번호 변경이 불가능 합니다.")
	}

	number := generateVerificationCode()

	subject := "[LOA TODO] 이메일 인증"
	body := `<h1>LoaTodo에 이용해주셔서 감사합니다.</h1>` +
		`<h3>인증 번호입니다.</h3>` +
		`<p>3분내로 입력해주세요.</p>` +
		fmt.Sprintf(`<h4> 인증번호 : %d</h4>`, number)

	if err := s.sendMail(email, subject, body); err != nil {
		return fmt.Errorf("메일 전송에 실패하였습니다: %w", err)
	}

	_, err = s.db.ExecContext(ctx,
		`INSERT INTO auth_mail (mail, number, is_auth, created_date, modified_date) VALUES (?, ?, false, NOW(), NOW())`,
		email, number)
	if err != nil {
		return fmt.Errorf("인증번호 저장에 실패하였습니다: %w", err)
	}

	return nil
}

func (s *EmailService) CheckMail(ctx context.Context, req MailCheckRequest) (*GlobalResponseDto, error) {
	var id int64
	var createdDate time.Time
	err := s.db.QueryRowContext(ctx,
		`SELECT auth_mail_id, created_date FROM auth_mail WHERE mail = ? AND number = ? ORDER BY created_date DESC LIMIT 1`,
		req.Mail, req.Number).Scan(&id, &createdDate)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("유효하지 않은 인증번호 입니다.")
		}
		return nil, err
	}

	// Check within 3 minutes
	if time.Since(createdDate).Minutes() <= 3 {
		_, err = s.db.ExecContext(ctx, `UPDATE auth_mail SET is_auth = true WHERE auth_mail_id = ?`, id)
		if err != nil {
			return nil, err
		}
		return &GlobalResponseDto{Result: true, Message: "이메일 인증번호 성공"}, nil
	}

	return &GlobalResponseDto{Result: false, Message: "인증번호가 일치하지 않거나 만료되었습니다."}, nil
}

func generateVerificationCode() int {
	return rand.Intn(90000) + 100000
}

func (s *EmailService) sendMail(to, subject, htmlBody string) error {
	headers := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n",
		s.senderEmail, to, subject)
	msg := []byte(headers + htmlBody)

	auth := smtp.PlainAuth("", s.senderEmail, s.senderPass, s.smtpHost)

	tlsConfig := &tls.Config{
		ServerName: s.smtpHost,
	}

	conn, err := tls.Dial("tcp", s.smtpHost+":"+s.smtpPort, tlsConfig)
	if err != nil {
		return err
	}

	client, err := smtp.NewClient(conn, s.smtpHost)
	if err != nil {
		return err
	}
	defer client.Close()

	if err := client.Auth(auth); err != nil {
		return err
	}
	if err := client.Mail(s.senderEmail); err != nil {
		return err
	}
	if err := client.Rcpt(to); err != nil {
		return err
	}

	w, err := client.Data()
	if err != nil {
		return err
	}
	_, err = w.Write(msg)
	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}

	return client.Quit()
}
