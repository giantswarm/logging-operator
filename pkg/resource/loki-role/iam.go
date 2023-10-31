package lokirole

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/smithy-go"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"

	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"
)

type IamAdapter struct {
	accountId   string
	cloudDomain string
	iamClient   *iam.Client
	log         logr.Logger
}

func NewIamService(accountId string, cloudDomain string, iamClient *iam.Client, log logr.Logger) IamAdapter {
	return IamAdapter{
		accountId:   accountId,
		cloudDomain: cloudDomain,
		iamClient:   iamClient,
		log:         log,
	}
}

func (s IamAdapter) GetRole(ctx context.Context, roleName string) (*types.Role, error) {
	l := s.log.WithValues("role_name", roleName)
	output, err := s.iamClient.GetRole(ctx, &iam.GetRoleInput{
		RoleName: aws.String(roleName),
	})

	if err != nil {
		var apiError smithy.APIError
		if errors.As(err, &apiError) {
			switch apiError.(type) {
			case *types.NoSuchEntityException:
				l.Info("IAM role does not exist")
				return nil, nil
			default:
				l.Error(err, "Failed to fetch IAM Role")
				return nil, errors.WithStack(err)
			}
		}
		return nil, errors.WithStack(err)
	}
	return output.Role, nil
}

func (s IamAdapter) ConfigureRole(ctx context.Context, lc loggedcluster.Interface) error {
	roleName := getRoleName(lc)
	l := s.log.WithValues("role_name", roleName)

	role, err := s.GetRole(ctx, roleName)
	if err != nil {
		return err
	}

	if role == nil {
		_, err := s.iamClient.CreateRole(ctx, &iam.CreateRoleInput{
			RoleName:                 aws.String(roleName),
			AssumeRolePolicyDocument: aws.String(templateTrustPolicy(lc, s.accountId, s.cloudDomain)),
			Description:              aws.String("Role for Giant Swarm managed Loki"),
		})

		if err != nil {
			return errors.WithStack(err)
		}
		l.Info("IAM Role created")
	} else {
		_, err = s.iamClient.UpdateAssumeRolePolicy(ctx, &iam.UpdateAssumeRolePolicyInput{
			RoleName:       aws.String(roleName),
			PolicyDocument: aws.String(templateTrustPolicy(lc, s.accountId, s.cloudDomain)),
		})
		if err != nil {
			return errors.WithStack(err)
		}
	}

	_, err = s.iamClient.PutRolePolicy(ctx, &iam.PutRolePolicyInput{
		RoleName:       aws.String(roleName),
		PolicyName:     aws.String(roleName),
		PolicyDocument: aws.String(templateRolePolicy(lc)),
	})
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (s IamAdapter) DeleteRole(ctx context.Context, roleName string) error {
	l := s.log.WithValues("role_name", roleName)

	role, err := s.GetRole(ctx, roleName)
	if err != nil {
		return err
	}

	if role == nil {
		l.Info("IAM role does not exist, skipping deletion")
		return nil
	}
	// TODO add ownship tag and validate ownership?

	// clean any attached policies, otherwise deletion of role will not work
	err = s.cleanAttachedPolicies(ctx, roleName)
	if err != nil {
		return err
	}

	_, err = s.iamClient.RemoveRoleFromInstanceProfile(ctx, &iam.RemoveRoleFromInstanceProfileInput{
		InstanceProfileName: aws.String(roleName),
		RoleName:            aws.String(roleName),
	})
	if err != nil {
		var apiError smithy.APIError
		if errors.As(err, &apiError) {
			switch apiError.(type) {
			case *types.NoSuchEntityException:
				l.Info("no instance profile attached to role, skipping")
			default:
				l.Error(err, "failed to remove role from instance profile")
				return errors.WithStack(err)
			}
		}
	}

	_, err = s.iamClient.DeleteInstanceProfile(ctx, &iam.DeleteInstanceProfileInput{
		InstanceProfileName: aws.String(roleName),
	})
	if err != nil {
		var apiError smithy.APIError
		if errors.As(err, &apiError) {
			switch apiError.(type) {
			case *types.NoSuchEntityException:
				l.Info("no instance profile to delete, skipping")
			default:
				l.Error(err, "failed to delete instance profile")
				return errors.WithStack(err)
			}
		}
	}
	_, err = s.iamClient.DeleteRole(ctx, &iam.DeleteRoleInput{
		RoleName: aws.String(roleName),
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *IamAdapter) cleanAttachedPolicies(ctx context.Context, roleName string) error {
	l := s.log.WithValues("role_name", roleName)

	{
		o, err := s.iamClient.ListAttachedRolePolicies(ctx, &iam.ListAttachedRolePoliciesInput{
			RoleName: aws.String(roleName),
		})
		if err != nil {
			return err
		} else {
			for _, p := range o.AttachedPolicies {
				policy := p
				l.Info(fmt.Sprintf("detaching policy %s", *policy.PolicyName))

				_, err := s.iamClient.DetachRolePolicy(ctx, &iam.DetachRolePolicyInput{
					PolicyArn: policy.PolicyArn,
					RoleName:  aws.String(roleName),
				})
				if err != nil {
					l.Error(err, fmt.Sprintf("failed to detach policy %s", *policy.PolicyName))
					return err
				}

				l.Info(fmt.Sprintf("detached policy %s", *policy.PolicyName))
			}
		}
	}

	// clean inline policies
	{
		o, err := s.iamClient.ListRolePolicies(ctx, &iam.ListRolePoliciesInput{
			RoleName: aws.String(roleName),
		})
		if err != nil {
			l.Error(err, "failed to list inline policies")
			return err
		}

		for _, p := range o.PolicyNames {
			policy := p
			l.Info(fmt.Sprintf("deleting inline policy %s", policy))
			_, err := s.iamClient.DeleteRolePolicy(ctx, &iam.DeleteRolePolicyInput{
				RoleName:   aws.String(roleName),
				PolicyName: aws.String(policy),
			})
			if err != nil {
				l.Error(err, fmt.Sprintf("failed to delete inline policy %s", policy))
				return err
			}
			l.Info(fmt.Sprintf("deleted inline policy %s", policy))
		}
	}

	l.Info("cleaned attached and inline policies from IAM Role")
	return nil
}

func (r *Reconciler) createIamClient(ctx context.Context, roleToAssume string, region string) (*iam.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, err
	}

	// Assume role
	stsClient := sts.NewFromConfig(cfg)
	credentials := stscreds.NewAssumeRoleProvider(stsClient, roleToAssume)
	cfg.Credentials = aws.NewCredentialsCache(credentials)

	return iam.NewFromConfig(cfg), nil
}
func getRoleName(lc loggedcluster.Interface) string {
	return fmt.Sprintf("giantswarm-%s-loki", lc.GetInstallationName())
}

func templateRolePolicy(lc loggedcluster.Interface) string {
	return strings.ReplaceAll(rolePolicy, "@INSTALLATION@", lc.GetInstallationName())
}

func templateTrustPolicy(lc loggedcluster.Interface, accountId string, cloudDomain string) string {
	return strings.ReplaceAll(
		strings.ReplaceAll(
			strings.ReplaceAll(
				trustIdentityPolicy, "@CLOUD_DOMAIN@", cloudDomain),
			"@INSTALLATION@", lc.GetInstallationName()),
		"@ACCOUNT_ID@", accountId)
}
