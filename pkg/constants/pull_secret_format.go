package constants

// PullSecretFormat ImagePullSecret
// EXAMPLE
//
//	{
//		"auths": {
//			"harbor(harbor.power.sd.k9s.space)": {
//				"auth": "base64(admin:password)"
//			}
//		}
//	}
const PullSecretFormat = `
{
	"auths": {
		"@@harbor_with_ssl@@": {
			"auth": "@@harbor_admin_username_password@@"
		}
	}
}`
