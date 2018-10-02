package cmd

import (
	"github.com/gumieri/ecsctl/compose"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func taskDefinitionsCreateRun(cmd *cobra.Command, args []string) {
	dockerCompose, err := compose.FromFile("docker-compose.yml")
	typist.Must(err)

	containerDefinitions, err := dockerCompose.ToAWSContainerDefinitions()
	typist.Must(err)

	typist.Println(containerDefinitions)

	// result, err := ecsI.RegisterTaskDefinition(&ecs.RegisterTaskDefinitionInput{
	// 	ContainerDefinitions: []*ecs.ContainerDefinition{
	// 		{
	// 			Command: []*string{
	// 				aws.String("sleep"),
	// 				aws.String("360"),
	// 			},
	// 			Cpu:       aws.Int64(10),
	// 			Essential: aws.Bool(true),
	// 			Image:     aws.String("busybox"),
	// 			Memory:    aws.Int64(10),
	// 			Name:      aws.String("sleep"),
	// 		},
	// 	},
	// 	Family:      aws.String("sleep360"),
	// 	TaskRoleArn: aws.String(""),
	// })

	// typist.Must(err)

	// typist.Println(result)
}

var taskDefinitionsCreateCmd = &cobra.Command{
	Use:   "create [task-definition]",
	Short: "Create a Task Definition",
	Args:  cobra.ExactArgs(1),
	Run:   taskDefinitionsCreateRun,
}

func init() {
	taskDefinitionsCmd.AddCommand(taskDefinitionsCreateCmd)

	flags := taskDefinitionsCreateCmd.Flags()

	flags.StringVarP(&cluster, "cluster", "c", "", requiredSpec+clusterSpec)
	// flags.StringVarP(&family, "family", "f", "", requiredSpec+familySpec)

	taskDefinitionsCreateCmd.MarkFlagRequired("cluster")

	viper.BindPFlag("cluster", taskDefinitionsCreateCmd.Flags().Lookup("cluster"))
}
