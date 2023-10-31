package lokirole

const rolePolicy = `{
	"Version": "2012-10-17",
	"Statement": [
		{
			"Effect": "Allow",
			"Action": [
				"s3:ListBucket",
				"s3:PutObject",
				"s3:GetObject",
				"s3:DeleteObject"
			],
			"Resource": [
			"arn:aws:s3:::@INSTALLATION@-g8s-loki",
			"arn:aws:s3:::@INSTALLATION@-g8s-loki/*"
			]
		},
		{
			"Effect": "Allow",
			"Action": [
			"s3:GetAccessPoint",
			"s3:GetAccountPublicAccessBlock",
			"s3:ListAccessPoints"
			],
			"Resource": "*"
		}
	]
}`

const trustIdentityPolicy = `{
	"Version": "2012-10-17",
	"Statement": [
		{
			"Effect": "Allow",
			"Principal": {
				"Federated": "arn:aws:iam::@ACCOUNT_ID@:oidc-provider/irsa.@INSTALLATION@.@CLOUD_DOMAIN@"
			},
			"Action": "sts:AssumeRoleWithWebIdentity",
			"Condition": {
				"StringEquals": {
					"irsa.@INSTALLATION@.@CLOUD_DOMAIN@:sub": "system:serviceaccount:loki:loki"
				}
			}
		}
	]
}`
