package ssl

import (
	"github.com/iris-contrib/middleware/secure"
	"github.com/spf13/viper"
)

func New() *secure.Secure {
	s := secure.New(secure.Options{
		// AllowedHosts是允许的完全限定域名列表。默认为空列表，允许任何和所有主机名。
		AllowedHosts: viper.GetStringSlice("ssl.AllowedHosts"),

		//如果SSLRedirect设置为true，则仅允许HTTPS请求。默认值为false。
		SSLRedirect: viper.GetBool("ssl.SSLRedirect"),

		//如果SSLTemporaryRedirect为true，则在重定向时将使用a 302。默认值为false（301）。
		SSLTemporaryRedirect: viper.GetBool("ssl.SSLTemporaryRedirect"),

		// SSLHost是用于将HTTP请求重定向到HTTPS的主机名。默认值为“”，表示使用相同的主机。
		SSLHost: viper.GetString("ssl.SSLHost"),

		// STSSeconds是Strict-Transport-Security标头的max-age。默认值为0，不包括header。
		STSSeconds: viper.GetInt64("ssl.STSSeconds"),

		//如果STSIncludeSubdomains设置为true，则`includeSubdomains`将附加到Strict-Transport-Security标头。默认值为false。
		STSIncludeSubdomains: viper.GetBool("ssl.STSIncludeSubdomains"),

		//如果STSPreload设置为true，则`preload`标志将附加到Strict-Transport-Security标头。默认值为false。
		STSPreload: viper.GetBool("ssl.STSPreload"),

		//仅当连接是HTTPS时才包含STS标头。如果要强制始终添加，请设置为true."IsDevelopment"仍然覆盖了这一点。默认值为false。
		ForceSTSHeader: viper.GetBool("ssl.ForceSTSHeader"),

		//如果FrameDeny设置为true，则添加值为"DENY"的X-Frame-Options标头。默认值为false。
		FrameDeny: viper.GetBool("ssl.FrameDeny"),

		// CustomFrameOptionsValue允许使用自定义值设置X-Frame-Options标头值。这会覆盖FrameDeny选项。
		CustomFrameOptionsValue: viper.GetString("ssl.CustomFrameOptionsValue"),

		//如果ContentTypeNosniff为true，则使用值nosniff添加X-Content-Type-Options标头。默认值为false。
		ContentTypeNosniff: viper.GetBool("ssl.ContentTypeNosniff"),

		//如果BrowserXssFilter为true，则添加值为1的X-XSS-Protection标头;模式= block`。默认值为false。
		BrowserXSSFilter: viper.GetBool("ssl.BrowserXSSFilter"),

		// ContentSecurityPolicy允许使用自定义值设置Content-Security-Policy标头值。默认为""。
		ContentSecurityPolicy: viper.GetString("ssl.ContentSecurityPolicy"),

		// PublicKey实现HPKP以防止伪造证书的MITM攻击。默认为""。
		PublicKey: viper.GetString("ssl.PublicKey"),

		//这将导致在开发期间忽略AllowedHosts，SSLRedirect和STSSeconds/STSIncludeSubdomains选项。 部署到生产时，请务必将其设置为false。
		IsDevelopment: viper.GetBool("ssl.IsDevelopment"),
	})

	return s
}
