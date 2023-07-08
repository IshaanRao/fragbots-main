package servers

import (
	"context"
	"fragbotsbackend/constants"
	"fragbotsbackend/database"
	"fragbotsbackend/logging"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
	"go.mongodb.org/mongo-driver/bson"
	"strings"
)

var dockerClient *client.Client
var ctx = context.Background()

var runningBots []*constants.BotInfo

func loadDockerClient() {
	c, err := client.NewClientWithOpts(client.WithHostFromEnv())
	if err != nil {
		logging.LogFatal("error loading docker client: " + err.Error())
	}
	dockerClient = c
	GetJoinCommand()
}

func GetJoinCommand() string {
	swarm, err := dockerClient.SwarmInspect(ctx)
	if err != nil {
		logging.Log(err.Error())
	}
	cmd := "docker swarm join --token " + swarm.JoinTokens.Worker + " fragbots.xyz:2377"
	return cmd
}

func RunFragbotsService(botId string) error {
	_, err := dockerClient.ServiceCreate(ctx, swarm.ServiceSpec{
		Annotations: swarm.Annotations{
			Name: botId,
		},
		TaskTemplate: swarm.TaskSpec{
			ContainerSpec: &swarm.ContainerSpec{
				Image: "ishaanrao/fragbots:latest",
				Env: []string{
					"BOT_ID=" + botId,
					"ACCESS_TOKEN=" + constants.AccessToken,
					"BACKEND_URI=" + constants.BackendUrl,
				},
			},
			Placement: &swarm.Placement{
				Constraints: []string{
					"node.hostname==" + strings.ToLower(botId),
				},
			},
		},
	}, types.ServiceCreateOptions{})
	if err != nil {
		logging.LogWarn("Failed to run fragbots service")
		return err
	}
	err = database.UpdateDocument("accounts", bson.D{{"botId", botId}}, bson.D{{"running", true}})
	if err != nil {
		logging.LogWarn("Failed to set bot to running: " + err.Error())
		return err
	}
	return nil
}

func StartServerManagement() {
	loadDockerClient()
	loadRunningBots()
}

func loadRunningBots() {
	var allBots = make([]constants.BotInfo, 0)
	err := database.GetAllDocuments("accounts", bson.D{{"botId", bson.D{{"$regex", "^(?!archive_(.*))"}}}}, &allBots)
	if err != nil {
		logging.LogWarn("Error loading running bots: " + err.Error())
		return
	}
	for _, info := range allBots {
		if info.Running {
			runningBots = append(runningBots, &info)
		}
	}
}
