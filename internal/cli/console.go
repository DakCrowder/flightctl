package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"

	api "github.com/flightctl/flightctl/api/v1alpha1"
	"github.com/flightctl/flightctl/internal/client"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"golang.org/x/term"
	api_remotecommand "k8s.io/apimachinery/pkg/util/remotecommand"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/client-go/util/exec"
	k8sTerm "k8s.io/kubectl/pkg/util/term"
)

type ConsoleOptions struct {
	GlobalOptions
	tty       bool
	noTTY     bool
	remoteTTY bool
	protocols []string
}

func DefaultConsoleOptions() *ConsoleOptions {
	return &ConsoleOptions{
		GlobalOptions: DefaultGlobalOptions(),
		protocols: []string{
			api_remotecommand.StreamProtocolV5Name,
		},
	}
}

func NewConsoleCmd() *cobra.Command {
	o := DefaultConsoleOptions()

	cmd := &cobra.Command{
		Use:   "console device/NAME",
		Short: "Connect a console to the remote device through the server.",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Split args at "--"
			var flagArgs, passThroughArgs []string
			for i, arg := range args {
				if arg == "--" {
					flagArgs = args[:i]
					passThroughArgs = args[i+1:]
					break
				}
			}

			// If no "--" was found, all args are flag arguments
			if flagArgs == nil {
				flagArgs = args
			}

			// Manually parse only the flag arguments
			if err := cmd.Flags().Parse(flagArgs); err != nil {
				return err
			}

			cleanArgs := cmd.Flags().Args()
			// Process console command normally
			if err := o.Complete(cmd, cleanArgs); err != nil {
				return err
			}
			if err := o.Validate(cleanArgs); err != nil {
				return err
			}

			// Run the main console command
			return o.Run(cmd.Context(), cleanArgs, passThroughArgs)
		},
		SilenceUsage:       true,
		DisableFlagParsing: true,
	}

	o.Bind(cmd.Flags())

	return cmd
}

func (o *ConsoleOptions) Bind(fs *pflag.FlagSet) {
	o.GlobalOptions.Bind(fs)
	fs.BoolVarP(&o.tty, "tty", "", o.tty, "Allocate remote pseudo terminal")
	fs.BoolVarP(&o.noTTY, "notty", "", o.noTTY, "Don't allocate remote pseudo terminal")
}

func (o *ConsoleOptions) Complete(cmd *cobra.Command, args []string) error {
	return o.GlobalOptions.Complete(cmd, args)
}

func (o *ConsoleOptions) Validate(args []string) error {
	kind, name, err := parseAndValidateKindName(args[0])
	if err != nil {
		return err
	}

	if kind != DeviceKind {
		return fmt.Errorf("only devices can be connected to a console")
	}

	if len(name) == 0 {
		return fmt.Errorf("device name is required")
	}

	if o.tty && o.noTTY {
		return fmt.Errorf("--tty and --notty are mutually exclusive")
	}
	return nil
}

func (o *ConsoleOptions) Run(ctx context.Context, flagArgs, passThroughArgs []string) error {
	config, err := client.ParseConfigFile(o.ConfigFilePath)
	if err != nil {
		return fmt.Errorf("parsing config file: %w", err)
	}

	_, name, err := parseAndValidateKindName(flagArgs[0])
	if err != nil {
		return err
	}

	o.analyzeResponseAndExit(o.connectViaWS(ctx, config, name, client.GetAccessToken(config, o.ConfigFilePath), passThroughArgs))

	// unreachable
	return nil
}

// NewWebSocketExecClient creates a WebSocketExecutor
func (o *ConsoleOptions) newWebSocketExecClient(url string, restConfig *rest.Config) (remotecommand.Executor, error) {
	// Create WebSocket executor.  In case we want to support multiple version protocols, they should
	// be added here
	exec, err := remotecommand.NewWebSocketExecutorForProtocols(restConfig, "GET", url, o.protocols...)
	if err != nil {
		return nil, fmt.Errorf("failed to create WebSocket executor: %v", err)
	}

	return exec, nil
}

func (o *ConsoleOptions) SetupTTY() k8sTerm.TTY {
	t := k8sTerm.TTY{
		In:  os.Stdin,
		Out: os.Stdout,
	}

	o.remoteTTY = o.tty || t.IsTerminalIn() && t.IsTerminalOut() && !o.noTTY
	t.Raw = o.remoteTTY && t.IsTerminalIn()

	return t
}

func (o *ConsoleOptions) buildURL(baseURL, metadata string) (string, error) {
	// Initialize a URL object
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse base URL %q: %v", baseURL, err)
	}

	// Create query parameters
	query := url.Values{}
	query.Set(api.DeviceQueryConsoleSessionMetadata, metadata)

	// Encode the query parameters and attach them to the URL
	u.RawQuery = query.Encode()
	return u.String(), nil
}

func (o *ConsoleOptions) asTerminalSize(size *remotecommand.TerminalSize) *api.TerminalSize {
	if size == nil {
		return nil
	}
	return &api.TerminalSize{
		Width:  size.Width,
		Height: size.Height,
	}
}

func (o *ConsoleOptions) createSessionMetadata(t k8sTerm.TTY, passThroughArgs []string) string {
	metadata := api.DeviceConsoleSessionMetadata{
		InitialDimensions: o.asTerminalSize(t.GetSize()),
	}
	termEnv := os.Getenv("TERM")
	if termEnv != "" {
		metadata.Term = &termEnv
	}
	if len(passThroughArgs) > 0 {
		metadata.Command = &api.DeviceCommand{
			Command: passThroughArgs[0],
			Args:    passThroughArgs[1:],
		}
	}
	metadata.TTY = o.remoteTTY
	b, _ := json.Marshal(&metadata)
	return string(b)
}

func (o *ConsoleOptions) connectViaWS(ctx context.Context, config *client.Config, deviceName, token string, passThroughArgs []string) error {

	t := o.SetupTTY()

	options := remotecommand.StreamOptions{
		Stdout: os.Stdout,
		Tty:    o.remoteTTY,
	}
	if o.remoteTTY || !t.IsTerminalIn() {
		options.Stdin = os.Stdin
	}
	if !o.remoteTTY {
		options.Stderr = os.Stderr
	}

	// this call spawns a goroutine to monitor/update the terminal size
	if t.Raw {
		options.TerminalSizeQueue = t.MonitorSize(t.GetSize())
	}

	if t.Raw {

		oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
		if err != nil {
			fmt.Fprintf(os.Stderr, "error making terminal raw: %s\n", err)
			return err
		}
		defer func() {
			err := term.Restore(int(os.Stdin.Fd()), oldState)
			// reset terminal and clear screen
			if err != nil {
				fmt.Printf("error restoring terminal: %v", err)
			}
		}()
	}

	restConfig := &rest.Config{
		BearerToken: token,
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: config.Service.InsecureSkipVerify,
			CertData: config.AuthInfo.ClientCertificateData,
			CAData:   config.Service.CertificateAuthorityData,
		},
	}

	connURL, err := o.buildURL(fmt.Sprintf("%s/ws/v1/devices/%s/console", config.Service.Server, deviceName),
		o.createSessionMetadata(t, passThroughArgs))
	if err != nil {
		return err
	}
	wsClient, err := o.newWebSocketExecClient(connURL, restConfig)
	if err != nil {
		return err
	}
	return wsClient.StreamWithContext(ctx, options)
}

func (o *ConsoleOptions) analyzeResponseAndExit(err error) {
	var exitCode int
	switch concreteErr := err.(type) {
	case nil:
	case exec.CodeExitError:
		exitCode = concreteErr.Code
	default:
		exitCode = 255
		fmt.Fprintf(os.Stderr, "Unexpected error type %T: %+v\n", err, err)
	}
	os.Exit(exitCode)
}
