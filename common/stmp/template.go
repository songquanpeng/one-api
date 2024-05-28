package stmp

import (
	"one-api/common/config"
)

func getLogo() string {
	if config.Logo == "" {
		return ""
	}
	return `<table class="logo" width="100%">
	<tr>
	  <td>
		<img src="` + config.Logo + `" width="130" style="max-width: 100%"
		/>
	  </td>
	</tr>
  </table>`
}

func getSystemName() string {
	if config.SystemName == "" {
		return "One API"
	}

	return config.SystemName
}

func getDefaultTemplate(content string) string {
	return `<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Transitional//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd">
	<html xmlns="http://www.w3.org/1999/xhtml">
	  <head>
		<meta http-equiv="Content-Type" content="text/html; charset=utf-8" />
		<meta http-equiv="X-UA-Compatible" content="IE=edge" />
		<meta name="viewport" content="width=device-width, initial-scale=1.0" />
		<style type="text/css">
		  body {
			margin: 0;
			background-color: #eceff1;
			font-family: Helvetica, sans-serif;
		  }
		  table {
			border-spacing: 0;
		  }
		  td {
			padding: 0;
		  }
		  img {
			border: 0;
		  }
		  .wrapper {
			width: 100%;
			table-layout: fixed;
			background-color: #eceff1;
			padding-bottom: 60px;
			padding-top: 60px;
		  }
		  .main {
			background-color: #ffffff;
			border-spacing: 0;
			color: #000000;
			border-radius: 10px;
			border-color: #ebebeb;
			border-width: 1px;
			border-style: solid;
			padding: 10px 30px;
			line-height: 25px;
			font-size: 16px;
			text-align: start;
			width: 600px;
			}
		  .button {
			background-color: #000000;
			color: #ffffff;
			text-decoration: none;
			padding: 12px 20px;
			font-weight: bold;
			border-radius: 5px;
		  }
		  .logo {
			text-align: center;
			margin: 10px auto;
		  }
		  .footer {
			text-align: center;
			color: #858585
		  }
		</style>
	  </head>
	  <body>
		<center class="wrapper">
		  ` + getLogo() + `
		  <table class="main" width="100%">
			<tr>
			  <td>
				` + content + `
			  </td>
			</tr>
		  </table>
		  <table class="footer" width="100%">
			<tr>
			  <td width="100%">
				<p>© ` + getSystemName() + `</p>
			  </td>
			</tr>
		  </table>
		</center>
	  </body>
	</html>`
}
