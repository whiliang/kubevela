package commands

import (
	"flag"
	"fmt"
	"os"
	"runtime"

	"github.com/oam-dev/kubevela/pkg/utils/common"

	"github.com/gosuri/uitable"
	"github.com/spf13/cobra"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	"github.com/oam-dev/kubevela/api/types"
	"github.com/oam-dev/kubevela/cmd/vela/fake"
	"github.com/oam-dev/kubevela/pkg/commands/util"
	"github.com/oam-dev/kubevela/pkg/utils/system"
	"github.com/oam-dev/kubevela/version"
)

func NewCommand() *cobra.Command {
	ioStream := util.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}

	cmds := &cobra.Command{
		Use:                "vela",
		DisableFlagParsing: true,
		Run: func(cmd *cobra.Command, args []string) {
			allCommands := cmd.Commands()
			cmd.Printf("✈️  The Extensible Application Platform Based on Kubernetes and Open Application Model (OAM).\n\nUsage:\n  vela [flags]\n  vela [command]\n\nAvailable Commands:\n\n")
			PrintHelpByTag(cmd, allCommands, types.TypeStart)
			PrintHelpByTag(cmd, allCommands, types.TypeApp)
			PrintHelpByTag(cmd, allCommands, types.TypeCap)
			PrintHelpByTag(cmd, allCommands, types.TypeSystem)
			cmd.Println("Flags:")
			cmd.Println("  -h, --help   help for vela")
			cmd.Println()
			cmd.Println(`Use "vela [command] --help" for more information about a command.`)
		},
		SilenceUsage: true,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
	}
	cmds.PersistentFlags().StringP("env", "e", "", "specify environment name for application")
	restConf, err := config.GetConfig()
	if err != nil {
		fmt.Println("get kubeconfig err", err)
		os.Exit(1)
	}

	commandArgs := types.Args{
		Config: restConf,
		Schema: common.Scheme,
	}

	if err := system.InitDirs(); err != nil {
		fmt.Println("InitDir err", err)
		os.Exit(1)
	}

	cmds.AddCommand(
		// Getting Start
		NewInstallCommand(commandArgs, fake.ChartSource, ioStream),
		NewInitCommand(commandArgs, ioStream),
		NewUpCommand(commandArgs, ioStream),

		// Apps
		NewListCommand(commandArgs, ioStream),
		NewDeleteCommand(commandArgs, ioStream),
		NewAppShowCommand(ioStream),
		NewAppStatusCommand(commandArgs, ioStream),
		NewExecCommand(commandArgs, ioStream),
		NewPortForwardCommand(commandArgs, ioStream),
		NewLogsCommand(commandArgs, ioStream),
		NewEnvCommand(commandArgs, ioStream),
		NewConfigCommand(commandArgs, ioStream),

		// Capabilities
		CapabilityCommandGroup(commandArgs, ioStream),
		NewTemplateCommand(commandArgs, ioStream),
		NewTraitsCommand(commandArgs, ioStream),
		NewWorkloadsCommand(commandArgs, ioStream),

		// Helper
		SystemCommandGroup(commandArgs, ioStream),
		NewDashboardCommand(commandArgs, ioStream, fake.FrontendSource),
		NewCompletionCommand(),
		NewVersionCommand(),

		AddCompCommands(commandArgs, ioStream),
	)

	// Traits
	if err := AddTraitCommands(cmds, commandArgs, ioStream); err != nil {
		fmt.Println("Add trait commands from traitDefinition err", err)
		os.Exit(1)
	}

	// this is for mute klog
	fset := flag.NewFlagSet("logs", flag.ContinueOnError)
	klog.InitFlags(fset)
	_ = fset.Set("v", "-1")

	return cmds
}

func PrintHelpByTag(cmd *cobra.Command, all []*cobra.Command, tag string) {
	cmd.Printf("  %s:\n\n", tag)
	table := uitable.New()
	for _, c := range all {
		if val, ok := c.Annotations[types.TagCommandType]; ok && val == tag {
			table.AddRow("    "+c.Use, c.Long)
		}
	}
	cmd.Println(table.String())
	cmd.Println()
}

func NewVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Prints out build version information",
		Long:  "Prints out build version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf(`Version: %v
GitRevision: %v
GolangVersion: %v
`,
				version.VelaVersion,
				version.GitRevision,
				runtime.Version())
		},
		Annotations: map[string]string{
			types.TagCommandType: types.TypeSystem,
		},
	}
}
