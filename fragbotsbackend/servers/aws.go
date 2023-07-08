package servers

import (
	"bufio"
	"context"
	"encoding/base64"
	"fragbotsbackend/constants"
	"fragbotsbackend/logging"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"os"
)

type EC2CreateInstanceAPI interface {
	RunInstances(ctx context.Context,
		params *ec2.RunInstancesInput,
		optFns ...func(*ec2.Options)) (*ec2.RunInstancesOutput, error)

	CreateTags(ctx context.Context,
		params *ec2.CreateTagsInput,
		optFns ...func(*ec2.Options)) (*ec2.CreateTagsOutput, error)
}

var AwsClient *ec2.Client

func init() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		logging.LogFatal("Failed to initialize AWS Configuration: " + err.Error())
	}

	AwsClient = ec2.NewFromConfig(cfg)

}

func MakeFragBotServer(botId string) (string, error) {
	templateId := "lt-0b61c41adf6ba92f2"

	userData, err := getAwsUserData()
	if err != nil {
		logging.LogWarn("Failed to get userdata file error: " + err.Error())
		return "", err
	}
	userData += "\n" +
		"hostnamectl set-hostname " + botId +
		"\nsudo systemctl restart docker\n" +
		GetJoinCommand() + "\n" +
		"curl --request PUT '" + constants.BackendUrl + " + /bots/" + botId + "' --header 'access-token: " + constants.AccessToken + "'"

	userDataEncoded := base64.StdEncoding.EncodeToString([]byte(userData))
	input := &ec2.RunInstancesInput{
		LaunchTemplate: &types.LaunchTemplateSpecification{
			LaunchTemplateId: &templateId,
		},
		MinCount: aws.Int32(1),
		MaxCount: aws.Int32(1),
		UserData: &userDataEncoded,
	}
	result, err := MakeInstance(context.TODO(), AwsClient, input)
	if err != nil {
		logging.LogWarn("Encountered error creating instance: " + err.Error())
		return "", err
	}
	logging.Log("Made aws server with id: " + *result.Instances[0].InstanceId)
	return *result.Instances[0].InstanceId, nil
}

func getAwsUserData() (string, error) {
	file, err := os.Open("fragbotstartup.sh")
	if err != nil {
		return "", err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			logging.LogWarn("Error while closing aws file: " + err.Error())
		}
	}(file)

	scanner := bufio.NewScanner(file)
	scanner.Scan()
	userData := scanner.Text()
	for scanner.Scan() {
		userData += "\n" + scanner.Text()
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}
	return userData, nil
}

func DeleteInstance(serverId string) error {
	instanceIds := []string{serverId}
	input := &ec2.TerminateInstancesInput{InstanceIds: instanceIds}
	_, err := AwsClient.TerminateInstances(context.TODO(), input)
	if err != nil {
		return err
	}
	return nil
}
func MakeInstance(c context.Context, api EC2CreateInstanceAPI, input *ec2.RunInstancesInput) (*ec2.RunInstancesOutput, error) {
	return api.RunInstances(c, input)
}
