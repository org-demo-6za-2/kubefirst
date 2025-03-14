package k3d

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

var (
	// Create
	cloudRegionFlag            string
	clusterNameFlag            string
	clusterTypeFlag            string
	dryRun                     bool
	githubOwnerFlag            string
	gitlabOwnerFlag            string
	gitProviderFlag            string
	gitopsTemplateURLFlag      string
	gitopsTemplateBranchFlag   string
	metaphorTemplateBranchFlag string
	metaphorTemplateURLFlag    string
	kbotPasswordFlag           string
	useTelemetryFlag           bool

	// Supported git providers
	supportedGitProviders = []string{"github", "gitlab"}

	// Quota
	quotaShowAllFlag bool
)

func NewCommand() *cobra.Command {

	k3dCmd := &cobra.Command{
		Use:   "k3d",
		Short: "kubefirst k3d installation",
		Long:  "kubefirst k3d",
	}

	// on error, doesnt show helper/usage
	k3dCmd.SilenceUsage = true

	// wire up new commands
	k3dCmd.AddCommand(Create(), Destroy())

	return k3dCmd
}

func Create() *cobra.Command {
	createCmd := &cobra.Command{
		Use:              "create",
		Short:            "create the kubefirst platform running in k3d on your localhost",
		TraverseChildren: true,
		RunE:             runK3d,
	}

	// todo review defaults and update descriptions
	createCmd.Flags().StringVar(&clusterNameFlag, "cluster-name", "kubefirst", "the name of the cluster to create")
	err := createCmd.MarkFlagRequired("cluster-name")
	if err != nil {
		log.Fatalf("error marking flag required: %s", err)
	}
	createCmd.Flags().StringVar(&clusterTypeFlag, "cluster-type", "mgmt", "the type of cluster to create (i.e. mgmt|workload)")
	createCmd.Flags().BoolVar(&dryRun, "dry-run", false, "don't execute the installation")
	createCmd.Flags().StringVar(&gitProviderFlag, "git-provider", "github", fmt.Sprintf("the git provider - one of: %s", supportedGitProviders))
	createCmd.Flags().StringVar(&githubOwnerFlag, "github-owner", "", "the GitHub owner of the new gitops and metaphor repositories - required if using github")
	createCmd.Flags().StringVar(&gitlabOwnerFlag, "gitlab-owner", "", "the GitLab owner (group) of the new gitops and metaphor projects - required if using gitlab")
	createCmd.Flags().StringVar(&gitopsTemplateBranchFlag, "gitops-template-branch", "main", "the branch to clone for the gitops-template repository")
	createCmd.Flags().StringVar(&gitopsTemplateURLFlag, "gitops-template-url", "https://github.com/kubefirst/gitops-template.git", "the fully qualified url to the gitops-template repository to clone")
	createCmd.Flags().StringVar(&kbotPasswordFlag, "kbot-password", "", "the default password to use for the kbot user")
	createCmd.Flags().StringVar(&metaphorTemplateBranchFlag, "metaphor-template-branch", "main", "the branch to clone for the metaphor-template repository")
	createCmd.Flags().StringVar(&metaphorTemplateURLFlag, "metaphor-template-url", "https://github.com/kubefirst/metaphor-frontend-template.git", "the fully qualified url to the metaphor-template repository to clone")
	createCmd.Flags().BoolVar(&useTelemetryFlag, "use-telemetry", true, "whether to emit telemetry")
	return createCmd
}

func Destroy() *cobra.Command {
	destroyCmd := &cobra.Command{
		Use:   "destroy",
		Short: "destroy the kubefirst platform",
		Long:  "deletes the GitHub resources, k3d resources, and local content to re-provision",
		RunE:  destroyK3d,
	}

	return destroyCmd
}
