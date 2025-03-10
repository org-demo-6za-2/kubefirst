package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/kubefirst/kubefirst/configs"
	"github.com/kubefirst/kubefirst/internal/aws"
	"github.com/kubefirst/kubefirst/internal/reports"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// cleanCmd removes all kubefirst resources created with the init command
var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "removes all kubefirst resources created with the init command",
	Long: `Kubefirst creates files, folders, cloud buckets and download tools during installation at your environment. 
This command removes and re-create Kubefirst base files. 
To destroy cloud resources you need to specify additional flags (--destroy-buckets)
To preserve the tools downloaded you need to specify additional flag (--preserve-tools).`,
	RunE: func(cmd *cobra.Command, args []string) error {

		config := configs.ReadConfig()

		destroyBuckets, err := cmd.Flags().GetBool("destroy-buckets")
		if err != nil {
			return err
		}
		destroyConfirm, err := cmd.Flags().GetBool("destroy-confirm")
		if err != nil {
			return err
		}
		if destroyBuckets && !destroyConfirm {
			return errors.New("this process will fully delete cloud buckets, and we would like you to confirm the deletion providing the --destroy-confirm when calling the clean command")
		}

		err = aws.DestroyBucketsInUse(false, destroyBuckets && destroyConfirm)
		if err != nil {
			return err
		}

		// command line flags
		rmLogsFolder, err := cmd.Flags().GetBool("rm-logs")
		if err != nil {
			return err
		}

		// remove logs folder if flag is enabled
		var logFolderLocation string
		if rmLogsFolder {
			logFolderLocation = viper.GetString("logs-location")

			err := os.RemoveAll(logFolderLocation)
			if err != nil {
				return fmt.Errorf("unable to delete %q file, error is: ", err)
			}
		}

		preserveTools, err := cmd.Flags().GetBool("preserve-tools")
		if err != nil {
			return err
		}

		if preserveTools {
			log.Info().Msg("Cleaning with --preserve-tools enabled")
			// delete gitops path and argo init values - caching tools to avoid re-download
			err = os.RemoveAll(config.GitOpsRepoPath)
			if err != nil {
				return fmt.Errorf("unable to delete %q folder, error is: %s", config.GitOpsRepoPath, err)
			}
			err = os.Remove(config.ArgoCDInitValuesYamlPath)
			if err != nil {
				return fmt.Errorf("unable to delete %q file, error is: ", err)
			}

			log.Debug().Msgf("Removing SSL folder")
			err = os.RemoveAll(fmt.Sprintf("%s/ssl", config.K1FolderPath))
			if err != nil {
				log.Debug().Msgf("unable to delete %q file, error is: ", err)
			}

			var metaphorFolders = []string{"metaphor", "metaphor-frontend", "metaphor-go"}
			for _, f := range metaphorFolders {
				log.Debug().Msgf("Removing metaphors folders: %s", f)
				err = os.RemoveAll(fmt.Sprintf("%s/%s", config.K1FolderPath, f))
				if err != nil {
					log.Debug().Msgf("unable to delete %q file, error is: ", err)
				}
			}

		} else {
			err = os.RemoveAll(config.K1FolderPath)
			if err != nil {
				return fmt.Errorf("unable to delete %q folder, error is: %s", config.K1FolderPath, err)
			}

			if err := os.Mkdir(fmt.Sprintf("%s", config.K1FolderPath), os.ModePerm); err != nil {
				return fmt.Errorf("error: could not create directory %q - it must exist to continue. error is: %s", config.K1FolderPath, err)
			}
		}

		err = os.Remove(config.KubefirstConfigFilePath)
		if err != nil {
			return fmt.Errorf("unable to delete %q file, error is: ", err)
		}

		// re-create .kubefirst file
		kubefirstFile, err := os.Create(config.KubefirstConfigFilePath)
		if err != nil {
			return fmt.Errorf("error: could not create `$HOME/.kubefirst` file: %v", err)
		}
		err = kubefirstFile.Close()
		if err != nil {
			return err
		}

		// re-create base
		log.Info().Msgf("%q config file and %q folder were deleted and re-created", config.KubefirstConfigFilePath, config.K1FolderPath)

		var cleanSummary bytes.Buffer
		cleanSummary.WriteString(strings.Repeat("-", 70))
		cleanSummary.WriteString("\nclean summary:\n")
		cleanSummary.WriteString(strings.Repeat("-", 70))
		cleanSummary.WriteString("\n\nFiles and folders deleted:\n\n")

		cleanSummary.WriteString(fmt.Sprintf("   %q\n", config.KubefirstConfigFilePath))
		cleanSummary.WriteString(fmt.Sprintf("   %q\n", config.K1FolderPath))

		if rmLogsFolder {
			cleanSummary.WriteString(fmt.Sprintf("   %q\n", logFolderLocation))
		}

		cleanSummary.WriteString("\nRe-created empty folder: \n\n")
		cleanSummary.WriteString(fmt.Sprintf("   %q\n\n", config.K1FolderPath))

		cleanSummary.WriteString("Re-created empty config file: \n\n")
		cleanSummary.WriteString(fmt.Sprintf("   %q", config.KubefirstConfigFilePath))

		fmt.Println(reports.StyleMessage(cleanSummary.String()))

		return nil
	},
}

func init() {
	rootCmd.AddCommand(cleanCmd)
	cleanCmd.Flags().Bool("rm-logs", false, "remove logs folder")
	cleanCmd.Flags().Bool("destroy-buckets", false, "destroy buckets created by init cmd")
	cleanCmd.Flags().Bool("destroy-confirm", false, "when detroy-buckets flag is provided, we must provide this flag as well to confirm the destroy operation")
	cleanCmd.Flags().Bool("preserve-tools", false, "preserve all downloaded tools (avoid re-downloading)")
}
