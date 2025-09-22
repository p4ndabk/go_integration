package email

// GetDefaultEmailHTML returns the HTML template for regular emails using payload content
func GetDefaultEmailHTML(subject, body, companyName string) string {
	template := `<!doctype html>
<html lang="pt-BR">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width,initial-scale=1">
  <title>` + subject + `</title>
  <style>
    body,table,td {font-family: Arial, Helvetica, sans-serif; margin:0; padding:0;}
    img {border:0; display:block;}
    a {color:#ffffff; text-decoration:none}

    .wrapper {width:100%; background:#f0f2f5; padding:30px 0;}
    .content {max-width:600px; background:#ffffff; margin:0 auto; border-radius:10px; overflow:hidden; box-shadow:0 4px 12px rgba(0,0,0,0.08)}

    .header {background:#1a73e8; padding:30px; text-align:center; color:#fff;}
    .header h1 {margin:0; font-size:24px;}
    .header img {max-width:200px; height:auto; margin:0 auto 20px auto; display:block; background:#ffffff; padding:10px; border-radius:8px;}

    .body {padding:30px; color:#333; line-height:1.6;}
    .body h2 {margin-top:0; color:#1a73e8;}

    .btn {display:inline-block; background:#1a73e8; padding:12px 20px; border-radius:6px; font-weight:bold; color:#ffffff;}

    .footer {background:#f7f7f7; padding:20px; font-size:12px; text-align:center; color:#666;}

    @media only screen and (max-width:480px) {
      .header h1 {font-size:20px;}
      .body h2 {font-size:18px;}
    }
  </style>
</head>
<body>
  <table role="presentation" class="wrapper" width="100%" cellspacing="0" cellpadding="0">
    <tr>
      <td align="center">
        <table role="presentation" class="content" width="100%" cellspacing="0" cellpadding="0">
          
          <!-- Header -->
          <tr>
            <td class="header">
              <img src="https://northfi.com.br/img/logoNorthPreto.png" alt="` + companyName + `" style="max-width:200px; height:auto; margin-bottom:20px;">
              <h1>` + subject + `</h1>
            </td>
          </tr>

          <!-- Body -->
          <tr>
            <td class="body">
              <div style="white-space: pre-line;">` + body + `</div>
            </td>
          </tr>

          <!-- Footer -->
          <tr>
            <td class="footer">
              <p>Você recebeu este e-mail de ` + companyName + `.</p>
            </td>
          </tr>

        </table>
      </td>
    </tr>
  </table>
</body>
</html>`

	return template
}

// GetWelcomeEmailHTML returns the HTML template for welcome emails
func GetWelcomeEmailHTML(username, companyName string) string {
	template := `<!doctype html>
<html lang="pt-BR">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width,initial-scale=1">
  <title>Bem-vindo</title>
  <style>
    body,table,td {font-family: Arial, Helvetica, sans-serif; margin:0; padding:0;}
    img {border:0; display:block;}
    a {color:#ffffff; text-decoration:none}

    .wrapper {width:100%; background:#f0f2f5; padding:30px 0;}
    .content {max-width:600px; background:#ffffff; margin:0 auto; border-radius:10px; overflow:hidden; box-shadow:0 4px 12px rgba(0,0,0,0.08)}

    .header {background:#1a73e8; padding:30px; text-align:center; color:#fff;}
    .header h1 {margin:0; font-size:24px;}
    .header img {max-width:200px; height:auto; margin:0 auto 20px auto; display:block; background:#ffffff; padding:10px; border-radius:8px;}

    .body {padding:30px; color:#333; line-height:1.6;}
    .body h2 {margin-top:0; color:#1a73e8;}

    .btn {display:inline-block; background:#1a73e8; padding:12px 20px; border-radius:6px; font-weight:bold; color:#ffffff;}

    .footer {background:#f7f7f7; padding:20px; font-size:12px; text-align:center; color:#666;}

    @media only screen and (max-width:480px) {
      .header h1 {font-size:20px;}
      .body h2 {font-size:18px;}
    }
  </style>
</head>
<body>
  <table role="presentation" class="wrapper" width="100%" cellspacing="0" cellpadding="0">
    <tr>
      <td align="center">
        <table role="presentation" class="content" width="100%" cellspacing="0" cellpadding="0">
          
          <!-- Header -->
          <tr>
            <td class="header">
              <img src="https://northfi.com.br/img/logoNorthPreto.png" alt="` + companyName + `" style="max-width:200px; height:auto; margin-bottom:20px;">
              <h1>Bem-vindo(a) à ` + companyName + `</h1>
            </td>
          </tr>

          <!-- Body -->
          <tr>
            <td class="body">
              <h2>Estamos muito felizes em ter você conosco!</h2>
              <p>Agora você faz parte da nossa comunidade e terá acesso a todas as vantagens que preparamos para você.</p>

              <p>Para começar, recomendamos:</p>
              <ul>
                <li>Completar seu perfil;</li>
                <li>Explorar os recursos principais;</li>
                <li>Ativar notificações para não perder nenhuma novidade.</li>
              </ul>

              <p style="margin:20px 0; text-align:center;">
                <a href="https://northfi.com.br" target="_blank" class="btn">Acessar minha conta</a>
              </p>

              <p>Se precisar de ajuda, nossa equipe está à disposição. Basta responder este e-mail ou acessar nossa central de suporte.</p>
            </td>
          </tr>

          <!-- Footer -->
          <tr>
            <td class="footer">
              <p>Você recebeu este e-mail porque se cadastrou em ` + companyName + `.</p>
            </td>
          </tr>

        </table>
      </td>
    </tr>
  </table>
</body>
</html>`

	return template
}

// GetVerificationEmailHTML returns the HTML template for email verification
func GetVerificationEmailHTML(username, companyName, verifyURL string) string {
	template := `<!doctype html>
<html lang="pt-BR">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width,initial-scale=1">
  <title>Verificação de Email</title>
  <style>
    body,table,td {font-family: Arial, Helvetica, sans-serif; margin:0; padding:0;}
    img {border:0; display:block;}
    a {color:#ffffff; text-decoration:none}

    .wrapper {width:100%; background:#f0f2f5; padding:30px 0;}
    .content {max-width:600px; background:#ffffff; margin:0 auto; border-radius:10px; overflow:hidden; box-shadow:0 4px 12px rgba(0,0,0,0.08)}

    .header {background:#1a73e8; padding:30px; text-align:center; color:#fff;}
    .header h1 {margin:0; font-size:24px;}
    .header img {max-width:200px; height:auto; margin:0 auto 20px auto; display:block; background:#ffffff; padding:10px; border-radius:8px;}

    .body {padding:30px; color:#333; line-height:1.6;}
    .body h2 {margin-top:0; color:#1a73e8;}

    .btn {display:inline-block; background:#1a73e8; padding:12px 20px; border-radius:6px; font-weight:bold; color:#ffffff !important; text-decoration:none !important; border:none; cursor:pointer;}
    .btn:hover {background:#0d5aa7;}

    .footer {background:#f7f7f7; padding:20px; font-size:12px; text-align:center; color:#666;}

    @media only screen and (max-width:480px) {
      .header h1 {font-size:20px;}
      .body h2 {font-size:18px;}
    }
  </style>
</head>
<body>
  <table role="presentation" class="wrapper" width="100%" cellspacing="0" cellpadding="0">
    <tr>
      <td align="center">
        <table role="presentation" class="content" width="100%" cellspacing="0" cellpadding="0">
          
          <!-- Header -->
          <tr>
            <td class="header">
              <img src="https://northfi.com.br/img/logoNorthPreto.png" alt="` + companyName + `" style="max-width:200px; height:auto; margin-bottom:20px;">
              <h1>Verificar Email</h1>
            </td>
          </tr>

          <!-- Body -->
          <tr>
            <td class="body">
              <h2>Olá, ` + username + `!</h2>
              <p>Para completar seu cadastro na ` + companyName + `, precisamos verificar seu endereço de email.</p>

              <p>Clique no botão abaixo para confirmar seu email:</p>

              <p style="margin:30px 0; text-align:center;">
                <a href="` + verifyURL + `" target="_blank" class="btn" style="background:#1a73e8; color:#ffffff; padding:15px 30px; border-radius:6px; font-weight:bold; text-decoration:none; display:inline-block;">Verificar meu email</a>
              </p>

              <p>Se você não conseguir clicar no botão, copie e cole o link abaixo no seu navegador:</p>
              <p style="word-break: break-all; color: #666; background: #f8f9fa; padding: 10px; border-radius: 4px; font-family: monospace;">` + verifyURL + `</p>

              <p><strong>Este link expira em 24 horas.</strong></p>
            </td>
          </tr>

          <!-- Footer -->
          <tr>
            <td class="footer">
              <p>Se você não se cadastrou na ` + companyName + `, ignore este email.</p>
              <p>Este email foi enviado automaticamente, não responda.</p>
            </td>
          </tr>

        </table>
      </td>
    </tr>
  </table>
</body>
</html>`

	return template
}
