package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate completion script",
	Long: `Generate shell completion script for aether.

To load completions:

Bash:
  $ source <(aether completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ aether completion bash > /etc/bash_completion.d/aether
  # macOS:
  $ aether completion bash > $(brew --prefix)/etc/bash_completion.d/aether

Zsh:
  # If shell completion is not already enabled in your environment,
  # you will need to enable it.  You can execute the following once:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ aether completion zsh > "${fpath[1]}/_aether"

  # For oh-my-zsh users:
  $ mkdir -p ~/.oh-my-zsh/custom/plugins/aether
  $ aether completion zsh > ~/.oh-my-zsh/custom/plugins/aether/_aether
  # Then add 'aether' to your plugins array in ~/.zshrc:
  # plugins=(... aether)

  # You will need to start a new shell for this setup to take effect.

Fish:
  $ aether completion fish | source

  # To load completions for each session, execute once:
  $ aether completion fish > ~/.config/fish/completions/aether.fish

PowerShell:
  PS> aether completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> aether completion powershell > aether.ps1
  # and source this file from your PowerShell profile.
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	Run: func(cmd *cobra.Command, args []string) {
		switch args[0] {
		case "bash":
			cmd.Root().GenBashCompletion(os.Stdout)
		case "zsh":
			cmd.Root().GenZshCompletion(os.Stdout)
		case "fish":
			cmd.Root().GenFishCompletion(os.Stdout, true)
		case "powershell":
			cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
		}
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)
}
