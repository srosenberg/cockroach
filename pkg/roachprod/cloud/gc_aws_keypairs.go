package cloud

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
)

func getEC2Client(region string) (*ec2.Client, error) {
	__antithesis_instrumentation__.Notify(180168)
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		__antithesis_instrumentation__.Notify(180170)
		return nil, errors.Wrap(err, "getEC2Client: failed to get EC2 client")
	} else {
		__antithesis_instrumentation__.Notify(180171)
	}
	__antithesis_instrumentation__.Notify(180169)
	return ec2.NewFromConfig(cfg), nil
}

func getIAMClient(region string) (*iam.Client, error) {
	__antithesis_instrumentation__.Notify(180172)
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		__antithesis_instrumentation__.Notify(180174)
		return nil, errors.Wrap(err, "getIAMClient: failed to get IAM client")
	} else {
		__antithesis_instrumentation__.Notify(180175)
	}
	__antithesis_instrumentation__.Notify(180173)
	return iam.NewFromConfig(cfg), nil
}

func getIAMUsers(IAMClient *iam.Client) ([]iamtypes.User, error) {
	__antithesis_instrumentation__.Notify(180176)
	var users []iamtypes.User
	input := iam.ListUsersInput{}

	isTruncated := true
	for isTruncated {
		__antithesis_instrumentation__.Notify(180178)
		resp, err := IAMClient.ListUsers(context.TODO(), &input)
		if err != nil {
			__antithesis_instrumentation__.Notify(180180)
			return nil, errors.Wrap(err, "getIAMUsers: failed to list IAM users")
		} else {
			__antithesis_instrumentation__.Notify(180181)
		}
		__antithesis_instrumentation__.Notify(180179)
		users = append(users, resp.Users...)

		isTruncated = resp.IsTruncated
		if isTruncated {
			__antithesis_instrumentation__.Notify(180182)

			input = iam.ListUsersInput{Marker: resp.Marker}
		} else {
			__antithesis_instrumentation__.Notify(180183)
		}
	}
	__antithesis_instrumentation__.Notify(180177)
	return users, nil
}

func getUsersWithActiveAccessKey(
	IAMClient *iam.Client, users []iamtypes.User,
) (map[string]bool, error) {
	__antithesis_instrumentation__.Notify(180184)
	usersWithActiveAccessKey := make(map[string]bool)
	for _, user := range users {
		__antithesis_instrumentation__.Notify(180186)
		input := iam.ListAccessKeysInput{UserName: user.UserName}

		isTruncated := true
	outerLoop:
		for isTruncated {
			__antithesis_instrumentation__.Notify(180187)
			resp, err := IAMClient.ListAccessKeys(context.TODO(), &input)
			if err != nil {
				__antithesis_instrumentation__.Notify(180190)
				return nil, errors.Wrap(err, "getUsersWithActiveAccessKey: failed to list access keys")
			} else {
				__antithesis_instrumentation__.Notify(180191)
			}
			__antithesis_instrumentation__.Notify(180188)
			for _, key := range resp.AccessKeyMetadata {
				__antithesis_instrumentation__.Notify(180192)
				if key.Status == "Active" {
					__antithesis_instrumentation__.Notify(180193)
					usersWithActiveAccessKey[*user.UserName] = true
					break outerLoop
				} else {
					__antithesis_instrumentation__.Notify(180194)
				}
			}
			__antithesis_instrumentation__.Notify(180189)

			isTruncated = resp.IsTruncated
			if isTruncated {
				__antithesis_instrumentation__.Notify(180195)

				input = iam.ListAccessKeysInput{UserName: user.UserName, Marker: resp.Marker}
			} else {
				__antithesis_instrumentation__.Notify(180196)
			}
		}
	}
	__antithesis_instrumentation__.Notify(180185)
	return usersWithActiveAccessKey, nil
}

func getUsersWithConsoleAccess(
	IAMClient *iam.Client, users []iamtypes.User,
) (map[string]bool, error) {
	__antithesis_instrumentation__.Notify(180197)
	usersWithConsoleAccess := make(map[string]bool)
	for _, user := range users {
		__antithesis_instrumentation__.Notify(180199)
		input := iam.GetLoginProfileInput{UserName: user.UserName}
		_, err := IAMClient.GetLoginProfile(context.TODO(), &input)
		if err != nil {
			__antithesis_instrumentation__.Notify(180201)

			var nse *iamtypes.NoSuchEntityException
			if errors.As(err, &nse) {
				__antithesis_instrumentation__.Notify(180203)
				continue
			} else {
				__antithesis_instrumentation__.Notify(180204)
			}
			__antithesis_instrumentation__.Notify(180202)
			return nil, errors.Wrap(err, "getUsersWithConsoleAccess: failed to get login profile")
		} else {
			__antithesis_instrumentation__.Notify(180205)
		}
		__antithesis_instrumentation__.Notify(180200)
		usersWithConsoleAccess[*user.UserName] = true
	}
	__antithesis_instrumentation__.Notify(180198)
	return usersWithConsoleAccess, nil
}

func getUsersWithMFAEnabled(IAMClient *iam.Client, users []iamtypes.User) (map[string]bool, error) {
	__antithesis_instrumentation__.Notify(180206)
	usersWithMFAEnabled := make(map[string]bool)
	for _, user := range users {
		__antithesis_instrumentation__.Notify(180208)
		input := iam.ListMFADevicesInput{UserName: user.UserName}
		resp, err := IAMClient.ListMFADevices(context.TODO(), &input)
		if err != nil {
			__antithesis_instrumentation__.Notify(180210)
			return nil, errors.Wrap(err, "getUsersWithMFAEnabled: failed to list mfa devices")
		} else {
			__antithesis_instrumentation__.Notify(180211)
		}
		__antithesis_instrumentation__.Notify(180209)
		if len(resp.MFADevices) > 0 {
			__antithesis_instrumentation__.Notify(180212)
			usersWithMFAEnabled[*user.UserName] = true
		} else {
			__antithesis_instrumentation__.Notify(180213)
		}
	}
	__antithesis_instrumentation__.Notify(180207)
	return usersWithMFAEnabled, nil
}

func getRegions() ([]ec2types.Region, error) {
	__antithesis_instrumentation__.Notify(180214)

	EC2Client, err := getEC2Client("")
	if err != nil {
		__antithesis_instrumentation__.Notify(180217)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(180218)
	}
	__antithesis_instrumentation__.Notify(180215)
	input := ec2.DescribeRegionsInput{}

	resp, err := EC2Client.DescribeRegions(context.TODO(), &input)
	if err != nil {
		__antithesis_instrumentation__.Notify(180219)
		return nil, errors.Wrap(err, "getRegions: failed to describe regions")
	} else {
		__antithesis_instrumentation__.Notify(180220)
	}
	__antithesis_instrumentation__.Notify(180216)
	return resp.Regions, nil
}

func getTagsValues(tags []ec2types.Tag) (string, string) {
	__antithesis_instrumentation__.Notify(180221)
	IAMUserName := ""
	createdAt := ""
	for _, tag := range tags {
		__antithesis_instrumentation__.Notify(180223)
		if *tag.Key == "IAMUserName" {
			__antithesis_instrumentation__.Notify(180224)
			IAMUserName = *tag.Value
		} else {
			__antithesis_instrumentation__.Notify(180225)
			if *tag.Key == "CreatedAt" {
				__antithesis_instrumentation__.Notify(180226)
				createdAt = *tag.Value
			} else {
				__antithesis_instrumentation__.Notify(180227)
			}
		}
	}
	__antithesis_instrumentation__.Notify(180222)
	return IAMUserName, createdAt
}

func getIAMUserNameFromKeyname(keyName string) string {
	__antithesis_instrumentation__.Notify(180228)
	if len(keyName) > 29 && func() bool {
		__antithesis_instrumentation__.Notify(180230)
		return keyName[len(keyName)-29:len(keyName)-29+1] == "-" == true
	}() == true {
		__antithesis_instrumentation__.Notify(180231)
		return keyName[:len(keyName)-29]
	} else {
		__antithesis_instrumentation__.Notify(180232)
	}
	__antithesis_instrumentation__.Notify(180229)
	return ""
}

func getKeyPairs(EC2Client *ec2.Client) ([]ec2types.KeyPairInfo, error) {
	__antithesis_instrumentation__.Notify(180233)
	input := ec2.DescribeKeyPairsInput{}
	resp, err := EC2Client.DescribeKeyPairs(context.TODO(), &input)
	if err != nil {
		__antithesis_instrumentation__.Notify(180235)
		return nil, errors.Wrap(err, "getKeyPairs: failed to describe key pairs")
	} else {
		__antithesis_instrumentation__.Notify(180236)
	}
	__antithesis_instrumentation__.Notify(180234)
	return resp.KeyPairs, nil
}

func tagKeyPairIfUntagged(
	EC2Client *ec2.Client,
	keyPair ec2types.KeyPairInfo,
	IAMUserName string,
	timestamp time.Time,
	dryrun bool,
) (string, error) {
	__antithesis_instrumentation__.Notify(180237)
	IAMUserNameTag, createdAtTag := getTagsValues(keyPair.Tags)
	var tags []ec2types.Tag
	if IAMUserNameTag == "" {
		__antithesis_instrumentation__.Notify(180241)
		IAMUserNameKey := "IAMUserName"
		tags = append(tags, ec2types.Tag{Key: &IAMUserNameKey, Value: &IAMUserName})
		log.Infof(context.Background(), "Tagging %s with IAMUserName: %s\n", *keyPair.KeyName, IAMUserName)
	} else {
		__antithesis_instrumentation__.Notify(180242)
	}
	__antithesis_instrumentation__.Notify(180238)
	createdAtValue := timestamp.Format(time.RFC3339)
	if createdAtTag == "" {
		__antithesis_instrumentation__.Notify(180243)
		createdAtKey := "CreatedAt"
		tags = append(tags, ec2types.Tag{Key: &createdAtKey, Value: &createdAtValue})
		log.Infof(context.Background(), "Tagging %s with CreatedAt: %s\n", *keyPair.KeyName, createdAtValue)
	} else {
		__antithesis_instrumentation__.Notify(180244)
		createdAtValue = createdAtTag
	}
	__antithesis_instrumentation__.Notify(180239)

	if !dryrun && func() bool {
		__antithesis_instrumentation__.Notify(180245)
		return len(tags) > 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(180246)
		input := ec2.CreateTagsInput{Resources: []string{*keyPair.KeyPairId}, Tags: tags}
		_, err := EC2Client.CreateTags(context.TODO(), &input)
		if err != nil {
			__antithesis_instrumentation__.Notify(180247)
			return "", errors.Wrapf(err, "tagKeyPair: failed to create tags for key: %s", *keyPair.KeyName)
		} else {
			__antithesis_instrumentation__.Notify(180248)
		}
	} else {
		__antithesis_instrumentation__.Notify(180249)
	}
	__antithesis_instrumentation__.Notify(180240)
	return createdAtValue, nil
}

func deleteKeyPair(EC2Client *ec2.Client, keyPairName string) error {
	__antithesis_instrumentation__.Notify(180250)
	input := ec2.DeleteKeyPairInput{KeyName: &keyPairName}
	_, err := EC2Client.DeleteKeyPair(context.TODO(), &input)
	if err != nil {
		__antithesis_instrumentation__.Notify(180252)
		return errors.Wrap(err, "deleteKeyPair: failed to delete key pair")
	} else {
		__antithesis_instrumentation__.Notify(180253)
	}
	__antithesis_instrumentation__.Notify(180251)
	return nil
}

func GCAWSKeyPairs(dryrun bool) error {
	__antithesis_instrumentation__.Notify(180254)
	timestamp := timeutil.Now()

	IAMClient, err := getIAMClient("")
	if err != nil {
		__antithesis_instrumentation__.Notify(180262)
		return err
	} else {
		__antithesis_instrumentation__.Notify(180263)
	}
	__antithesis_instrumentation__.Notify(180255)

	IAMUsers, err := getIAMUsers(IAMClient)
	if err != nil {
		__antithesis_instrumentation__.Notify(180264)
		return err
	} else {
		__antithesis_instrumentation__.Notify(180265)
	}
	__antithesis_instrumentation__.Notify(180256)
	usersWithActiveAccessKey, err := getUsersWithActiveAccessKey(IAMClient, IAMUsers)
	if err != nil {
		__antithesis_instrumentation__.Notify(180266)
		return err
	} else {
		__antithesis_instrumentation__.Notify(180267)
	}
	__antithesis_instrumentation__.Notify(180257)
	usersWithMFAEnabled, err := getUsersWithMFAEnabled(IAMClient, IAMUsers)
	if err != nil {
		__antithesis_instrumentation__.Notify(180268)
		return err
	} else {
		__antithesis_instrumentation__.Notify(180269)
	}
	__antithesis_instrumentation__.Notify(180258)
	usersWithConsoleAccess, err := getUsersWithConsoleAccess(IAMClient, IAMUsers)
	if err != nil {
		__antithesis_instrumentation__.Notify(180270)
		return err
	} else {
		__antithesis_instrumentation__.Notify(180271)
	}
	__antithesis_instrumentation__.Notify(180259)
	regions, err := getRegions()
	if err != nil {
		__antithesis_instrumentation__.Notify(180272)
		return err
	} else {
		__antithesis_instrumentation__.Notify(180273)
	}
	__antithesis_instrumentation__.Notify(180260)
	for _, region := range regions {
		__antithesis_instrumentation__.Notify(180274)
		log.Infof(context.Background(), "%s", *region.RegionName)
		EC2Client, err := getEC2Client(*region.RegionName)
		if err != nil {
			__antithesis_instrumentation__.Notify(180277)
			return err
		} else {
			__antithesis_instrumentation__.Notify(180278)
		}
		__antithesis_instrumentation__.Notify(180275)
		keyPairs, err := getKeyPairs(EC2Client)
		if err != nil {
			__antithesis_instrumentation__.Notify(180279)
			return err
		} else {
			__antithesis_instrumentation__.Notify(180280)
		}
		__antithesis_instrumentation__.Notify(180276)
		for _, keyPair := range keyPairs {
			__antithesis_instrumentation__.Notify(180281)

			IAMUserName := getIAMUserNameFromKeyname(*keyPair.KeyName)
			if IAMUserName == "" {
				__antithesis_instrumentation__.Notify(180285)

				continue
			} else {
				__antithesis_instrumentation__.Notify(180286)
			}
			__antithesis_instrumentation__.Notify(180282)
			createdAt, err := tagKeyPairIfUntagged(EC2Client, keyPair, IAMUserName, timestamp, dryrun)
			if err != nil {
				__antithesis_instrumentation__.Notify(180287)
				return err
			} else {
				__antithesis_instrumentation__.Notify(180288)
			}
			__antithesis_instrumentation__.Notify(180283)

			if IAMUserName == "teamcity-runner" {
				__antithesis_instrumentation__.Notify(180289)
				createdAtTimestamp, err := time.Parse(time.RFC3339, createdAt)
				if err != nil {
					__antithesis_instrumentation__.Notify(180292)
					return err
				} else {
					__antithesis_instrumentation__.Notify(180293)
				}
				__antithesis_instrumentation__.Notify(180290)

				if timestamp.Sub(createdAtTimestamp).Hours() >= 240 {
					__antithesis_instrumentation__.Notify(180294)
					log.Infof(context.Background(), "Deleting %s because it is a teamcity-runner key created at %s.\n",
						*keyPair.KeyName, createdAtTimestamp)
					if !dryrun {
						__antithesis_instrumentation__.Notify(180295)
						err := deleteKeyPair(EC2Client, *keyPair.KeyName)
						if err != nil {
							__antithesis_instrumentation__.Notify(180296)
							return err
						} else {
							__antithesis_instrumentation__.Notify(180297)
						}
					} else {
						__antithesis_instrumentation__.Notify(180298)
					}
				} else {
					__antithesis_instrumentation__.Notify(180299)
				}
				__antithesis_instrumentation__.Notify(180291)
				continue
			} else {
				__antithesis_instrumentation__.Notify(180300)
			}
			__antithesis_instrumentation__.Notify(180284)

			if usersWithConsoleAccess[IAMUserName] && func() bool {
				__antithesis_instrumentation__.Notify(180301)
				return !usersWithMFAEnabled[IAMUserName] == true
			}() == true {
				__antithesis_instrumentation__.Notify(180302)
				log.Infof(context.Background(), "Deleting %s because %s has console access but MFA disabled.\n", *keyPair.KeyName, IAMUserName)
				if !dryrun {
					__antithesis_instrumentation__.Notify(180303)
					err := deleteKeyPair(EC2Client, *keyPair.KeyName)
					if err != nil {
						__antithesis_instrumentation__.Notify(180304)
						return err
					} else {
						__antithesis_instrumentation__.Notify(180305)
					}
				} else {
					__antithesis_instrumentation__.Notify(180306)
				}

			} else {
				__antithesis_instrumentation__.Notify(180307)
				if !usersWithActiveAccessKey[IAMUserName] {
					__antithesis_instrumentation__.Notify(180308)
					log.Infof(context.Background(), "Deleting %s because %s does not have an active access key.\n",
						*keyPair.KeyName, IAMUserName)
					if !dryrun {
						__antithesis_instrumentation__.Notify(180309)
						err := deleteKeyPair(EC2Client, *keyPair.KeyName)
						if err != nil {
							__antithesis_instrumentation__.Notify(180310)
							return err
						} else {
							__antithesis_instrumentation__.Notify(180311)
						}
					} else {
						__antithesis_instrumentation__.Notify(180312)
					}
				} else {
					__antithesis_instrumentation__.Notify(180313)
				}
			}
		}
	}
	__antithesis_instrumentation__.Notify(180261)
	return nil
}
