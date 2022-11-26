// Package commands contains the CLI commands
package commands

import (
	"fmt"
	"os"
	"runtime/debug"
	"time"

	"github.com/spf13/cobra"

	"github.com/MichaelMure/git-bug/commands/bridge"
	usercmd "github.com/MichaelMure/git-bug/commands/user"

	"github.com/MichaelMure/git-bug/commands/bug"
	"github.com/MichaelMure/git-bug/commands/execenv"
)

var GitCommit = "unknown"

func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   execenv.RootCommandName,
		Short: "A bug tracker embedded in Git",
		Long: `git-bug is a bug tracker embedded in git.

git-bug use git objects to store the bug tracking separated from the files
history. As bugs are regular git objects, they can be pushed and pulled from/to
the same git remote you are already using to collaborate with other people.

`,

		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			root := cmd.Root()

			if bi, ok := debug.ReadBuildInfo(); !ok {
				root.Version = "unknown"
			} else {
				modified := false
				timestamp := "unknown"

				for _, s := range bi.Settings {
					switch s.Key {
					case "vcs.modified":
						modified = s.Value == "true"
					case "vcs.revision":
						GitCommit = s.Value
					case "vcs.time":
						if t, err := time.Parse(time.RFC3339, s.Value); err == nil {
							timestamp = t.Format("060102150405")
						}
					}
				}

				root.Version = bi.Main.Version

				if root.Version == "(devel)" {
					root.Version = fmt.Sprintf("dev-%s-%.12s", timestamp, GitCommit)
				}

				if modified {
					root.Version += " (modified)"
				}
			}
		},

		// For the root command, force the execution of the PreRun
		// even if we just display the help. This is to make sure that we check
		// the repository and give the user early feedback.
		Run: func(cmd *cobra.Command, args []string) {
			if err := cmd.Help(); err != nil {
				os.Exit(1)
			}
		},

		SilenceUsage:      true,
		DisableAutoGenTag: true,
	}

	const entityGroup = "entity"
	const uiGroup = "ui"
	const remoteGroup = "remote"

	cmd.AddGroup(&cobra.Group{ID: entityGroup, Title: "Entities"})
	cmd.AddGroup(&cobra.Group{ID: uiGroup, Title: "User interfaces"})
	cmd.AddGroup(&cobra.Group{ID: remoteGroup, Title: "Interaction with the outside world"})

	addCmdWithGroup := func(child *cobra.Command, groupID string) {
		cmd.AddCommand(child)
		child.GroupID = groupID
	}

	addCmdWithGroup(bugcmd.NewBugCommand(), entityGroup)
	addCmdWithGroup(usercmd.NewUserCommand(), entityGroup)
	addCmdWithGroup(newLabelCommand(), entityGroup)

	addCmdWithGroup(newTermUICommand(), uiGroup)
	addCmdWithGroup(newWebUICommand(), uiGroup)

	addCmdWithGroup(newPullCommand(), remoteGroup)
	addCmdWithGroup(newPushCommand(), remoteGroup)
	addCmdWithGroup(bridgecmd.NewBridgeCommand(), remoteGroup)

	cmd.AddCommand(newCommandsCommand())
	cmd.AddCommand(newVersionCommand())

	return cmd
}

func Execute() {
	if err := NewRootCommand().Execute(); err != nil {
		os.Exit(1)
	}
}
