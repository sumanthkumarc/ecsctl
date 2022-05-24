package cmd

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var applyCmd = &cobra.Command{
	Use:   "apply [flags]",
	Short: "Commands to apply resources",
	Run:   applyResources,
}

var file string
var sep = regexp.MustCompile("(?:^|\\s*\n)---\\s*")

func init() {
	rootCmd.AddCommand(applyCmd)
	applyCmd.Flags().StringVarP(&file, "file", "f", "", "Path to file to apply. Pass `-` to read from stdin")
	applyCmd.MarkFlagRequired("file")
}

// Copied from https://github.com/kumahq/kuma/blob/master/pkg/util/yaml/split.go
// SplitYAML takes YAMLs separated by `---` line and splits it into multiple YAMLs. Empty entries are ignored
func splitYAML(yamls string) []string {
	var result []string
	// Making sure that any extra whitespace in YAML stream doesn't interfere in splitting documents correctly.
	trimYAMLs := strings.TrimSpace(yamls)
	docs := sep.Split(trimYAMLs, -1)
	for _, doc := range docs {
		if doc == "" {
			continue
		}
		doc = strings.TrimSpace(doc)
		result = append(result, doc)
	}
	return result
}

func applyResources(cmd *cobra.Command, args []string) {
	var b []byte
	var err error

	if file == "-" {
		b, err = io.ReadAll(cmd.InOrStdin())
		if err != nil {
			cmd.PrintErrln(err)
		}
	} else {
		b, err = os.ReadFile(file)

		if err != nil {
			errors.Wrap(err, "error while reading provided file")
		}
	}

	if len(b) == 0 {
		fmt.Print("no resource(s) passed to apply")
	}

	rawResources := splitYAML(string(b))

	aws_path, err := exec.LookPath("aws")
	if err != nil {
		log.Fatalf("AWS not found in $PATH")
	}

	for _, rawResource := range rawResources {
		if len(rawResource) == 0 {
			continue
		}

		var resource map[interface{}]interface{}

		err := yaml.Unmarshal([]byte(rawResource), &resource)
		if err != nil {
			log.Fatalf("error: %v", err)
		}

		//  if resource has containerDefinitions, create/upsert task-def
		if _, ok := resource["containerDefinitions"]; ok {
			// fmt.Println("This is task definition")

			writeData("manifests/task_definition.yaml", []byte(rawResource))

			err := runAWSCli([]string{
				aws_path,
				"ecs",
				"register-task-definition",
				"--cli-input-yaml",
				fmt.Sprintf("file://%s/manifests/task_definition.yaml", getCWD()),
			})

			if err != nil {
				log.Fatal(err)
			}

			// var jsonResource []byte

			// jsonResource, err := k8syaml.YAMLToJSON([]byte(rawResource))
			// if err != nil {
			// 	log.Fatalf("error: %v", err)
			// }

			// var taskDef ecs.RegisterTaskDefinitionInput

			// if err := json.Unmarshal([]byte(jsonResource), &taskDef); err != nil {
			// 	log.Fatalf("error: %v", err)
			// }

			// fmt.Println(taskDef)

			// ecsI.RegisterTaskDefinition(&taskDef)

		} else if _, ok := resource["serviceName"]; ok {
			// if resource has serviceName, create/upsert service
			// fmt.Println("This is service  definition")
			// fmt.Println(resource)
			writeData("manifests/svc_definition.yaml", []byte(rawResource))

			err := runAWSCli([]string{
				aws_path,
				"ecs",

				"create-service",
				"--cli-input-yaml",
				fmt.Sprintf("file://%s/manifests/svc_definition.yaml", getCWD()),
			})

			if err != nil {
				log.Fatal(err)
			}

		}

	}

}

func writeData(path string, data []byte) {
	f, err := createPath(path)
	if err != nil {
		fmt.Println(err)
	}

	defer f.Close()
	f.Write(data)

}

func createPath(p string) (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(p), 0770); err != nil {
		return nil, err
	}
	return os.Create(p)
}

func runAWSCli(command []string) (err error) {

	aws_path, _ := exec.LookPath("aws")
	cmd := &exec.Cmd{
		Path:   aws_path,
		Args:   command,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}

	log.Println("Executing command ", cmd)

	if err := cmd.Start(); err != nil {
		return err
	}

	if err = cmd.Wait(); err != nil {
		return err
	}

	// fmt.Println(cmd.StdoutPipe())

	return nil
}

func getCWD() (cwd string) {
	mydir, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
	}

	return mydir
}
