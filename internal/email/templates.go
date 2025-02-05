package email

type VerificationEmailData struct {
	FirstName        string
	VerificationLink string
}

const verificationEmailTemplate = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Verify Your Email</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h2>Welcome to Options Manager!</h2>
        <p>Hello {{.FirstName}},</p>
        <p>Thank you for registering. Please verify your email address by clicking the button below:</p>
        <p style="text-align: center;">
            <a href="{{.VerificationLink}}" 
               style="display: inline-block; padding: 12px 24px; background-color: #3b82f6; color: white; 
                      text-decoration: none; border-radius: 4px; font-weight: bold;">
                Verify Email Address
            </a>
        </p>
        <p>If the button doesn't work, you can copy and paste this link into your browser:</p>
        <p>{{.VerificationLink}}</p>
        <p>This link will expire in 24 hours.</p>
        <p>If you didn't create an account, you can safely ignore this email.</p>
        <br>
        <p>Best regards,<br>The Options Manager Team</p>
    </div>
</body>
</html>`
